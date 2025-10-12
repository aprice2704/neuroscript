// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Corrected ruleMatches to properly handle a universal "*" wildcard.
// filename: pkg/policygate/policygate.go
// nlines: 75
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
		return lang.NewRuntimeError(lang.ErrorCodePolicy, "action denied: no execution policy is set", policy.ErrPolicy)
	}

	// Deny list is checked first and is an absolute override.
	for _, denied := range p.Deny {
		if ruleMatches(denied, cap) {
			return lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("action denied by policy rule: %s", denied), policy.ErrPolicy)
		}
	}

	// If context is not privileged, check for trust requirements.
	if p.Context != policy.ContextConfig {
		// This is a placeholder for a more robust trust check. For now, we assume
		// the tool metadata would be passed in. This check is handled in the interpreter for now.
	}

	// Check against the allow list.
	allowedByName := false
	for _, allowed := range p.Allow {
		if ruleMatches(allowed, cap) {
			allowedByName = true
			break
		}
	}

	if !allowedByName {
		return lang.NewRuntimeError(lang.ErrorCodePolicy, "action not on allow list", policy.ErrPolicy)
	}

	// Finally, check if the grants satisfy the capability.
	if p.Grants.Check(cap) {
		return nil
	}

	return lang.NewRuntimeError(lang.ErrorCodePolicy, "action denied: missing required grants", policy.ErrCapability)
}

// ruleMatches checks if a policy rule (e.g., "tool.fs.*" or "*") matches a capability.
func ruleMatches(rule string, cap capability.Capability) bool {
	if rule == "*" {
		return true
	}
	// A rule can match against any verb in the capability.
	for _, verb := range cap.Verbs {
		capString := fmt.Sprintf("%s.%s", cap.Resource, verb)
		if strings.HasSuffix(rule, ".*") {
			prefix := strings.TrimSuffix(rule, ".*")
			if strings.HasPrefix(capString, prefix) {
				return true
			}
		}
		if rule == capString {
			return true
		}
	}
	return false
}
