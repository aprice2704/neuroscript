// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Correctly constructs a privileged ExecPolicy using the builder to fix 'missing required grants' error.
// Latest change: Re-wrote test to check for capsuleStore propagation, not adminCapsuleRegistry.
// filename: pkg/interpreter/clone_repro_internal_test.go
// nlines: 104
// risk_rating: HIGH

package interpreter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- THE FIX: This tool now probes for the single, unified CapsuleStore ---
var capsuleStoreProbeTool = tool.ToolImplementation{
	Spec: tool.ToolSpec{
		Name:  "probeCapsuleStore",
		Group: "test",
	},
	Func: func(rt tool.Runtime, args []any) (any, error) {
		interp, ok := rt.(*interpreter.Interpreter)
		if !ok {
			return nil, errors.New("TEST ERROR: Could not assert tool.Runtime to *interpreter.Interpreter")
		}

		if interp.CapsuleStore() == nil {
			return nil, errors.New("BUG REPRODUCED: capsuleStore is nil in the cloned interpreter")
		}
		return true, nil
	},
}

func TestInterpreter_ClonePropagatesCapsuleStore(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting TestInterpreter_ClonePropagatesCapsuleStore.")
	script := `
func check_clone() means
    must tool.test.probeCapsuleStore()
endfunc
`
	// --- THE FIX: Create a store to inject ---
	customStore := capsule.NewStore(capsule.NewRegistry()) // A custom, non-default store

	h := NewTestHarness(t)

	// For this specific test, we need a custom interpreter with a specific
	// policy and registry. We use the builder to ensure the policy is valid.
	privilegedPolicy := policy.NewBuilder(policy.ContextConfig).
		Allow("*").
		Grant("tool:exec:*"). // Grant permission to execute the probe tool.
		Build()

	// Create a new interpreter with the specific configuration needed for this test,
	// reusing the HostContext from the harness.
	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(h.HostContext),
		interpreter.WithExecPolicy(privilegedPolicy),
		// --- THE FIX: Inject the single, unified store ---
		interpreter.WithCapsuleStore(customStore),
	)

	t.Logf("[DEBUG] Turn 2: Harness and custom interpreter created; capsule store set via options.")

	if _, err := interp.ToolRegistry().RegisterTool(capsuleStoreProbeTool); err != nil {
		t.Fatalf("Failed to register probe tool: %v", err)
	}
	t.Logf("[DEBUG] Turn 3: Probe tool registered.")

	tree, pErr := h.Parser.Parse(script)
	if pErr != nil {
		t.Fatalf("Parser failed: %v", pErr)
	}
	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		t.Fatalf("AST Build failed: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	t.Logf("[DEBUG] Turn 4: Script loaded.")

	_, err := interp.Run("check_clone")
	t.Logf("[DEBUG] Turn 5: 'check_clone' procedure executed.")

	if err != nil {
		var rtErr *lang.RuntimeError
		if errors.As(err, &rtErr) {
			// --- THE FIX: Check for the new error message ---
			if strings.Contains(rtErr.Unwrap().Error(), "BUG REPRODUCED: capsuleStore is nil") {
				t.Errorf("FAILURE CONFIRMED: The clone() method did not propagate the capsuleStore. Error: %v", err)
				return
			}
		}
		t.Fatalf("RunProcedure failed with an unexpected error: %v", err)
	} else {
		t.Log("SUCCESS: Test passed, indicating the bug in interpreter_clone.go has been fixed.")
	}
	t.Logf("[DEBUG] Turn 6: Test completed.")
}
