// NeuroScript Version: 0.5.2
// File version: 8.0.0
// Purpose: Refactored to use the centralized TestHarness and a direct execution method, removing dependency on obsolete suite helpers.
// filename: pkg/interpreter/interpreter_string_escaping_test.go
// nlines: 90
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestInterpretStringEscaping verifies that the interpreter correctly
// handles strings that have been unescaped by the AST builder.
func TestInterpretStringEscaping(t *testing.T) {
	testCases := []struct {
		name         string
		literalValue string
		expected     string
	}{
		{"Interpret Backspace", "text\bback", "text\bback"},
		{"Interpret Tab", "col1\tcol2", "col1\tcol2"},
		{"Interpret Newline", "first\nsecond", "first\nsecond"},
		{"Interpret Double Quote", `a "quoted" string`, `a "quoted" string`},
		{"Interpret Backslash", `a path C:\folder`, `a path C:\folder`},
		{"Interpret Unicode BMP", "currency: €", "currency: €"},
		{"Interpret Unicode Surrogate Pair", "face: 😀", "face: 😀"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[DEBUG] Turn 1: Starting string escaping test for '%s'.", tc.name)
			h := NewTestHarness(t)
			// We need to quote the literal value for it to be a valid string in the script.
			script := fmt.Sprintf(`func main() returns result means set result = "%s" return result endfunc`, tc.literalValue)

			t.Logf("[DEBUG] Turn 2: Executing script:\n%s", script)
			result, err := h.Interpreter.ExecuteScriptString("main", script, nil)
			if err != nil {
				t.Fatalf("ExecuteScriptString failed: %v", err)
			}
			t.Logf("[DEBUG] Turn 3: Script executed.")

			got, ok := result.(lang.StringValue)
			if !ok {
				t.Fatalf("Expected a StringValue result, but got %T", result)
			}

			if got.Value != tc.expected {
				t.Errorf("Result mismatch:\n  Expected: %q\n       Got: %q", tc.expected, got.Value)
			}
			t.Logf("[DEBUG] Turn 4: Assertion passed.")
		})
	}
}
