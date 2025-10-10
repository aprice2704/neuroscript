// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Refactored to use the new fluent policy builder and the correct WithExecPolicy option.
// filename: pkg/api/config.go
// nlines: 26
// risk_rating: MEDIUM

package api

import (
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
)

// WithTrustedPolicy creates an interpreter option that applies a pre-configured
// execution policy suitable for running trusted configuration scripts.
func WithTrustedPolicy(allowedTools []string, grants ...capability.Capability) interpreter.InterpreterOption {
	builder := NewPolicyBuilder(ContextConfig).Allow(allowedTools...)
	for _, g := range grants {
		builder.GrantCap(g)
	}
	policy := builder.Build()

	// FIX: Use the public WithExecPolicy option to apply the policy instead of
	// trying to access the unexported field directly.
	return WithExecPolicy(policy)
}

// NewConfigInterpreter is a convenience function that creates a new interpreter
// pre-configured with a trusted policy. This is the recommended entrypoint for
// hosts that need to run setup or initialization scripts.
func NewConfigInterpreter(allowedTools []string, grants []capability.Capability, otherOpts ...Option) *Interpreter {
	// FIX: The 'otherOpts' were being appended but not passed to New().
	// This corrects the call to ensure all options, including WithCapsuleAdminRegistry, are applied.
	opts := []Option{
		WithTrustedPolicy(allowedTools, grants...),
	}
	opts = append(opts, otherOpts...)
	return New(opts...)
}
