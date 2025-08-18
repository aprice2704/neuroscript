// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Provides high-level helpers for creating interpreters configured to run trusted setup scripts.
// filename: pkg/api/config.go
// nlines: 35
// risk_rating: MEDIUM

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/runtime"
)

// WithTrustedPolicy creates an interpreter option that applies a pre-configured
// execution policy suitable for running trusted configuration scripts.
// It sets the context to 'config', allowing privileged tools, and applies the
// specified tool allow-list and capability grants.
func WithTrustedPolicy(allowedTools []string, grants ...capability.Capability) interpreter.InterpreterOption {
	return func(i *interpreter.Interpreter) {
		policy := &runtime.ExecPolicy{
			Context: runtime.ContextConfig,
			Allow:   allowedTools,
			Deny:    []string{}, // Start with no denials
			Grants:  capability.NewGrantSet(grants, capability.Limits{}),
		}
		i.ExecPolicy = policy
	}
}

// NewConfigInterpreter is a convenience function that creates a new interpreter
// pre-configured with a trusted policy. This is the recommended entrypoint for
// hosts that need to run setup or initialization scripts.
func NewConfigInterpreter(allowedTools []string, grants []capability.Capability, otherOpts ...Option) *Interpreter {
	opts := []Option{
		WithTrustedPolicy(allowedTools, grants...),
	}
	opts = append(opts, otherOpts...)
	return New(opts...)
}
