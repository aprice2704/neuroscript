// NeuroScript Version: 0.8.0
// File version: 6.0.0
// Purpose: Enables all tests for built-in type-checking functions.
// filename: pkg/interpreter/functions_types_test.go
// nlines: 161
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

type dummyTool struct{}

func (d dummyTool) IsTool()              {}
func (d dummyTool) Name() types.FullName { return "dummyTool" }

func TestBuiltinTypeCheckFunctions(t *testing.T) {
	dummyProc := &ast.Procedure{}
	dummyProc.SetName("dummy")

	testCases := []struct {
		funcName string
		argName  string
		argValue lang.Value
		want     bool
	}{
		{"is_string", "v", lang.StringValue{Value: "hello"}, true},
		{"is_number", "v", lang.NumberValue{Value: 123}, true},
		{"is_int", "v", lang.NumberValue{Value: 123.0}, true},
		{"is_int", "v_float", lang.NumberValue{Value: 123.45}, false}, // Test false case
		{"is_float", "v", lang.NumberValue{Value: 123.45}, true},
		{"is_float", "v_int", lang.NumberValue{Value: 123.0}, false}, // Test false case
		{"is_bool", "v", lang.BoolValue{Value: true}, true},
		{"is_list", "v", lang.NewListValue(nil), true},
		{"is_map", "v", lang.NewMapValue(nil), true},
		{"is_error", "v", lang.NewErrorValue("code", "msg", nil), true},
		{"is_function", "v", lang.FunctionValue{Value: dummyProc}, true},
		{"is_tool", "v", lang.ToolValue{Value: dummyTool{}}, true},
		{"is_event", "v", lang.EventValue{}, true},
		{"is_timedate", "v", lang.TimedateValue{Value: time.Now()}, true},
		{"is_fuzzy", "v", lang.NewFuzzyValue(0.5), true},
		{"is_string", "v_num", lang.NumberValue{Value: 123}, false},
		{"is_number", "v_str", lang.StringValue{Value: "123"}, false},
		{"is_bool", "v_str", lang.StringValue{Value: "true"}, false},
		{"is_list", "v_map", lang.NewMapValue(nil), false},
		{"is_map", "v_list", lang.NewListValue(nil), false},
		{"is_error", "v_map", lang.NewMapValue(map[string]lang.Value{"key": lang.StringValue{"val"}}), false},
		{"is_function", "v_tool", lang.ToolValue{Value: dummyTool{}}, false},
		{"is_tool", "v_func", lang.FunctionValue{Value: dummyProc}, false},
		{"is_event", "v_err", lang.NewErrorValue("code", "msg", nil), false},
		{"is_timedate", "v_fuzzy", lang.NewFuzzyValue(0.5), false},
		{"is_fuzzy", "v_time", lang.TimedateValue{Value: time.Now()}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_with_%s", tc.funcName, tc.argName), func(t *testing.T) {
			// A number of these specialized type checkers are not yet implemented.
			// Skip them so the test suite can pass, leaving the tests as a
			// specification for future implementation.
			/*
				unimplemented := map[string]bool{
					"is_int":      true,
					"is_float":    true,
					"is_error":    true,
					"is_function": true,
					"is_tool":     true,
					"is_event":    true,
					"is_timedate": true,
					"is_fuzzy":    true,
				}
				if unimplemented[tc.funcName] {
					t.Skipf("Skipping test for unimplemented built-in function '%s'", tc.funcName)
				}
			*/
			// REMOVED skip logic - all functions are now implemented in pkg/eval/functions.go

			t.Logf("[DEBUG] Turn 1: Starting '%s' test.", tc.funcName)
			h := NewTestHarness(t)
			h.Interpreter.SetVariable(tc.argName, tc.argValue)

			// Updated function syntax to the modern 'needs/returns' format.
			script := fmt.Sprintf("func main(needs %s returns result) means\nreturn %s(%s)\nendfunc", tc.argName, tc.funcName, tc.argName)
			t.Logf("[DEBUG] Turn 2: Executing script:\n%s", script)

			result, err := h.Interpreter.ExecuteScriptString("main", script, nil)
			if err != nil {
				t.Fatalf("ExecuteScriptString failed for %s: %v", tc.funcName, err)
			}
			t.Logf("[DEBUG] Turn 3: Script executed.")

			got, ok := result.(lang.BoolValue)
			if !ok {
				t.Fatalf("Expected BoolValue, got %T", result)
			}
			if got.Value != tc.want {
				t.Errorf("Function '%s' with arg '%v' (%T): got %v, want %v", tc.funcName, tc.argValue, tc.argValue, got.Value, tc.want)
			}
			t.Logf("[DEBUG] Turn 4: Assertion passed.")
		})
	}
}
