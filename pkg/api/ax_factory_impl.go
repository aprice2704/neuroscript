// NeuroScript Version: 0.8.0
// File version: 13
// Purpose: FIX: Rewrote factory to use the new Parcel/Catalogs model, resolving all compiler errors.
// filename: pkg/api/ax_factory_impl.go
// nlines: 68
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
	root     *Interpreter // The root interpreter is still needed for its function definitions
}

// Compile-time checks to ensure interface satisfaction.
var _ ax.RunnerFactory = (*axFactory)(nil)

// NewAXFactory creates a new factory.
func NewAXFactory(ctx context.Context, rootOpts ax.RunnerOpts, baseRt Runtime, id ax.ID) (*axFactory, error) {
	// 1. Create the root interpreter, which holds the shared function library.
	configPolicy := policy.NewBuilder(policy.ContextConfig).Allow("*").Build()
	root := New(WithExecPolicy(configPolicy))
	root.SetRuntime(baseRt) // baseRt is used for boot-time tool calls.

	// 2. Create the two core components from the new model.
	parcel := contract.NewParcel(id, configPolicy, root.InternalRuntime().GetLogger(), nil)
	catalogs := NewSharedCatalogs(root)

	return &axFactory{
		parcel:   parcel,
		catalogs: catalogs,
		root:     root,
	}, nil
}

// Env is deprecated in the parcel model but kept for compatibility during transition.
// func (f *axFactory) Env() ax.RunEnv {
// 	// This can be refactored or removed once all consumers use parcel/catalogs.
// 	return &axRunEnv{root: f.root}
// }

// NewRunner creates a new ax.Runner.
func (f *axFactory) NewRunner(ctx context.Context, mode ax.RunnerMode, opts ax.RunnerOpts) (ax.Runner, error) {
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
	}

	return r, nil
}
