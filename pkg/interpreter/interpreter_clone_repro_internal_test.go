// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: FIX: Updates the test to work with the refactored interpreter by removing the obsolete WithCapsuleAdminRegistry option and verifying the default registry is propagated correctly.
// filename: pkg/interpreter/clone_repro_internal_test.go
// nlines: 95
// risk_rating: HIGH

package interpreter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// adminRegistryProbeTool is a custom tool designed to act as a probe. It runs inside
// a sandboxed procedure call (which uses a clone) and checks if the admin capsule
// registry was correctly propagated from the parent interpreter.
var adminRegistryProbeTool = tool.ToolImplementation{
	Spec: tool.ToolSpec{
		Name:  "probeAdminRegistry",
		Group: "test",
	},
	Func: func(rt tool.Runtime, args []any) (any, error) {
		// This is the core of the test. The 'rt' is the cloned interpreter.
		interp, ok := rt.(*interpreter.Interpreter)
		if !ok {
			return nil, errors.New("TEST ERROR: Could not assert tool.Runtime to *interpreter.Interpreter")
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
	// 1. A minimal script that calls our probe tool from inside a procedure.
	script := `
func check_clone() means
    # This tool call will fail if the clone is broken because the 'must' will fail.
    must tool.test.probeAdminRegistry()
endfunc
`
	// 2. Create a parent interpreter. It will create its own default capsule registry.
	interp := interpreter.NewInterpreter(
		interpreter.WithLogger(logging.NewTestLogger(t)),
		interpreter.WithExecPolicy(&interfaces.ExecPolicy{Context: policy.ContextConfig, Allow: []string{"*"}}),
	)

	// 3. Register our probe tool.
	if _, err := interp.ToolRegistry().RegisterTool(adminRegistryProbeTool); err != nil {
		t.Fatalf("Failed to register probe tool: %v", err)
	}

	// 4. Load the script.
	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(script)
	if pErr != nil {
		t.Fatalf("Parser failed: %v", pErr)
	}
	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		t.Fatalf("AST Build failed: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 5. Run the procedure. This will trigger a clone() internally.
	_, err := interp.Run("check_clone")

	// 6. Assert the outcome.
	if err != nil {
		var rtErr *lang.RuntimeError
		if errors.As(err, &rtErr) {
			// Check if the *wrapped* error from our probe tool is the cause.
			if strings.Contains(rtErr.Unwrap().Error(), "BUG REPRODUCED: adminCapsuleRegistry is nil") {
				// This is the expected failure. The test fails, proving the bug.
				t.Errorf("FAILURE CONFIRMED: The clone() method did not propagate the adminCapsuleRegistry. Error: %v", err)
				return
			}
		}
		// Any other error is an unexpected test setup failure.
		t.Fatalf("RunProcedure failed with an unexpected error: %v", err)
	} else {
		// If there's no error, it means the probe tool succeeded, and the bug is fixed.
		t.Log("SUCCESS: Test passed, indicating the bug in interpreter_clone.go has been fixed.")
	}
}
