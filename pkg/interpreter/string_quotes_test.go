// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// filename: pkg/interpreter/interpreter_string_quotes_test.go
// nlines: 80
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestStringQuoting(t *testing.T) {
	testCases := []struct {
		name               string
		inputStringLiteral string
		expectedOutput     string
	}{
		{
			name:               "Double-quoted string",
			inputStringLiteral: `"This is a standard double-quoted string."`,
			expectedOutput:     `This is a standard double-quoted string.`,
		},
		{
			name:               "Single-quoted string with JSON",
			inputStringLiteral: `'{"key":"value","message":"JSON in single quotes"}'`,
			expectedOutput:     `{"key":"value","message":"JSON in single quotes"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("[DEBUG] Turn 1: Starting string quoting test for '%s'.", tc.name)
			script := fmt.Sprintf(`
func main() means
    emit %s
endfunc
`, tc.inputStringLiteral)

			var capturedEmits []string
			h := NewTestHarness(t)
			h.HostContext.EmitFunc = func(v lang.Value) {
				capturedEmits = append(capturedEmits, v.String())
			}
			t.Logf("[DEBUG] Turn 2: Test harness created and EmitFunc configured.")

			_, execErr := h.Interpreter.ExecuteScriptString("main", script, nil)
			t.Logf("[DEBUG] Turn 3: Script executed.")

			if execErr != nil {
				t.Fatalf("Script execution failed unexpectedly: %v", execErr)
			}

			if len(capturedEmits) != 1 {
				t.Fatalf("Expected 1 emit, but got %d", len(capturedEmits))
			}

			if capturedEmits[0] != tc.expectedOutput {
				t.Errorf("Mismatch on emitted string.\n  Got:      '%s'\n  Expected: '%s'", capturedEmits[0], tc.expectedOutput)
			}
			t.Logf("[DEBUG] Turn 4: Assertion passed.")
		})
	}
}
