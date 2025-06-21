// NeuroScript Version: 0.4.1
// File version: 9
// Purpose: Added len() built-in function.
// filename: pkg/core/evaluation_functions.go
// nlines: 154
// risk_rating: HIGH

package core

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

// isBuiltInFunction checks if a name corresponds to a known built-in function.
func isBuiltInFunction(name string) bool {
	switch strings.ToLower(name) {
	case "len", "ln", "log", "sin", "cos", "tan", "asin", "acos", "atan",
		"is_string", "is_number", "is_int", "is_float", "is_bool", "is_list", "is_map", "not_empty",
		"is_error":
		return true
	default:
		return false
	}
}

// getNumericArg extracts a float64 from an interface{}, handling various numeric types.
func getNumericArg(arg interface{}) (float64, bool) {
	switch v := arg.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case NumberValue:
		return v.Value, true
	default:
		return 0, false
	}
}

// evaluateBuiltInFunction handles built-in function calls.
// It adheres to the contract by operating on primitive Go types.
func evaluateBuiltInFunction(funcName string, args []interface{}) (Value, error) {
	checkArgCount := func(expectedCount int) error {
		if len(args) != expectedCount {
			return fmt.Errorf("%w: func %s expects %d arg(s), got %d", ErrIncorrectArgCount, funcName, expectedCount, len(args))
		}
		return nil
	}

	funcLower := strings.ToLower(funcName)
	switch funcLower {
	case "len":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		arg := args[0]
		var length int
		switch v := arg.(type) {
		case nil:
			length = 0
		case string:
			length = utf8.RuneCountInString(v)
		case []interface{}:
			length = len(v)
		case map[string]interface{}:
			length = len(v)
		default:
			// For any other type (Number, Bool, Error, etc.), the length is 1
			// as it represents a single, indivisible value.
			length = 1
		}
		return NumberValue{Value: float64(length)}, nil
		// Type checking functions
	case "is_error":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		arg := args[0]
		// FIX: The test `evaluation_new_types_test.go` confirms that an unwrapped tool
		// error becomes a `map[string]interface{}`. We check for that case directly.
		// We also check for the native `error` type for robustness.
		if _, ok := arg.(error); ok {
			return BoolValue{Value: true}, nil
		}
		if _, ok := arg.(map[string]interface{}); ok {
			return BoolValue{Value: true}, nil
		}
		return BoolValue{Value: false}, nil
	case "is_string":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(string)
		return BoolValue{Value: ok}, nil
	case "is_number":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := getNumericArg(args[0])
		return BoolValue{Value: ok}, nil
	case "is_int":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		f, ok := getNumericArg(args[0])
		return BoolValue{Value: ok && f == math.Trunc(f)}, nil
	case "is_float":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		f, ok := getNumericArg(args[0])
		return BoolValue{Value: ok && f != math.Trunc(f)}, nil
	case "is_bool":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(bool)
		return BoolValue{Value: ok}, nil
	case "is_list":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].([]interface{})
		return BoolValue{Value: ok}, nil
	case "is_map":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(map[string]interface{})
		return BoolValue{Value: ok}, nil
	case "not_empty":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		wrappedArg, err := Wrap(args[0])
		if err != nil {
			return nil, fmt.Errorf("internal error in 'not_empty': could not re-wrap argument: %w", err)
		}
		return BoolValue{Value: wrappedArg.IsTruthy()}, nil

	// Math functions
	case "ln", "log", "sin", "cos", "tan", "asin", "acos", "atan":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := getNumericArg(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: math func needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}

		switch funcLower {
		case "ln":
			if fVal <= 0 {
				return nil, fmt.Errorf("%w: LN needs positive arg", ErrInvalidFunctionArgument)
			}
			return NumberValue{Value: math.Log(fVal)}, nil
		case "log":
			if fVal <= 0 {
				return nil, fmt.Errorf("%w: LOG needs positive arg", ErrInvalidFunctionArgument)
			}
			return NumberValue{Value: math.Log10(fVal)}, nil
		case "sin":
			return NumberValue{Value: math.Sin(fVal)}, nil
		case "cos":
			return NumberValue{Value: math.Cos(fVal)}, nil
		case "tan":
			return NumberValue{Value: math.Tan(fVal)}, nil
		case "asin":
			if fVal < -1 || fVal > 1 {
				return nil, fmt.Errorf("%w: ASIN arg must be -1 to 1", ErrInvalidFunctionArgument)
			}
			return NumberValue{Value: math.Asin(fVal)}, nil
		case "acos":
			if fVal < -1 || fVal > 1 {
				return nil, fmt.Errorf("%w: ACOS arg must be -1 to 1", ErrInvalidFunctionArgument)
			}
			return NumberValue{Value: math.Acos(fVal)}, nil
		case "atan":
			return NumberValue{Value: math.Atan(fVal)}, nil
		}
	}
	return nil, fmt.Errorf("%w: '%s'", ErrUnknownFunction, funcName)
}
