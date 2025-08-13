// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Provides comprehensive, table-driven unit tests for the runtime.ExecPolicy gate, covering trust, allow/deny, capabilities, and limits.
// filename: pkg/interpreter/policy_gate_extended_test.go
// nlines: 250
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/runtime"
)

// --- Test Suite Setup ---

var (
	// Mock Tools for Testing
	trustedAdminTool = runtime.ToolMeta{
		Name:          "tool.agentmodel.register",
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
		},
	}
	normalReadTool = runtime.ToolMeta{
		Name:          "tool.fs.read",
		RequiresTrust: false,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/data/input.txt"}},
		},
	}
	networkTool = runtime.ToolMeta{
		Name:          "tool.http.get",
		RequiresTrust: false,
		RequiredCaps: []capability.Capability{
			{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"api.example.com:443"}},
		},
	}
	untrustedDangerousTool = runtime.ToolMeta{
		Name:          "tool.debug.run_command",
		RequiresTrust: true,
	}
)

func TestExecPolicy_CanCall(t *testing.T) {
	testCases := []struct {
		name        string
		policy      *runtime.ExecPolicy
		tool        runtime.ToolMeta
		expectErrIs error
	}{
		// --- Trust Context Tests ---
		{
			name: "Success: Trusted tool in config context",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextConfig,
				Allow:   []string{"*"},
				Grants: capability.NewGrantSet([]capability.Capability{
					{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
				}, capability.Limits{}),
			},
			tool:        trustedAdminTool,
			expectErrIs: nil,
		},
		{
			name: "Failure: Trusted tool in normal context",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
				Grants: capability.NewGrantSet([]capability.Capability{
					{Resource: "model", Verbs: []string{"admin"}, Scopes: []string{"*"}},
				}, capability.Limits{}),
			},
			tool:        trustedAdminTool,
			expectErrIs: runtime.ErrTrust,
		},
		{
			name: "Success: Normal tool in normal context",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
				Grants: capability.NewGrantSet([]capability.Capability{
					{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/data/*"}},
				}, capability.Limits{}),
			},
			tool:        normalReadTool,
			expectErrIs: nil,
		},

		// --- Allow/Deny Pattern Tests ---
		{
			name: "Failure: Deny rule overrides allow rule",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextConfig,
				Allow:   []string{"tool.agentmodel.*"},
				Deny:    []string{"tool.agentmodel.register"},
				Grants:  capability.NewGrantSet(nil, capability.Limits{}),
			},
			tool:        trustedAdminTool,
			expectErrIs: runtime.ErrPolicy,
		},
		{
			name: "Failure: Not in allow list",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"tool.fs.*"},
				Grants:  capability.NewGrantSet(nil, capability.Limits{}),
			},
			tool:        networkTool,
			expectErrIs: runtime.ErrPolicy,
		},
		{
			name: "Success: Allow list with wildcard",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"tool.http.*"},
				Grants: capability.NewGrantSet([]capability.Capability{
					{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"*"}},
				}, capability.Limits{}),
			},
			tool:        networkTool,
			expectErrIs: nil,
		},
		{
			name: "Failure: Deny all",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Deny:    []string{"*"},
				Grants:  capability.NewGrantSet(nil, capability.Limits{}),
			},
			tool:        normalReadTool,
			expectErrIs: runtime.ErrPolicy,
		},

		// --- Capability Grant Tests ---
		{
			name: "Failure: Missing required capability",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
				Grants:  capability.NewGrantSet(nil, capability.Limits{}),
			},
			tool:        normalReadTool,
			expectErrIs: runtime.ErrCapability,
		},
		{
			name: "Success: Exact capability grant",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
				Grants: capability.NewGrantSet([]capability.Capability{
					{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/data/input.txt"}},
				}, capability.Limits{}),
			},
			tool:        normalReadTool,
			expectErrIs: nil,
		},
		{
			name: "Success: Wildcard scope grant satisfies specific need",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
				Grants: capability.NewGrantSet([]capability.Capability{
					{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"*.example.com:443"}},
				}, capability.Limits{}),
			},
			tool:        networkTool,
			expectErrIs: nil,
		},
		{
			name: "Failure: Correct resource, wrong verb",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
				Grants: capability.NewGrantSet([]capability.Capability{
					{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"/data/input.txt"}},
				}, capability.Limits{}),
			},
			tool:        normalReadTool,
			expectErrIs: runtime.ErrCapability,
		},

		// --- Limits and Counters Tests ---
		{
			name: "Failure: Per-tool call limit exceeded",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Allow:   []string{"*"},
				Grants: capability.NewGrantSet(
					[]capability.Capability{{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"*"}}},
					capability.Limits{ToolMaxCalls: map[string]int{"tool.http.get": 1}},
				),
			},
			tool:        networkTool,
			expectErrIs: capability.ErrToolExceeded, // Expect this on the *second* call
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// First call
			err := tc.policy.CanCall(tc.tool)

			// For the tool limit test, we need to call it twice.
			if tc.name == "Failure: Per-tool call limit exceeded" {
				if err != nil {
					t.Fatalf("Expected first call to succeed, but got: %v", err)
				}
				// Second call should fail
				err = tc.policy.CanCall(tc.tool)
			}

			if !errors.Is(err, tc.expectErrIs) {
				t.Errorf("Expected error '%v', but got '%v'", tc.expectErrIs, err)
			}
		})
	}
}

func TestAgentModelEnvelopeValidation(t *testing.T) {
	testCases := []struct {
		name        string
		policy      *runtime.ExecPolicy
		envelope    runtime.AgentModelEnvelope
		expectErrIs error
	}{
		{
			name: "Success: All grants and budget sufficient",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Grants: capability.NewGrantSet(
					[]capability.Capability{
						{Resource: "model", Verbs: []string{"use"}, Scopes: []string{"gpt-4"}},
						{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"openai_api_key"}},
						{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"*.openai.com:443"}},
					},
					capability.Limits{
						BudgetPerCallCents: map[string]int{"CAD": 100},
						BudgetPerRunCents:  map[string]int{"CAD": 5000},
					},
				),
			},
			envelope: runtime.AgentModelEnvelope{
				Name:            "gpt-4",
				Hosts:           []string{"api.openai.com:443"},
				SecretEnvKeys:   []string{"OPENAI_API_KEY"},
				BudgetCurrency:  "CAD",
				MinPerCallCents: 50,
				MinPerRunCents:  2000,
			},
			expectErrIs: nil,
		},
		{
			name: "Failure: Missing model:use grant",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Grants: capability.NewGrantSet(
					[]capability.Capability{
						{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"openai_api_key"}},
						{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"*.openai.com:443"}},
					},
					capability.Limits{},
				),
			},
			envelope: runtime.AgentModelEnvelope{
				Name: "gpt-4",
			},
			expectErrIs: runtime.ErrCapability,
		},
		{
			name: "Failure: Insufficient per-call budget",
			policy: &runtime.ExecPolicy{
				Context: runtime.ContextNormal,
				Grants: capability.NewGrantSet(
					[]capability.Capability{{Resource: "model", Verbs: []string{"use"}, Scopes: []string{"*"}}},
					capability.Limits{BudgetPerCallCents: map[string]int{"CAD": 10}},
				),
			},
			envelope: runtime.AgentModelEnvelope{
				Name:            "gpt-4",
				BudgetCurrency:  "CAD",
				MinPerCallCents: 50,
			},
			expectErrIs: runtime.ErrCapability,
		},
		{
			name:   "Failure: Policy is nil",
			policy: nil,
			envelope: runtime.AgentModelEnvelope{
				Name: "gpt-4",
			},
			expectErrIs: runtime.ErrPolicy,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.ValidateAgentModelEnvelope(tc.envelope)
			if !errors.Is(err, tc.expectErrIs) {
				t.Errorf("Expected error '%v', but got '%v'", tc.expectErrIs, err)
			}
		})
	}
}
