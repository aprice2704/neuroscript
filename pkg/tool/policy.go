// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Centralizes all tool execution policy checks.
// filename: pkg/tool/policy.go
// nlines: 50
// risk_rating: HIGH

package tool

import (
	"fmt"

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

	// 2. Allow/Deny and Grant Checks via the policygate.
	// If a tool requires no specific capabilities, we still check its name against the allow/deny list.
	if len(tool.RequiredCaps) == 0 {
		nameOnlyCap := capability.Capability{Resource: "tool", Verbs: []string{"exec"}, Scopes: []string{string(tool.FullName)}}
		return policygate.Check(rt, nameOnlyCap)
	}

	// If the tool requires capabilities, check each one.
	for _, cap := range tool.RequiredCaps {
		if err := policygate.Check(rt, cap); err != nil {
			// Return the specific error from the policy gate.
			return lang.NewRuntimeError(lang.ErrorCodePolicy, fmt.Sprintf("permission denied: tool '%s' requires capabilities that are not granted", tool.FullName), err)
		}
	}

	return nil
}
