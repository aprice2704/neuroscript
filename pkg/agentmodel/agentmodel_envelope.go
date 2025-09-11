// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Converted ValidateAgentModelEnvelope from a method to a standalone function to resolve compiler errors.
// filename: pkg/agentmodel/agentmodel_envelope.go
// nlines: 127
// risk_rating: MEDIUM

package agentmodel

import (
	"fmt"
	"strings"

	cap "github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// AgentModelEnvelope declares the external effects and requirements for using
// a registered AgentModel. It is persisted with the model registration and
// checked at ask-time against the current ExecPolicy.
type AgentModelEnvelope struct {
	// Logical handle used by scripts: ask "<Name>", ...
	Name string

	// Provider + ModelID are informational for operators and logs.
	Provider string
	ModelID  string

	// Hosts the model will contact (host or host:port). Wildcards allowed (e.g., "*.openai.com:443").
	Hosts []string

	// SecretEnvKeys lists env var names expected to hold credentials (e.g., OPENAI_API_KEY).
	SecretEnvKeys []string

	// Budget expectations (optional). Zero means "no minimum expected".
	// If set, Validate ensures the policy's limits meet or exceed these minima;
	// a zero policy value is treated as "unlimited" and therefore sufficient.
	BudgetCurrency  string // e.g., "CAD"
	MinPerCallCents int    // e.g., 50 == $0.50 CAD
	MinPerRunCents  int    // e.g., 1500 == $15.00 CAD
}

// ValidateAgentModelEnvelope checks that the current ExecPolicy grants are
// sufficient to use this AgentModel: env keys, network hosts, model use grant,
// and (optionally) budget currency/limits.
func ValidateAgentModelEnvelope(p *policy.ExecPolicy, env AgentModelEnvelope) error {
	if p == nil {
		return fmt.Errorf("policy: %w", policy.ErrPolicy)
	}

	// Build required capability set from the envelope.
	needs := envelopeNeeds(env)

	if !cap.CapsSatisfied(needs, p.Grants.Grants) {
		return fmt.Errorf("policy: %w (missing env/net/model grants for %q)", policy.ErrCapability, env.Name)
	}

	// Budget checks: currency must exist if specified, and limits must meet minima when both sides non-zero.
	if c := strings.TrimSpace(env.BudgetCurrency); c != "" {
		// Per-call
		if env.MinPerCallCents > 0 {
			pc := p.Grants.Limits.BudgetPerCallCents[c]
			if pc > 0 && pc < env.MinPerCallCents {
				return fmt.Errorf("policy: %w (per-call budget %s %d < required %d)", policy.ErrCapability, c, pc, env.MinPerCallCents)
			}
		}
		// Per-run
		if env.MinPerRunCents > 0 {
			rc := p.Grants.Limits.BudgetPerRunCents[c]
			if rc > 0 && rc < env.MinPerRunCents {
				return fmt.Errorf("policy: %w (per-run budget %s %d < required %d)", policy.ErrCapability, c, rc, env.MinPerRunCents)
			}
		}
	}

	return nil
}

// envelopeNeeds converts the envelope into a minimal set of required capabilities.
func envelopeNeeds(env AgentModelEnvelope) []cap.Capability {
	var needs []cap.Capability

	// env:read for each secret key
	if len(env.SecretEnvKeys) > 0 {
		needs = append(needs, cap.Capability{
			Resource: "env",
			Verbs:    []string{"read"},
			Scopes:   dedupLower(env.SecretEnvKeys),
		})
	}

	// net:read for each host
	if len(env.Hosts) > 0 {
		needs = append(needs, cap.Capability{
			Resource: "net",
			Verbs:    []string{"read"},
			Scopes:   dedupLower(env.Hosts),
		})
	}

	// model:use for this logical model name (if provided)
	if strings.TrimSpace(env.Name) != "" {
		needs = append(needs, cap.Capability{
			Resource: "model",
			Verbs:    []string{"use"},
			Scopes:   []string{strings.TrimSpace(env.Name)},
		})
	}

	return needs
}

func dedupLower(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		l := strings.ToLower(strings.TrimSpace(s))
		if l == "" {
			continue
		}
		if _, ok := seen[l]; ok {
			continue
		}
		seen[l] = struct{}{}
		out = append(out, l)
	}
	return out
}
