// NeuroScript Version: 0.8.0
// File version: 51.0.0
// Purpose: Updated to use the new fork() method for creating sandboxed procedure environments.
// filename: pkg/interpreter/procedures.go
// nlines: 50
// risk_rating: HIGH

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

const maxCallDepth = 500 // Prevents stack overflow

// runProcedure executes a defined procedure with the given arguments.
// This is the internal implementation.
func (i *Interpreter) runProcedure(procName string, args ...lang.Value) (lang.Value, error) {
	proc, exists := i.KnownProcedures()[procName]
	if !exists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeProcNotFound, fmt.Sprintf("procedure '%s' not found", procName), lang.ErrProcedureNotFound)
	}

	if len(i.state.stackFrames) >= maxCallDepth {
		return nil, lang.NewRuntimeError(
			lang.ErrorCodeResourceExhaustion,
			fmt.Sprintf("maximum call depth of %d exceeded", maxCallDepth),
			lang.ErrMaxCallDepthExceeded,
		)
	}

	procInterpreter := i.fork() // Use fork() to create the sandboxed environment
	procInterpreter.state.currentProcName = procName
	procInterpreter.state.stackFrames = append(i.state.stackFrames, procName)

	if len(proc.ErrorHandlers) > 0 {
		procInterpreter.state.errorHandlerStack = append(procInterpreter.state.errorHandlerStack, proc.ErrorHandlers)
	}

	for idx, paramName := range proc.RequiredParams {
		if idx < len(args) {
			procInterpreter.SetVariable(paramName, args[idx])
		} else {
			procInterpreter.SetVariable(paramName, &lang.NilValue{})
		}
	}

	result, _, _, err := procInterpreter.executeSteps(proc.Steps, false, nil)

	return result, err
}
