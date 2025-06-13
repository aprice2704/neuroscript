// NeuroScript Version: 0.4.1
// File version: 3
// Purpose: Corrected built-in function logic for type assertions and return values.
// filename: pkg/core/evaluation_functions.go
// nlines: 105
// risk_rating: MEDIUM

package core

import (
	"fmt"
	"math"
	"strings"
)

// isBuiltInFunction checks if a name corresponds to a known built-in function.
func isBuiltInFunction(name string) bool {
	switch strings.ToLower(name) {
	case "ln", "log", "sin", "cos", "tan", "asin", "acos", "atan",
		"is_string", "is_number", "is_int", "is_float", "is_bool", "is_list", "is_map", "not_empty",
		"is_error":
		return true
	default:
		return false
	}
}

// evaluateBuiltInFunction handles built-in function calls.
func evaluateBuiltInFunction(funcName string, args []interface{}) (Value, error) {
	checkArgCount := func(expectedCount int) error {
		if len(args) != expectedCount {
			// FIX: Return only a single error value.
			return fmt.Errorf("%w: func %s expects %d arg(s), got %d", ErrIncorrectArgCount, funcName, expectedCount, len(args))
		}
		return nil
	}

	// FIX: Declare 'arg' as a Value interface, initialized to NilValue.
	var arg Value = NilValue{}
	if len(args) > 0 {
		var ok bool
		// FIX: Correctly assign the interface value from the assertion.
		arg, ok = args[0].(Value)
		if !ok {
			if args[0] == nil {
				arg = NilValue{}
			} else {
				return nil, fmt.Errorf("internal error: built-in function %s received non-Value type argument: %T", funcName, args[0])
			}
		}
	}

	funcLower := strings.ToLower(funcName)
	switch funcLower {
	case "is_error":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := arg.(ErrorValue)
		return BoolValue{Value: ok}, nil
	case "is_string":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := arg.(StringValue)
		return BoolValue{Value: ok}, nil
	case "is_number":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := arg.(NumberValue)
		return BoolValue{Value: ok}, nil
	case "is_int":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		num, ok := arg.(NumberValue)
		isInt := ok && num.Value == math.Trunc(num.Value)
		return BoolValue{Value: isInt}, nil
	case "is_float":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		num, ok := arg.(NumberValue)
		isFloat := ok && num.Value != math.Trunc(num.Value)
		return BoolValue{Value: isFloat}, nil
	case "is_bool":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := arg.(BoolValue)
		return BoolValue{Value: ok}, nil
	case "is_list":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := arg.(ListValue)
		return BoolValue{Value: ok}, nil
	case "is_map":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := arg.(MapValue)
		return BoolValue{Value: ok}, nil
	case "not_empty":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		return BoolValue{Value: arg.IsTruthy()}, nil
	case "ln", "log", "sin", "cos", "tan", "asin", "acos", "atan":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		num, ok := ToNumeric(arg)
		if !ok {
			return nil, fmt.Errorf("%w: math func needs number, got %s", ErrInvalidFunctionArgument, TypeOf(arg))
		}
		fVal := num.Value
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
