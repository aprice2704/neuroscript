// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Verifies that the policy gate blocks privileged tools in a normal context.
// filename: pkg/interpreter/priviledged_tools_policy_test.go
// nlines: 83
// risk_rating: MEDIUM

package interpreter

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/policy"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // Ensure all tools are registered for the test
)

// TestPolicyGate_BlocksAllPrivilegedToolsInNormalContext verifies that the policy gate
// correctly denies all registered tools that are marked with 'RequiresTrust = true'
// when the interpreter is operating in a non-privileged (ContextNormal) context.
func TestPolicyGate_BlocksAllPrivilegedToolsInNormalContext(t *testing.T) {
	p := &policy.ExecPolicy{
		Context: policy.ContextNormal,
		Allow:   []string{"*"},
	}

	interp, err := NewTestInterpreter(t, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}
	if interp.NTools() == 0 {
		t.Fatal("Tool registry is empty. Ensure tools are being registered in the test environment.")
	}

	allTools := interp.ToolRegistry().ListTools()
	t.Logf("Found %d registered tools to check against the policy gate.", len(allTools))

	privilegedToolsFound := 0
	for _, toolImpl := range allTools {
		if !toolImpl.RequiresTrust {
			continue
		}

		privilegedToolsFound++
		toolName := string(toolImpl.FullName)

		t.Run(toolName, func(t *testing.T) {
			meta := policy.ToolMeta{
				Name:          toolName,
				RequiresTrust: toolImpl.RequiresTrust,
				RequiredCaps:  toolImpl.RequiredCaps,
				Effects:       toolImpl.Effects,
			}

			// Directly check the policy gate's decision.
			err := p.CanCall(meta)

			if err == nil {
				t.Errorf("Tool '%s' is privileged but was NOT blocked by the policy gate in a normal context.", toolName)
			} else if !errors.Is(err, policy.ErrTrust) {
				t.Errorf("Tool '%s' was blocked, but with the wrong error. Got '%v', expected to wrap '%v'", toolName, err, policy.ErrTrust)
			}
		})
	}

	if privilegedToolsFound == 0 {
		t.Log("WARNING: No privileged tools were found in the registry to test. The test passed vacuously.")
	} else {
		t.Logf("Successfully verified that %d privileged tools are correctly blocked by the policy gate.", privilegedToolsFound)
	}
}
