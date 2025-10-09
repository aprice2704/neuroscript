// NeuroScript Version: 0.8.0
// File version: 66
// Purpose: Updates the tool call dispatch logic to use the configurable 'runtime' field instead of the interpreter itself.
// filename: pkg/interpreter/evaluation_main.go
// nlines: 280
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

var placeholderRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)

// evaluation holds the methods for evaluating AST expression nodes.
type evaluation struct {
	i *Interpreter
}

// Expression evaluates an AST node representing an expression.
func (e *evaluation) Expression(node ast.Expression) (lang.Value, error) {
	if node == nil {
		return &lang.NilValue{}, nil
	}

	switch n := node.(type) {
	case *ast.StringLiteralNode:
		return e.evaluateStringLiteral(n)
	case *ast.NumberLiteralNode:
		val, _ := lang.ToFloat64(n.Value)
		return lang.NumberValue{Value: val}, nil
	case *ast.BooleanLiteralNode:
		return lang.BoolValue{Value: n.Value}, nil
	case *ast.NilLiteralNode:
		return &lang.NilValue{}, nil
	case *ast.VariableNode:
		return e.i.resolveVariable(n)
	case *ast.PlaceholderNode:
		return e.i.resolvePlaceholder(n)
	case *ast.LastNode:
		return e.i.lastCallResult, nil
	case *ast.EvalNode:
		return e.i.lastCallResult, nil
	case *ast.ListLiteralNode:
		return e.evaluateListLiteral(n)
	case *ast.MapLiteralNode:
		return e.evaluateMapLiteral(n)
	case *ast.UnaryOpNode:
		operandVal, err := e.Expression(n.Operand)
		if err != nil {
			return nil, err
		}
		return e.i.EvaluateUnaryOp(n.Operator, operandVal)
	case *ast.BinaryOpNode:
		leftVal, errL := e.Expression(n.Left)
		if errL != nil {
			return nil, errL
		}
		rightVal, errR := e.Expression(n.Right)
		if errR != nil {
			return nil, errR
		}
		return e.i.EvaluateBinaryOp(leftVal, rightVal, n.Operator)
	case *ast.TypeOfNode:
		return e.evaluateTypeOf(n)
	case *ast.CallableExprNode:
		return e.evaluateCall(n)
	case *ast.ElementAccessNode:
		return e.i.evaluateElementAccess(n)
	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("unhandled expression type: %T", node), nil)
	}
}

func (e *evaluation) evaluateStringLiteral(n *ast.StringLiteralNode) (lang.Value, error) {
	if n.IsRaw {
		resolvedStr, resolveErr := e.i.resolvePlaceholdersWithError(n.Value)
		if resolveErr != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeEvaluation, "evaluating raw string literal", resolveErr).WithPosition(n.StartPos)
		}
		return lang.StringValue{Value: resolvedStr}, nil
	}
	return lang.StringValue{Value: n.Value}, nil
}

func (i *Interpreter) resolveVariable(n *ast.VariableNode) (lang.Value, error) {
	if val, exists := i.GetVariable(n.Name); exists {
		return val, nil
	}

	if proc, procExists := i.KnownProcedures()[n.Name]; procExists {
		return lang.FunctionValue{Value: proc}, nil
	}
	if tool, toolExists := i.tools.GetTool(types.FullName(n.Name)); toolExists {
		return lang.ToolValue{Value: &tool}, nil
	}
	if typeVal, typeExists := GetTypeConstant(n.Name); typeExists {
		return lang.StringValue{Value: typeVal}, nil
	}
	if isBuiltInFunction(n.Name) {
		return lang.StringValue{Value: fmt.Sprintf("<built-in function: %s>", n.Name)}, nil
	}
	return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("variable or function '%s' not found", n.Name), lang.ErrVariableNotFound).WithPosition(n.StartPos)
}

func (i *Interpreter) resolvePlaceholder(n *ast.PlaceholderNode) (lang.Value, error) {
	var refValue lang.Value
	var exists bool
	if n.Name == "LAST" {
		refValue = i.lastCallResult
		exists = refValue != nil
	} else {
		refValue, exists = i.GetVariable(n.Name)
	}
	if !exists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' for placeholder not found", n.Name), lang.ErrVariableNotFound).WithPosition(n.StartPos)
	}
	return refValue, nil
}

func (e *evaluation) evaluateListLiteral(n *ast.ListLiteralNode) (lang.Value, error) {
	evaluatedElements := make([]lang.Value, len(n.Elements))
	for idx, elemNode := range n.Elements {
		var err error
		evaluatedElements[idx], err = e.Expression(elemNode)
		if err != nil {
			return nil, err
		}
	}
	return lang.ListValue{Value: evaluatedElements}, nil
}

func (e *evaluation) evaluateMapLiteral(n *ast.MapLiteralNode) (lang.Value, error) {
	evaluatedMap := make(map[string]lang.Value)
	for _, entry := range n.Entries {
		mapKey := entry.Key.Value
		elemVal, err := e.Expression(entry.Value)
		if err != nil {
			return nil, err
		}
		evaluatedMap[mapKey] = elemVal
	}
	return lang.NewMapValue(evaluatedMap), nil
}

func (e *evaluation) evaluateTypeOf(n *ast.TypeOfNode) (lang.Value, error) {
	argValue, err := e.Expression(n.Argument)
	if err != nil {
		if errors.Is(err, lang.ErrVariableNotFound) {
			argValue = &lang.NilValue{}
		} else {
			return nil, err
		}
	}
	return lang.StringValue{Value: string(lang.TypeOf(argValue))}, nil
}

func (e *evaluation) evaluateCall(n *ast.CallableExprNode) (lang.Value, error) {
	if isBuiltInFunction(n.Target.Name) {
		evaluatedArgs := make([]lang.Value, len(n.Arguments))
		for i, argNode := range n.Arguments {
			var err error
			evaluatedArgs[i], err = e.Expression(argNode)
			if err != nil {
				return nil, err
			}
		}
		return e.evaluateUserOrBuiltInFunction(n.Target.Name, evaluatedArgs, n.StartPos)
	}

	if n.Target.IsTool {
		toolNameForLookup, err := resolveToolName(n)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to resolve tool name", err).WithPosition(n.StartPos)
		}

		toolImpl, found := e.i.tools.GetTool(toolNameForLookup)
		if !found {
			errMessage := fmt.Sprintf("tool '%s' not found (looked up as '%s')", n.Target.Name, toolNameForLookup)
			e.i.logger.Errorf("[DEBUG] Point A: Tool not found error created in evaluateCall: %s", errMessage) // DEBUG
			return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, errMessage, lang.ErrToolNotFound).WithPosition(n.StartPos)
		}

		// --- POLICY GATE ---
		if e.i.ExecPolicy != nil {
			meta := policy.ToolMeta{
				Name:          strings.ToLower(string(toolImpl.FullName)),
				RequiresTrust: toolImpl.RequiresTrust,
				RequiredCaps:  toolImpl.RequiredCaps,
				Effects:       toolImpl.Effects,
			}
			specFetcher := func(name string) (policy.ToolSpecProvider, bool) {
				impl, found := e.i.tools.GetTool(types.FullName(name))
				if !found {
					return nil, false
				}
				return impl.Spec, true
			}
			if err := policy.CanCall(e.i.ExecPolicy, meta, specFetcher); err != nil {
				return nil, lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("tool call '%s' rejected by policy", toolImpl.FullName), err).WithPosition(n.StartPos)
			}
		}
		// --- END POLICY GATE ---

		specArgs := toolImpl.Spec.Args
		if len(n.Arguments) > len(specArgs) && !toolImpl.Spec.Variadic {
			return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("tool '%s' expects at most %d arguments, got %d", toolImpl.Spec.Name, len(specArgs), len(n.Arguments)), lang.ErrArgumentMismatch).WithPosition(n.StartPos)
		}

		evaluatedArgs := make([]lang.Value, len(n.Arguments))
		for i, argNode := range n.Arguments {
			var err error
			evaluatedArgs[i], err = e.Expression(argNode)
			if err != nil {
				return nil, err
			}
		}

		unwrappedArgs := make([]interface{}, len(evaluatedArgs))
		for i, v := range evaluatedArgs {
			unwrappedArgs[i] = lang.Unwrap(v)
		}

		result, err := toolImpl.Func(e.i.runtime, unwrappedArgs)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("tool '%s' execution failed", toolImpl.Spec.Name), err).WithPosition(n.StartPos)
		}

		return lang.Wrap(result)
	}

	evaluatedArgs := make([]lang.Value, len(n.Arguments))
	for idx, argNode := range n.Arguments {
		var err error
		evaluatedArgs[idx], err = e.Expression(argNode)
		if err != nil {
			return nil, err
		}
	}
	return e.evaluateUserOrBuiltInFunction(n.Target.Name, evaluatedArgs, n.StartPos)
}

func (e *evaluation) evaluateUserOrBuiltInFunction(funcName string, args []lang.Value, pos *types.Position) (lang.Value, error) {
	if isBuiltInFunction(funcName) {
		unwrappedArgs := make([]interface{}, len(args))
		for i, v := range args {
			unwrappedArgs[i] = lang.Unwrap(v)
		}
		result, err := evaluateBuiltInFunction(funcName, unwrappedArgs)
		if err != nil {
			if _, ok := err.(*lang.RuntimeError); !ok {
				err = lang.NewRuntimeError(lang.ErrorCodeGeneric, fmt.Sprintf("built-in function '%s' failed", funcName), err).WithPosition(pos)
			}
			return nil, err
		}
		wrappedResult, wrapErr := lang.Wrap(result)
		if wrapErr != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "wrapping built-in function result failed", wrapErr).WithPosition(pos)
		}
		return wrappedResult, nil
	}
	procResult, procErr := e.i.RunProcedure(funcName, args...)
	if procErr != nil {
		if re, ok := procErr.(*lang.RuntimeError); ok {
			return nil, re.WithPosition(pos)
		}
		return nil, lang.NewRuntimeError(lang.ErrorCodeProcNotFound, fmt.Sprintf("error calling procedure '%s'", funcName), procErr).WithPosition(pos)
	}
	e.i.lastCallResult = procResult
	return procResult, nil
}

func (i *Interpreter) resolvePlaceholdersWithError(raw string) (string, error) {
	var firstErr error
	resolved := placeholderRegex.ReplaceAllStringFunc(raw, func(match string) string {
		if firstErr != nil {
			return ""
		}
		varName := strings.TrimSpace(match[2 : len(match)-2])
		val, exists := i.GetVariable(varName)
		if !exists {
			firstErr = lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' not found in placeholder", varName), lang.ErrVariableNotFound)
			return ""
		}
		return val.String()
	})
	if firstErr != nil {
		return "", firstErr
	}
	return resolved, nil
}

func GetTypeConstant(name string) (string, bool) {
	switch name {
	case "TYPE_STRING":
		return string(lang.TypeString), true
	case "TYPE_NUMBER":
		return string(lang.TypeNumber), true
	case "TYPE_BOOLEAN":
		return string(lang.TypeBoolean), true
	case "TYPE_LIST":
		return string(lang.TypeList), true
	case "TYPE_MAP":
		return string(lang.TypeMap), true
	case "TYPE_NIL":
		return string(lang.TypeNil), true
	case "TYPE_FUNCTION":
		return string(lang.TypeFunction), true
	case "TYPE_TOOL":
		return string(lang.TypeTool), true
	case "TYPE_ERROR":
		return string(lang.TypeError), true
	case "TYPE_TIMEDATE":
		return string(lang.TypeTimedate), true
	case "TYPE_EVENT":
		return string(lang.TypeEvent), true
	case "TYPE_FUZZY":
		return string(lang.TypeFuzzy), true
	case "TYPE_UNKNOWN":
		return string(lang.TypeUnknown), true
	case "TYPE_BYTES":
		return string(lang.TypeBytes), true
	}
	return "", false
}
