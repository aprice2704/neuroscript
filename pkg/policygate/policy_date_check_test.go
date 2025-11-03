// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Adds dedicated tests for the policygate.Check function.
// filename: pkg/policygate/policy_gate_check_test.go
// nlines: 83
// risk_rating: LOW

package policygate

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// mockRuntime implements the Runtime interface for testing.
type mockRuntime struct {
	policy *policy.ExecPolicy
}

func (m *mockRuntime) GetExecPolicy() *policy.ExecPolicy {
	return m.policy
}

func TestPolicyGate_Check(t *testing.T) {
	grant := capability.MustParse("model:admin:*")
	requiredCap := capability.MustParse("model:admin:*")

	testCases := []struct {
		name        string
		policy      *policy.ExecPolicy
		capToCheck  capability.Capability
		expectErrIs error
	}{
		{
			name:        "Success: Capability is granted",
			policy:      policy.NewBuilder(policy.ContextConfig).GrantCap(grant).Build(),
			capToCheck:  requiredCap,
			expectErrIs: nil,
		},
		{
			name:        "Failure: No policy",
			policy:      nil,
			capToCheck:  requiredCap,
			expectErrIs: policy.ErrPolicy,
		},
		{
			name: "Failure: Policy has no grants",
			policy: policy.NewBuilder(policy.ContextConfig).
				Allow("tool.agentmodel.register"). // Allow list is irrelevant
				Build(),
			capToCheck:  requiredCap,
			expectErrIs: policy.ErrCapability,
		},
		{
			name: "Failure: Policy has wrong grants",
			policy: policy.NewBuilder(policy.ContextConfig).
				Grant("fs:read:*").
				Build(),
			capToCheck:  requiredCap,
			expectErrIs: policy.ErrCapability,
		},
		{
			name: "Success: Policy has wildcard grant",
			policy: policy.NewBuilder(policy.ContextConfig).
				Grant("*:*:*").
				Build(),
			capToCheck:  requiredCap,
			expectErrIs: nil,
		},
		{
			name: "Failure: Allow list is present but grants are not",
			policy: policy.NewBuilder(policy.ContextConfig).
				Allow("model:admin:*"). // This should be ignored by Check()
				Build(),
			capToCheck:  requiredCap,
			expectErrIs: policy.ErrCapability,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rt := &mockRuntime{policy: tc.policy}
			err := Check(rt, tc.capToCheck)

			if !errors.Is(err, tc.expectErrIs) {
				t.Errorf("Expected error '%v', but got '%v'", tc.expectErrIs, err)
			}

			// Check error message for grant failure
			if tc.expectErrIs == policy.ErrCapability {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
				var rtErr *lang.RuntimeError
				if !errors.As(err, &rtErr) {
					t.Fatalf("Expected a *lang.RuntimeError, got %T", err)
				}
				if !strings.Contains(rtErr.Message, "missing required grants") {
					t.Errorf("Expected error message to contain 'missing required grants', got: %s", rtErr.Message)
				}
			}
		})
	}
}
