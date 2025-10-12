// NeuroScript Version: 0.5.2
// File version: 3.0.0
// Purpose: Refactored to use the centralized TestHarness and script execution for a more robust and realistic test.
// filename: pkg/interpreter/interpreter_functions_types_test.go
// nlines: 160
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
		{"is_float", "v", lang.NumberValue{Value: 123.45}, true},
		{"is_bool", "v", lang.BoolValue{Value: true}, true},
		{"is_list", "v", lang.NewListValue(nil), true},
		{"is_map", "v", lang.NewMapValue(nil), true},
		{"is_error", "v", lang.NewErrorValue("code", "msg", nil), true},
		{"is_function", "v", lang.FunctionValue{Value: dummyProc}, true},
		{"is_tool", "v", lang.ToolValue{Value: dummyTool{}}, true},
		{"is_event", "v", lang.EventValue{}, true},
		{"is_timedate", "v", lang.TimedateValue{Value: time.Now()}, true},
		{"is_fuzzy", "v", lang.NewFuzzyValue(0.5), true},
		{"is_string", "v", lang.NumberValue{Value: 123}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_with_%T", tc.funcName, tc.argValue), func(t *testing.T) {
			t.Logf("[DEBUG] Turn 1: Starting '%s' test.", tc.funcName)
			h := NewTestHarness(t)
			h.Interpreter.SetVariable(tc.argName, tc.argValue)
			script := fmt.Sprintf(`func main() returns result means return %s(%s) endfunc`, tc.funcName, tc.argName)
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
