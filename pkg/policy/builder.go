// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Provides a fluent builder for creating ExecPolicy instances. Changed default to deny-by-default.
// filename: pkg/policy/builder.go
// nlines: 119
// risk_rating: MEDIUM

package policy

import (
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
)

// Builder is a fluent API for constructing ExecPolicy objects.
type Builder struct {
	policy *ExecPolicy
}

// NewBuilder creates a new builder for an ExecPolicy. By default, the policy
// is configured to "deny-by-default".
func NewBuilder(context ExecContext) *Builder {
	p := &ExecPolicy{
		Context: context,
		// CRITICAL: Initialize with a non-nil, empty slice. This activates the
		// "deny-by-default" logic in the disallowed() function.
		Allow: []string{},
		Deny:  []string{},
		Grants: capability.GrantSet{
			Grants:   []capability.Capability{},
			Limits:   capability.Limits{},
			Counters: capability.NewCounters(),
		},
	}
	return &Builder{policy: p}
}

// Allow sets the tool allow list. If this method is never called, all tools
// are allowed unless explicitly denied. If called (even with no args),
// the policy switches to "deny by default".
func (b *Builder) Allow(tools ...string) *Builder {
	if b.policy.Allow == nil {
		b.policy.Allow = []string{}
	}
	b.policy.MergeAllows(tools...)
	return b
}

// Deny adds tools to the deny list. Deny rules always take precedence.
func (b *Builder) Deny(tools ...string) *Builder {
	b.policy.MergeDenies(tools...)
	return b
}

// Grant adds a capability grant by parsing a capability string.
// It will panic if the string is invalid. Use GrantCap for pre-validated caps.
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
func (b *Builder) Build() *ExecPolicy {
	// Ensure counters are non-nil on the final product.
	if b.policy.Grants.Counters == nil {
		b.policy.Grants.Counters = capability.NewCounters()
	}
	return b.policy
}
