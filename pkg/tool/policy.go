// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Centralizes all tool execution policy checks with a correct two-phase validation and adds limit enforcement.
// filename: pkg/tool/policy.go
// nlines: 83
// risk_rating: HIGH

package tool

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/policygate"
)

// CanCall performs a full policy check for a given tool against a runtime's policy.
func CanCall(rt policygate.Runtime, tool ToolImplementation) error {
	p := rt.GetExecPolicy()
	if p == nil {
		return lang.NewRuntimeError(lang.ErrorCodePolicy, "action denied: no execution policy is set", policy.ErrPolicy)
	}

	// 1. Trust Check: This is the first and most important check.
	if tool.RequiresTrust && p.Context != policy.ContextConfig {
		return lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("tool '%s' requires a privileged context", tool.FullName), policy.ErrTrust)
	}

	// 2. Allow List Check: Verify the tool's NAME is on the allow list.
	// We create a synthetic capability representing the tool call itself to check against the allow/deny lists.
	toolNameCap := capability.Capability{Resource: "tool", Verbs: []string{"exec"}, Scopes: []string{string(tool.FullName)}}

	// Check deny list first
	for _, denied := range p.Deny {
		if policygate.RuleMatches(denied, toolNameCap) {
			return lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("tool '%s' denied by policy rule: %s", tool.FullName, denied), policy.ErrPolicy)
		}
	}

	// Check allow list
	allowedByName := false
	for _, allowed := range p.Allow {
		if policygate.RuleMatches(allowed, toolNameCap) {
			allowedByName = true
			break
		}
	}

	if !allowedByName {
		errMsg := fmt.Sprintf("tool '%s' was not found in the policy's allow list. Allow list contains: [%s]", tool.FullName, strings.Join(p.Allow, ", "))
		return lang.NewRuntimeError(lang.ErrorCodePolicy, errMsg, policy.ErrPolicy)
	}

	// 3. Grant Check: If the tool requires capabilities, check them against the grants.
	for _, cap := range tool.RequiredCaps {
		if !p.Grants.Check(cap) {
			var hadGrants []string
			for _, grant := range p.Grants.Grants {
				hadGrants = append(hadGrants, grant.String())
			}
			hadStr := strings.Join(hadGrants, ", ")
			if hadStr == "" {
				hadStr = "none"
			}
			errMsg := fmt.Sprintf("permission denied for tool '%s': missing required grants. Required: [%s], Had: [%s]", tool.FullName, cap.String(), hadStr)
			return lang.NewRuntimeError(lang.ErrorCodePolicy, errMsg, policy.ErrCapability)
		}
	}

	// 4. Limit Check: Enforce tool call limits.
	toolName := string(tool.FullName)
	if max, ok := p.Grants.Limits.ToolMaxCalls[toolName]; ok {
		if p.Grants.Counters == nil {
			p.Grants.Counters = capability.NewCounters()
		}
		// Increment must happen before the check.
		count := p.Grants.Counters.ToolCalls[toolName] + 1
		p.Grants.Counters.ToolCalls[toolName] = count
		if count > max {
			errMsg := fmt.Sprintf("tool '%s' exceeded its call limit of %d", tool.FullName, max)
			return lang.NewRuntimeError(lang.ErrorCodePolicy, errMsg, policy.ErrPolicy)
		}
	}

	return nil
}
