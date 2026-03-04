// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 5
// :: description: Implements built-in functions including char(n) and ord(s) for Unicode handling.
// :: latestChange: Added char(n) and ord(s) built-in functions.
// :: filename: pkg/eval/functions.go
// :: serialization: go

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
	case "len", "typeof", "char", "ord", "is_string", "is_number", "is_bool", "is_list", "is_map", "is_nil",
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
			if arg == nil {
				length = 0
			} else {
				length = 1
			}
		}
		return lang.NumberValue{Value: float64(length)}, nil

	case "char":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		num, ok := lang.ToFloat64(args[0])
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "char() expects a numeric codepoint", lang.ErrArgumentMismatch).WithPosition(pos)
		}
		return lang.StringValue{Value: string(rune(int(num)))}, nil

	case "ord":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		s, ok := lang.ToString(args[0])
		if !ok || utf8.RuneCountInString(s) == 0 {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, "ord() expects a non-empty string", lang.ErrArgumentMismatch).WithPosition(pos)
		}
		r, _ := utf8.DecodeRuneInString(s)
		return lang.NumberValue{Value: float64(r)}, nil

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
			return lang.BoolValue{Value: args[0] == nil}, nil
		}

	case "is_int":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		num, ok := args[0].(lang.NumberValue)
		if !ok {
			return lang.BoolValue{Value: false}, nil
		}
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
		isFloat := num.Value != math.Trunc(num.Value)
		return lang.BoolValue{Value: isFloat}, nil

	case "is_error":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		if _, ok := args[0].(lang.ErrorValue); ok {
			return lang.BoolValue{Value: true}, nil
		}
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
