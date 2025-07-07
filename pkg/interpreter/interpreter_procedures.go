// NeuroScript Version: 0.5.2
// File version: 44
// Purpose: Corrected error propagation by REMOVING the copying of the parent's error handler stack to the child interpreter. This ensures errors are handled in the correct scope.
// filename: pkg/interpreter/interpreter_procedures.go
// nlines: 115
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

	// --- DEBUG ---
	//	fmt.Printf("[DEBUG] Entering RunProcedure: '%s'. Parent Interpreter: %p\n", procName, i)

	procInterpreter := i.CloneWithNewVariables()

	// --- DEBUG ---
	//	fmt.Printf("[DEBUG] Created new interpreter for '%s': %p\n", procName, procInterpreter)

	// Inherit and manage the call stack correctly.
	procInterpreter.state.stackFrames = append(i.state.stackFrames, procName)

	// FIX: DO NOT copy the parent's error handler stack. A procedure should only
	// be aware of its OWN 'on error' handlers. Any unhandled error will
	// naturally propagate up to the caller when this function returns.
	// procInterpreter.state.errorHandlerStack = make([][]*ast.Step, len(i.state.errorHandlerStack)) // <-- REMOVED
	// copy(procInterpreter.state.errorHandlerStack, i.state.errorHandlerStack) // <-- REMOVED

	// Pass arguments by setting them in the new, isolated scope.
	for idx, paramName := range proc.RequiredParams {
		if idx < len(args) {
			procInterpreter.SetVariable(paramName, args[idx])
		} else {
			procInterpreter.SetVariable(paramName, &lang.NilValue{})
		}
	}

	// Add any error handlers that are defined *inside this specific procedure*.
	if len(proc.ErrorHandlers) > 0 {
		procInterpreter.state.errorHandlerStack = append(procInterpreter.state.errorHandlerStack, proc.ErrorHandlers)
	}

	// The execution now happens in the sandboxed interpreter.
	result, _, _, err := procInterpreter.executeSteps(proc.Steps, false, nil)

	// --- DEBUG ---
	//	fmt.Printf("[DEBUG] Exiting RunProcedure: '%s'.\n", procName)

	// Return the result and any unhandled error. The CALLER's executeSteps loop
	// will be responsible for catching this error and invoking its own handler.
	return result, err
}
