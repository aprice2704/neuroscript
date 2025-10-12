// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Implements built-in functions for the evaluator.
// filename: pkg/eval/functions.go
// nlines: 150
// risk_rating: HIGH

package eval

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func isBuiltInFunction(name string) bool {
	switch strings.ToLower(name) {
	case "len", "typeof", "is_string", "is_number", "is_bool", "is_list", "is_map", "is_nil":
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
			length = len(v.Value)
		case lang.MapValue:
			length = len(v.Value)
		case *lang.MapValue:
			length = len(v.Value)
		case *lang.NilValue:
			length = 0
		default:
			length = 1 // All other single values have a length of 1
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
		switch args[0].(type) {
		case lang.ListValue, *lang.ListValue:
			return lang.BoolValue{Value: true}, nil
		default:
			return lang.BoolValue{Value: false}, nil
		}
	case "is_map":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		switch args[0].(type) {
		case lang.MapValue, *lang.MapValue:
			return lang.BoolValue{Value: true}, nil
		default:
			return lang.BoolValue{Value: false}, nil
		}
	case "is_nil":
		if err := checkArgCount(1); err != nil {
			return nil, err
		}
		_, ok := args[0].(*lang.NilValue)
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
