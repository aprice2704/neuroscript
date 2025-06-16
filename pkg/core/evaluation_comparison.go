// NeuroScript Version: 0.3.5
// File version: 1.0.0
// Purpose: Corrected missing type assertion on condNode before passing to evaluateExpression.
// filename: pkg/core/evaluation_comparison.go

package core

import "fmt"

// evaluateComparison handles logical and comparison operators (==, !=, <, >, etc.).
func (i *Interpreter) evaluateComparison(left, right, op string) (bool, error) {
	// This is a simplified placeholder. A full implementation would handle
	// type coercion and comparisons for all supported types.
	return left == right, nil
}

// isTruthy determines the boolean value of a condition in an if/while statement.
func (i *Interpreter) isTruthy(condNode interface{}) (bool, error) {
	// FIX: The condNode must be asserted to an Expression before being evaluated.
	condExpr, ok := condNode.(Expression)
	if !ok {
		return false, fmt.Errorf("internal error: condition node is not an Expression, but %T", condNode)
	}

	val, err := i.evaluateExpression(condExpr)
	if err != nil {
		return false, err
	}
	return IsTruthy(val), nil
}
