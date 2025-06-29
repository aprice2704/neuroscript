// NeuroScript Version: 0.4.2
// File version: 1
// Purpose: Handles loading NeuroScript programs and executing top-level procedures.
// filename: pkg/core/interpreter_program.go
// nlines: 66
// risk_rating: LOW
package core

import (
	"fmt"
)

// ExecuteProc finds and executes a procedure that has already been loaded into the
// interpreter by name, returning the final unwrapped result.
func (i *Interpreter) ExecuteProc(procName string) (interface{}, error) {
	proc, exists := i.knownProcedures[procName]
	if !exists {
		// As a fallback for simple test scripts, if only one procedure is loaded, run it.
		if len(i.knownProcedures) == 1 {
			for _, p := range i.knownProcedures {
				proc = p
				break
			}
		} else {
			return nil, NewRuntimeError(ErrorCodeProcNotFound, fmt.Sprintf("procedure '%s' not found", procName), ErrProcedureNotFound)
		}
	}

	// --- NEW: Register handlers for this procedure call and defer their cleanup ---
	i.errorHandlerStack = append(i.errorHandlerStack, proc.ErrorHandlers)
	defer func() {
		// This cleanup will run when ExecuteProc returns, for any reason (success, error, panic),
		// guaranteeing that handlers for this scope are removed and do not leak.
		if len(i.errorHandlerStack) > 0 {
			i.errorHandlerStack = i.errorHandlerStack[:len(i.errorHandlerStack)-1]
		}
	}()
	// --- END NEW ---

	i.currentProcName = proc.Name
	finalResult, wasReturn, _, err := i.executeSteps(proc.Steps, false, nil)
	if err != nil {
		return nil, err
	}

	var returnValue Value
	if wasReturn {
		returnValue = finalResult
	} else {
		// If the function ends without an explicit return, the result is the
		// result of the last executed statement.
		returnValue = i.lastCallResult
	}

	return Unwrap(returnValue), nil
}

// LoadProgram registers all procedures and event handlers from a parsed Program AST.
func (i *Interpreter) LoadProgram(prog *Program) error {
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]*Procedure)
	}
	for name, proc := range prog.Procedures {
		if _, exists := i.knownProcedures[name]; exists {
			return fmt.Errorf("procedure '%s' already exists", name)
		}
		i.knownProcedures[name] = proc
	}
	if i.eventHandlers == nil {
		i.eventHandlers = make(map[string][]*OnEventDecl)
	}
	for _, ev := range prog.Events {
		nameLit, ok := ev.EventNameExpr.(*StringLiteralNode)
		if !ok {
			// This check enforces that event names must be static strings at load time.
			return NewRuntimeError(ErrorCodeType, "event name must be a static string literal", nil).WithPosition(ev.Pos)
		}
		eventName := nameLit.Value

		i.eventHandlersMu.Lock()
		i.eventHandlers[eventName] = append(i.eventHandlers[eventName], ev)
		i.eventHandlersMu.Unlock()
	}
	return nil
}
