// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Updates the persistence test to use the new HostContextBuilder for configuration and the correct ExecuteCommands method.
// filename: pkg/api/interpreter_persistence_test.go
// nlines: 67
// risk_rating: LOW

package api_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
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
	var capturedEmit string
	hc, err := api.NewHostContextBuilder().
		WithLogger(logging.NewNoOpLogger()).
		WithStdout(nil).
		WithStdin(nil).
		WithStderr(nil).
		WithEmitFunc(func(v api.Value) {
			val, _ := v.(lang.Value)
			capturedEmit = val.String()
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}
	interp := api.New(api.WithHostContext(hc))

	// 3. Phase 1: Parse and append all scripts to build the state.
	defsTree, err := api.Parse([]byte(defsScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Failed to parse defs script: %v", err)
	}
	commandTree, err := api.Parse([]byte(commandScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Failed to parse command script: %v", err)
	}

	if err := interp.AppendScript(defsTree); err != nil {
		t.Fatalf("Failed to append defs script: %v", err)
	}
	if err := interp.AppendScript(commandTree); err != nil {
		t.Fatalf("Failed to append command script: %v", err)
	}

	// 4. Phase 2: Execute the loaded command blocks.
	_, err = interp.ExecuteCommands()
	if err != nil {
		// This is the critical check. If the function wasn't persisted,
		// this will fail with a "procedure not found" error.
		t.Fatalf("ExecuteCommands failed with an unexpected error: %v", err)
	}

	// 5. Verify the function was called correctly.
	expected := "boot sequence successful"
	if !strings.Contains(capturedEmit, expected) {
		t.Errorf("Execution did not produce the correct output.\n  Expected to contain: %q\n  Got: %q", expected, capturedEmit)
	}
}
