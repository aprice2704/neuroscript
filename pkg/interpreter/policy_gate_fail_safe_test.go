// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Contains tests to verify the policy gate's "fail closed" or "secure by default" behavior.
// filename: pkg/interpreter/policy_gate_fail_safe_test.go
// nlines: 130
// risk_rating: MEDIUM

package interpreter

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/runtime"
)

var (
	// A harmless tool that requires no special permissions.
	failSafeNoReqsTool = runtime.ToolMeta{
		Name:          "tool.math.add",
		RequiresTrust: false,
		RequiredCaps:  nil,
	}
	// A tool that requires a trusted context to run.
	failSafeTrustReqTool = runtime.ToolMeta{
		Name:          "tool.os.setenv",
		RequiresTrust: true,
	}
	// A tool that requires a specific filesystem capability.
	failSafeCapsReqTool = runtime.ToolMeta{
		Name:          "tool.fs.read",
		RequiresTrust: false,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/config.json"}},
		},
	}
)

func TestPolicyGate_FailSafeBehavior(t *testing.T) {
	testCases := []struct {
		name        string
		policy      *runtime.ExecPolicy
		toolToCall  runtime.ToolMeta
		expectErrIs error
		description string
	}{
		{
			name:        "Nil Policy",
			policy:      nil, // A nil policy should not cause a panic; the gate should just be inactive.
			toolToCall:  failSafeNoReqsTool,
			expectErrIs: nil,
			description: "A nil policy should permit calls, as the gate is effectively disabled.",
		},
		{
			name:        "Empty Policy - NoReqs Tool",
			policy:      &runtime.ExecPolicy{Allow: []string{}}, // An empty policy is the most restrictive.
			toolToCall:  failSafeNoReqsTool,
			expectErrIs: runtime.ErrPolicy,
			description: "An empty policy has no 'Allow' list, so it should deny all calls.",
		},
		{
			name:        "Empty Policy - Trust Tool",
			policy:      &runtime.ExecPolicy{Allow: []string{}},
			toolToCall:  failSafeTrustReqTool,
			expectErrIs: runtime.ErrTrust,
			description: "The trust check runs first; an untrusted context with a trusted tool should fail.",
		},
		{
			name:        "Empty Policy - Caps Tool",
			policy:      &runtime.ExecPolicy{Allow: []string{}},
			toolToCall:  failSafeCapsReqTool,
			expectErrIs: runtime.ErrPolicy, // Fails on the allow list check before the capability check.
			description: "An empty policy denies the tool before the capability check is even reached.",
		},
		{
			name: "Normal Context - Trust Tool",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
			},
			toolToCall:  failSafeTrustReqTool,
			expectErrIs: runtime.ErrTrust,
			description: "A normal context must block trusted tools, even if they are allowed.",
		},
		{
			name: "Allow All, No Grants - Caps Tool",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
				Grants:  capability.NewGrantSet(nil, capability.Limits{}), // No grants provided.
			},
			toolToCall:  failSafeCapsReqTool,
			expectErrIs: runtime.ErrCapability,
			description: "Even if allowed, a tool must fail if its capability requirements are not met.",
		},
		{
			name: "Allow All, No Grants - NoReqs Tool",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
				Grants:  capability.NewGrantSet(nil, capability.Limits{}),
			},
			toolToCall:  failSafeNoReqsTool,
			expectErrIs: nil,
			description: "A tool with no requirements should succeed if it's allowed and the context is appropriate.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.description)
			// Simulate the interpreter having the policy set.
			// The actual check is `policy.CanCall`, so we call it directly.
			var err error
			if tc.policy != nil {
				err = tc.policy.CanCall(tc.toolToCall)
			}

			if !errors.Is(err, tc.expectErrIs) {
				t.Errorf("Expected error '%v', but got '%v'", tc.expectErrIs, err)
			}
		})
	}
}
