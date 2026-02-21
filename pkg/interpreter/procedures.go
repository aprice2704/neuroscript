// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 53
// :: description: Fixes bug where optional parameters were not bound to the execution scope.
// :: latestChange: Iterate over proc.OptionalParams and bind arguments or fallback to nil.
// :: filename: pkg/interpreter/procedures.go
// :: serialization: go

package interpreter

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast" // Added import
	"github.com/aprice2704/neuroscript/pkg/lang"
)

const maxCallDepth = 500 // Prevents stack overflow

// runProcedure executes a defined procedure with the given arguments.
// This is the internal implementation.
func (i *Interpreter) runProcedure(procName string, args ...lang.Value) (lang.Value, error) {
	// FIX: Implement provider-aware procedure lookup, just like GetVariable.
	// 1. Check local procedures
	proc, exists := i.state.knownProcedures[procName]

	// 2. If not found, check symbol provider
	if !exists {
		// Use the helper method from interpreter_load.go
		if provider := i.symbolProvider(); provider != nil {
			procAny, existsProvider := provider.GetProcedure(procName)
			if existsProvider {
				if p, ok := procAny.(*ast.Procedure); ok {
					proc = p
					exists = true
				} else if procAny != nil {
					// This would be a bad state, log it.
					i.Logger().Error("Symbol provider returned non-procedure for GetProcedure",
						"name", procName, "type", fmt.Sprintf("%T", procAny))
				}
			}
		}
	}
	// --- End of FIX ---

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

	for idx, optParam := range proc.OptionalParams {
		argIdx := len(proc.RequiredParams) + idx
		if argIdx < len(args) {
			procInterpreter.SetVariable(optParam.Name, args[argIdx])
		} else {
			procInterpreter.SetVariable(optParam.Name, &lang.NilValue{})
		}
	}

	result, _, _, err := procInterpreter.executeSteps(proc.Steps, false, nil)

	return result, err
}
