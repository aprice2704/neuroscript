// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Updated test to call ValidateAgentModelEnvelope as a standalone function.
// filename: pkg/agentmodel/agentmodel_envelope_test.go
// nlines: 115
// risk_rating: MEDIUM

package agentmodel

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

func TestValidateAgentModelEnvelope(t *testing.T) {
	baseEnvelope := AgentModelEnvelope{
		Name:          "test-model",
		Hosts:         []string{"api.openai.com"},
		SecretEnvKeys: []string{"OPENAI_API_KEY"},
	}

	validGrants := []capability.Capability{
		{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"api.openai.com"}},
		{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"openai_api_key"}},
		{Resource: "model", Verbs: []string{"use"}, Scopes: []string{"test-model"}},
	}

	testCases := []struct {
		name     string
		policy   *policy.ExecPolicy
		envelope AgentModelEnvelope
		wantErr  error
	}{
		{
			name: "Success - Valid policy",
			policy: &policy.ExecPolicy{
				Grants: capability.GrantSet{Grants: validGrants},
			},
			envelope: baseEnvelope,
		},
		{
			name: "Fail - Missing network grant",
			policy: &policy.ExecPolicy{
				Grants: capability.GrantSet{Grants: []capability.Capability{
					{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"openai_api_key"}},
					{Resource: "model", Verbs: []string{"use"}, Scopes: []string{"test-model"}},
				}},
			},
			envelope: baseEnvelope,
			wantErr:  policy.ErrCapability,
		},
		{
			name: "Fail - Missing env grant",
			policy: &policy.ExecPolicy{
				Grants: capability.GrantSet{Grants: []capability.Capability{
					{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"api.openai.com"}},
					{Resource: "model", Verbs: []string{"use"}, Scopes: []string{"test-model"}},
				}},
			},
			envelope: baseEnvelope,
			wantErr:  policy.ErrCapability,
		},
		{
			name: "Fail - Missing model grant",
			policy: &policy.ExecPolicy{
				Grants: capability.GrantSet{Grants: []capability.Capability{
					{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"api.openai.com"}},
					{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"openai_api_key"}},
				}},
			},
			envelope: baseEnvelope,
			wantErr:  policy.ErrCapability,
		},
		{
			name: "Fail - Per-call budget too low",
			policy: &policy.ExecPolicy{
				Grants: capability.GrantSet{
					Grants: validGrants,
					Limits: capability.Limits{BudgetPerCallCents: map[string]int{"CAD": 49}},
				},
			},
			envelope: AgentModelEnvelope{
				Name:            "test-model",
				Hosts:           []string{"api.openai.com"},
				SecretEnvKeys:   []string{"OPENAI_API_KEY"},
				BudgetCurrency:  "CAD",
				MinPerCallCents: 50,
			},
			wantErr: policy.ErrCapability,
		},
		{
			name: "Fail - Per-run budget too low",
			policy: &policy.ExecPolicy{
				Grants: capability.GrantSet{
					Grants: validGrants,
					Limits: capability.Limits{BudgetPerRunCents: map[string]int{"CAD": 1499}},
				},
			},
			envelope: AgentModelEnvelope{
				Name:           "test-model",
				Hosts:          []string{"api.openai.com"},
				SecretEnvKeys:  []string{"OPENAI_API_KEY"},
				BudgetCurrency: "CAD",
				MinPerRunCents: 1500,
			},
			wantErr: policy.ErrCapability,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAgentModelEnvelope(tc.policy, tc.envelope)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("ValidateAgentModelEnvelope() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
