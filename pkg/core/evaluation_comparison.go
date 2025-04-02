// pkg/core/evaluation_comparison.go
package core

import (
	"fmt"
	"strings"
)

// performComparison performs comparisons (==, !=, >, <, >=, <=) between evaluated values.
// Moved from evaluation.go during refactoring.
// It uses string comparison for ==/!= and attempts numeric comparison for others.
func performComparison(leftVal, rightVal interface{}, operator string) (bool, error) {
	// Handle nil comparisons first
	// ==: true if both are nil, false otherwise
	// !=: false if both are nil, true otherwise
	// >, <, >=, <=: error if either is nil
	leftIsNil := leftVal == nil
	rightIsNil := rightVal == nil

	if leftIsNil || rightIsNil {
		switch operator {
		case "==":
			return leftIsNil && rightIsNil, nil // True only if both are nil
		case "!=":
			return !(leftIsNil && rightIsNil), nil // False only if both are nil
		case ">", "<", ">=", "<=":
			// Numeric comparisons are invalid if either operand is nil
			return false, fmt.Errorf("operator '%s' requires non-nil numeric operands, received %T (%v) and %T (%v)",
				operator, leftVal, leftVal, rightVal, rightVal)
		default:
			// Should not happen if grammar is correct, but handle defensively
			return false, fmt.Errorf("unsupported comparison operator '%s' with nil operand(s)", operator)
		}
	}

	// If neither is nil, proceed with type-based comparison
	// Convert non-nil values to string representation first for general comparison
	leftStr := fmt.Sprintf("%v", leftVal)
	rightStr := fmt.Sprintf("%v", rightVal)

	switch operator {
	case "==":
		// Consider type for non-numeric comparison? For now, string compare is fine.
		return leftStr == rightStr, nil
	case "!=":
		return leftStr != rightStr, nil
	case ">", "<", ">=", "<=":
		// Attempt numeric comparison for >, <, >=, <=
		leftFloat, leftOk := tryParseFloat(leftStr)    // Uses helper from evaluation_helpers.go
		rightFloat, rightOk := tryParseFloat(rightStr) // Uses helper from evaluation_helpers.go

		if leftOk && rightOk { // Both successfully parsed as numbers
			switch operator {
			case ">":
				return leftFloat > rightFloat, nil
			case "<":
				return leftFloat < rightFloat, nil
			case ">=":
				return leftFloat >= rightFloat, nil
			case "<=":
				return leftFloat <= rightFloat, nil
			}
		}
		// If one or both could not be parsed as float, numeric comparison is invalid
		return false, fmt.Errorf("operator '%s' requires numeric operands, but received %T (%v) and %T (%v)",
			operator, leftVal, leftVal, rightVal, rightVal)

	default:
		return false, fmt.Errorf("unsupported comparison operator '%s'", operator)
	}
	// Should not be reached, but satisfy compiler
	// return false, fmt.Errorf("internal error during comparison logic")
}

// evaluateCondition evaluates an AST node intended to be a boolean condition for IF/WHILE.
// Now returns (bool, error) to propagate evaluation errors.
func (i *Interpreter) evaluateCondition(condNode interface{}) (bool, error) {

	// Check if it's a comparison operation first
	if compNode, ok := condNode.(ComparisonNode); ok {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        Evaluating Condition (ComparisonNode): Left=%T, Op=%q, Right=%T", compNode.Left, compNode.Operator, compNode.Right)
		}
		// Evaluate the left and right hand side expressions, checking for errors
		leftValue, errLeft := i.evaluateExpression(compNode.Left)
		if errLeft != nil {
			// Check if error is due to 'variable not found' - treat this as evaluating to nil for comparison context
			if strings.Contains(errLeft.Error(), "not found") {
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]          Comparison LHS evaluated to nil due to error: %v", errLeft)
				}
				leftValue = nil // Treat as nil for comparison
			} else {
				// Propagate other evaluation errors
				return false, fmt.Errorf("evaluating left side of comparison: %w", errLeft)
			}
		}
		rightValue, errRight := i.evaluateExpression(compNode.Right)
		if errRight != nil {
			// Check if error is due to 'variable not found' - treat this as evaluating to nil for comparison context
			if strings.Contains(errRight.Error(), "not found") {
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]          Comparison RHS evaluated to nil due to error: %v", errRight)
				}
				rightValue = nil // Treat as nil for comparison
			} else {
				// Propagate other evaluation errors
				return false, fmt.Errorf("evaluating right side of comparison: %w", errRight)
			}
		}

		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]          Comparison LHS => %v (%T)", leftValue, leftValue)
			i.logger.Printf("[DEBUG-INTERP]          Comparison RHS => %v (%T)", rightValue, rightValue)
		}

		// Perform the actual comparison
		result, compErr := performComparison(leftValue, rightValue, compNode.Operator)
		if compErr != nil {
			// Comparison itself failed (e.g., non-numeric types for >)
			return false, fmt.Errorf("condition comparison failed: %w", compErr)
		}
		return result, nil // Return comparison result and nil error
	}

	// --- Single Expression Condition ---
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]        Evaluating Condition (Single Expression): Node=%T", condNode)
	}
	// Evaluate the expression, checking for errors
	evaluatedValue, errEval := i.evaluateExpression(condNode)
	if errEval != nil {
		// Check if error is due to 'variable not found' - treat this as evaluating to nil -> false condition
		if strings.Contains(errEval.Error(), "not found") {
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]          Single Expression Condition evaluated to false due to error: %v", errEval)
			}
			return false, nil // Condition is false, no error propagated here
		}
		// Propagate other evaluation errors as condition evaluation failure
		return false, fmt.Errorf("evaluating condition expression: %w", errEval)
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]          Single Expression => %v (%T)", evaluatedValue, evaluatedValue)
	}

	// Determine truthiness based on the evaluated value
	switch v := evaluatedValue.(type) {
	case bool:
		return v, nil // Direct boolean value
	case int64:
		return v != 0, nil // Numeric: 0 is false, others true
	case float64:
		return v != 0.0, nil // Numeric: 0.0 is false, others true
	case string:
		// Handle "true" and "false" strings case-insensitively
		lowerV := strings.ToLower(v)
		if lowerV == "true" {
			return true, nil
		}
		if lowerV == "false" {
			return false, nil
		}
		// Non-true/false strings are considered falsy, return false without error
		return false, nil
	case nil:
		// Evaluating nil returns false condition without error
		return false, nil
	default:
		// All other types (lists, maps etc.) are considered falsy without error
		return false, nil
	}
}
