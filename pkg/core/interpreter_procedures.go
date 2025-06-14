// NeuroScript Version: 0.3.1
// File version: 12
// Purpose: Corrects all calls to the new standalone Wrap function to handle two return values.
// filename: pkg/core/interpreter_procedures.go
// nlines: 168
// risk_rating: HIGH

package core

import (
	"errors"
	"fmt"
	"reflect"
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

func (i *Interpreter) RunProcedure(procName string, args ...interface{}) (interface{}, error) {
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

	for idx := 0; idx < numRequired; idx++ {
		paramName := proc.RequiredParams[idx]
		wrappedArg, err := Wrap(args[idx])
		if err != nil {
			return nil, fmt.Errorf("failed to wrap required parameter '%s': %w", paramName, err)
		}
		if setErr := procInterpreter.SetVariable(paramName, wrappedArg); setErr != nil {
			return nil, fmt.Errorf("failed to set required parameter '%s': %w", paramName, setErr)
		}
	}

	if numProvided > numRequired {
		for idx := 0; idx < numOptional && (numRequired+idx) < numProvided; idx++ {
			paramSpec := proc.OptionalParams[idx]
			paramName := paramSpec.Name
			wrappedArg, err := Wrap(args[numRequired+idx])
			if err != nil {
				return nil, fmt.Errorf("failed to wrap optional parameter '%s': %w", paramName, err)
			}
			if setErr := procInterpreter.SetVariable(paramName, wrappedArg); setErr != nil {
				return nil, fmt.Errorf("failed to set optional parameter '%s': %w", paramName, setErr)
			}
		}
	}

	if proc.Variadic && proc.VariadicParamName != "" && numProvided > numTotalParams {
		variadicArgsRaw := args[numTotalParams:]
		variadicArgsValue := make([]Value, len(variadicArgsRaw))
		for idx, arg := range variadicArgsRaw {
			var err error
			variadicArgsValue[idx], err = Wrap(arg)
			if err != nil {
				return nil, fmt.Errorf("failed to wrap variadic parameter #%d: %w", idx+1, err)
			}
		}
		if setErr := procInterpreter.SetVariable(proc.VariadicParamName, NewListValue(variadicArgsValue)); setErr != nil {
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

	expectedReturnCount := len(proc.ReturnVarNames)
	if expectedReturnCount == 0 {
		return NilValue{}, nil
	}

	var finalResult interface{}
	if list, ok := result.(ListValue); ok && expectedReturnCount > 1 {
		if len(list.Value) != expectedReturnCount {
			return nil, fmt.Errorf("%w: procedure '%s' expected %d return values, but returned %d", ErrReturnMismatch, procName, expectedReturnCount, len(list.Value))
		}
		finalResult = list
	} else if expectedReturnCount == 1 {
		finalResult = result
	} else if result == nil || (reflect.ValueOf(result).IsValid() && reflect.ValueOf(result).IsNil()) {
		if expectedReturnCount > 0 {
			return nil, fmt.Errorf("%w: procedure '%s' expected %d return values, but returned nil", ErrReturnMismatch, procName, expectedReturnCount)
		}
	} else {
		return nil, fmt.Errorf("%w: procedure '%s' expected %d return values, but returned a single value of type %s", ErrReturnMismatch, procName, expectedReturnCount, TypeOf(result))
	}

	return finalResult, nil
}
