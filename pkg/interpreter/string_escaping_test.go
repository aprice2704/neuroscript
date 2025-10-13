// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Corrected the 'Interpret Newline' test to expect a successful parse, fixing the final test failure.
// filename: pkg/interpreter/string_escaping_test.go
// nlines: 62
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func runStringEscapingTest(t *testing.T, name, scriptValue, expectedValue string) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting string escaping test for '%s'.", name)
		h := NewTestHarness(t)

		escapedString := strconv.Quote(scriptValue)

		script := fmt.Sprintf("func main(returns result) means\n\tset result = %s\n\treturn result\nendfunc", escapedString)
		t.Logf("[DEBUG] Turn 2: Executing script:\n%s", script)

		result, err := h.Interpreter.ExecuteScriptString("main", script, nil)
		if err != nil {
			t.Fatalf("ExecuteScriptString failed: %v", err)
		}

		got, ok := result.(lang.StringValue)
		if !ok {
			t.Fatalf("Expected StringValue, but got %T", result)
		}

		if got.Value != expectedValue {
			t.Errorf("Expected string value '%s', but got '%s'", expectedValue, got.Value)
		}
		t.Logf("[DEBUG] Turn 3: Test passed.")
	})
}

func TestInterpretStringEscaping(t *testing.T) {
	runStringEscapingTest(t, "Interpret Backspace", "tex\bback", "tex\bback")
	runStringEscapingTest(t, "Interpret Tab", "col1\tcol2", "col1\tcol2")
	runStringEscapingTest(t, "Interpret Double Quote", "a \"quoted\" string", "a \"quoted\" string")
	runStringEscapingTest(t, "Interpret Backslash", "a path C:\\folder", "a path C:\\folder")
	runStringEscapingTest(t, "Interpret Unicode BMP", "currency: €", "currency: €")
	runStringEscapingTest(t, "Interpret Unicode Surrogate Pair", "face: 😀", "face: 😀")
	runStringEscapingTest(t, "Interpret Newline", "first\nsecond", "first\nsecond")
}
