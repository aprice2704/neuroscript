// pkg/core/evaluation_comparison.go
package core

import (
	"fmt"
	"strings" // Keep this import
)

// performComparison is NO LONGER USED directly by evaluateCondition.
// Its logic is integrated into evaluateBinaryOp in evaluation_logic.go.
// func performComparison(leftVal, rightVal interface{}, operator string) (bool, error) { ... }

// evaluateCondition evaluates an expression node used in IF/WHILE contexts.
// It now relies on evaluateExpression and the isTruthy helper.
func (i *Interpreter) evaluateCondition(condNode interface{}) (bool, error) {
	// The condition node is now just a standard expression node (e.g., BinaryOpNode, VariableNode, LiteralNode).
	// We evaluate it directly. The result of comparisons, AND, OR etc. will be a boolean.
	// For other expression types, we check their truthiness.

	evaluatedValue, errEval := i.evaluateExpression(condNode) // Evaluate the whole condition expression
	if errEval != nil {
		// Allow "variable not found" to evaluate to false in conditions
		if strings.Contains(errEval.Error(), "not found") {
			if i.logger != nil {
				i.logger.Debug("-INTERP]        Condition variable not found, evaluating as false: %v", errEval)
			}
			return false, nil
		} else {
			// Propagate other evaluation errors
			return false, fmt.Errorf("evaluating condition expression: %w", errEval)
		}
	}

	// Determine truthiness of the final evaluated value
	result := isTruthy(evaluatedValue)
	if i.logger != nil {
		i.logger.Debug("-INTERP]        Condition node %T evaluated to %v (%T), truthiness: %t", condNode, evaluatedValue, evaluatedValue, result)
	}
	return result, nil
}

// tryParseFloat moved to evaluation_helpers.go or evaluation_logic.go (if still needed there)
// func tryParseFloat(s string) (float64, bool) { ... }
