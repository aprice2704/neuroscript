// NeuroScript Version: 0.8.0
// File version: 15
// Purpose: FIX: Implemented ListTools and GetTool on the axTools struct to fully satisfy the ax.Tools interface, resolving the final compiler errors.
// filename: pkg/api/ax_runner_impl.go
// nlines: 130
// risk_rating: HIGH

package api

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/ax/contract"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
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

// axTools adapts the internal ToolRegistry to the ax.Tools interface.
type axTools struct{ itp *Interpreter }

func (t *axTools) Register(name string, impl any) error {
	if ti, ok := impl.(ToolImplementation); ok {
		_, err := t.itp.internal.ToolRegistry().RegisterTool(ti)
		return err
	}
	// Fallback for other registration types if necessary in the future.
	return errors.New("unsupported tool implementation type for ax registration")
}

func (t *axTools) Lookup(name string) (any, bool) {
	return t.itp.internal.ToolRegistry().GetTool(types.FullName(name))
}

func (t *axTools) ListTools() []any {
	tools := t.itp.internal.ToolRegistry().ListTools()
	anys := make([]any, len(tools))
	for i, tool := range tools {
		anys[i] = tool
	}
	return anys
}

func (t *axTools) GetTool(name string) (any, bool) {
	return t.itp.internal.ToolRegistry().GetTool(types.FullName(name))
}

// Compile-time checks
var _ ax.Runner = (*axRunner)(nil)
var _ ax.CloneCap = (*axRunner)(nil)
var _ ax.IdentityCap = (*hostRuntime)(nil)
var _ ax.Tools = (*axTools)(nil)
var _ contract.ParcelProvider = (*axRunner)(nil)

// func (r *axRunner) Env() ax.RunEnv { return &axRunEnv{root: r.itp} }

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
	vs := make([]lang.Value, len(args))
	for i, a := range args {
		v, err := lang.Wrap(a)
		if err != nil {
			return nil, err
		}
		vs[i] = v
	}
	return r.itp.Run(proc, vs...)
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
