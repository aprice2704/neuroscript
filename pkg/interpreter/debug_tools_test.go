// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Updated to use the new exported ID() method.
// filename: pkg/interpreter/debug_tools_test.go
// nlines: 40
// risk_rating: LOW

package interpreter_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestDebugTools(t *testing.T) {
	t.Run("tool.debug.dumpClones", func(t *testing.T) {
		h := NewTestHarness(t)
		rootInterpreter := h.Interpreter

		// Create a clone to ensure it appears in the dump.
		forkedInterpreter := rootInterpreter.Clone()

		// The script simply calls the tool.
		script := `
			func main() means
				return tool.debug.dumpClones()
			endfunc
		`
		// We need to use ExecuteScriptString here as the forked interpreter
		// does not have the 'main' procedure loaded.
		result, err := forkedInterpreter.ExecuteScriptString("main", script, nil)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		report, ok := result.(lang.StringValue)
		if !ok {
			t.Fatalf("Expected tool to return a string, but got %T", result)
		}

		// Assert that the report contains the IDs of both interpreters.
		if !strings.Contains(report.Value, rootInterpreter.ID()) {
			t.Errorf("Debug report is missing the root interpreter ID: %s", rootInterpreter.ID())
		}
		if !strings.Contains(report.Value, forkedInterpreter.ID()) {
			t.Errorf("Debug report is missing the forked interpreter ID: %s", forkedInterpreter.ID())
		}
	})
}
