// NeuroScript Version: 0.4.0
// File version: 25
// Purpose: A complete, corrected version that properly handles the unwrap/wrap boundary for built-in functions.
// filename: pkg/core/evaluation_main.go

package core

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var placeholderRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)

// evaluateExpression evaluates an AST node representing an expression.
func (i *Interpreter) evaluateExpression(node Expression) (Value, error) {
	if node == nil {
		return NilValue{}, nil
	}

	switch n := node.(type) {
	case *StringLiteralNode:
		return i.evaluateStringLiteral(n)
	case *NumberLiteralNode:
		val, _ := toFloat64(n.Value)
		return NumberValue{Value: val}, nil
	case *BooleanLiteralNode:
		return BoolValue{Value: n.Value}, nil
	case *NilLiteralNode:
		return NilValue{}, nil
	case *VariableNode:
		return i.resolveVariable(n)
	case *PlaceholderNode:
		return i.resolvePlaceholder(n)
	case *LastNode:
		return i.lastCallResult, nil
	case *ListLiteralNode:
		return i.evaluateListLiteral(n)
	case *MapLiteralNode:
		return i.evaluateMapLiteral(n)
	case *EvalNode:
		return i.evaluateEvalNode(n)
	case *UnaryOpNode:
		operandVal, err := i.evaluateExpression(n.Operand)
		if err != nil {
			return nil, err
		}
		return i.evaluateUnaryOp(n.Operator, operandVal)
	case *BinaryOpNode:
		leftVal, errL := i.evaluateExpression(n.Left)
		if errL != nil {
			return nil, errL
		}
		rightVal, errR := i.evaluateExpression(n.Right)
		if errR != nil {
			return nil, errR
		}
		return i.evaluateBinaryOp(leftVal, rightVal, n.Operator)
	case *TypeOfNode:
		return i.evaluateTypeOf(n)
	case *CallableExprNode:
		return i.evaluateCall(n)
	case *ElementAccessNode:
		return i.evaluateElementAccess(n)
	default:
		// This can happen if a non-expression type is passed by mistake.
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("unhandled expression type: %T", node), nil)
	}
}

func (i *Interpreter) evaluateStringLiteral(n *StringLiteralNode) (Value, error) {
	if n.IsRaw {
		resolvedStr, resolveErr := i.resolvePlaceholdersWithError(n.Value)
		if resolveErr != nil {
			return nil, NewRuntimeError(ErrorCodeEvaluation, "evaluating raw string literal", resolveErr).WithPosition(n.Pos)
		}
		return StringValue{Value: resolvedStr}, nil
	}
	return StringValue{Value: n.Value}, nil
}

func (i *Interpreter) resolveVariable(n *VariableNode) (Value, error) {
	if val, exists := i.GetVariable(n.Name); exists {
		return val, nil
	}
	if proc, procExists := i.KnownProcedures()[n.Name]; procExists {
		return FunctionValue{Value: *proc}, nil
	}
	if tool, toolExists := i.ToolRegistry().GetTool(n.Name); toolExists {
		return ToolValue{Value: tool}, nil
	}
	if typeVal, typeExists := GetTypeConstant(n.Name); typeExists {
		return StringValue{Value: typeVal}, nil
	}
	return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' not found", n.Name), ErrVariableNotFound).WithPosition(n.Pos)
}

func (i *Interpreter) resolvePlaceholder(n *PlaceholderNode) (Value, error) {
	var refValue Value
	var exists bool
	if n.Name == "LAST" {
		refValue = i.lastCallResult
		exists = refValue != nil
	} else {
		refValue, exists = i.GetVariable(n.Name)
	}
	if !exists {
		return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' for placeholder not found", n.Name), ErrVariableNotFound).WithPosition(n.Pos)
	}
	return refValue, nil
}

func (i *Interpreter) evaluateListLiteral(n *ListLiteralNode) (Value, error) {
	evaluatedElements := make([]Value, len(n.Elements))
	for idx, elemNode := range n.Elements {
		var err error
		evaluatedElements[idx], err = i.evaluateExpression(elemNode)
		if err != nil {
			return nil, err
		}
	}
	return NewListValue(evaluatedElements), nil
}

func (i *Interpreter) evaluateMapLiteral(n *MapLiteralNode) (Value, error) {
	evaluatedMap := make(map[string]Value)
	for _, entry := range n.Entries {
		mapKey := entry.Key.Value
		elemVal, err := i.evaluateExpression(entry.Value)
		if err != nil {
			return nil, err
		}
		evaluatedMap[mapKey] = elemVal
	}
	return NewMapValue(evaluatedMap), nil
}

func (i *Interpreter) evaluateEvalNode(n *EvalNode) (Value, error) {
	argValueRaw, err := i.evaluateExpression(n.Argument)
	if err != nil {
		return nil, err
	}
	argStr, _ := toString(argValueRaw)
	resolvedStr, resolveErr := i.resolvePlaceholdersWithError(argStr)
	if resolveErr != nil {
		return nil, NewRuntimeError(ErrorCodeEvaluation, "resolving placeholders during EVAL", resolveErr).WithPosition(n.Pos)
	}
	return StringValue{Value: resolvedStr}, nil
}

func (i *Interpreter) evaluateTypeOf(n *TypeOfNode) (Value, error) {
	argValue, err := i.evaluateExpression(n.Argument)
	if err != nil {
		if errors.Is(err, ErrVariableNotFound) {
			argValue = NilValue{}
		} else {
			return nil, err
		}
	}
	return StringValue{Value: string(TypeOf(argValue))}, nil
}

func (i *Interpreter) evaluateCall(n *CallableExprNode) (Value, error) {
	if n.Target.IsTool {
		tool, found := i.ToolRegistry().GetTool(n.Target.Name)
		if !found {
			return nil, NewRuntimeError(ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", n.Target.Name), ErrToolNotFound).WithPosition(n.Pos)
		}
		namedArgs := make(map[string]Value)
		specArgs := tool.Spec.Args
		if len(n.Arguments) > len(specArgs) && !tool.Spec.Variadic {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("tool '%s' expects at most %d arguments, got %d", tool.Spec.Name, len(specArgs), len(n.Arguments)), ErrArgumentMismatch).WithPosition(n.Pos)
		}
		for idx, argNode := range n.Arguments {
			if idx < len(specArgs) {
				argName := specArgs[idx].Name
				argValue, err := i.evaluateExpression(argNode)
				if err != nil {
					return nil, err
				}
				namedArgs[argName] = argValue
			}
		}
		return i.ExecuteTool(n.Target.Name, namedArgs)
	}
	evaluatedArgs := make([]Value, len(n.Arguments))
	for idx, argNode := range n.Arguments {
		var err error
		evaluatedArgs[idx], err = i.evaluateExpression(argNode)
		if err != nil {
			return nil, err
		}
	}
	return i.evaluateUserOrBuiltInFunction(n.Target.Name, evaluatedArgs, n.Pos)
}

func (i *Interpreter) evaluateUserOrBuiltInFunction(funcName string, args []Value, pos *Position) (Value, error) {
	if isBuiltInFunction(funcName) {
		unwrappedArgs := make([]interface{}, len(args))
		for i, v := range args {
			unwrappedArgs[i] = Unwrap(v)
		}
		result, err := evaluateBuiltInFunction(funcName, unwrappedArgs)
		if err != nil {
			if _, ok := err.(*RuntimeError); !ok {
				err = NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("built-in function '%s' failed", funcName), err).WithPosition(pos)
			}
			return nil, err
		}
		return Wrap(result)
	}
	procResult, procErr := i.RunProcedure(funcName, args...)
	if procErr != nil {
		if re, ok := procErr.(*RuntimeError); ok {
			return nil, re.WithPosition(pos)
		}
		return nil, NewRuntimeError(ErrorCodeProcNotFound, fmt.Sprintf("error calling procedure '%s'", funcName), procErr).WithPosition(pos)
	}
	i.lastCallResult = procResult
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
			firstErr = NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' not found in placeholder", varName), ErrVariableNotFound)
			return ""
		}
		return val.String()
	})
	if firstErr != nil {
		return "", firstErr
	}
	return resolved, nil
}

// GetTypeConstant is now a standalone function as it does not depend on interpreter state.
func GetTypeConstant(name string) (string, bool) {
	switch name {
	case "TYPE_STRING":
		return string(TypeString), true
	case "TYPE_NUMBER":
		return string(TypeNumber), true
	case "TYPE_BOOLEAN":
		return string(TypeBoolean), true
	case "TYPE_LIST":
		return string(TypeList), true
	case "TYPE_MAP":
		return string(TypeMap), true
	case "TYPE_NIL":
		return string(TypeNil), true
	case "TYPE_FUNCTION":
		return string(TypeFunction), true
	case "TYPE_TOOL":
		return string(TypeTool), true
	case "TYPE_ERROR":
		return string(TypeError), true
	case "TYPE_EVENT":
		return string(TypeEvent), true
	case "TYPE_TIMEDATE":
		return string(TypeTimedate), true
	case "TYPE_FUZZY":
		return string(TypeFuzzy), true
	case "TYPE_UNKNOWN":
		return string(TypeUnknown), true
	case "TYPE_BYTES":
		return string(TypeBytes), true
	}
	return "", false
}
