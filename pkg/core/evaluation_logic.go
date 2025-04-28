// filename: pkg/core/evaluation_logic.go
package core

import (
	"fmt"
	"math"
	"reflect"
	"strings" // Added for function names comparison
	// Assuming errors like ErrNilOperand, ErrInvalidOperandType*, ErrUnsupportedOperator,
	// ErrIncorrectArgCount, ErrInvalidFunctionArgument, ErrUnknownFunction
	// are defined in errors.go
)

// --- Evaluation Logic for Operations ---

// isZeroValue checks if a value is the zero value for its type.
func isZeroValue(val interface{}) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	// Use val.IsValid() check first for safety, especially with interfaces
	if !v.IsValid() {
		return true // Treat invalid reflect.Value as zero/nil equivalent
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.String, reflect.Array: // Added Array
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr, reflect.UnsafePointer: // Added more nillable types
		return v.IsNil()
	// case reflect.Struct: // Structs are zero if all fields are zero - more complex, skip for now
	default:
		// For other types (like struct), assume non-zero unless it's the zero value instance.
		// This check might be sufficient for many cases.
		return v.IsZero()
	}
}

// evaluateUnaryOp performs prefix unary operations (not, -, no, some, ~).
func evaluateUnaryOp(op string, operand interface{}) (interface{}, error) {
	if operand == nil && (op == "-" || op == "~") {
		return nil, fmt.Errorf("%w: unary operator '%s'", ErrNilOperand, op)
	}

	switch strings.ToLower(op) { // Convert op to lower for case-insensitivity
	case "not":
		// isTruthy should be defined in evaluation_helpers.go
		return !isTruthy(operand), nil
	case "-":
		iVal, isInt := toInt64(operand)
		if isInt {
			return -iVal, nil
		}
		fVal, isFloat := toFloat64(operand)
		if isFloat {
			return -fVal, nil
		}
		return nil, fmt.Errorf("%w: unary operator '-' needs number, got %T", ErrInvalidOperandTypeNumeric, operand)
	case "~":
		iVal, isInt := toInt64(operand)
		if isInt {
			return ^iVal, nil
		} // Go uses ^ for bitwise NOT
		return nil, fmt.Errorf("%w: unary operator '~' needs integer, got %T", ErrInvalidOperandTypeInteger, operand)
	case "no":
		return isZeroValue(operand), nil
	case "some":
		return !isZeroValue(operand), nil
	default:
		return nil, fmt.Errorf("%w: '%s'", ErrUnsupportedOperator, op)
	}
}

// evaluateBinaryOp performs infix binary operations (dispatching).
func evaluateBinaryOp(left, right interface{}, op string) (interface{}, error) {
	opLower := strings.ToLower(op) // Convert op to lower once

	switch opLower {
	case "and":
		leftBool := isTruthy(left)
		if !leftBool {
			return false, nil
		} // Short-circuit
		return isTruthy(right), nil
	case "or":
		leftBool := isTruthy(left)
		if leftBool {
			return true, nil
		} // Short-circuit
		return isTruthy(right), nil

	// Comparison (handle nil checks first)
	case "==", "!=":
		// performComparison handles nil checks internally now
		result, err := performComparison(left, right, op) // Pass original op for correct != logic
		if err != nil {
			return nil, fmt.Errorf("compare %T %s %T: %w", left, op, right, err)
		}
		return result, nil
	case "<", ">", "<=", ">=":
		// performComparison handles nil checks internally now
		result, err := performComparison(left, right, op)
		if err != nil {
			return nil, fmt.Errorf("compare %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	// Arithmetic (includes string concat via '+')
	case "+":
		result, err := performStringConcatOrNumericAdd(left, right)
		if err != nil {
			return nil, fmt.Errorf("op %T + %T: %w", left, right, err)
		}
		return result, nil
	case "-", "*", "/", "%", "**":
		result, err := performArithmetic(left, right, op) // Pass original op for "**"
		if err != nil {
			return nil, fmt.Errorf("arithmetic %T %s %T: %w", left, op, right, err)
		}
		return result, nil

	// Bitwise
	case "&", "|", "^":
		result, err := performBitwise(left, right, op)
		if err != nil {
			return nil, fmt.Errorf("bitwise %T %s %T: %w", left, op, right, err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("%w: '%s'", ErrUnsupportedOperator, op)
	}
}

// evaluateFunctionCall handles built-in function calls.
// ADDED: Cases for is_string, is_number, is_int, is_float, is_bool, is_list, is_map, not_empty
func evaluateFunctionCall(funcName string, args []interface{}) (interface{}, error) {
	// Helper to check arg count
	checkArgCount := func(expectedCount int) error {
		if len(args) != expectedCount {
			return fmt.Errorf("%w: func %s expects %d arg(s), got %d", ErrIncorrectArgCount, funcName, expectedCount, len(args))
		}
		return nil
	}
	// Helper to check arg count and type (basic check, specific checks below)
	// Deprecated in favor of specific checks within cases
	_ = func(expectedCount int, argTypes ...string) error { /* ... */ return nil }

	funcLower := strings.ToLower(funcName) // Convert func name for case-insensitive matching

	switch funcLower {
	// --- Type Check Functions (for mustBe) ---
	case "is_string":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(string)
		return ok, nil
	case "is_number": // Checks if it's any numeric Go type (int*, uint*, float*)
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return false, nil
		} // nil is not a number
		k := reflect.ValueOf(args[0]).Kind()
		isNum := (k >= reflect.Int && k <= reflect.Int64) ||
			(k >= reflect.Uint && k <= reflect.Uintptr) || // Include Uintptr? Maybe not. Let's exclude Uintptr.
			(k >= reflect.Uint && k <= reflect.Uint64) ||
			k == reflect.Float32 || k == reflect.Float64
		return isNum, nil
	case "is_int": // Checks specifically for integer types (signed or unsigned)
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		isInt := (k >= reflect.Int && k <= reflect.Int64) ||
			(k >= reflect.Uint && k <= reflect.Uint64) // Exclude Uintptr
		return isInt, nil
	case "is_float": // Checks specifically for float types
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		isFloat := k == reflect.Float32 || k == reflect.Float64
		return isFloat, nil
	case "is_bool":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(bool)
		return ok, nil
	case "is_list": // Checks for slice or array
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return false, nil
		} // nil is not a list
		k := reflect.ValueOf(args[0]).Kind()
		isList := k == reflect.Slice || k == reflect.Array
		return isList, nil
	case "is_map": // Checks for map kind
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if args[0] == nil {
			return false, nil
		} // nil is not a map
		k := reflect.ValueOf(args[0]).Kind()
		isMap := k == reflect.Map
		return isMap, nil
	case "not_empty": // Checks if list, map, or string is not empty (opposite of isZeroValue for these types)
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		// Use !isZeroValue, but only for types where "empty" makes sense (list, map, string)
		if args[0] == nil {
			return false, nil
		} // nil is considered empty
		v := reflect.ValueOf(args[0])
		k := v.Kind()
		if k == reflect.Slice || k == reflect.Map || k == reflect.String || k == reflect.Array {
			return v.Len() > 0, nil
		}
		// For other types, is "not_empty" meaningful? Maybe equivalent to isTruthy?
		// Let's return false for types where length isn't applicable.
		// Or should it error? Let's error for clarity.
		return nil, fmt.Errorf("%w: func %s expects list, map, or string argument, got %T", ErrInvalidFunctionArgument, funcName, args[0])

	// --- Built-in Math Functions ---
	case "ln":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: LN needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		if fVal <= 0 {
			return nil, fmt.Errorf("%w: LN needs positive arg, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Log(fVal), nil
	case "log":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: LOG needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		if fVal <= 0 {
			return nil, fmt.Errorf("%w: LOG needs positive arg, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Log10(fVal), nil
	case "sin":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: SIN needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		return math.Sin(fVal), nil
	case "cos":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: COS needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		return math.Cos(fVal), nil
	case "tan":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: TAN needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		return math.Tan(fVal), nil
	case "asin":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: ASIN needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		if fVal < -1 || fVal > 1 {
			return nil, fmt.Errorf("%w: ASIN needs arg between -1 and 1, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Asin(fVal), nil
	case "acos":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: ACOS needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		if fVal < -1 || fVal > 1 {
			return nil, fmt.Errorf("%w: ACOS needs arg between -1 and 1, got %v", ErrInvalidFunctionArgument, fVal)
		}
		return math.Acos(fVal), nil
	case "atan":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: ATAN needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		return math.Atan(fVal), nil

	default:
		// Return specific error for unknown functions
		return nil, fmt.Errorf("%w: '%s'", ErrUnknownFunction, funcName)
	}
}

// --- Placeholders for other evaluation logic functions ---
// func performComparison(left, right interface{}, op string) (bool, error) { ... } // Assumed defined in evaluation_comparison.go
// func performStringConcatOrNumericAdd(left, right interface{}) (interface{}, error) { ... } // Assumed defined in evaluation_operators.go
// func performArithmetic(left, right interface{}, op string) (interface{}, error) { ... } // Assumed defined in evaluation_operators.go
// func performBitwise(left, right interface{}, op string) (interface{}, error) { ... } // Assumed defined in evaluation_operators.go
// func toInt64(v interface{}) (int64, bool) { ... } // Assumed defined in evaluation_helpers.go
// func toFloat64(v interface{}) (float64, bool) { ... } // Assumed defined in evaluation_helpers.go
// func isTruthy(value interface{}) bool { ... } // Assumed defined in evaluation_helpers.go
