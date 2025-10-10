// NeuroScript Version: 0.8.0
// File version: 18
// Purpose: FIX: Removed duplicate axTools adapter and corrected ParcelProvider method signature to resolve compiler errors.
// filename: pkg/api/ax_runner_impl.go
// nlines: 130
// risk_rating: HIGH

package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/ax/contract"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// hostRuntime augments the existing tool.Runtime with the ax.IdentityCap.
type hostRuntime struct {
	Runtime
	id ax.ID
}

func (h *hostRuntime) Identity() ax.ID { return h.id }

// axRunner is the concrete implementation of the ax.Runner interface.
type axRunner struct {
	parcel   contract.RunnerParcel
	catalogs contract.SharedCatalogs
	itp      *Interpreter
}

// Compile-time checks
var _ ax.Runner = (*axRunner)(nil)
var _ ax.CloneCap = (*axRunner)(nil)
var _ ax.IdentityCap = (*hostRuntime)(nil)
var _ contract.ParcelProvider = (*axRunner)(nil)

// --- ParcelProvider Implementation ---
func (r *axRunner) GetParcel() contract.RunnerParcel  { return r.parcel }
func (r *axRunner) SetParcel(p contract.RunnerParcel) { r.parcel = p }

// --- RunnerCore Implementation ---

func (r *axRunner) LoadScript(script []byte) error {
	tree, err := Parse(script, ParseSkipComments)
	if err != nil {
		return fmt.Errorf("ax load script: failed to parse source: %w", err)
	}
	if err := r.itp.AppendScript(tree); err != nil {
		return fmt.Errorf("ax load script: failed to load definitions: %w", err)
	}
	return nil
}

func (r *axRunner) Execute() (any, error) { return r.itp.Execute() }

func (r *axRunner) Run(proc string, args ...any) (any, error) {
	// RunProcedure is used because it handles the required internal cloning.
	return RunProcedure(context.Background(), r.itp, proc, args...)
}

func (r *axRunner) EmitEvent(name, src string, payload any) {
	pv, _ := lang.Wrap(payload)
	r.itp.EmitEvent(name, src, pv)
}

// --- Capability Interfaces ---

func (r *axRunner) Identity() ax.ID { return r.parcel.Identity() }
func (r *axRunner) Tools() ax.Tools { return r.catalogs.Tools() }

func (r *axRunner) CopyFunctionsFrom(src ax.RunnerCore) error {
	other, ok := src.(*axRunner)
	if !ok || other == nil {
		return errors.New("CopyFunctionsFrom: incompatible src runner type")
	}
	return r.itp.CopyFunctionsFrom(other.itp)
}

// Clone implements the ax.CloneCap interface.
func (r *axRunner) Clone() ax.Runner {
	clonedItp := r.itp.Clone()

	// The parcel is copied by reference. The engine handles sandboxing exec state.
	return &axRunner{
		parcel:   r.parcel,
		catalogs: r.catalogs,
		itp:      clonedItp,
	}
}
