// NeuroScript Version: 0.5.2
// File version: 48
// Purpose: Adds high-visibility debugging to trace the return value as it exits the sandboxed procedure execution.
// filename: pkg/interpreter/interpreter_procedures.go
// nlines: 120
// risk_rating: HIGH

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

const maxCallDepth = 500 // Prevents stack overflow

// RunProcedure executes a defined procedure with the given arguments.
func (i *Interpreter) RunProcedure(procName string, args ...lang.Value) (lang.Value, error) {
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

	procInterpreter := i.CloneWithNewVariables()
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

	// =========================================================================
	fmt.Printf(">>>> [DEBUG] RunProcedure ('%s'): Value coming OUT of executeSteps is: %#v\n", procName, result)
	// =========================================================================

	if err == nil {
		i.lastCallResult = result
	}

	return result, err
}
