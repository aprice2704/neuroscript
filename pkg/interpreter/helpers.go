// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Slimmed down by removing or un-exporting convenience methods.
// filename: pkg/interpreter/helpers.go
// nlines: 30
// risk_rating: MEDIUM

package interpreter

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// defaultWhisperFunc is the built-in whisper implementation.
func (i *Interpreter) defaultWhisperFunc(handle, data lang.Value) {
	i.bufferManager.Write(handle.String(), data.String()+"\n")
}

// addProcedure programmatically adds a single procedure to the interpreter's registry.
func (i *Interpreter) addProcedure(proc ast.Procedure) error {
	if i.state.knownProcedures == nil {
		i.state.knownProcedures = make(map[string]*ast.Procedure)
	}
	if proc.Name() == "" {
		return errors.New("cannot add procedure with empty name")
	}
	if _, exists := i.state.knownProcedures[proc.Name()]; exists {
		return fmt.Errorf("%w: '%s'", lang.ErrProcedureExists, proc.Name())
	}
	i.state.knownProcedures[proc.Name()] = &proc
	return nil
}

// getAllVariables returns a copy of all variables in the current scope for testing.
func (i *Interpreter) getAllVariables() (map[string]lang.Value, error) {
	i.state.variablesMu.RLock()
	defer i.state.variablesMu.RUnlock()
	clone := make(map[string]lang.Value)
	for k, v := range i.state.variables {
		clone[k] = v
	}
	return clone, nil
}
