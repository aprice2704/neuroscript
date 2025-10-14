// NeuroScript Version: 0.8.0
// File version: 11
// Purpose: Updated to use the re-exported HostContextBuilder and options for a clean, canonical implementation.
// filename: pkg/api/config.go
// nlines: 42
// risk_rating: MEDIUM

package api

import (
	"io"
	"os"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// Option is an alias for the internal interpreter's option type.
type Option = interpreter.InterpreterOption

// WithTrustedPolicy creates an interpreter option that applies a pre-configured
// execution policy suitable for running trusted configuration scripts.
func WithTrustedPolicy(allowedTools []string, grants ...capability.Capability) Option {
	builder := NewPolicyBuilder(ContextConfig).Allow(allowedTools...)
	for _, g := range grants {
		builder.GrantCap(g)
	}
	policy := builder.Build()
	return WithExecPolicy(policy)
}

// NewConfigInterpreter is a convenience function that creates a new interpreter
// pre-configured with a trusted policy. It ensures a minimal HostContext is always present.
func NewConfigInterpreter(allowedTools []string, grants []capability.Capability, otherOpts ...Option) *Interpreter {
	// Start with a default, minimal HostContext using the canonical builder.
	defaultHC, err := NewHostContextBuilder().
		WithLogger(logging.NewNoOpLogger()).
		WithStdout(io.Discard).
		WithStdin(os.Stdin).
		WithStderr(io.Discard).
		Build()
	if err != nil {
		// This should not happen with default values.
		panic("failed to build default HostContext in NewConfigInterpreter: " + err.Error())
	}

	// The default context is the first option. If the user passes their own
	// WithHostContext in otherOpts, it will be applied later and override this one.
	opts := []Option{
		WithHostContext(defaultHC),
		WithTrustedPolicy(allowedTools, grants...),
	}
	opts = append(opts, otherOpts...)

	return New(opts...)
}
