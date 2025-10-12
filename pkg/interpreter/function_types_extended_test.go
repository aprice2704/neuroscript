// NeuroScript Version: 0.5.2
// File version: 2.0.0
// Purpose: Refactored to use the centralized TestHarness and script execution for a more robust and realistic test.
// filename: pkg/interpreter/interpreter_functions_types_extended_test.go
// nlines: 70
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestIsErrorFunction_Extended(t *testing.T) {
	testCases := []struct {
		name     string
		argName  string
		argValue lang.Value
		want     bool
	}{
		{"is_error with a complete error map", "v", lang.NewMapValue(map[string]lang.Value{lang.ErrorKeyCode: lang.StringValue{Value: "E_CODE"}, lang.ErrorKeyMessage: lang.StringValue{Value: "An error occurred"}}), true},
		{"is_error with incomplete error map", "v", lang.NewMapValue(map[string]lang.Value{lang.ErrorKeyCode: lang.StringValue{Value: "E_CODE"}}), false},
		{"is_error with a non-error map", "v", lang.NewMapValue(map[string]lang.Value{"key": lang.StringValue{Value: "value"}}), false},
		{"is_error with a timedate value", "v", lang.TimedateValue{Value: time.Now()}, false},
		{"is_error with a fuzzy value", "v", lang.NewFuzzyValue(0.5), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[DEBUG] Turn 1: Starting '%s' test.", tc.name)
			h := NewTestHarness(t)
			h.Interpreter.SetVariable(tc.argName, tc.argValue)
			script := fmt.Sprintf(`func main() returns result means return is_error(%s) endfunc`, tc.argName)
			t.Logf("[DEBUG] Turn 2: Executing script:\n%s", script)

			result, err := h.Interpreter.ExecuteScriptString("main", script, nil)
			if err != nil {
				t.Fatalf("ExecuteScriptString failed: %v", err)
			}
			t.Logf("[DEBUG] Turn 3: Script executed.")

			got, ok := result.(lang.BoolValue)
			if !ok {
				t.Fatalf("Expected BoolValue, got %T", result)
			}
			if got.Value != tc.want {
				t.Errorf("Function 'is_error' with arg '%v': got %v, want %v", tc.argValue, got.Value, tc.want)
			}
			t.Logf("[DEBUG] Turn 4: Assertion passed.")
		})
	}
}
