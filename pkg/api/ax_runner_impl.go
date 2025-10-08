// NeuroScript Version: 0.7.4
// File version: 8
// Purpose: FIX: Implemented the Clone() method for the ax.Runner interface. ADD: Implemented ax.CloneCap and the Lookup method for ax.Tools.
// filename: pkg/api/ax_runner_impl.go
// nlines: 99
// risk_rating: HIGH

package api

import (
	"errors"

	"github.com/aprice2704/neuroscript/pkg/ax"
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
	env  *axRunEnv
	host *hostRuntime
	itp  *Interpreter
}

// axTools adapts the internal ToolRegistry to the ax.Tools interface.
type axTools struct{ itp *Interpreter }

func (t *axTools) Register(name string, impl any) error {
	if ti, ok := impl.(ToolImplementation); ok {
		_, err := t.itp.ToolRegistry().RegisterTool(ti)
		return err
	}
	// Fallback for other registration types if necessary in the future.
	return errors.New("unsupported tool implementation type for ax registration")
}

func (t *axTools) Lookup(name string) (any, bool) {
	return t.itp.ToolRegistry().GetTool(types.FullName(name))
}

// Compile-time checks
var _ ax.Runner = (*axRunner)(nil)
var _ ax.CloneCap = (*axRunner)(nil)
var _ ax.IdentityCap = (*hostRuntime)(nil)
var _ ax.Tools = (*axTools)(nil)

func (r *axRunner) Env() ax.RunEnv { return r.env }

// --- RunnerCore Implementation ---

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

func (r *axRunner) Identity() ax.ID { return r.host.id }
func (r *axRunner) Tools() ax.Tools { return &axTools{itp: r.itp} }

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
	// The new host runtime points to the new cloned interpreter's internal instance
	newHost := &hostRuntime{
		Runtime: clonedItp.internal,
		id:      r.host.id,
	}
	clonedItp.SetRuntime(newHost)

	return &axRunner{
		env:  r.env, // Env is shared
		host: newHost,
		itp:  clonedItp,
	}
}
