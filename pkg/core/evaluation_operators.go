// pkg/core/evaluation_operators.go
package core

import (
	"fmt"
	"math"
	//"reflect" // Can use type assertion instead
)

// --- Operator Evaluation Logic ---

// performArithmetic handles +, -, *, /, %, **
func performArithmetic(left, right interface{}, op string) (interface{}, error) {
	// (This function remains the same as the previous version)
	leftF, leftIsNumeric := toFloat64(left)
	rightF, rightIsNumeric := toFloat64(right)
	leftI, leftIsInt := toInt64(left)
	rightI, rightIsInt := toInt64(right)

	if !leftIsNumeric || !rightIsNumeric {
		if op == "+" {
			return performStringConcatOrNumericAdd(left, right)
		}
		return nil, fmt.Errorf("operator '%s' requires numeric operands, got %T and %T", op, left, right)
	}

	useFloat := !leftIsInt || !rightIsInt

	switch op {
	case "+":
		if useFloat {
			return leftF + rightF, nil
		}
		return leftI + rightI, nil
	case "-":
		if useFloat {
			return leftF - rightF, nil
		}
		return leftI - rightI, nil
	case "*":
		if useFloat {
			return leftF * rightF, nil
		}
		return leftI * rightI, nil
	case "/":
		if rightF == 0.0 {
			return nil, fmt.Errorf("division by zero")
		}
		if leftIsInt && rightIsInt && leftI%rightI == 0 {
			return leftI / rightI, nil
		}
		return leftF / rightF, nil
	case "%":
		if leftIsInt && rightIsInt {
			if rightI == 0 {
				return nil, fmt.Errorf("division by zero in modulo")
			}
			return leftI % rightI, nil
		}
		return nil, fmt.Errorf("operator '%%' requires integer operands, got %T and %T", left, right)
	case "**":
		return math.Pow(leftF, rightF), nil
	default:
		return nil, fmt.Errorf("unknown arithmetic operator '%s'", op)
	}
}

// performStringConcatOrNumericAdd decides between string concat and numeric add for '+'
func performStringConcatOrNumericAdd(left, right interface{}) (interface{}, error) {
	// (This function remains the same as the previous version)
	_, leftIsString := left.(string)
	_, rightIsString := right.(string)
	leftF, leftIsNumeric := toFloat64(left)
	rightF, rightIsNumeric := toFloat64(right)
	leftI, leftIsInt := toInt64(left)
	rightI, rightIsInt := toInt64(right)

	if leftIsString || rightIsString {
		if leftIsNumeric && rightIsNumeric { // Both convertible to number? Prefer numeric add.
			if leftIsInt && rightIsInt {
				return leftI + rightI, nil
			}
			return leftF + rightF, nil
		}
		// Otherwise, concatenate (handle nil correctly)
		leftStr := ""
		if left != nil {
			leftStr = fmt.Sprintf("%v", left)
		}
		rightStr := ""
		if right != nil {
			rightStr = fmt.Sprintf("%v", right)
		}
		return leftStr + rightStr, nil
	}

	// Neither is string, perform numeric addition
	if leftIsNumeric && rightIsNumeric {
		if leftIsInt && rightIsInt {
			return leftI + rightI, nil
		}
		return leftF + rightF, nil
	}

	return nil, fmt.Errorf("operator '+' cannot operate on types %T and %T", left, right)
}

// performComparison handles ==, !=, >, <, >=, <=
func performComparison(left, right interface{}, op string) (bool, error) {
	// (This function remains the same as the previous version)
	leftF, leftIsNumeric := toFloat64(left)
	rightF, rightIsNumeric := toFloat64(right)
	leftI, leftIsInt := toInt64(left)
	rightI, rightIsInt := toInt64(right)

	switch op {
	case "==":
		if leftIsNumeric && rightIsNumeric {
			if leftIsInt && rightIsInt {
				return leftI == rightI, nil
			}
			return leftF == rightF, nil
		}
		return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right), nil
	case "!=":
		if leftIsNumeric && rightIsNumeric {
			if leftIsInt && rightIsInt {
				return leftI != rightI, nil
			}
			return leftF != rightF, nil
		}
		return fmt.Sprintf("%v", left) != fmt.Sprintf("%v", right), nil
	case "<", ">", "<=", ">=":
		if leftIsNumeric && rightIsNumeric {
			switch op {
			case "<":
				return leftF < rightF, nil
			case ">":
				return leftF > rightF, nil
			case "<=":
				return leftF <= rightF, nil
			case ">=":
				return leftF >= rightF, nil
			}
		}
		return false, fmt.Errorf("comparison operator '%s' requires numeric operands, got %T and %T", op, left, right)
	default:
		return false, fmt.Errorf("unknown comparison operator '%s'", op)
	}
}

// performLogical handles AND, OR
func performLogical(left, right interface{}, op string) (bool, error) {
	// (This function remains the same as the previous version)
	switch op {
	case "AND":
		return isTruthy(left) && isTruthy(right), nil // Non-short-circuit version
	case "OR":
		return isTruthy(left) || isTruthy(right), nil // Non-short-circuit version
	default:
		return false, fmt.Errorf("unknown logical operator '%s'", op)
	}
}

// performBitwise handles &, |, ^
// *** MODIFIED: Stricter type check ***
func performBitwise(left, right interface{}, op string) (int64, error) {
	// Use type assertion to ensure both operands are actually int64
	leftI, leftOk := left.(int64)
	rightI, rightOk := right.(int64)

	// Proceed only if both type assertions succeeded
	if leftOk && rightOk {
		switch op {
		case "&":
			return leftI & rightI, nil
		case "|":
			return leftI | rightI, nil
		case "^":
			return leftI ^ rightI, nil
		default:
			return 0, fmt.Errorf("unknown bitwise operator '%s'", op) // Should not happen
		}
	}

	// If type assertion failed, return an error
	// Determine which operand(s) had the wrong type for the error message
	errMsg := "bitwise operator '%s' requires integer (int64) operands"
	if !leftOk && !rightOk {
		errMsg += ", got %T and %T"
		return 0, fmt.Errorf(errMsg, op, left, right)
	} else if !leftOk {
		errMsg += ", got %T for left operand"
		return 0, fmt.Errorf(errMsg, op, left)
	} else { // !rightOk
		errMsg += ", got %T for right operand"
		return 0, fmt.Errorf(errMsg, op, right)
	}
}
