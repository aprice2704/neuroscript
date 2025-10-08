// NeuroScript Version: 0.7.4
// File version: 1
// Purpose: Contains helper and ax-support methods for the Interpreter facade.
// filename: pkg/api/interpreter_helpers.go
// nlines: 60
// risk_rating: LOW

package api

import (
	"errors"

	"github.com/aprice2704/neuroscript/pkg/ax"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// HasEmitFunc returns true if a custom emit handler has been set.
func (i *Interpreter) HasEmitFunc() bool {
	return i.internal.HasEmitFunc()
}

// Unwrap converts a NeuroScript api.Value back into a standard Go `any` type.
func Unwrap(v Value) (any, error) {
	if val, ok := v.(lang.Value); ok {
		return lang.Unwrap(val), nil
	}
	return v, nil
}

// ParseLoopControl is deprecated.
func ParseLoopControl(output string) (*LoopControl, error) {
	return nil, errors.New("ParseLoopControl is deprecated; use the AEIOU v3 LoopController")
}

// GetVariable retrieves a variable from the interpreter's current state.
func (i *Interpreter) GetVariable(name string) (Value, bool) {
	val, exists := i.internal.GetVariable(name)
	return val, exists
}

// CopyFunctionsFrom copies only function definitions from a source interpreter.
func (i *Interpreter) CopyFunctionsFrom(source *Interpreter) error {
	if source == nil || source.internal == nil {
		return errors.New("source interpreter cannot be nil")
	}
	return i.internal.CopyProceduresFrom(source.internal)
}

// Identity inspects the interpreter's runtime to extract the actor's identity
// for the ax factory.
func (i *Interpreter) Identity() ax.ID {
	if i.runtime == nil {
		return nil
	}
	if hr, ok := i.runtime.(*hostRuntime); ok {
		return hr.id
	}
	return nil
}
