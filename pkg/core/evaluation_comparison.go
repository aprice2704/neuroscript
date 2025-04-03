// pkg/core/evaluation_comparison.go
package core

import (
	"fmt"
	"strings"
)

// performComparison performs comparisons (==, !=, >, <, >=, <=) between evaluated values.
func performComparison(leftVal, rightVal interface{}, operator string) (bool, error) {
	leftIsNil := leftVal == nil
	rightIsNil := rightVal == nil

	if leftIsNil || rightIsNil { /* ... nil handling as before ... */
		switch operator {
		case "==":
			return leftIsNil && rightIsNil, nil
		case "!=":
			return !(leftIsNil && rightIsNil), nil
		case ">", "<", ">=", "<=":
			return false, fmt.Errorf("operator '%s' requires non-nil operands", operator)
		default:
			return false, fmt.Errorf("unsupported comparison operator '%s' with nil", operator)
		}
	}

	// *** ADDED: Explicit string conversion before fmt.Sprintf for safety ***
	var leftStr, rightStr string
	if s, ok := leftVal.(string); ok {
		leftStr = s
	} else {
		leftStr = fmt.Sprintf("%v", leftVal)
	}
	if s, ok := rightVal.(string); ok {
		rightStr = s
	} else {
		rightStr = fmt.Sprintf("%v", rightVal)
	}

	switch operator {
	case "==":
		// Compare string representations
		return leftStr == rightStr, nil
	case "!=":
		return leftStr != rightStr, nil
	case ">", "<", ">=", "<=":
		// Attempt numeric comparison using the string representations
		leftFloat, leftOk := tryParseFloat(leftStr)
		rightFloat, rightOk := tryParseFloat(rightStr)
		if leftOk && rightOk {
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
		return false, fmt.Errorf("operator '%s' requires numeric operands, got %T(%v) and %T(%v)", operator, leftVal, leftVal, rightVal, rightVal)
	default:
		return false, fmt.Errorf("unsupported comparison operator '%s'", operator)
	}
}

// evaluateCondition: Uses performComparison. String truthiness check remains strict ("true" or "1").
func (i *Interpreter) evaluateCondition(condNode interface{}) (bool, error) {
	if compNode, ok := condNode.(ComparisonNode); ok {
		leftValue, errLeft := i.evaluateExpression(compNode.Left)
		if errLeft != nil {
			if strings.Contains(errLeft.Error(), "not found") {
				leftValue = nil
			} else {
				return false, fmt.Errorf("evaluating left side: %w", errLeft)
			}
		}
		rightValue, errRight := i.evaluateExpression(compNode.Right)
		if errRight != nil {
			if strings.Contains(errRight.Error(), "not found") {
				rightValue = nil
			} else {
				return false, fmt.Errorf("evaluating right side: %w", errRight)
			}
		}

		result, compErr := performComparison(leftValue, rightValue, compNode.Operator)
		if compErr != nil {
			return false, fmt.Errorf("condition comparison: %w", compErr)
		}
		return result, nil
	}

	evaluatedValue, errEval := i.evaluateExpression(condNode)
	if errEval != nil {
		if strings.Contains(errEval.Error(), "not found") {
			return false, nil
		} else {
			return false, fmt.Errorf("evaluating condition expression: %w", errEval)
		}
	}

	switch v := evaluatedValue.(type) {
	case bool:
		return v, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0.0, nil
	case string:
		lowerV := strings.ToLower(v)
		if lowerV == "true" || v == "1" {
			return true, nil
		} // Strict check
		return false, nil
	default:
		return false, nil
	}
}
