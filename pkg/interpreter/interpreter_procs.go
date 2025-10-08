// NeuroScript Version: 0.7.3
// File version: 3
// Purpose: FIX: Correctly initializes the destination procedure map if it is nil before copying.
// filename: pkg/interpreter/interpreter_procs.go
// nlines: 32
// risk_rating: HIGH

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// CopyProceduresFrom copies the procedure definitions from a source interpreter
// into the receiver. It returns an error if any procedure being copied already
// exists in the destination to prevent accidental overwrites.
func (i *Interpreter) CopyProceduresFrom(source *Interpreter) error {
	if source == nil {
		return nil
	}

	// FIX: If the destination interpreter's internal map is nil, we must
	// initialize it here. Otherwise, the call to KnownProcedures() below
	// would return a temporary map, and the copy would be lost.
	if i.state.knownProcedures == nil {
		i.state.knownProcedures = make(map[string]*ast.Procedure)
	}

	sourceProcs := source.KnownProcedures()
	destProcs := i.KnownProcedures() // Now this will return the actual state map

	for name, proc := range sourceProcs {
		if _, exists := destProcs[name]; exists {
			return lang.NewRuntimeError(lang.ErrorCodeDuplicate, fmt.Sprintf("procedure '%s' already exists in destination interpreter", name), nil)
		}
		destProcs[name] = proc
	}
	return nil
}
