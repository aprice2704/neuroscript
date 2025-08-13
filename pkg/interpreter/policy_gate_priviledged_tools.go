// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Adds a test to verify that all privileged tools are blocked by the policy gate in a non-privileged context. Corrected package to 'interpreter' to resolve build errors.
// filename: pkg/interpreter/policy_gate_privileged_tools_test.go
// nlines: 83
// risk_rating: MEDIUM

package interpreter

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/runtime"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // Ensure all tools are registered for the test
)

// TestPolicyGate_BlocksAllPrivilegedToolsInNormalContext verifies that the policy gate
// correctly denies all registered tools that are marked with 'RequiresTrust = true'
// when the interpreter is operating in a non-privileged (ContextNormal) context.
// This test is crucial for ensuring the security sandbox is fail-closed by default for
// sensitive operations. It tests the gate's logic without executing the tools,
// thus avoiding any side effects.
func TestPolicyGate_BlocksAllPrivilegedToolsInNormalContext(t *testing.T) {
	// 1. Setup an interpreter with a restrictive policy.
	// This policy simulates a standard, untrusted execution environment.
	policy := &runtime.ExecPolicy{
		Context: runtime.ContextNormal,
		Allow:   []string{"*"}, // Allow all tools by name, so only trust/capability checks are tested.
	}

	// 2. Create an interpreter to get access to a fully populated tool registry.
	interp := NewInterpreter()
	if interp.NTools() == 0 {
		t.Fatal("Tool registry is empty. Ensure tools are being registered in the test environment (e.g., via toolbundles).")
	}

	// 3. Get all registered tools.
	allTools := interp.ToolRegistry().ListTools()
	t.Logf("Found %d registered tools to check against the policy gate.", len(allTools))

	// 4. Iterate and test each tool against the policy.
	privilegedToolsFound := 0
	for _, toolImpl := range allTools {
		if !toolImpl.RequiresTrust {
			continue
		}

		privilegedToolsFound++
		toolName := string(toolImpl.FullName)

		// We run this as a sub-test for better reporting.
		t.Run(toolName, func(t *testing.T) {
			// Construct the metadata that the policy gate evaluates.
			meta := runtime.ToolMeta{
				Name:          toolName,
				RequiresTrust: toolImpl.RequiresTrust,
				RequiredCaps:  toolImpl.RequiredCaps,
				Effects:       toolImpl.Effects,
			}

			// 5. Directly check the policy gate's decision, avoiding tool execution.
			err := policy.CanCall(meta)

			// 6. Assert that the tool was blocked with the correct error.
			if err == nil {
				t.Errorf("Tool '%s' is privileged but was NOT blocked by the policy gate in a normal context.", toolName)
			} else if !errors.Is(err, runtime.ErrTrust) {
				t.Errorf("Tool '%s' was blocked, but with the wrong error. Got '%v', expected to wrap '%v'", toolName, err, runtime.ErrTrust)
			}
		})
	}

	if privilegedToolsFound == 0 {
		t.Log("WARNING: No privileged tools were found in the registry to test. The test passed vacuously.")
	} else {
		t.Logf("Successfully verified that %d privileged tools are correctly blocked by the policy gate.", privilegedToolsFound)
	}
}
