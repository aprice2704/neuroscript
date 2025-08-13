// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Contains policy gate tests for trust contexts and allow/deny patterns.
// filename: pkg/interpreter/policy_gate_access_test.go
// nlines: 150
// risk_rating: MEDIUM

package interpreter

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/runtime"
)

var (
	// Mock tools with different trust levels
	accessTestTrustedTool = runtime.ToolMeta{
		Name:          "tool.os.setenv",
		RequiresTrust: true,
	}
	accessTestNormalTool = runtime.ToolMeta{
		Name:          "tool.str.contains",
		RequiresTrust: false,
	}
	accessTestAnotherNormalTool = runtime.ToolMeta{
		Name:          "tool.math.add",
		RequiresTrust: false,
	}
)

func TestPolicyGate_AccessControl(t *testing.T) {
	testCases := []struct {
		name        string
		policy      *runtime.ExecPolicy
		tool        runtime.ToolMeta
		expectErrIs error
	}{
		// --- Trust Context Scenarios ---
		{
			name: "[Trust] Trusted tool succeeds in config context",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextConfig,
				Allow:   []string{"*"},
			},
			tool:        accessTestTrustedTool,
			expectErrIs: nil,
		},
		{
			name: "[Trust] Trusted tool fails in normal context",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
			},
			tool:        accessTestTrustedTool,
			expectErrIs: runtime.ErrTrust,
		},
		{
			name: "[Trust] Trusted tool fails in test context",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextTest,
				Allow:   []string{"*"},
			},
			tool:        accessTestTrustedTool,
			expectErrIs: runtime.ErrTrust,
		},
		{
			name: "[Trust] Normal tool succeeds in normal context",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
			},
			tool:        accessTestNormalTool,
			expectErrIs: nil,
		},
		{
			name: "[Trust] Normal tool succeeds in config context",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextConfig,
				Allow:   []string{"*"},
			},
			tool:        accessTestNormalTool,
			expectErrIs: nil,
		},

		// --- Allow/Deny Pattern Scenarios ---
		{
			name: "[Allow/Deny] Deny all overrides everything",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextConfig,
				Allow:   []string{"*"},
				Deny:    []string{"*"},
			},
			tool:        accessTestTrustedTool,
			expectErrIs: runtime.ErrPolicy,
		},
		{
			name: "[Allow/Deny] Exact deny matches",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
				Deny:    []string{"tool.str.contains"},
			},
			tool:        accessTestNormalTool,
			expectErrIs: runtime.ErrPolicy,
		},
		{
			name: "[Allow/Deny] Wildcard deny matches",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"tool.*"},
				Deny:    []string{"tool.str.*"},
			},
			tool:        accessTestNormalTool,
			expectErrIs: runtime.ErrPolicy,
		},
		{
			name: "[Allow/Deny] Deny overrides specific allow",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"tool.str.contains"},
				Deny:    []string{"tool.str.contains"},
			},
			tool:        accessTestNormalTool,
			expectErrIs: runtime.ErrPolicy,
		},
		{
			name: "[Allow/Deny] Success with specific allow and no deny",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"tool.math.add"},
			},
			tool:        accessTestAnotherNormalTool,
			expectErrIs: nil,
		},
		{
			name: "[Allow/Deny] Failure because not in specific allow list",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"tool.math.add"}, // str.contains is not in this list
			},
			tool:        accessTestNormalTool,
			expectErrIs: runtime.ErrPolicy,
		},
		{
			name: "[Allow/Deny] Default deny when allow list is empty",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{}, // Empty allow list means deny everything
			},
			tool:        accessTestNormalTool,
			expectErrIs: runtime.ErrPolicy,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.CanCall(tc.tool)
			if !errors.Is(err, tc.expectErrIs) {
				t.Errorf("Expected error '%v', but got '%v'", tc.expectErrIs, err)
			}
		})
	}
}
