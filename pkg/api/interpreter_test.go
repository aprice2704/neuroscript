// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Corrects a parser error by adding a statement to an empty function block in the test source.
// filename: pkg/api/interpreter_test.go
// nlines: 29
// risk_rating: LOW

package api_test

import (
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// TestInterpreter_RunNonExistentProcedure verifies that calling a function
// that hasn't been defined returns a clear error.
func TestInterpreter_RunNonExistentProcedure(t *testing.T) {
	// **FIX:** Add a 'return' statement to make the function body non-empty.
	src := "func do_work() means\n  return\nendfunc"
	tree, err := api.Parse([]byte(src), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Setup failed: api.Parse returned an error: %v", err)
	}

	interp := api.New()
	// Load the program, which defines `do_work`.
	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Attempt to run a completely different function.
	_, runErr := interp.Run("this_function_does_not_exist")
	if runErr == nil {
		t.Fatal("Expected an error when running a non-existent procedure, but got nil")
	}
}
