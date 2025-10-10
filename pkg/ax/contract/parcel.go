// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: FIX: Ensures AEIOU() returns a non-nil struct and Fork() correctly copies the AEIOU context.
// filename: pkg/ax/contract/parcel.go
// nlines: 61
// risk_rating: LOW

package contract

import (
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// parcel is the concrete implementation of the RunnerParcel interface.
type parcel struct {
	aeiou   *aeiou.HostContext
	id      ax.ID
	log     interfaces.Logger
	pol     *interfaces.ExecPolicy
	globals map[string]any // The private, mutable backing store
}

var _ RunnerParcel = (*parcel)(nil)

// NewParcel creates a new parcel. Globals are defensively copied on creation.
func NewParcel(id ax.ID, pol *interfaces.ExecPolicy, log interfaces.Logger, globals map[string]any) RunnerParcel {
	g := make(map[string]any, len(globals))
	for k, v := range globals {
		g[k] = v
	}
	return &parcel{
		id:      id,
		pol:     pol,
		log:     log,
		globals: g,
	}
}

func (p *parcel) AEIOU() aeiou.HostContext {
	if p.aeiou == nil {
		return aeiou.HostContext{} // Return a zero-value struct if nil
	}
	return *p.aeiou
}
func (p *parcel) Identity() ax.ID                { return p.id }
func (p *parcel) Logger() interfaces.Logger      { return p.log }
func (p *parcel) Policy() *interfaces.ExecPolicy { return p.pol }

// Globals returns a defensive, read-only copy of the globals map.
func (p *parcel) Globals() map[string]any {
	if p.globals == nil {
		return nil
	}
	g := make(map[string]any, len(p.globals))
	for k, v := range p.globals {
		g[k] = v
	}
	return g
}

// Fork creates a shallow copy, applies mutations, and returns the new parcel.
func (p *parcel) Fork(mut func(*ParcelMut)) RunnerParcel {
	// Create a shallow copy of the parcel.
	cp := *p

	// Create a mutable struct for the mutation function.
	m := ParcelMut{AEIOU: cp.aeiou, ID: cp.id, Logger: cp.log, Policy: cp.pol}

	// Apply the mutations.
	mut(&m)

	// Update the copied parcel with the mutated values.
	cp.aeiou, cp.id, cp.log, cp.pol = m.AEIOU, m.ID, m.Logger, m.Policy
	return &cp
}
