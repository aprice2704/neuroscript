// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: FIX: Imports reader interfaces from pkg/interfaces to finally break the import cycle with pkg/api.
// filename: pkg/ax/contract/contracts.go
// nlines: 45
// risk_rating: LOW

package contract

import (
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// RunnerParcel is the "who/where/rights" bundle for an execution context.
type RunnerParcel interface {
	AEIOU() aeiou.HostContext
	Identity() ax.ID
	Logger() interfaces.Logger
	Policy() *interfaces.ExecPolicy
	Globals() map[string]any

	Fork(mut func(*ParcelMut)) RunnerParcel
}

// ParcelMut provides a mutable view of a parcel's fields for use within Fork().
type ParcelMut struct {
	AEIOU  *aeiou.HostContext
	ID     ax.ID
	Logger interfaces.Logger
	Policy *interfaces.ExecPolicy
}

// ParcelProvider is an interface for objects that carry a RunnerParcel.
type ParcelProvider interface {
	GetParcel() RunnerParcel
	SetParcel(RunnerParcel)
}

// SharedCatalogs provides a single facade for accessing all long-lived resources.
type SharedCatalogs interface {
	Accounts() interfaces.AccountReader
	AgentModels() interfaces.AgentModelReader
	Tools() ax.Tools
	Capsules() *capsule.Registry
}
