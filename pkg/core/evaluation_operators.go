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
	// Use reflect.Kind for robustness against different int types (int, int8, int16, int32, int64, uint etc.)
	// Ensure we handle potential pointers if needed, though unlikely for numeric literals/vars
	valType := reflect.TypeOf(v)
	if valType.Kind() == reflect.Ptr {
		valType = valType.Elem() // Dereference pointer type
	}
	k := valType.Kind()
	return k >= reflect.Int && k <= reflect.Uint64 // Covers all standard signed/unsigned ints
}

// performArithmetic handles -, *, /, %, **
func performArithmetic(left, right interface{}, op string) (interface{}, error) {
	// Check for nil operands EARLY
	if op != "+" { // '+' is handled separately
		if left == nil || right == nil {
			return nil, fmt.Errorf("%w: operator '%s' requires non-nil operands", ErrNilOperand, op)
		}
	}

	// --- Special Handling for Modulo (%) ---
	if op == "%" {
		// Modulo specifically requires integer operands. Check original types FIRST.
		if !isIntegerType(left) || !isIntegerType(right) {
			return nil, fmt.Errorf("operator '%%' requires integer operands, got %T and %T: %w", left, right, ErrInvalidOperandTypeInteger)
		}
		// If types are correct, attempt conversion to int64
		leftI, leftOk := toInt64(left)
		rightI, rightOk := toInt64(right)
		if !leftOk || !rightOk {
			// This suggests an internal inconsistency if isIntegerType passed
			return nil, fmt.Errorf("internal error: failed int64 conversion for modulo operands '%v' (%T) and '%v' (%T) despite passing integer type check", left, left, right, right)
		}
		if rightI == 0 {
			return nil, fmt.Errorf("modulo by zero: %w", ErrDivisionByZero)
		}
		return leftI % rightI, nil // Perform modulo
	}
	// --- End Special Handling for Modulo (%) ---

	// --- Handling for other arithmetic operators -, *, /, ** ---
	leftF, leftFloatOk := toFloat64(left)
	rightF, rightFloatOk := toFloat64(right)
	leftI, leftIntOk := toInt64(left)
	rightI, rightIntOk := toInt64(right)

	// Check if *originally* numeric before arithmetic
	_, leftOriginallyNumeric := ToNumeric(left)
	_, rightOriginallyNumeric := ToNumeric(right)

	if !leftOriginallyNumeric || !rightOriginallyNumeric {
		// This check is now primarily for -, *, /, **
		return nil, fmt.Errorf("operator '%s' requires numeric operands, got %T and %T: %w", op, left, right, ErrInvalidOperandTypeNumeric)
	}

	// Determine if float arithmetic is needed for -, *, /
	// Use float if either operand is float or if integer division would truncate
	useFloat := !leftIntOk || !rightIntOk // Use float if either failed direct int conversion

	switch op {
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
			return nil, fmt.Errorf("division by zero: %w", ErrDivisionByZero)
		}
		// Determine if float division is necessary
		// Need float if either operand isn't an int OR if it's int division with remainder
		performFloatDivision := useFloat || (rightI != 0 && leftI%rightI != 0)

		if performFloatDivision {
			// Ensure we use float representations even if originally int
			if !leftFloatOk { // Should ideally not happen if originally numeric
				return nil, fmt.Errorf("internal error: failed float64 conversion for numeric division operand '%v' (%T)", left, left)
			}
			if !rightFloatOk {
				return nil, fmt.Errorf("internal error: failed float64 conversion for numeric division operand '%v' (%T)", right, right)
			}
			return leftF / rightF, nil
		}

		// Perform integer division
		if rightI == 0 { // Belt-and-suspenders check
			return nil, fmt.Errorf("division by zero: %w", ErrDivisionByZero)
		}
		return leftI / rightI, nil
	case "**":
		// Ensure operands convert to float for Pow
		if !leftFloatOk || !rightFloatOk {
			return nil, fmt.Errorf("operator '**' requires operands convertible to float, got %T and %T: %w", left, right, ErrInvalidOperandTypeNumeric)
		}
		return math.Pow(leftF, rightF), nil // Exponentiation uses float
	default:
		// Should not be reached if '+' and '%' are handled earlier
		return nil, fmt.Errorf("unknown arithmetic operator '%s': %w", op, ErrUnsupportedOperator)
	}
}

// performStringConcatOrNumericAdd decides between string concat and numeric add for '+'
func performStringConcatOrNumericAdd(left, right interface{}) (interface{}, error) {
	// Attempt numeric conversions FIRST
	leftF, leftIsNumeric := toFloat64(left)
	rightF, rightIsNumeric := toFloat64(right)

	// Case 1: Both operands convert successfully to numeric -> Perform numeric addition
	if leftIsNumeric && rightIsNumeric {
		leftI, leftIsInt := toInt64(left)
		rightI, rightIsInt := toInt64(right)
		if leftIsInt && rightIsInt { // Prefer integer addition if possible
			return leftI + rightI, nil
		}
		return leftF + rightF, nil // Otherwise, use float addition
	}

	// Case 2: At least one operand does NOT convert to numeric. Check for strings.
	leftStr, leftWasString := toString(left) // toString handles nil -> ""
	rightStr, rightWasString := toString(right)

	// Case 2a: If one was originally string and the other *was* numeric -> ERROR
	// Need to check original numeric status, not just conversion success
	_, leftOriginallyNumeric := ToNumeric(left)
	_, rightOriginallyNumeric := ToNumeric(right)
	if (leftWasString && rightOriginallyNumeric) || (rightWasString && leftOriginallyNumeric) {
		return nil, fmt.Errorf("operator '+' cannot mix numeric and string types (%T + %T): %w", left, right, ErrInvalidOperandType)
	}

	// Case 2b: If at least one was originally string, and the other was NOT numeric -> CONCAT
	if leftWasString || rightWasString {
		return leftStr + rightStr, nil
	}

	// Case 3: Neither converts to numeric, neither was originally string -> ERROR
	// Example: bool + bool, list + map, nil + bool etc.
	return nil, fmt.Errorf("operator '+' cannot operate on types %T and %T: %w", left, right, ErrUnsupportedOperator)
}

// performComparison handles ==, !=, >, <, >=, <=
func performComparison(left, right interface{}, op string) (bool, error) {
	// Handle == and != separately (allows comparing different types via string representation)
	if op == "==" || op == "!=" {
		leftIsNil := left == nil
		rightIsNil := right == nil
		if leftIsNil || rightIsNil {
			isEqual := leftIsNil == rightIsNil
			return (op == "==") == isEqual, nil
		}

		// Try numeric comparison first if both are numeric
		leftF, leftIsNumeric := toFloat64(left)
		rightF, rightIsNumeric := toFloat64(right)
		if leftIsNumeric && rightIsNumeric {
			leftI, leftIsInt := toInt64(left)
			rightI, rightIsInt := toInt64(right)
			if leftIsInt && rightIsInt {
				isEqual := leftI == rightI
				return (op == "==") == isEqual, nil // Use int comparison
			}
			isEqual := leftF == rightF
			return (op == "==") == isEqual, nil // Use float comparison
		}

		// Fallback: Compare string representations using toString helper
		// Consider if deep equality is needed for lists/maps? Current spec uses string compare.
		leftStr, _ := toString(left)
		rightStr, _ := toString(right)
		isEqual := leftStr == rightStr
		return (op == "==") == isEqual, nil
	}

	// Handle >, <, >=, <= (Strictly numeric comparison)
	if left == nil || right == nil {
		return false, fmt.Errorf("comparison operator '%s' requires non-nil operands: %w", op, ErrNilOperand)
	}

	// Check if *originally* numeric
	_, leftOriginallyNumeric := ToNumeric(left)
	_, rightOriginallyNumeric := ToNumeric(right)

	if !leftOriginallyNumeric || !rightOriginallyNumeric {
		return false, fmt.Errorf("comparison operator '%s' requires numeric operands, got %T and %T: %w", op, left, right, ErrInvalidOperandTypeNumeric)
	}

	// Now safe to use converted float values
	leftF, leftOk := toFloat64(left)
	rightF, rightOk := toFloat64(right)
	if !leftOk || !rightOk { // Should not happen if originally numeric, but check anyway
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
		// Should not be reached if parser validates operators
		return false, fmt.Errorf("unknown comparison operator '%s': %w", op, ErrUnsupportedOperator)
	}
}

// performLogical - Likely unused.
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

	// Check *original types* before attempting conversion/operation
	if !isIntegerType(left) || !isIntegerType(right) {
		errMsgFormat := "bitwise operator '%s' requires integer operands"
		if !isIntegerType(left) && !isIntegerType(right) {
			errMsgFormat += ", got %T and %T: %w"
			return 0, fmt.Errorf(errMsgFormat, op, left, right, ErrInvalidOperandTypeInteger)
		} else if !isIntegerType(left) {
			errMsgFormat += ", got %T for left operand: %w"
			return 0, fmt.Errorf(errMsgFormat, op, left, ErrInvalidOperandTypeInteger)
		} else { // !isIntegerType(right)
			errMsgFormat += ", got %T for right operand: %w"
			return 0, fmt.Errorf(errMsgFormat, op, right, ErrInvalidOperandTypeInteger)
		}
	}

	// Now that we know they are originally integer types, attempt conversion to int64
	leftI, leftOk := toInt64(left)
	rightI, rightOk := toInt64(right)

	// This check should ideally not fail if isIntegerType passed, but added for safety.
	if !leftOk || !rightOk {
		return 0, fmt.Errorf("internal error: failed int64 conversion for bitwise operands '%v' (%T) and '%v' (%T) despite passing integer type check", left, left, right, right)
	}

	// Perform the bitwise operation
	switch op {
	case "&":
		return leftI & rightI, nil
	case "|":
		return leftI | rightI, nil
	case "^":
		return leftI ^ rightI, nil
	default:
		// This should ideally be caught by the parser/AST builder
		return 0, fmt.Errorf("unknown bitwise operator '%s': %w", op, ErrUnsupportedOperator)
	}
}

// --- Placeholders for helpers (ensure these exist, e.g., in evaluation_helpers.go) ---
// func toInt64(v interface{}) (int64, bool)
// func toFloat64(v interface{}) (float64, bool)
// func toString(v interface{}) (string, bool) // Returns string representation and bool if original was string
// func isTruthy(value interface{}) bool
// func ToNumeric(v interface{}) (interface{}, bool) // Checks if convertible to int64 or float64
