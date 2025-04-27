// filename: pkg/core/evaluation_operators.go
package core

import (
	"fmt"
	"math"
	// "reflect" // Not needed
)

// --- Operator Evaluation Logic ---

// performArithmetic handles +, -, *, /, %, **
// UPDATED: '+' case removed, now handled entirely by performStringConcatOrNumericAdd
func performArithmetic(left, right interface{}, op string) (interface{}, error) {
	leftF, leftIsNumeric := toFloat64(left)
	rightF, rightIsNumeric := toFloat64(right)
	leftI, leftIsInt := toInt64(left)
	rightI, rightIsInt := toInt64(right)

	// Check if both operands are numeric for arithmetic operations (excluding '+')
	if !leftIsNumeric || !rightIsNumeric {
		// Use specific sentinel error
		return nil, fmt.Errorf("%w: operator '%s' requires numeric operands, got %T and %T", ErrInvalidOperandTypeNumeric, op, left, right)
	}

	useFloat := !leftIsInt || !rightIsInt // Use float if either operand is float

	switch op {
	// '+' is handled by performStringConcatOrNumericAdd
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
		// Check for division by zero (using float representation)
		if rightF == 0.0 {
			return nil, fmt.Errorf("%w", ErrDivisionByZero) // Use sentinel error
		}
		// Perform float division if either operand is float OR if int division has remainder
		if useFloat || (leftI%rightI != 0) {
			return leftF / rightF, nil
		}
		// Otherwise perform integer division
		return leftI / rightI, nil
	case "%":
		// Modulo requires integers
		if leftIsInt && rightIsInt {
			if rightI == 0 {
				return nil, fmt.Errorf("%w in modulo", ErrDivisionByZero) // Use sentinel error
			}
			return leftI % rightI, nil
		}
		// Use specific sentinel error
		return nil, fmt.Errorf("%w: operator '%%' requires integer operands, got %T and %T", ErrInvalidOperandTypeInteger, left, right)
	case "**":
		// Power always results in float
		return math.Pow(leftF, rightF), nil
	default:
		// Use specific sentinel error
		return nil, fmt.Errorf("%w: unknown arithmetic operator '%s'", ErrUnsupportedOperator, op)
	}
}

// performStringConcatOrNumericAdd decides between string concat and numeric add for '+'
// UPDATED: Returns error for incompatible types.
func performStringConcatOrNumericAdd(left, right interface{}) (interface{}, error) {
	// Attempt conversions first
	leftStr, leftIsString := toString(left) // Helper assumes toString converts nil to ""
	rightStr, rightIsString := toString(right)
	leftF, leftIsNumeric := toFloat64(left)
	rightF, rightIsNumeric := toFloat64(right)
	leftI, leftIsInt := toInt64(left)
	rightI, rightIsInt := toInt64(right)

	// Case 1: Both are numeric -> Perform numeric addition
	if leftIsNumeric && rightIsNumeric {
		if leftIsInt && rightIsInt {
			return leftI + rightI, nil
		}
		return leftF + rightF, nil
	}

	// Case 2: At least one is string AND the *other* is NOT numeric -> String Concatenation
	// (e.g., string + string, string + bool, string + nil, numeric + bool, numeric + nil)
	// Numeric + String is handled below as an error now.
	if (leftIsString && !rightIsNumeric) || (rightIsString && !leftIsNumeric) || (!leftIsNumeric && !rightIsNumeric) {
		// Convert both to string for concatenation (handles nil via toString helper)
		return leftStr + rightStr, nil
	}

	// Case 3: One is string, the other is numeric -> ERROR
	if (leftIsString && rightIsNumeric) || (rightIsString && leftIsNumeric) {
		// Use specific sentinel error
		return nil, fmt.Errorf("%w: operator '+' cannot mix numeric and string types (%T + %T)", ErrInvalidOperandType, left, right) // Assuming ErrInvalidOperandType exists
	}

	// Should not be reached if logic above is correct, but acts as a fallback.
	// Covers cases like bool + bool, list + list etc. which are invalid for '+'
	return nil, fmt.Errorf("%w: operator '+' cannot operate on types %T and %T", ErrUnsupportedOperator, left, right)
}

// performComparison handles ==, !=, >, <, >=, <=
// UPDATED: Uses sentinel errors for numeric comparison failures.
func performComparison(left, right interface{}, op string) (bool, error) {
	// Handle == and != first, as they allow non-numeric types
	if op == "==" || op == "!=" {
		// Attempt numeric comparison if both are numeric
		leftF, leftIsNumeric := toFloat64(left)
		rightF, rightIsNumeric := toFloat64(right)
		if leftIsNumeric && rightIsNumeric {
			// Compare as integers if both are integers
			leftI, leftIsInt := toInt64(left)
			rightI, rightIsInt := toInt64(right)
			if leftIsInt && rightIsInt {
				result := leftI == rightI
				return op == "==" == result, nil // Simpler return logic
			}
			// Otherwise compare as floats
			result := leftF == rightF // Use tolerance? Probably not for equality.
			return op == "==" == result, nil
		}
		// If not both numeric, compare string representations
		result := fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right)
		return op == "==" == result, nil
	}

	// Handle >, <, >=, <= which require numeric operands
	leftF, leftIsNumeric := toFloat64(left)
	rightF, rightIsNumeric := toFloat64(right)

	if !leftIsNumeric || !rightIsNumeric {
		// Use specific sentinel error
		return false, fmt.Errorf("%w: comparison operator '%s' requires numeric operands, got %T and %T", ErrInvalidOperandTypeNumeric, op, left, right)
	}

	// Perform numeric comparison
	switch op {
	case "<":
		return leftF < rightF, nil
	case ">":
		return leftF > rightF, nil
	case "<=":
		return leftF <= rightF, nil
	case ">=":
		return leftF >= rightF, nil
	default: // Should not happen if called from evaluateBinaryOp
		// Use specific sentinel error
		return false, fmt.Errorf("%w: unknown comparison operator '%s'", ErrUnsupportedOperator, op)
	}
}

// performLogical handles AND, OR (non-short-circuiting version)
// UPDATED: Returns error if operands are not convertible to boolean per isTruthy.
// Note: This might be redundant if evaluateExpression handles truthiness.
// Keeping it simple for now, assumes isTruthy handles types.
func performLogical(left, right interface{}, op string) (bool, error) {
	// isTruthy handles type conversion implicitly
	leftBool := isTruthy(left)
	rightBool := isTruthy(right)

	switch op {
	case "and", "AND":
		return leftBool && rightBool, nil
	case "or", "OR":
		return leftBool || rightBool, nil
	default: // Should not happen
		// Use specific sentinel error
		return false, fmt.Errorf("%w: unknown logical operator '%s'", ErrUnsupportedOperator, op)
	}
}

// performBitwise handles &, |, ^
// UPDATED: Uses specific sentinel errors for type issues.
func performBitwise(left, right interface{}, op string) (int64, error) {
	leftI, leftOk := left.(int64)
	rightI, rightOk := right.(int64)

	if !leftOk || !rightOk {
		// Use specific sentinel error
		errMsgFormat := "%w: bitwise operator '%s' requires integer (int64) operands"
		if !leftOk && !rightOk {
			errMsgFormat += ", got %T and %T"
			return 0, fmt.Errorf(errMsgFormat, ErrInvalidOperandTypeInteger, op, left, right)
		} else if !leftOk {
			errMsgFormat += ", got %T for left operand"
			return 0, fmt.Errorf(errMsgFormat, ErrInvalidOperandTypeInteger, op, left)
		} else { // !rightOk
			errMsgFormat += ", got %T for right operand"
			return 0, fmt.Errorf(errMsgFormat, ErrInvalidOperandTypeInteger, op, right)
		}
	}

	// Both are int64, proceed
	switch op {
	case "&":
		return leftI & rightI, nil
	case "|":
		return leftI | rightI, nil
	case "^":
		return leftI ^ rightI, nil
	default: // Should not happen
		// Use specific sentinel error
		return 0, fmt.Errorf("%w: unknown bitwise operator '%s'", ErrUnsupportedOperator, op)
	}
}

// Helper assumed to exist in evaluation_helpers.go
// func toString(v interface{}) (string, bool)
