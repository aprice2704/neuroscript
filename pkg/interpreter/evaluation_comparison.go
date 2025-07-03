// NeuroScript Version: 0.5.2
// File version: 2.0.0
// Purpose: Corrected missing type assertion on condNode before passing to evaluate.Expression.
// filename: pkg/interpreter/evaluation_comparison.go
// nlines: 30
// risk_rating: MEDIUM

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// evaluateComparison is a placeholder for a more complex implementation.
// The primary logic is now in evaluateBinaryOp and the performComparison helper.
func (i *Interpreter) evaluateComparison(left, right, op string) (bool, error) {
	// This is a simplified placeholder. A full implementation would handle
	// type coercion and comparisons for all supported types.
	return left == right, nil
}

// isTruthy determines the boolean value of a condition in an if/while statement.
func (i *Interpreter) isTruthy(condNode interface{}) (bool, error) {
	// FIX: The condNode must be asserted to an ast.Expression before being evaluated.
	condExpr, ok := condNode.(ast.Expression)
	if !ok {
		return false, fmt.Errorf("internal error: condition node is not an ast.Expression, but %T", condNode)
	}

	val, err := i.evaluate.Expression(condExpr)
	if err != nil {
		return false, err
	}
	return lang.IsTruthy(val), nil
}