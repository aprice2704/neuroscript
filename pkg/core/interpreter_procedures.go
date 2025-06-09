// NeuroScript Version: 0.3.1
// File version: 6
// Purpose: Removed obsolete event logic and fixed KnownProcedures function signature.
// filename: pkg/core/interpreter_procedures.go
// nlines: 90
// risk_rating: MEDIUM

package core

import (
	"errors"
	"fmt"
	"reflect"
)

// AddProcedure programmatically adds a single procedure to the interpreter's registry.
// Note: For loading full scripts, use interpreter.LoadProgram(...)
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

func (i *Interpreter) RunProcedure(procName string, args ...interface{}) (result interface{}, err error) {
	originalProcName := i.currentProcName
	i.Logger().Debug("Running procedure", "name", procName, "arg_count", len(args))
	i.currentProcName = procName
	defer func() {
		if r := recover(); r != nil {
			err = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("panic occurred during procedure '%s': %v", procName, r), errors.New("panic"))
			i.Logger().Error("Panic recovered during procedure execution", "proc_name", procName, "panic_value", r, "error", err)
			result = nil
		}
		i.currentProcName = originalProcName
	}()

	proc, exists := i.knownProcedures[procName]
	if !exists {
		err = fmt.Errorf("%w: '%s'", ErrProcedureNotFound, procName)
		i.Logger().Error("Procedure definition not found", "name", procName)
		return nil, err
	}

	numRequired := len(proc.RequiredParams)
	numOptional := len(proc.OptionalParams)
	numTotalParams := numRequired + numOptional
	numProvided := len(args)

	if numProvided < numRequired {
		err = fmt.Errorf("%w: procedure '%s' requires %d arguments, but received %d", ErrArgumentMismatch, procName, numRequired, numProvided)
		i.Logger().Error("Argument count mismatch (too few)", "proc_name", procName, "required", numRequired, "provided", numProvided)
		return nil, err
	}
	if numProvided > numTotalParams && !proc.Variadic {
		i.Logger().Warn("Procedure called with extra arguments.", "proc_name", procName, "provided", numProvided, "defined_max", numTotalParams)
	}

	procScope := make(map[string]interface{})
	if i.variables != nil {
		for k, v := range i.variables {
			procScope[k] = v
		}
	}
	originalScope := i.variables
	i.variables = procScope
	defer func() {
		i.variables = originalScope
		i.Logger().Debug("Restored parent variable scope.", "proc_name", procName)
	}()

	for idx, paramName := range proc.RequiredParams {
		if idx < len(args) {
			if setErr := i.SetVariable(paramName, args[idx]); setErr != nil {
				return nil, fmt.Errorf("failed to set required parameter '%s': %w", paramName, setErr)
			}
		} else {
			return nil, fmt.Errorf("internal error: insufficient arguments for required parameter '%s'", paramName)
		}
	}

	for idx, paramSpec := range proc.OptionalParams {
		paramName := paramSpec.Name
		valueToSet := paramSpec.DefaultValue
		originalOptionalArgIndex := numRequired + idx
		if originalOptionalArgIndex < numProvided {
			valueToSet = args[originalOptionalArgIndex]
		}
		if setErr := i.SetVariable(paramName, valueToSet); setErr != nil {
			return nil, fmt.Errorf("failed to set optional parameter '%s': %w", paramName, setErr)
		}
	}

	if proc.Variadic && proc.VariadicParamName != "" && numProvided > numTotalParams {
		variadicArgs := args[numTotalParams:]
		if setErr := i.SetVariable(proc.VariadicParamName, variadicArgs); setErr != nil {
			return nil, fmt.Errorf("failed to set variadic parameter '%s': %w", proc.VariadicParamName, setErr)
		}
	}

	result, _, _, err = i.executeSteps(proc.Steps, false, nil)
	if err != nil {
		if _, ok := err.(*RuntimeError); !ok {
			err = fmt.Errorf("error executing steps for procedure '%s': %w", procName, err)
		}
		return nil, err
	}

	expectedReturnCount := len(proc.ReturnVarNames)
	actualReturnCount := 0
	var finalResult interface{}

	if result != nil {
		resultValue := reflect.ValueOf(result)
		kind := resultValue.Kind()
		if kind == reflect.Ptr || kind == reflect.Interface {
			if !resultValue.IsNil() {
				resultValue = resultValue.Elem()
				kind = resultValue.Kind()
			} else {
				kind = reflect.Invalid
			}
		}
		if kind == reflect.Slice {
			actualReturnCount = resultValue.Len()
			finalResult = result
		} else if resultValue.IsValid() {
			actualReturnCount = 1
			finalResult = result
		}
	}

	if actualReturnCount != expectedReturnCount {
		if !(expectedReturnCount == 0 && actualReturnCount == 0) {
			err = fmt.Errorf("%w: procedure '%s' expected %d return values, but yielded %d", ErrReturnMismatch, procName, expectedReturnCount, actualReturnCount)
			return nil, err
		}
	}
	i.lastCallResult = finalResult
	return finalResult, nil
}
