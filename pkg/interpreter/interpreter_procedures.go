// NeuroScript Version: 0.5.2
// File version: 31
// Purpose: Final correction to error propagation to ensure the specific RuntimeError type is preserved.
// filename: pkg/interpreter/interpreter_procedures.go
// nlines: 120
// risk_rating: HIGH

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

const maxCallDepth = 500 // Prevents stack overflow

// RunProcedure executes a defined procedure with the given arguments.
func (i *Interpreter) RunProcedure(procName string, args ...lang.Value) (lang.Value, error) {
	proc, exists := i.KnownProcedures()[procName]
	if !exists {
		// This now returns the correct error type thanks to the change in interpreter.go
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
	if len(i.state.errorHandlerStack) > 0 {
		newStack := make([][]*ast.Step, len(i.state.errorHandlerStack))
		copy(newStack, i.state.errorHandlerStack)
		procInterpreter.state.errorHandlerStack = newStack
	}

	procInterpreter.state.stackFrames = append(i.state.stackFrames, procName)
	defer func() {
		if len(procInterpreter.state.stackFrames) > 0 {
			procInterpreter.state.stackFrames = procInterpreter.state.stackFrames[:len(procInterpreter.state.stackFrames)-1]
		}
	}()

	for idx, paramName := range proc.RequiredParams {
		if idx < len(args) {
			procInterpreter.SetVariable(paramName, args[idx])
		} else {
			return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("missing required argument '%s'", paramName), lang.ErrArgumentMismatch)
		}
	}

	if len(proc.ErrorHandlers) > 0 {
		procInterpreter.state.errorHandlerStack = append(procInterpreter.state.errorHandlerStack, proc.ErrorHandlers)
		defer func() {
			if len(procInterpreter.state.errorHandlerStack) > 0 {
				procInterpreter.state.errorHandlerStack = procInterpreter.state.errorHandlerStack[:len(procInterpreter.state.errorHandlerStack)-1]
			}
		}()
	}

	result, _, _, err := procInterpreter.executeSteps(proc.Steps, false, nil)

	// FINAL FIX: The error from executeSteps is guaranteed to be a *lang.RuntimeError
	// or nil. We propagate it directly to prevent it from being re-wrapped.
	return result, err
}
