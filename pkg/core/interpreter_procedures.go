// NeuroScript Version: 0.3.1
// File version: 14
// Purpose: Reviewed for compliance with value-wrapping contract; file remains compliant with no changes needed.
// filename: pkg/core/interpreter_procedures.go
// nlines: 154
// risk_rating: HIGH

package core

import (
	"errors"
	"fmt"
)

// AddProcedure programmatically adds a single procedure to the interpreter's registry.
func (i *Interpreter) AddProcedure(proc Procedure) error {
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]*Procedure)
	}
	if proc.Name == "" {
		return errors.New("cannot add procedure with empty name")
	}
	if _, exists := i.knownProcedures[proc.Name]; exists {
		return fmt.Errorf("%w: '%s'", ErrProcedureExists, proc.Name)
	}
	i.knownProcedures[proc.Name] = &proc
	i.Logger().Debug("Added procedure definition.", "name", proc.Name)
	return nil
}

// KnownProcedures returns the map of known procedures.
func (i *Interpreter) KnownProcedures() map[string]*Procedure {
	if i.knownProcedures == nil {
		return make(map[string]*Procedure)
	}
	return i.knownProcedures
}

// RunProcedure executes a defined procedure with the given arguments.
// In accordance with the value wrapping contract, this core interpreter function
// accepts and returns only core.Value types. The caller is responsible for wrapping
// any primitive Go types into core.Value before calling this function.
func (i *Interpreter) RunProcedure(procName string, args ...Value) (Value, error) {
	originalProcName := i.currentProcName
	i.Logger().Debug("Running procedure", "name", procName, "caller", originalProcName)
	defer func() {
		i.currentProcName = originalProcName
		i.Logger().Debug("Finished procedure", "name", procName, "caller", originalProcName)
	}()

	i.currentProcName = procName

	proc, exists := i.knownProcedures[procName]
	if !exists {
		return nil, fmt.Errorf("%w: '%s'", ErrProcedureNotFound, procName)
	}

	numRequired := len(proc.RequiredParams)
	numOptional := len(proc.OptionalParams)
	numTotalParams := numRequired + numOptional
	numProvided := len(args)

	if numProvided < numRequired {
		return nil, fmt.Errorf("%w: procedure '%s' requires %d args, got %d", ErrArgumentMismatch, procName, numRequired, numProvided)
	}
	if !proc.Variadic && numProvided > numTotalParams {
		return nil, fmt.Errorf("%w: procedure '%s' expects max %d args, got %d", ErrArgumentMismatch, procName, numTotalParams, numProvided)
	}

	procInterpreter := i.CloneWithNewVariables()

	// Assign required parameters. Arguments are already core.Value.
	for idx := 0; idx < numRequired; idx++ {
		paramName := proc.RequiredParams[idx]
		if setErr := procInterpreter.SetVariable(paramName, args[idx]); setErr != nil {
			return nil, fmt.Errorf("failed to set required parameter '%s': %w", paramName, setErr)
		}
	}

	// Assign optional parameters.
	if numProvided > numRequired {
		for idx := 0; idx < numOptional && (numRequired+idx) < numProvided; idx++ {
			paramSpec := proc.OptionalParams[idx]
			paramName := paramSpec.Name
			if setErr := procInterpreter.SetVariable(paramName, args[numRequired+idx]); setErr != nil {
				return nil, fmt.Errorf("failed to set optional parameter '%s': %w", paramName, setErr)
			}
		}
	}

	// Assign variadic parameters, if any.
	if proc.Variadic && proc.VariadicParamName != "" && numProvided > numTotalParams {
		variadicArgs := args[numTotalParams:] // This is already a []Value slice
		variadicList := NewListValue(variadicArgs)
		if setErr := procInterpreter.SetVariable(proc.VariadicParamName, variadicList); setErr != nil {
			return nil, fmt.Errorf("failed to set variadic parameter '%s': %w", proc.VariadicParamName, setErr)
		}
	}

	result, _, _, err := procInterpreter.executeSteps(proc.Steps, false, nil)
	if err != nil {
		if _, ok := err.(*RuntimeError); !ok {
			err = fmt.Errorf("error executing steps for procedure '%s': %w", procName, err)
		}
		return nil, err
	}

	// Validate and shape the return value according to the procedure's definition.
	expectedReturnCount := len(proc.ReturnVarNames)
	if expectedReturnCount == 0 {
		return NilValue{}, nil
	}

	// Normalize a nil interface from executeSteps to a proper NilValue type for consistency.
	if result == nil {
		result = NilValue{}
	}

	if expectedReturnCount == 1 {
		// If one return value is expected, return the result as-is.
		return result, nil
	}

	// If multiple return values are expected, the result must be a ListValue.
	list, ok := result.(ListValue)
	if !ok {
		return nil, fmt.Errorf("%w: procedure '%s' expected %d return values, but returned a single value of type %s", ErrReturnMismatch, procName, expectedReturnCount, TypeOf(result))
	}

	// The list must contain the exact number of expected return values.
	if len(list.Value) != expectedReturnCount {
		return nil, fmt.Errorf("%w: procedure '%s' expected %d return values, but returned a list with %d items", ErrReturnMismatch, procName, expectedReturnCount, len(list.Value))
	}

	return list, nil
}
