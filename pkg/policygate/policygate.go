// NeuroScript Version: 0.8.0
// File version: 13
// Purpose: Corrects a fundamental bug by removing the Allow/Deny list checks. Capabilities are checked against Grants only.
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
// This check is for programmatic Go-level calls (e.g., admin functions),
// not for script-level tool execution.
// It should only check for capability grants, not the tool allow/deny list.
func Check(rt Runtime, cap capability.Capability) error {
	p := rt.GetExecPolicy()
	if p == nil {
		return lang.NewRuntimeError(lang.ErrorCodePolicy, "action denied: no execution policy is set", policy.ErrPolicy)
	}

	// If context is not privileged, check for trust requirements.
	// This is a placeholder; trust is primarily handled at the tool-call level.
	if p.Context != policy.ContextConfig {
		// This check may need to be more robust, e.g., checking if the
		// capability itself is considered "trusted".
	}

	// FIX: Removed the incorrect checks against p.Allow and p.Deny.
	// The Allow/Deny lists are for *tool names* (checked by policy.CanCall),
	// not for programmatic capability checks.
	// This function should *only* check the Grants list.

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

	// This is now the only failure path, as intended.
	errMsg := fmt.Sprintf("action denied: missing required grants. Required: [%s], Had: [%s]", cap.String(), hadStr)
	return lang.NewRuntimeError(lang.ErrorCodePolicy, errMsg, policy.ErrCapability)
}

// RuleMatches is no longer used by Check. It is preserved only in case
// external packages were depending on it, but its logic is flawed
// for matching capability strings.
func RuleMatches(rule string, cap capability.Capability) bool {
	// A capability can have multiple scopes (e.g., tool name and effects).
	// If the rule matches any of them, the capability is considered matched.
	for _, scope := range cap.Scopes {
		// This logic is suspect: it matches the scope (e.g. "*") against the rule (e.g. "model:admin:*")
		if policy.PatMatch(scope, rule) {
			return true
		}
	}
	return false
}
