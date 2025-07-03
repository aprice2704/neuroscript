// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Removed duplicate method declarations to resolve compiler errors.
// filename: pkg/interpreter/evaluation_resolve.go
// nlines: 35
// risk_rating: LOW

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// resolveValue handles resolving variable names, placeholders, and literals to their actual values.
// This function is now simplified as the more complex resolution logic has been
// consolidated into methods on the Interpreter in evaluation_main.go.
func (e *evaluation) resolveValue(node ast.Expression) (lang.Value, error) {
	switch n := node.(type) {
	case *ast.VariableNode:
		return e.i.resolveVariable(n)
	case *ast.PlaceholderNode:
		return e.i.resolvePlaceholder(n)
	case *ast.EvalNode:
		return e.i.lastCallResult, nil
	case *ast.StringLiteralNode:
		return e.evaluateStringLiteral(n)
	case *ast.NumberLiteralNode:
		val, _ := lang.ToFloat64(n.Value)
		return lang.NumberValue{Value: val}, nil
	case *ast.BooleanLiteralNode:
		return lang.BoolValue{Value: n.Value}, nil
	case *ast.NilLiteralNode:
		return &lang.NilValue{}, nil
	default:
		return nil, fmt.Errorf("internal error: resolveValue received unexpected node type %T", n)
	}
}