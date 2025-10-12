// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Centralizes all execution policy and capability checks.
// filename: pkg/policygate/policygate.go
// nlines: 70
// risk_rating: HIGH

package policygate

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// Runtime defines the interface the policy gate needs to inspect the interpreter's state.
type Runtime interface {
	GetExecPolicy() *policy.ExecPolicy
}

// Check verifies if a given capability is allowed by the runtime's execution policy.
func Check(rt Runtime, cap capability.Capability) error {
	p := rt.GetExecPolicy()
	if p == nil {
		// No policy means default deny.
		return lang.NewRuntimeError(lang.ErrorCodePolicy, "action denied: no execution policy is set", policy.ErrTrust)
	}

	// 1. Check if the entire context is trusted. If so, allow.
	if p.Context == policy.ContextConfig {
		return nil // Config context allows all actions.
	}

	// 2. Check Deny list - these are explicit hard-stops.
	for _, denied := range p.Deny {
		if ruleMatches(denied, cap) {
			return lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("action denied by policy: %s %s::%s", cap.Verb, cap.Resource, cap.Scope), policy.ErrTrust)
		}
	}

	// 3. Check Allow list - if anything matches here, it's an immediate pass.
	for _, allowed := range p.Allow {
		if ruleMatches(allowed, cap) {
			return nil
		}
	}

	// 4. Check specific grants.
	if p.Grants.Check(cap) {
		return nil
	}

	// 5. If we fall through, the default is to deny.
	return lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("action denied by policy: %s %s::%s", cap.Verb, cap.Resource, cap.Scope), policy.ErrTrust)
}

// ruleMatches checks if a policy rule (e.g., "tool.fs.*" or "model.admin") matches a capability.
func ruleMatches(rule string, cap capability.Capability) bool {
	capString := fmt.Sprintf("%s.%s", cap.Resource, cap.Verb)
	if strings.HasSuffix(rule, ".*") {
		prefix := strings.TrimSuffix(rule, ".*")
		return strings.HasPrefix(capString, prefix)
	}
	return rule == capString
}
