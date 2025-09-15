// NeuroScript Version: 0.7.2
// File version: 1
// Purpose: Confirms that function definitions persist across multiple AppendScript calls.
// filename: pkg/api/interpreter_persistence_test.go
// nlines: 65
// risk_rating: LOW

package api_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestStatePersistence_AppendScriptAndExecute(t *testing.T) {
	// 1. Define the separate script files for the boot sequence.
	defsScript := `
		func get_boot_message() means
			return "boot sequence successful"
		endfunc
	`
	commandScript := `
		command
			set msg = get_boot_message()
			emit msg
		endcommand
	`

	// 2. Create a single, persistent interpreter instance.
	interp := api.New()
	var capturedEmit string
	interp.SetEmitFunc(func(v api.Value) {
		val, _ := v.(lang.Value)
		capturedEmit = val.String()
	})

	parserAPI := parser.NewParserAPI(nil)

	// 3. Phase 1: Append all scripts to build the state.
	defsTree, pErr1 := parserAPI.Parse(defsScript)
	if pErr1 != nil {
		t.Fatalf("Failed to parse defs script: %v", pErr1)
	}
	defsProgram, _, bErr1 := parser.NewASTBuilder(nil).Build(defsTree)
	if bErr1 != nil {
		t.Fatalf("Failed to build defs AST: %v", bErr1)
	}

	commandTree, pErr2 := parserAPI.Parse(commandScript)
	if pErr2 != nil {
		t.Fatalf("Failed to parse command script: %v", pErr2)
	}
	commandProgram, _, bErr2 := parser.NewASTBuilder(nil).Build(commandTree)
	if bErr2 != nil {
		t.Fatalf("Failed to build command AST: %v", bErr2)
	}

	if err := interp.AppendScript(&interfaces.Tree{Root: defsProgram}); err != nil {
		t.Fatalf("Failed to append defs script: %v", err)
	}
	if err := interp.AppendScript(&interfaces.Tree{Root: commandProgram}); err != nil {
		t.Fatalf("Failed to append command script: %v", err)
	}

	// 4. Phase 2: Execute the loaded command blocks.
	_, err := interp.Execute()
	if err != nil {
		// This is the critical check. If the function wasn't persisted,
		// this will fail with a "procedure not found" error.
		t.Fatalf("Execute failed with an unexpected error: %v", err)
	}

	// 5. Verify the function was called correctly.
	expected := "boot sequence successful"
	if !strings.Contains(capturedEmit, expected) {
		t.Errorf("Execution did not produce the correct output.\n  Expected to contain: %q\n  Got: %q", expected, capturedEmit)
	}
}
