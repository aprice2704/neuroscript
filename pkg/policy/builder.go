// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: FIX: Refactored to build and return the neutral *interfaces.ExecPolicy type.
// filename: pkg/policy/builder.go
// nlines: 119
// risk_rating: MEDIUM

package policy

import (
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// Builder is a fluent API for constructing ExecPolicy objects.
type Builder struct {
	policy *interfaces.ExecPolicy
}

// NewBuilder creates a new builder for an ExecPolicy.
func NewBuilder(context interfaces.ExecContext) *Builder {
	p := &interfaces.ExecPolicy{
		Context: context,
		Allow:   []string{},
		Deny:    []string{},
		Grants: capability.GrantSet{
			Grants:   []capability.Capability{},
			Limits:   capability.Limits{},
			Counters: capability.NewCounters(),
		},
	}
	return &Builder{policy: p}
}

// Allow sets the tool allow list.
func (b *Builder) Allow(tools ...string) *Builder {
	if b.policy.Allow == nil {
		b.policy.Allow = []string{}
	}
	b.policy.Allow = dedupMerge(b.policy.Allow, tools...)
	return b
}

// Deny adds tools to the deny list.
func (b *Builder) Deny(tools ...string) *Builder {
	b.policy.Deny = dedupMerge(b.policy.Deny, tools...)
	return b
}

// Grant adds a capability grant by parsing a capability string.
func (b *Builder) Grant(capStr string) *Builder {
	capa := capability.MustParse(capStr)
	b.policy.Grants.Grants = append(b.policy.Grants.Grants, capa)
	return b
}

// GrantCap adds a pre-constructed capability grant.
func (b *Builder) GrantCap(capa capability.Capability) *Builder {
	b.policy.Grants.Grants = append(b.policy.Grants.Grants, capa)
	return b
}

// LimitPerRunCents sets a per-run budget limit for a currency.
func (b *Builder) LimitPerRunCents(currency string, cents int) *Builder {
	if b.policy.Grants.Limits.BudgetPerRunCents == nil {
		b.policy.Grants.Limits.BudgetPerRunCents = make(map[string]int)
	}
	b.policy.Grants.Limits.BudgetPerRunCents[currency] = cents
	return b
}

// LimitPerCallCents sets a per-call budget limit for a currency.
func (b *Builder) LimitPerCallCents(currency string, cents int) *Builder {
	if b.policy.Grants.Limits.BudgetPerCallCents == nil {
		b.policy.Grants.Limits.BudgetPerCallCents = make(map[string]int)
	}
	b.policy.Grants.Limits.BudgetPerCallCents[currency] = cents
	return b
}

// LimitNet sets network usage limits.
func (b *Builder) LimitNet(maxCalls int, maxBytes int64) *Builder {
	b.policy.Grants.Limits.NetMaxCalls = maxCalls
	b.policy.Grants.Limits.NetMaxBytes = maxBytes
	return b
}

// LimitFS sets filesystem usage limits.
func (b *Builder) LimitFS(maxCalls int, maxBytes int64) *Builder {
	b.policy.Grants.Limits.FSMaxCalls = maxCalls
	b.policy.Grants.Limits.FSMaxBytes = maxBytes
	return b
}

// LimitToolCalls sets a per-tool call limit.
func (b *Builder) LimitToolCalls(tool string, maxCalls int) *Builder {
	if b.policy.Grants.Limits.ToolMaxCalls == nil {
		b.policy.Grants.Limits.ToolMaxCalls = make(map[string]int)
	}
	b.policy.Grants.Limits.ToolMaxCalls[tool] = maxCalls
	return b
}

// Build finalizes and returns the constructed ExecPolicy.
func (b *Builder) Build() *interfaces.ExecPolicy {
	if b.policy.Grants.Counters == nil {
		b.policy.Grants.Counters = capability.NewCounters()
	}
	return b.policy
}
