// NeuroScript Version: 0.5.2
// File version: 12
// Purpose: Corrected all type-checking functions (is_string, etc.) to recognize both primitive types and their lang.Value wrappers.
// filename: pkg/interpreter/evaluation_functions.go
// nlines: 160
// risk_rating: HIGH

package interpreter

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// isBuiltInFunction checks if a name corresponds to a known built-in function.
func isBuiltInFunction(name string) bool {
	switch strings.ToLower(name) {
	case "len", "ln", "log", "sin", "cos", "tan", "asin", "acos", "atan",
		"is_string", "is_number", "is_int", "is_float", "is_bool", "is_list", "is_map", "not_empty",
		"is_error", "is_function", "is_tool", "is_event", "is_timedate", "is_fuzzy": // Added new functions
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
	case lang.NumberValue:
		return v.Value, true
	default:
		return 0, false
	}
}

// evaluateBuiltInFunction handles built-in function calls.
func evaluateBuiltInFunction(funcName string, args []interface{}) (lang.Value, error) {
	checkArgCount := func(expectedCount int) error {
		if len(args) != expectedCount {
			return fmt.Errorf("%w: func %s expects %d arg(s), got %d", lang.ErrIncorrectArgCount, funcName, expectedCount, len(args))
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
			length = 1
		}
		return lang.NumberValue{Value: float64(length)}, nil

	// --- Type Checking Functions ---
	case "is_string":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, isPrim := args[0].(string)
		_, isWrap := args[0].(lang.StringValue)
		return lang.BoolValue{Value: isPrim || isWrap}, nil
	case "is_number":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := getNumericArg(args[0])
		return lang.BoolValue{Value: ok}, nil
	case "is_int":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		f, ok := getNumericArg(args[0])
		return lang.BoolValue{Value: ok && f == math.Trunc(f)}, nil
	case "is_float":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		f, ok := getNumericArg(args[0])
		return lang.BoolValue{Value: ok && f != math.Trunc(f)}, nil
	case "is_bool":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, isPrim := args[0].(bool)
		_, isWrap := args[0].(lang.BoolValue)
		return lang.BoolValue{Value: isPrim || isWrap}, nil
	case "is_list":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, isPrim := args[0].([]interface{})
		_, isWrap := args[0].(lang.ListValue)
		return lang.BoolValue{Value: isPrim || isWrap}, nil
	case "is_map":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, isPrim := args[0].(map[string]interface{})
		_, isWrap := args[0].(lang.MapValue)
		return lang.BoolValue{Value: isPrim || isWrap}, nil
	case "is_error":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, isPrim := args[0].(error)
		_, isWrap := args[0].(lang.ErrorValue)
		// Also check for the map structure
		isMapStruct := false
		if m, ok := args[0].(map[string]interface{}); ok {
			_, hasCode := m[lang.ErrorKeyCode]
			_, hasMsg := m[lang.ErrorKeyMessage]
			isMapStruct = hasCode && hasMsg
		}
		if m, ok := args[0].(lang.MapValue); ok {
			_, hasCode := m.Value[lang.ErrorKeyCode]
			_, hasMsg := m.Value[lang.ErrorKeyMessage]
			isMapStruct = hasCode && hasMsg
		}
		return lang.BoolValue{Value: isPrim || isWrap || isMapStruct}, nil
	case "is_function":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(lang.FunctionValue)
		return lang.BoolValue{Value: ok}, nil
	case "is_tool":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(lang.ToolValue)
		return lang.BoolValue{Value: ok}, nil
	case "is_event":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(lang.EventValue)
		return lang.BoolValue{Value: ok}, nil
	case "is_timedate":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(lang.TimedateValue)
		return lang.BoolValue{Value: ok}, nil
	case "is_fuzzy":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(lang.FuzzyValue)
		return lang.BoolValue{Value: ok}, nil
	case "not_empty":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		wrappedArg, err := lang.Wrap(args[0])
		if err != nil {
			return nil, fmt.Errorf("internal error in 'not_empty': could not re-wrap argument: %w", err)
		}
		return lang.BoolValue{Value: wrappedArg.IsTruthy()}, nil

	// Math functions
	case "ln", "log", "sin", "cos", "tan", "asin", "acos", "atan":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		fVal, ok := getNumericArg(args[0])
		if !ok {
			return nil, fmt.Errorf("%w: math func needs number, got %T", lang.ErrInvalidFunctionArgument, args[0])
		}

		switch funcLower {
		case "ln":
			if fVal <= 0 {
				return nil, fmt.Errorf("%w: LN needs positive arg", lang.ErrInvalidFunctionArgument)
			}
			return lang.NumberValue{Value: math.Log(fVal)}, nil
		case "log":
			if fVal <= 0 {
				return nil, fmt.Errorf("%w: LOG needs positive arg", lang.ErrInvalidFunctionArgument)
			}
			return lang.NumberValue{Value: math.Log10(fVal)}, nil
		case "sin":
			return lang.NumberValue{Value: math.Sin(fVal)}, nil
		case "cos":
			return lang.NumberValue{Value: math.Cos(fVal)}, nil
		case "tan":
			return lang.NumberValue{Value: math.Tan(fVal)}, nil
		case "asin":
			if fVal < -1 || fVal > 1 {
				return nil, fmt.Errorf("%w: ASIN arg must be -1 to 1", lang.ErrInvalidFunctionArgument)
			}
			return lang.NumberValue{Value: math.Asin(fVal)}, nil
		case "acos":
			if fVal < -1 || fVal > 1 {
				return nil, fmt.Errorf("%w: ACOS arg must be -1 to 1", lang.ErrInvalidFunctionArgument)
			}
			return lang.NumberValue{Value: math.Acos(fVal)}, nil
		case "atan":
			return lang.NumberValue{Value: math.Atan(fVal)}, nil
		}
	}
	return nil, fmt.Errorf("%w: '%s'", lang.ErrUnknownFunction, funcName)
}
