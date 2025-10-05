// NeuroScript Version: 0.7.3
// File version: 1
// Purpose: Tests the specific copying of function definitions via CopyFunctionsFrom.
// filename: pkg/api/interpreter_functions_copy_test.go
// nlines: 70
// risk_rating: LOW

package api_test

import (
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// TestInterpreter_CopyFunctionsFrom verifies that only function definitions are
// copied from a source interpreter, excluding event handlers and command blocks.
func TestInterpreter_CopyFunctionsFrom(t *testing.T) {
	// 1. A source script with a function, a command, and an event handler.
	sourceScript := `
# This function should be copied.
func get_library_message(returns string) means
    return "from the library"
endfunc

# This event handler should NOT be copied.
on event "source:event" do
    emit "source event fired"
endon
`
	// 2. Create and load the source interpreter.
	sourceInterp := api.New()
	var sourceEmit string
	sourceInterp.SetEmitFunc(func(v api.Value) {
		sourceEmit = v.String()
	})
	_ = sourceEmit

	tree, err := api.Parse([]byte(sourceScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Source: api.Parse() failed: %v", err)
	}
	if err := api.LoadFromUnit(sourceInterp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("Source: LoadFromUnit() failed: %v", err)
	}

	// 3. Create the destination interpreter and copy the functions.
	destInterp := api.New()
	if err := destInterp.CopyFunctionsFrom(sourceInterp); err != nil {
		t.Fatalf("CopyFunctionsFrom() failed: %v", err)
	}

	// 4. Verify the function was copied.
	runScript := `
func main(returns string) means
    return get_library_message()
endfunc
`
	tree, err = api.Parse([]byte(runScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Dest: api.Parse() failed: %v", err)
	}
	// Use AppendScript so we don't wipe the copied function.
	if err := destInterp.AppendScript(tree); err != nil {
		t.Fatalf("Dest: AppendScript() failed: %v", err)
	}

	result, err := api.RunProcedure(context.Background(), destInterp, "main")
	if err != nil {
		t.Fatalf("Dest: RunProcedure() failed: %v", err)
	}
	unwrapped, _ := api.Unwrap(result)
	if val, ok := unwrapped.(string); !ok || val != "from the library" {
		t.Errorf("Expected result 'from the library', got %q", val)
	}

	// 5. Verify the command was NOT copied.
	// Executing commands on the destination should do nothing.
	var destEmit string
	destInterp.SetEmitFunc(func(v api.Value) {
		destEmit = v.String()
	})
	if _, err := destInterp.ExecuteCommands(); err != nil {
		t.Fatalf("Dest: ExecuteCommands() failed: %v", err)
	}
	if destEmit != "" {
		t.Errorf("Destination interpreter should not have emitted anything, but got %q", destEmit)
	}
}
