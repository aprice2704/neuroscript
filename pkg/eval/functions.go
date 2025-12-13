// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Implements all built-in functions with nil-pointer safety for len().
// filename: pkg/eval/functions.go
// nlines: 230
// risk_rating: HIGH

package eval

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func isBuiltInFunction(name string) bool {
	switch strings.ToLower(name) {
	case "len", "typeof", "is_string", "is_number", "is_bool", "is_list", "is_map", "is_nil",
		"is_int", "is_float", "is_error", "is_function", "is_tool", "is_event", "is_timedate", "is_fuzzy":
		return true
	default:
		return false
	}
}

func (e *evaluation) evaluateBuiltInFunction(funcName string, args []lang.Value, pos *types.Position) (lang.Value, error) {
	checkArgCount := func(expected int) error {
		if len(args) != expected {
			return lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("function '%s' expects %d argument(s), got %d", funcName, expected, len(args)), nil).WithPosition(pos)
		}
		return nil
	}

	switch strings.ToLower(funcName) {
	case "len":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		arg := args[0]
		var length int
		switch v := arg.(type) {
		case lang.StringValue:
			length = utf8.RuneCountInString(v.Value)
		case lang.ListValue:
			length = len(v.Value)
		case *lang.ListValue:
			if v == nil {
				length = 0
			} else {
				length = len(v.Value)
			}
		case lang.MapValue:
			length = len(v.Value)
		case *lang.MapValue:
			if v == nil {
				length = 0
			} else {
				length = len(v.Value)
			}
		case *lang.NilValue:
			length = 0
		default:
			// If it's a generic nil interface, it's 0
			if arg == nil {
				length = 0
			} else {
				length = 1 // All other single values have a length of 1
			}
		}
		return lang.NumberValue{Value: float64(length)}, nil
	case "typeof":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		return lang.StringValue{Value: string(lang.TypeOf(args[0]))}, nil
	case "is_string":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(lang.StringValue)
		return lang.BoolValue{Value: ok}, nil
	case "is_number":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(lang.NumberValue)
		return lang.BoolValue{Value: ok}, nil
	case "is_bool":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(lang.BoolValue)
		return lang.BoolValue{Value: ok}, nil
	case "is_list":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		switch v := args[0].(type) {
		case lang.ListValue:
			return lang.BoolValue{Value: true}, nil
		case *lang.ListValue:
			return lang.BoolValue{Value: v != nil}, nil
		default:
			return lang.BoolValue{Value: false}, nil
		}
	case "is_map":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		switch v := args[0].(type) {
		case lang.MapValue:
			return lang.BoolValue{Value: true}, nil
		case *lang.MapValue:
			return lang.BoolValue{Value: v != nil}, nil
		default:
			return lang.BoolValue{Value: false}, nil
		}
	case "is_nil":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		switch args[0].(type) {
		case *lang.NilValue:
			return lang.BoolValue{Value: true}, nil
		default:
			// Handle raw nil interface
			return lang.BoolValue{Value: args[0] == nil}, nil
		}

	// --- NEWLY IMPLEMENTED FUNCTIONS ---

	case "is_int":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		num, ok := args[0].(lang.NumberValue)
		if !ok {
			return lang.BoolValue{Value: false}, nil
		}
		// Check if the number has no fractional part
		isInt := num.Value == math.Trunc(num.Value)
		return lang.BoolValue{Value: isInt}, nil

	case "is_float":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		num, ok := args[0].(lang.NumberValue)
		if !ok {
			return lang.BoolValue{Value: false}, nil
		}
		// A number is a "float" if it has a fractional part
		isFloat := num.Value != math.Trunc(num.Value)
		return lang.BoolValue{Value: isFloat}, nil

	case "is_error":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		// Case 1: It's an actual ErrorValue type
		if _, ok := args[0].(lang.ErrorValue); ok {
			return lang.BoolValue{Value: true}, nil
		}
		// Case 2: It's a map that *looks* like an error
		var valMap map[string]lang.Value
		if mv, ok := args[0].(lang.MapValue); ok {
			valMap = mv.Value
		} else if mvPtr, ok := args[0].(*lang.MapValue); ok {
			if mvPtr != nil {
				valMap = mvPtr.Value
			}
		}

		if valMap != nil {
			_, hasCode := valMap[lang.ErrorKeyCode]
			_, hasMsg := valMap[lang.ErrorKeyMessage]
			// Per the test, it must have both keys to qualify
			if hasCode && hasMsg {
				return lang.BoolValue{Value: true}, nil
			}
		}
		return lang.BoolValue{Value: false}, nil

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
	}

	return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "unhandled built-in function in switch", nil)
}

func (e *evaluation) evaluateArgs(argNodes []ast.Expression) ([]lang.Value, error) {
	args := make([]lang.Value, len(argNodes))
	for i, argNode := range argNodes {
		val, err := e.Expression(argNode)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}
	return args, nil
}
