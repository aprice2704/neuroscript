// NeuroScript Version: 0.5.2
// File version: 16
// Purpose: Added error handler stack management to RunProcedure to support scoped error handling in nested calls.
// filename: pkg/interpreter/interpreter_procedures.go
// nlines: 140
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// AddProcedure programmatically adds a single procedure to the interpreter's registry.
func (i *Interpreter) AddProcedure(proc ast.Procedure) error {
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
	i.Logger().Debug("Added procedure definition.", "name", proc.Name())
	return nil
}

// KnownProcedures returns the map of known procedures.
func (i *Interpreter) KnownProcedures() map[string]*ast.Procedure {
	if i.state.knownProcedures == nil {
		return make(map[string]*ast.Procedure)
	}
	return i.state.knownProcedures
}

// RunProcedure executes a defined procedure with the given arguments.
func (i *Interpreter) RunProcedure(procName string, args ...lang.Value) (lang.Value, error) {
	originalProcName := i.state.currentProcName
	i.Logger().Debug("Running procedure", "name", procName, "caller", originalProcName)
	defer func() {
		i.state.currentProcName = originalProcName
		i.Logger().Debug("Finished procedure", "name", procName, "caller", originalProcName)
	}()

	i.state.currentProcName = procName

	proc, exists := i.state.knownProcedures[procName]
	if !exists {
		return nil, fmt.Errorf("%w: '%s'", lang.ErrProcedureNotFound, procName)
	}

	numRequired := len(proc.RequiredParams)
	numOptional := len(proc.OptionalParams)
	numTotalParams := numRequired + numOptional
	numProvided := len(args)

	if numProvided < numRequired {
		return nil, fmt.Errorf("%w: procedure '%s' requires %d args, got %d", lang.ErrArgumentMismatch, procName, numRequired, numProvided)
	}
	if !proc.Variadic && numProvided > numTotalParams {
		return nil, fmt.Errorf("%w: procedure '%s' expects max %d args, got %d", lang.ErrArgumentMismatch, procName, numTotalParams, numProvided)
	}

	// Create a new interpreter with a fresh variable scope for the procedure call.
	procInterpreter := i.CloneWithNewVariables()

	// Assign required and optional parameters.
	for idx := 0; idx < numRequired; idx++ {
		procInterpreter.SetVariable(proc.RequiredParams[idx], args[idx])
	}
	if numProvided > numRequired {
		for idx := 0; idx < numOptional && (numRequired+idx) < numProvided; idx++ {
			paramSpec := proc.OptionalParams[idx]
			procInterpreter.SetVariable(paramSpec.Name, args[numRequired+idx])
		}
	}

	// Assign variadic parameters.
	if proc.Variadic && proc.VariadicParamName != "" && numProvided > numTotalParams {
		variadicArgs := args[numTotalParams:]
		variadicList := lang.NewListValue(variadicArgs)
		procInterpreter.SetVariable(proc.VariadicParamName, variadicList)
	}

	// Push this procedure's error handlers onto the stack and defer their removal.
	procInterpreter.state.errorHandlerStack = append(procInterpreter.state.errorHandlerStack, proc.ErrorHandlers)
	defer func() {
		if len(procInterpreter.state.errorHandlerStack) > 0 {
			procInterpreter.state.errorHandlerStack = procInterpreter.state.errorHandlerStack[:len(procInterpreter.state.errorHandlerStack)-1]
		}
	}()

	result, _, _, err := procInterpreter.executeSteps(proc.Steps, false, nil)
	if err != nil {
		if _, ok := err.(*lang.RuntimeError); !ok {
			err = fmt.Errorf("error executing steps for procedure '%s': %w", procName, err)
		}
		return nil, err
	}

	// Validate return values against the procedure's definition.
	expectedReturnCount := len(proc.ReturnVarNames)
	if expectedReturnCount == 0 {
		return &lang.NilValue{}, nil
	}

	if result == nil {
		result = &lang.NilValue{}
	}

	if expectedReturnCount == 1 {
		return result, nil
	}

	list, ok := result.(lang.ListValue)
	if !ok {
		return nil, fmt.Errorf("%w: procedure '%s' expected %d return values, but returned a single value of type %s", lang.ErrReturnMismatch, procName, expectedReturnCount, lang.TypeOf(result))
	}

	if len(list.Value) != expectedReturnCount {
		return nil, fmt.Errorf("%w: procedure '%s' expected %d return values, but returned a list with %d items", lang.ErrReturnMismatch, procName, expectedReturnCount, len(list.Value))
	}

	return list, nil
}