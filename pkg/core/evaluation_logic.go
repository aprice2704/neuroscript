// filename: pkg/core/evaluation_logic.go
package core

import (
	"fmt"
	"math" // Keep for evaluateFunctionCall
	"reflect"
)

// --- Evaluation Logic for Operations ---

// NEW Helper: isZeroValue checks if a value is the zero value for its type.
func isZeroValue(val interface{}) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}

// evaluateUnaryOp performs prefix unary operations (not, -, no, some).
// Updated to handle NOT correctly based on truthiness.
func evaluateUnaryOp(op string, operand interface{}) (interface{}, error) {
	switch op {
	// ** Logical NOT **
	// Uses the language's truthiness rules defined in isTruthy (evaluation_helpers.go)
	case "NOT", "not": // Support uppercase and lowercase keywords
		return !isTruthy(operand), nil

	// ** Numeric Negation **
	case "-":
		iVal, isInt := toInt64(operand)
		if isInt {
			return -iVal, nil
		}
		fVal, isFloat := toFloat64(operand)
		if isFloat {
			return -fVal, nil
		}
		// Use specific error type if defined, otherwise wrap standard error
		return nil, fmt.Errorf("%w: unary operator '-' needs number, got %T", ErrInvalidOperandTypeNumeric, operand) // Assuming ErrInvalidOperandTypeNumeric is defined

	// ** Zero Value Checks **
	case "no": // Check if operand IS the zero value for its type
		return isZeroValue(operand), nil
	case "some": // Check if operand is NOT the zero value for its type
		return !isZeroValue(operand), nil

	default:
		// Use specific error type if defined
		return nil, fmt.Errorf("%w: '%s'", ErrUnsupportedOperator, op) // Assuming ErrUnsupportedOperator is defined
	}
}

// evaluateBinaryOp performs infix binary operations (dispatching).
// Updated to handle AND, OR directly with short-circuiting logic.
func evaluateBinaryOp(left, right interface{}, op string) (interface{}, error) {
	switch op {
	// ** Logical AND (Short-circuiting) **
	case "AND", "and": // Support uppercase and lowercase keywords
		leftBool := isTruthy(left)
		if !leftBool {
			return false, nil // Short-circuit: false AND anything is false
		}
		// If left is true, the result depends on the right side
		return isTruthy(right), nil

	// ** Logical OR (Short-circuiting) **
	case "OR", "or": // Support uppercase and lowercase keywords
		leftBool := isTruthy(left)
		if leftBool {
			return true, nil // Short-circuit: true OR anything is true
		}
		// If left is false, the result depends on the right side
		return isTruthy(right), nil

	// Comparison (handle nil checks first)
	case "==", "!=":
		leftIsNil := left == nil
		rightIsNil := right == nil
		if leftIsNil || rightIsNil {
			isEqual := leftIsNil && rightIsNil
			return op == "==" == isEqual, nil // Simpler way to return isEqual for == and !isEqual for !=
		}
		// If neither is nil, use the comparison helper
		result, err := performComparison(left, right, op)
		if err != nil {
			// Wrap the specific error from performComparison
			return nil, fmt.Errorf("compare %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	case "<", ">", "<=", ">=":
		if left == nil || right == nil {
			// Use specific error type if defined
			return nil, fmt.Errorf("%w: operator '%s' needs non-nil operands", ErrNilOperand, op) // Assuming ErrNilOperand is defined
		}
		result, err := performComparison(left, right, op)
		if err != nil {
			// Wrap the specific error from performComparison
			return nil, fmt.Errorf("compare %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	// Arithmetic (includes string concat via '+')
	case "+":
		result, err := performStringConcatOrNumericAdd(left, right)
		if err != nil {
			// Wrap the specific error from helper
			return nil, fmt.Errorf("op %T + %T: %w", left, right, err)
		}
		return result, nil

	case "-", "*", "/", "%", "**":
		if left == nil || right == nil {
			// Use specific error type if defined
			return nil, fmt.Errorf("%w: operator '%s' needs non-nil operands", ErrNilOperand, op)
		}
		result, err := performArithmetic(left, right, op)
		if err != nil {
			// Wrap the specific error from helper
			return nil, fmt.Errorf("arithmetic %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	// Bitwise
	case "&", "|", "^":
		if left == nil || right == nil {
			// Use specific error type if defined
			return nil, fmt.Errorf("%w: operator '%s' needs non-nil operands", ErrNilOperand, op)
		}
		result, err := performBitwise(left, right, op)
		if err != nil {
			// Wrap the specific error from helper
			return nil, fmt.Errorf("bitwise %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	default:
		// Use specific error type if defined
		return nil, fmt.Errorf("%w: '%s'", ErrUnsupportedOperator, op)
	}
}

// evaluateFunctionCall handles built-in function calls.
// (No changes needed here based on the logical operator fixes)
func evaluateFunctionCall(funcName string, args []interface{}) (interface{}, error) {
	// Helper to check arg count and types
	checkArgs := func(expectedCount int, argTypes ...string) error {
		if len(args) != expectedCount {
			// Use specific error type if defined
			return fmt.Errorf("%w: func %s expects %d arg(s), got %d", ErrIncorrectArgCount, funcName, expectedCount, len(args)) // Assuming ErrIncorrectArgCount
		}
		for i, expectedType := range argTypes {
			valid := false
			switch expectedType {
			case "number": // Check for any numeric type (int or float)
				_, isInt := toInt64(args[i])
				_, isFlt := toFloat64(args[i])
				valid = isInt || isFlt
				if !valid {
					return fmt.Errorf("%w: func %s arg %d needs number, got %T", ErrInvalidFunctionArgument, funcName, i+1, args[i])
				}
			case "int":
				_, isInt := toInt64(args[i])
				valid = isInt
				if !valid {
					return fmt.Errorf("%w: func %s arg %d needs integer, got %T", ErrInvalidFunctionArgument, funcName, i+1, args[i])
				}
			case "float": // Specifically check for float or convertible to float
				_, isFlt := toFloat64(args[i])
				valid = isFlt
				if !valid {
					return fmt.Errorf("%w: func %s arg %d needs float, got %T", ErrInvalidFunctionArgument, funcName, i+1, args[i])
				}
			default:
				// This indicates an internal issue with the test setup itself
				return fmt.Errorf("internal error: unknown type check '%s' for func %s", expectedType, funcName)
			}
		}
		return nil
	}

	// Built-in math functions
	switch funcName {
	case "LN":
		if err := checkArgs(1, "number"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0]) // Assume checkArgs ensures conversion is possible
		if fVal <= 0 {
			return nil, fmt.Errorf("%w: LN needs positive arg, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Log(fVal), nil
	case "LOG":
		if err := checkArgs(1, "number"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		if fVal <= 0 {
			return nil, fmt.Errorf("%w: LOG needs positive arg, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Log10(fVal), nil
	case "SIN":
		if err := checkArgs(1, "number"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		return math.Sin(fVal), nil
	case "COS":
		if err := checkArgs(1, "number"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		return math.Cos(fVal), nil
	case "TAN":
		if err := checkArgs(1, "number"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		return math.Tan(fVal), nil
	case "ASIN":
		if err := checkArgs(1, "number"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		if fVal < -1 || fVal > 1 {
			return nil, fmt.Errorf("%w: ASIN needs arg between -1 and 1, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Asin(fVal), nil
	case "ACOS":
		if err := checkArgs(1, "number"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		if fVal < -1 || fVal > 1 {
			return nil, fmt.Errorf("%w: ACOS needs arg between -1 and 1, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Acos(fVal), nil
	case "ATAN":
		if err := checkArgs(1, "number"); err != nil {
			return nil, err
		}
		fVal, _ := toFloat64(args[0])
		return math.Atan(fVal), nil
	// TODO: Add other built-in functions (string, list, etc.) handled here later?
	default:
		// If not a known built-in, it might be a user-defined procedure or tool handled elsewhere.
		// Return a specific error indicating it's not a built-in function handled by this evaluator.
		return nil, fmt.Errorf("%w: '%s'", ErrUnknownFunction, funcName) // Assuming ErrUnknownFunction
	}
}
