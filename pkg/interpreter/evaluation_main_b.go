// NeuroScript Version: 0.8.0
// File version: 68
// Purpose: FIX: Uses the ToolRegistry() accessor instead of the removed 'tools' field.
// filename: pkg/interpreter/evaluation_main_b.go
// nlines: 148
// risk_rating: HIGH

package interpreter

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

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

		toolImpl, found := e.i.ToolRegistry().GetTool(toolNameForLookup)
		if !found {
			errMessage := fmt.Sprintf("tool '%s' not found (looked up as '%s')", n.Target.Name, toolNameForLookup)
			e.i.Logger().Errorf("[DEBUG] Point A: Tool not found error created in evaluateCall: %s", errMessage) // DEBUG
			return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, errMessage, lang.ErrToolNotFound).WithPosition(n.StartPos)
		}

		// --- POLICY GATE ---
		execPolicy := e.i.parcel.Policy()
		if execPolicy != nil {
			meta := policy.ToolMeta{
				Name:          strings.ToLower(string(toolImpl.FullName)),
				RequiresTrust: toolImpl.RequiresTrust,
				RequiredCaps:  toolImpl.RequiredCaps,
				Effects:       toolImpl.Effects,
			}
			specFetcher := func(name string) (policy.ToolSpecProvider, bool) {
				impl, found := e.i.ToolRegistry().GetTool(types.FullName(name))
				if !found {
					return nil, false
				}
				return impl.Spec, true
			}
			if err := policy.CanCall(execPolicy, meta, specFetcher); err != nil {
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
