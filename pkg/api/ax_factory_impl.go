// NeuroScript Version: 0.7.4
// File version: 7
// Purpose: FIX: Use CopyFunctionsFrom for user runners and create the root with a trusted context.
// filename: pkg/api/ax_factory_impl.go
// nlines: 67
// risk_rating: HIGH

package api

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// axFactory is the concrete implementation of the ax.RunnerFactory.
type axFactory struct {
	env  *axRunEnv
	root *Interpreter
}

// Compile-time checks to ensure interface satisfaction.
var _ ax.RunnerFactory = (*axFactory)(nil)
var _ ax.EnvCap = (*axFactory)(nil)

// NewAXFactory creates a new factory.
func NewAXFactory(ctx context.Context, rootOpts ax.RunnerOpts, baseRt Runtime, id ax.ID) (*axFactory, error) {
	// FIX: The root interpreter for the factory must have a trusted context
	// to perform administrative actions that populate its state.
	configPolicy := policy.NewBuilder(policy.ContextConfig).Allow("*").Build()
	root := New(WithExecPolicy(configPolicy))
	host := &hostRuntime{Runtime: baseRt, id: id}

	root.SetRuntime(host)

	if rootOpts.SandboxDir != "" {
		root.SetSandboxDir(rootOpts.SandboxDir)
	}
	return &axFactory{env: &axRunEnv{root: root}, root: root}, nil
}

// Env satisfies the ax.EnvCap interface.
func (f *axFactory) Env() ax.RunEnv { return f.env }

// NewRunner creates a new ax.Runner.
func (f *axFactory) NewRunner(ctx context.Context, mode ax.RunnerMode, opts ax.RunnerOpts) (ax.Runner, error) {
	// All runners start as a fresh, clean interpreter.
	itp := New()

	id := f.root.Identity()
	host := &hostRuntime{Runtime: f.root.internal, id: id}
	itp.SetRuntime(host)

	if opts.SandboxDir != "" {
		itp.SetSandboxDir(opts.SandboxDir)
	}

	r := &axRunner{env: f.env, host: host, itp: itp}

	// FIX: If it's a user runner, surgically copy only the function definitions.
	if mode == ax.RunnerUser {
		if err := r.itp.CopyFunctionsFrom(f.root); err != nil {
			return nil, fmt.Errorf("failed to copy function definitions: %w", err)
		}
	} else if mode == ax.RunnerConfig {
		// Ensure config runners also have a trusted context.
		configPolicy := policy.NewBuilder(policy.ContextConfig).Allow("*").Build()
		itp.internal.ExecPolicy = configPolicy
	}

	return r, nil
}
