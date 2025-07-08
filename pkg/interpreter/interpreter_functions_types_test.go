// NeuroScript Version: 0.5.2
// File version: 2.1.0
// Purpose: Corrected is_tool test by creating a dummy struct that properly implements the lang.Tool interface.
// filename: pkg/interpreter/interpreter_functions_types_test.go
// nlines: 155
// risk_rating: LOW

package interpreter

import (
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// dummyTool is a local struct just for testing to satisfy the lang.Tool interface.
type dummyTool struct{}

func (d dummyTool) IsTool()              {}
func (d dummyTool) Name() types.FullName { return "dummyTool" }

// Test for is_string, is_number, is_int, is_float, is_bool, is_list, is_map, is_error, etc.
func TestBuiltinTypeCheckFunctions(t *testing.T) {
	dummyProc := &ast.Procedure{}
	dummyProc.SetName("dummy")

	testCases := []struct {
		funcName string
		arg      interface{}
		want     bool
	}{
		// is_string
		{"is_string", "hello", true},
		{"is_string", lang.StringValue{Value: "a"}, true},
		{"is_string", 123, false},

		// is_number
		{"is_number", 123, true},
		{"is_number", lang.NumberValue{Value: 5}, true},
		{"is_number", "123", false},

		// is_int
		{"is_int", 123, true},
		{"is_int", lang.NumberValue{Value: 123.0}, true},
		{"is_int", 123.45, false},

		// is_float
		{"is_float", 123.45, true},
		{"is_float", 123, false},
		{"is_float", lang.NumberValue{Value: 123.0}, false},

		// is_bool
		{"is_bool", true, true},
		{"is_bool", lang.BoolValue{Value: false}, true},
		{"is_bool", "true", false},

		// is_list
		{"is_list", []interface{}{1}, true},
		{"is_list", lang.NewListValue(nil), true},
		{"is_list", "[]", false},

		// is_map
		{"is_map", map[string]interface{}{"a": 1}, true},
		{"is_map", lang.NewMapValue(nil), true},
		{"is_map", "{}", false},

		// is_error
		{"is_error", lang.NewErrorValue("code", "msg", nil), true},
		{"is_error", lang.NewMapValue(map[string]lang.Value{lang.ErrorKeyCode: lang.StringValue{Value: "c"}, lang.ErrorKeyMessage: lang.StringValue{Value: "m"}}), true},
		{"is_error", lang.StringValue{Value: "error"}, false},

		// is_function
		{"is_function", lang.FunctionValue{Value: dummyProc}, true},
		{"is_function", "my_func", false},

		// is_tool
		// FIX: Use a struct that actually implements the lang.Tool interface.
		{"is_tool", lang.ToolValue{Value: dummyTool{}}, true},
		{"is_tool", "my_tool", false},

		// is_event
		{"is_event", lang.EventValue{}, true},
		{"is_event", lang.NewMapValue(nil), false},

		// is_timedate
		{"is_timedate", lang.TimedateValue{Value: time.Now()}, true},
		{"is_timedate", time.Now().String(), false},

		// is_fuzzy
		{"is_fuzzy", lang.NewFuzzyValue(0.5), true},
		{"is_fuzzy", 0.5, false},
	}

	for _, tc := range testCases {
		t.Run(tc.funcName+"_with_"+t.Name(), func(t *testing.T) {
			result, err := evaluateBuiltInFunction(tc.funcName, []interface{}{tc.arg})
			if err != nil {
				t.Fatalf("evaluateBuiltInFunction failed for %s: %v", tc.funcName, err)
			}
			got, ok := result.(lang.BoolValue)
			if !ok {
				t.Fatalf("Expected BoolValue, got %T", result)
			}
			if got.Value != tc.want {
				t.Errorf("Function '%s' with arg '%v' (%T): got %v, want %v", tc.funcName, tc.arg, tc.arg, got.Value, tc.want)
			}
		})
	}
}

// Test for not_empty
func TestNotEmptyFunction(t *testing.T) {
	testCases := []struct {
		name string
		arg  lang.Value
		want bool
	}{
		{"truthy string", lang.StringValue{Value: "hello"}, true},
		{"falsey string", lang.StringValue{Value: ""}, false},
		{"truthy number", lang.NumberValue{Value: 1}, true},
		{"falsey number", lang.NumberValue{Value: 0}, false},
		{"truthy bool", lang.BoolValue{Value: true}, true},
		{"falsey bool", lang.BoolValue{Value: false}, false},
		{"truthy list", lang.NewListValue([]lang.Value{lang.NumberValue{Value: 1}}), true},
		{"falsey list", lang.NewListValue(nil), false},
		{"truthy map", lang.NewMapValue(map[string]lang.Value{"a": lang.NumberValue{Value: 1}}), true},
		{"falsey map", lang.NewMapValue(nil), false},
		{"nil value", &lang.NilValue{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := evaluateBuiltInFunction("not_empty", []interface{}{tc.arg})
			if err != nil {
				t.Fatalf("evaluateBuiltInFunction failed for not_empty: %v", err)
			}
			got, ok := result.(lang.BoolValue)
			if !ok {
				t.Fatalf("Expected BoolValue, got %T", result)
			}
			if got.Value != tc.want {
				t.Errorf("Function 'not_empty' with arg '%v': got %v, want %v", tc.arg, got.Value, tc.want)
			}
		})
	}
}
