// NeuroScript Version: 0.7.2
// File version: 3
// Purpose: Refactored to use the new fluent policy builder.
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

	return func(i *interpreter.Interpreter) {
		i.ExecPolicy = policy
	}
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
