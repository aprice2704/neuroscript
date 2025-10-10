// NeuroScript Version: 0.8.0
// File version: 15
// Purpose: FIX: Passes the internal interpreter to the hostRuntime to satisfy the tool.Runtime interface.
// filename: pkg/api/ax_factory_impl.go
// nlines: 75
// risk_rating: HIGH

package api

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/ax/contract"
	"github.com/aprice2704/neuroscript/pkg/policy"
)

// axFactory is the concrete implementation of the ax.RunnerFactory.
type axFactory struct {
	parcel   contract.RunnerParcel
	catalogs contract.SharedCatalogs
	root     *Interpreter // The root interpreter holds the shared function library.
}

// Compile-time checks to ensure interface satisfaction.
var _ ax.RunnerFactory = (*axFactory)(nil)

// NewAXFactory creates a new factory.
func NewAXFactory(ctx context.Context, rootOpts ax.RunnerOpts, baseRt Runtime, id ax.ID) (*axFactory, error) {
	// 1. Create the root interpreter, which holds the shared function library
	// and provides access to the underlying stores.
	configPolicy := policy.NewBuilder(policy.ContextConfig).Allow("*").Build()
	root := New(WithExecPolicy(configPolicy))

	// The host runtime must be set on the root interpreter so that boot-time
	// tool calls have the correct context.
	// FIX: Pass the internal interpreter, which implements the Runtime interface.
	hostRT := &hostRuntime{Runtime: root.internal, id: id}
	root.SetRuntime(hostRT)

	// 2. Create the two core components from the new model.
	// The parcel carries the "who/where/rights" context.
	// The catalogs provide a facade to the root interpreter's shared stores.
	parcel := contract.NewParcel(id, configPolicy, root.internal.GetLogger(), nil)
	catalogs := NewSharedCatalogs(root)

	return &axFactory{
		parcel:   parcel,
		catalogs: catalogs,
		root:     root,
	}, nil
}

// NewRunner creates a new ax.Runner.
func (f *axFactory) NewRunner(ctx context.Context, mode ax.RunnerMode, opts ax.RunnerOpts) (ax.Runner, error) {
	// Create a fresh, clean interpreter for the new runner.
	itp := New()

	// 1. Every runner gets a pointer to the shared catalogs.
	// 2. Every runner gets the factory's base parcel by reference.
	// The caller can then Fork() it if they need a modified context.
	r := &axRunner{
		parcel:   f.parcel,
		catalogs: f.catalogs,
		itp:      itp,
	}

	// User runners inherit function definitions from the root interpreter.
	if mode == ax.RunnerUser {
		if err := r.itp.CopyFunctionsFrom(f.root); err != nil {
			return nil, fmt.Errorf("failed to copy function definitions: %w", err)
		}
		// User runners get a restrictive, deny-by-default policy.
		userPolicy := policy.NewBuilder(policy.ContextUser).Build()
		r.parcel = r.parcel.Fork(func(m *contract.ParcelMut) {
			m.Policy = userPolicy
		})
	}

	// The runner's interpreter needs its own hostRuntime that references itself.
	// FIX: Pass the internal interpreter, which implements the Runtime interface.
	runnerHostRT := &hostRuntime{Runtime: itp.internal, id: f.parcel.Identity()}
	r.itp.SetRuntime(runnerHostRT)

	return r, nil
}
