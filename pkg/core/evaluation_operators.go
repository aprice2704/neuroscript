// filename: pkg/core/evaluation_operators.go
package core

import (
	"fmt"
	"math"
	"reflect" // Import reflect for type checking
	// Assuming NsError, RuntimeError, error codes, and error constants are defined in errors.go
	// Assuming type conversion helpers (toInt64, toFloat64, toString), isTruthy, ToNumeric are defined in evaluation_helpers.go
)

// --- Operator Evaluation Logic ---

// isIntegerType checks if the underlying kind is any Go integer type.
func isIntegerType(v interface{}) bool {
	if v == nil {
		return false
	}
	valType := reflect.TypeOf(v)
	if valType.Kind() == reflect.Ptr {
		valType = valType.Elem()
	}
	k := valType.Kind()
	return k >= reflect.Int && k <= reflect.Uint64
}

// performArithmetic handles -, *, /, %, **
func performArithmetic(left, right interface{}, op string) (interface{}, error) {
	if left == nil || right == nil {
		return nil, fmt.Errorf("%w: operator '%s' requires non-nil operands, got %T and %T", ErrNilOperand, op, left, right)
	}

	if op == "%" {
		if !isIntegerType(left) || !isIntegerType(right) {
			return nil, fmt.Errorf("operator '%%' requires integer operands, got %T and %T: %w", left, right, ErrInvalidOperandTypeInteger)
		}
		leftI, leftOk := toInt64(left)
		rightI, rightOk := toInt64(right)
		if !leftOk || !rightOk {
			return nil, fmt.Errorf("internal error: failed int64 conversion for modulo operands '%v' (%T) and '%v' (%T)", left, left, right, right)
		}
		if rightI == 0 {
			return nil, fmt.Errorf("modulo by zero: %w", ErrDivisionByZero)
		}
		return leftI % rightI, nil
	}

	_, leftOriginallyNumeric := ToNumeric(left)
	_, rightOriginallyNumeric := ToNumeric(right)
	if !leftOriginallyNumeric || !rightOriginallyNumeric {
		return nil, fmt.Errorf("operator '%s' requires numeric operands, got %T and %T: %w", op, left, right, ErrInvalidOperandTypeNumeric)
	}

	leftF, leftFloatOk := toFloat64(left)
	rightF, rightFloatOk := toFloat64(right)
	leftI, leftIntOk := toInt64(left)
	rightI, rightIntOk := toInt64(right)
	useFloat := !leftIntOk || !rightIntOk

	switch op {
	case "-":
		if useFloat {
			if !leftFloatOk || !rightFloatOk {
				return nil, fmt.Errorf("internal error: failed float64 conversion for subtraction operands '%v' (%T) and '%v' (%T)", left, left, right, right)
			}
			return leftF - rightF, nil
		}
		return leftI - rightI, nil
	case "*":
		if useFloat {
			if !leftFloatOk || !rightFloatOk {
				return nil, fmt.Errorf("internal error: failed float64 conversion for multiplication operands '%v' (%T) and '%v' (%T)", left, left, right, right)
			}
			return leftF * rightF, nil
		}
		return leftI * rightI, nil
	case "/":
		if !rightFloatOk {
			return nil, fmt.Errorf("internal error: failed float64 conversion for division divisor '%v' (%T)", right, right)
		}
		if rightF == 0.0 {
			return nil, fmt.Errorf("division by zero: %w", ErrDivisionByZero)
		}
		performFloatDivision := useFloat || (rightI != 0 && leftI%rightI != 0)
		if performFloatDivision {
			if !leftFloatOk {
				return nil, fmt.Errorf("internal error: failed float64 conversion for division dividend '%v' (%T)", left, left)
			}
			return leftF / rightF, nil
		}
		return leftI / rightI, nil
	case "**":
		if !leftFloatOk || !rightFloatOk {
			return nil, fmt.Errorf("operator '**' requires operands convertible to float, got %T and %T: %w", left, right, ErrInvalidOperandTypeNumeric)
		}
		return math.Pow(leftF, rightF), nil
	default:
		return nil, fmt.Errorf("unknown arithmetic operator '%s': %w", op, ErrUnsupportedOperator)
	}
}

// performStringConcatOrNumericAdd decides between string concat and numeric add for '+'
// ** FINAL FIX: Removed incorrect checks on toString boolean return **
func performStringConcatOrNumericAdd(left, right interface{}) (interface{}, error) {
	_, leftIsNumeric := ToNumeric(left)
	_, rightIsNumeric := ToNumeric(right)

	// Case 1: Both operands are numeric -> Perform numeric addition
	if leftIsNumeric && rightIsNumeric {
		leftF, _ := toFloat64(left)
		rightF, _ := toFloat64(right)
		leftI, leftIsInt := toInt64(left)
		rightI, rightIsInt := toInt64(right)
		if leftIsInt && rightIsInt {
			return leftI + rightI, nil
		}
		return leftF + rightF, nil
	}

	// Case 2: At least one operand is NOT numeric. Check if string concatenation is possible.
	_, leftIsString := left.(string)
	_, rightIsString := right.(string)

	// If EITHER operand is a string, perform string concatenation.
	if leftIsString || rightIsString {
		// Convert BOTH operands to string using the helper.
		// We assume toString always returns a valid string representation.
		leftStr, _ := toString(left)   // Ignore the boolean return value
		rightStr, _ := toString(right) // Ignore the boolean return value

		// ** REMOVED THE INCORRECT !leftOk and !rightOk CHECKS **

		return leftStr + rightStr, nil // Concatenate the string representations
	}

	// Case 3: Not (both numeric) AND Not (at least one string). -> Error
	return nil, fmt.Errorf("operator '+' cannot operate on types %T and %T: %w", left, right, ErrInvalidOperandType)
}

// performComparison handles ==, !=, >, <, >=, <=
func performComparison(left, right interface{}, op string) (bool, error) {
	if op == "==" || op == "!=" {
		leftIsNil := left == nil
		rightIsNil := right == nil
		if leftIsNil != rightIsNil {
			return op == "!=", nil
		}
		if leftIsNil && rightIsNil {
			return op == "==", nil
		}

		leftNum, leftIsNumeric := ToNumeric(left)
		rightNum, rightIsNumeric := ToNumeric(right)
		if leftIsNumeric && rightIsNumeric {
			leftF, _ := toFloat64(leftNum)
			rightF, _ := toFloat64(rightNum)
			leftI, leftIsInt := toInt64(leftNum)
			rightI, rightIsInt := toInt64(rightNum)
			var isEqual bool
			if leftIsInt && rightIsInt {
				isEqual = leftI == rightI
			} else {
				isEqual = leftF == rightF
			}
			return (op == "==") == isEqual, nil
		}

		leftStr, _ := toString(left)
		rightStr, _ := toString(right)
		isEqual := leftStr == rightStr
		return (op == "==") == isEqual, nil
	}

	// Handle >, <, >=, <= (Strictly numeric comparison)
	if left == nil || right == nil {
		return false, fmt.Errorf("comparison operator '%s' requires non-nil operands: %w", op, ErrNilOperand)
	}
	_, leftOriginallyNumeric := ToNumeric(left)
	_, rightOriginallyNumeric := ToNumeric(right)
	if !leftOriginallyNumeric || !rightOriginallyNumeric {
		return false, fmt.Errorf("comparison operator '%s' requires numeric operands, got %T and %T: %w", op, left, right, ErrInvalidOperandTypeNumeric)
	}

	leftF, leftOk := toFloat64(left)
	rightF, rightOk := toFloat64(right)
	if !leftOk || !rightOk {
		return false, fmt.Errorf("internal error: failed float64 conversion for comparison operands '%v' (%T) and '%v' (%T)", left, left, right, right)
	}

	switch op {
	case "<":
		return leftF < rightF, nil
	case ">":
		return leftF > rightF, nil
	case "<=":
		return leftF <= rightF, nil
	case ">=":
		return leftF >= rightF, nil
	default:
		return false, fmt.Errorf("unknown comparison operator '%s': %w", op, ErrUnsupportedOperator)
	}
}

// performLogical handles 'and', 'or' (using truthiness)
func performLogical(left, right interface{}, op string) (bool, error) {
	leftBool := isTruthy(left)
	rightBool := isTruthy(right)
	switch op {
	case "and", "AND":
		return leftBool && rightBool, nil
	case "or", "OR":
		return leftBool || rightBool, nil
	default:
		return false, fmt.Errorf("unknown logical operator '%s': %w", op, ErrUnsupportedOperator)
	}
}

// performBitwise handles &, |, ^
func performBitwise(left, right interface{}, op string) (int64, error) {
	if left == nil || right == nil {
		return 0, fmt.Errorf("bitwise operator '%s' requires non-nil integer operands: %w", op, ErrNilOperand)
	}
	if !isIntegerType(left) || !isIntegerType(right) {
		errMsgFormat := "bitwise operator '%s' requires integer operands"
		if !isIntegerType(left) && !isIntegerType(right) {
			errMsgFormat += ", got %T and %T: %w"
			return 0, fmt.Errorf(errMsgFormat, op, left, right, ErrInvalidOperandTypeInteger)
		} else if !isIntegerType(left) {
			errMsgFormat += ", got %T for left operand: %w"
			return 0, fmt.Errorf(errMsgFormat, op, left, ErrInvalidOperandTypeInteger)
		} else {
			errMsgFormat += ", got %T for right operand: %w"
			return 0, fmt.Errorf(errMsgFormat, op, right, ErrInvalidOperandTypeInteger)
		}
	}
	leftI, leftOk := toInt64(left)
	rightI, rightOk := toInt64(right)
	if !leftOk || !rightOk {
		return 0, fmt.Errorf("internal error: failed int64 conversion for bitwise operands '%v' (%T) and '%v' (%T)", left, left, right, right)
	}
	switch op {
	case "&":
		return leftI & rightI, nil
	case "|":
		return leftI | rightI, nil
	case "^":
		return leftI ^ rightI, nil
	default:
		return 0, fmt.Errorf("unknown bitwise operator '%s': %w", op, ErrUnsupportedOperator)
	}
}

// --- Placeholders for helpers (ensure these exist, e.g., in evaluation_helpers.go) ---
// func toInt64(v interface{}) (int64, bool)
// func toFloat64(v interface{}) (float64, bool)
// func toString(v interface{}) (string, bool) // Returns string representation and bool if conversion was needed/possible
// func isTruthy(value interface{}) bool
// func ToNumeric(v interface{}) (interface{}, bool) // Checks if convertible to int64 or float64, returns the number if so
