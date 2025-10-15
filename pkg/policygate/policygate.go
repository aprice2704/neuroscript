// NeuroScript Version: 0.8.0
// File version: 11
// Purpose: Restored the exported `RuleMatches` function as a passthrough to `policy.PatMatch` to fix a build error in an external package. Centralized matching logic within it.
// filename: pkg/policygate/policygate.go
// nlines: 97
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
		if RuleMatches(denied, cap) {
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
		if RuleMatches(allowed, cap) {
			allowedByName = true
			break
		}
	}

	if !allowedByName {
		errMsg := fmt.Sprintf("action requiring capability [%s] was not found in the policy's allow list. Allow list contains: [%s]", cap.String(), strings.Join(p.Allow, ", "))
		return lang.NewRuntimeError(lang.ErrorCodePolicy, errMsg, policy.ErrPolicy)
	}

	// Finally, check if the grants satisfy the capability.
	if p.Grants.Check(cap) {
		return nil
	}

	// Create a detailed error message showing what was required vs. what was possessed.
	var hadGrants []string
	for _, grant := range p.Grants.Grants {
		hadGrants = append(hadGrants, grant.String())
	}
	hadStr := strings.Join(hadGrants, ", ")
	if hadStr == "" {
		hadStr = "none"
	}

	errMsg := fmt.Sprintf("action denied: missing required grants. Required: [%s], Had: [%s]", cap.String(), hadStr)
	return lang.NewRuntimeError(lang.ErrorCodePolicy, errMsg, policy.ErrCapability)
}

// RuleMatches checks if a policy rule (e.g., "tool.fs.*") matches any scope
// within a capability, using the case-insensitive logic from the policy package.
// This function is preserved for external callers but delegates its logic.
func RuleMatches(rule string, cap capability.Capability) bool {
	// A capability can have multiple scopes (e.g., tool name and effects).
	// If the rule matches any of them, the capability is considered matched.
	for _, scope := range cap.Scopes {
		if policy.PatMatch(scope, rule) {
			return true
		}
	}
	return false
}
