// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: This file now compiles as the 'Required' field has been added to eval.ArgSpec.
// filename: pkg/eval/helpers_eval.go
// nlines: 80
// risk_rating: HIGH

package eval

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (e *evaluation) mapArgsToSpec(toolName types.FullName, args []lang.Value, node *ast.CallableExprNode) (map[string]lang.Value, error) {
	spec, ok := e.rt.GetToolSpec(toolName)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", toolName), lang.ErrToolNotFound).WithPosition(node.GetPos())
	}
	namedArgs := make(map[string]lang.Value)
	for i, argSpec := range spec.Args {
		if i < len(args) {
			namedArgs[argSpec.Name] = args[i]
		} else if argSpec.Required {
			return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("missing required argument '%s' for tool '%s'", argSpec.Name, toolName), lang.ErrAssignCountMismatch).WithPosition(node.GetPos())
		}
	}
	return namedArgs, nil
}

func (e *evaluation) evaluateAccessorKey(accessor *ast.AccessorNode) (string, error) {
	if accessor.Type == ast.DotAccess {
		if strLiteral, ok := accessor.Key.(*ast.StringLiteralNode); ok {
			return strLiteral.Value, nil
		}
		return strings.TrimPrefix(accessor.Key.String(), "."), nil
	}
	keyVal, err := e.Expression(accessor.Key)
	if err != nil {
		return "", lang.WrapErrorWithPosition(err, accessor.Key.GetPos(), "evaluating map key")
	}
	key, _ := lang.ToString(keyVal)
	return key, nil
}

func (e *evaluation) evaluateAccessorIndex(accessor *ast.AccessorNode) (int64, error) {
	indexVal, err := e.Expression(accessor.Key)
	if err != nil {
		return 0, lang.WrapErrorWithPosition(err, accessor.Key.GetPos(), "evaluating list index")
	}
	index, isInt := lang.ToInt64(indexVal)
	if !isInt {
		return 0, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("list index must be an integer, got %s", lang.TypeOf(indexVal)), lang.ErrListInvalidIndexType).WithPosition(accessor.Key.GetPos())
	}
	if index < 0 {
		return 0, lang.NewRuntimeError(lang.ErrorCodeBounds, fmt.Sprintf("list index cannot be negative, got %d", index), lang.ErrListIndexOutOfBounds).WithPosition(accessor.Key.GetPos())
	}
	return index, nil
}

func resolveToolName(n *ast.CallableExprNode) (types.FullName, error) {
	if !n.Target.IsTool {
		return "", fmt.Errorf("internal error: resolveToolName called on a non-tool expression")
	}
	if n.Target.Name == "" {
		return "", fmt.Errorf("internal error: tool call expression has an empty target name")
	}
	if strings.HasPrefix(n.Target.Name, "tool.") {
		return types.FullName(strings.ToLower(n.Target.Name)), nil
	}
	return types.FullName(strings.ToLower("tool." + n.Target.Name)), nil
}
