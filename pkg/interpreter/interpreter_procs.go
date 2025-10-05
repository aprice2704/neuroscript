// NeuroScript Version: 0.7.3
// File version: 1
// Purpose: Implements the internal logic for copying procedure definitions between interpreters.
// filename: pkg/interpreter/interpreter_procs.go
// nlines: 25
// risk_rating: MEDIUM

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// CopyProceduresFrom copies the procedure definitions from a source interpreter
// into the receiver. It returns an error if any procedure being copied already
// exists in the destination to prevent accidental overwrites.
func (i *Interpreter) CopyProceduresFrom(source *Interpreter) error {
	if source == nil {
		return nil
	}

	sourceProcs := source.KnownProcedures()
	destProcs := i.KnownProcedures()

	for name, proc := range sourceProcs {
		if _, exists := destProcs[name]; exists {
			return lang.NewRuntimeError(lang.ErrorCodeDuplicate, fmt.Sprintf("procedure '%s' already exists in destination interpreter", name), nil)
		}
		destProcs[name] = proc
	}
	return nil
}
