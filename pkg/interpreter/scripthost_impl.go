// NeuroScript Version: 0.8.0
// File version: 84
// Purpose: Corrected AddProcedure signature to match the ScriptHost interface, resolving a compiler error.
// filename: pkg/interpreter/scripthost_impl.go
// nlines: 36
// risk_rating: LOW

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// AddProcedure programmatically adds a single procedure to the interpreter's registry.
// It is a public method to satisfy the scripthost.ScriptHost interface.
func (i *Interpreter) AddProcedure(proc ast.Procedure) error {
	i.state.variablesMu.Lock()
	defer i.state.variablesMu.Unlock()
	if _, exists := i.state.knownProcedures[proc.Name()]; exists {
		return fmt.Errorf("%w: '%s'", lang.ErrProcedureExists, proc.Name())
	}
	i.state.knownProcedures[proc.Name()] = &proc
	return nil
}

// RegisterEvent programmatically registers an event handler declaration.
// It is a public method to satisfy the scripthost.ScriptHost interface.
func (i *Interpreter) RegisterEvent(decl *ast.OnEventDecl) error {
	// The event manager handles its own locking.
	return i.eventManager.register(decl, i)
}

// NOTE: The `KnownProcedures` method that was previously in this file was a
// duplicate of the one in `api.go` and has been removed to resolve the
// compilation error. The canonical implementation remains in `api.go`.
