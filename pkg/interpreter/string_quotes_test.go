// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Refactored the string quoting tests into a single, table-driven test that asserts both single and double quotes should succeed, as per the language design. This test will currently fail for single quotes, highlighting the bug in the AST builder.
// filename: pkg/interpreter/string_quotes_test.go
// nlines: 75
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
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
			script := fmt.Sprintf(`
func main() means
    emit %s
endfunc
`, tc.inputStringLiteral)

			var capturedEmits []string
			interp, err := interpreter.NewTestInterpreter(t, nil, nil, false)
			if err != nil {
				t.Fatalf("Failed to create test interpreter: %v", err)
			}

			interp.SetEmitFunc(func(v lang.Value) {
				capturedEmits = append(capturedEmits, v.String())
			})

			_, execErr := interp.ExecuteScriptString("main", script, nil)

			// This is the key change: we now expect NO error for either case.
			if execErr != nil {
				t.Fatalf("Script execution failed unexpectedly: %v", execErr)
			}

			if len(capturedEmits) != 1 {
				t.Fatalf("Expected 1 emit, but got %d", len(capturedEmits))
			}

			if capturedEmits[0] != tc.expectedOutput {
				t.Errorf("Mismatch on emitted string.\n  Got:      '%s'\n  Expected: '%s'", capturedEmits[0], tc.expectedOutput)
			}
		})
	}
}
