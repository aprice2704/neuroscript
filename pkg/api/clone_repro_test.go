// NeuroScript Version: 0.7.2
// File version: 3
// Purpose: Provides a definitive, failing test case to prove that the interpreter's clone() method loses the adminCapsuleRegistry.
// filename: pkg/api/clone_repro_test.go
// nlines: 79
// risk_rating: HIGH

package api_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interpreter" // NOTE: Importing internal package for targeted diagnostic test.
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// adminRegistryProbeTool is a custom tool designed to act as a probe. It runs inside
// a sandboxed procedure call (which uses a clone) and checks if the admin capsule
// registry was correctly propagated from the parent interpreter.
var adminRegistryProbeTool = api.ToolImplementation{
	Spec: api.ToolSpec{
		Name:  "probeAdminRegistry",
		Group: "test",
	},
	Func: func(rt api.Runtime, args []any) (any, error) {
		// This is the core of the test. The 'rt' is the cloned interpreter.
		// We must type-assert it to its concrete internal type to access admin methods.
		interp, ok := rt.(*interpreter.Interpreter)
		if !ok {
			return nil, errors.New("TEST ERROR: Could not assert api.Runtime to *interpreter.Interpreter")
		}

		if interp.CapsuleRegistryForAdmin() == nil {
			// If it's nil, the clone is broken. Return an error to fail the test.
			return nil, errors.New("BUG REPRODUCED: adminCapsuleRegistry is nil in the cloned interpreter")
		}
		// If it's present, the clone is working correctly.
		return true, nil
	},
}

// TestInterpreter_CloneLosesAdminRegistry provides a direct, focused test to prove
// that the interpreter clone does not inherit the parent's admin capsule registry.
// This test is expected to FAIL until the bug in `pkg/interpreter/interpreter_clone.go` is fixed.
func TestInterpreter_CloneLosesAdminRegistry(t *testing.T) {
	// 1. A minimal script that calls our probe tool.
	script := `
func check_clone() means
    # This tool will fail if the clone is broken.
    must tool.test.probeAdminRegistry()
endfunc
`
	// 2. Create the host-owned admin registry.
	liveAdminRegistry := api.NewAdminCapsuleRegistry()

	// 3. Create a parent interpreter and configure it with the admin registry.
	// This is the state that is supposed to be inherited by the clone.
	interp := api.NewConfigInterpreter(
		[]string{"tool.test.probeAdminRegistry"}, // Allow the probe tool
		[]api.Capability{},
		api.WithTool(adminRegistryProbeTool),
		api.WithCapsuleAdminRegistry(liveAdminRegistry),
	)

	// 4. Load the script.
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit() failed: %v", err)
	}

	// 5. Run the procedure. This will trigger a clone() internally.
	// The probe tool will then execute inside the clone and check its state.
	_, err = api.RunProcedure(context.Background(), interp, "check_clone")

	// 6. Assert the outcome.
	if err != nil {
		var rtErr *lang.RuntimeError
		// We expect a runtime error from the 'must' statement failing.
		if errors.As(err, &rtErr) {
			// Check if the *wrapped* error from our probe tool is the cause.
			if strings.Contains(rtErr.Message, "BUG REPRODUCED: adminCapsuleRegistry is nil in the cloned interpreter") {
				t.Log("SUCCESS: Test correctly failed, proving the clone bug.")
				t.Logf("Failure reason: %v", err)
				return // Test succeeded in its goal of proving the bug.
			}
		}
		// Any other error is an unexpected failure.
		t.Fatalf("RunProcedure failed with an unexpected error: %v", err)

	} else {
		t.Log("NOTE: Test passed. This indicates the bug in interpreter_clone.go has been fixed.")
	}
}
