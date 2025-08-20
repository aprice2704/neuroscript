// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Provides additional self-contained unit tests for the ExecPolicy.CanCall gating function.
// filename: pkg/runtime/policy_gate_more_test.go
// nlines: 110
// risk_rating: MEDIUM

package runtime

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/policy/capability"
)

func TestExecPolicy_CanCall_FromScratch(t *testing.T) {
	// Define common tool metadata and capabilities to be reused in test cases.
	basicTool := ToolMeta{Name: "tool.basic.run"}
	trustedTool := ToolMeta{Name: "tool.admin.setConfig", RequiresTrust: true}
	capTool := ToolMeta{
		Name:         "tool.fs.writeFile",
		RequiredCaps: []capability.Capability{{Resource: "fs", Verbs: []string{"write"}}},
	}
	fsWriteCap := capability.Capability{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"*"}}

	testCases := []struct {
		name    string
		policy  *ExecPolicy
		tool    ToolMeta
		wantErr error
	}{
		{
			name: "Success - Simple allow",
			policy: &ExecPolicy{
				Context: ContextNormal,
				Allow:   []string{"tool.basic.run"},
			},
			tool:    basicTool,
			wantErr: nil,
		},
		{
			name: "Failure - Simple deny",
			policy: &ExecPolicy{
				Context: ContextNormal,
				Allow:   []string{"*"},
				Deny:    []string{"tool.basic.run"},
			},
			tool:    basicTool,
			wantErr: ErrPolicy,
		},
		{
			name: "Failure - Not in active allow list",
			policy: &ExecPolicy{
				Context: ContextNormal,
				Allow:   []string{"tool.other.thing"},
			},
			tool:    basicTool,
			wantErr: ErrPolicy,
		},
		{
			name: "Success - Wildcard allow",
			policy: &ExecPolicy{
				Context: ContextNormal,
				Allow:   []string{"tool.basic.*"},
			},
			tool:    basicTool,
			wantErr: nil,
		},
		{
			name: "Failure - Trust required in normal context",
			policy: &ExecPolicy{
				Context: ContextNormal,
				Allow:   []string{"*"},
			},
			tool:    trustedTool,
			wantErr: ErrTrust,
		},
		{
			name: "Success - Trust required in config context",
			policy: &ExecPolicy{
				Context: ContextConfig,
				Allow:   []string{"*"},
			},
			tool:    trustedTool,
			wantErr: nil,
		},
		{
			name: "Failure - Capability not granted",
			policy: &ExecPolicy{
				Context: ContextConfig,
				Allow:   []string{"*"},
				Grants:  capability.GrantSet{},
			},
			tool:    capTool,
			wantErr: ErrCapability,
		},
		{
			name: "Success - Capability granted",
			policy: &ExecPolicy{
				Context: ContextConfig,
				Allow:   []string{"*"},
				Grants: capability.GrantSet{
					Grants: []capability.Capability{fsWriteCap},
				},
			},
			tool:    capTool,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure counters are initialized for every run to prevent state leakage.
			if tc.policy.Grants.Counters == nil {
				tc.policy.Grants.Counters = capability.NewCounters()
			}

			err := tc.policy.CanCall(tc.tool)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("CanCall() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
