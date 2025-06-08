// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Defines logic for all built-in NeuroScript functions (e.g., is_error, sin).
// filename: core/evaluation_functions.go
// nlines: 100
// risk_rating: LOW

package core

import (
	"fmt"
	"math"
	"reflect"
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
func evaluateBuiltInFunction(funcName string, args []interface{}) (interface{}, error) {
	checkArgCount := func(expectedCount int) error {
		if len(args) != expectedCount {
			return fmt.Errorf("%w: func %s expects %d arg(s), got %d", ErrIncorrectArgCount, funcName, expectedCount, len(args))
		}
		return nil
	}
	funcLower := strings.ToLower(funcName)
	switch funcLower {
	case "is_error":
		_ = checkArgCount(1)
		_, ok := args[0].(ErrorValue)
		return ok, nil
	case "is_string":
		_ = checkArgCount(1)
		_, ok := args[0].(string)
		return ok, nil
	case "is_number":
		_ = checkArgCount(1)
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		return (k >= reflect.Int && k <= reflect.Float64), nil
	case "is_int":
		_ = checkArgCount(1)
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		return (k >= reflect.Int && k <= reflect.Uint64), nil
	case "is_float":
		_ = checkArgCount(1)
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		return (k == reflect.Float32 || k == reflect.Float64), nil
	case "is_bool":
		_ = checkArgCount(1)
		_, ok := args[0].(bool)
		return ok, nil
	case "is_list":
		_ = checkArgCount(1)
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		return (k == reflect.Slice || k == reflect.Array), nil
	case "is_map":
		_ = checkArgCount(1)
		if args[0] == nil {
			return false, nil
		}
		k := reflect.ValueOf(args[0]).Kind()
		return k == reflect.Map, nil
	case "not_empty":
		_ = checkArgCount(1)
		if args[0] == nil {
			return false, nil
		}
		v := reflect.ValueOf(args[0])
		k := v.Kind()
		if k == reflect.Slice || k == reflect.Map || k == reflect.String || k == reflect.Array {
			return v.Len() > 0, nil
		}
		return nil, fmt.Errorf("%w: func %s expects list, map, or string, got %T", ErrInvalidFunctionArgument, funcName, args[0])
	case "ln", "log", "sin", "cos", "tan", "asin", "acos", "atan":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := toFloat64(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: math func needs number, got %T", ErrInvalidFunctionArgument, args[0])
		}
		switch funcLower {
		case "ln":
			if fVal <= 0 {
				return nil, fmt.Errorf("%w: LN needs positive arg", ErrInvalidFunctionArgument)
			}
			return math.Log(fVal), nil
		case "log":
			if fVal <= 0 {
				return nil, fmt.Errorf("%w: LOG needs positive arg", ErrInvalidFunctionArgument)
			}
			return math.Log10(fVal), nil
		case "sin":
			return math.Sin(fVal), nil
		case "cos":
			return math.Cos(fVal), nil
		case "tan":
			return math.Tan(fVal), nil
		case "asin":
			if fVal < -1 || fVal > 1 {
				return nil, fmt.Errorf("%w: ASIN arg must be -1 to 1", ErrInvalidFunctionArgument)
			}
			return math.Asin(fVal), nil
		case "acos":
			if fVal < -1 || fVal > 1 {
				return nil, fmt.Errorf("%w: ACOS arg must be -1 to 1", ErrInvalidFunctionArgument)
			}
			return math.Acos(fVal), nil
		case "atan":
			return math.Atan(fVal), nil
		}
	}
	return nil, fmt.Errorf("%w: '%s'", ErrUnknownFunction, funcName)
}
