// NeuroScript Version: 0.3.1
// File version: 10
// Purpose: Corrected variable shadowing bug in the Wrap method.
// filename: pkg/core/interpreter_procedures.go
// nlines: 160
// risk_rating: HIGH

package core

import (
	"errors"
	"fmt"
	"reflect"
)

// Wrap converts a native Go type into its corresponding NeuroScript Value type.
// If the input is already a Value type, it is returned unchanged.
func (i *Interpreter) Wrap(v interface{}) Value {
	if val, ok := v.(Value); ok {
		return val // It's already a Value, do nothing.
	}
	switch val := v.(type) {
	case string:
		return StringValue{Value: val}
	case int:
		return NumberValue{Value: float64(val)}
	case int64:
		return NumberValue{Value: float64(val)}
	case float64:
		return NumberValue{Value: val}
	case bool:
		return BoolValue{Value: val}
	case nil:
		return NilValue{}
	case []interface{}:
		list := make([]Value, len(val))
		// FIX: Use 'idx' to avoid shadowing the interpreter 'i'.
		for idx, item := range val {
			list[idx] = i.Wrap(item)
		}
		return NewListValue(list)
	case map[string]interface{}:
		newMap := make(map[string]Value)
		// FIX: Use a different variable name for the value to avoid confusion.
		for k, item := range val {
			newMap[k] = i.Wrap(item)
		}
		return NewMapValue(newMap)
	default:
		i.Logger().Warn("Attempted to wrap an unhandled native type; returning NilValue.", "type", fmt.Sprintf("%T", v))
		return NilValue{}
	}
}

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

func (i *Interpreter) RunProcedure(procName string, args ...interface{}) (result interface{}, err error) {
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

	procInterpreter := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: i.knownProcedures,
		stdout:          i.stdout,
		logger:          i.logger,
		toolRegistry:    i.toolRegistry,
		llmClient:       i.llmClient,
		fileAPI:         i.fileAPI,
	}

	for k, v := range i.variables {
		procInterpreter.variables[k] = v
	}

	for idx := 0; idx < numRequired; idx++ {
		paramName := proc.RequiredParams[idx]
		wrappedArg := i.Wrap(args[idx])
		if setErr := procInterpreter.SetVariable(paramName, wrappedArg); setErr != nil {
			return nil, fmt.Errorf("failed to set required parameter '%s': %w", paramName, setErr)
		}
	}

	if numProvided > numRequired {
		for idx := 0; idx < numOptional && (numRequired+idx) < numProvided; idx++ {
			paramSpec := proc.OptionalParams[idx]
			paramName := paramSpec.Name
			wrappedArg := i.Wrap(args[numRequired+idx])
			if setErr := procInterpreter.SetVariable(paramName, wrappedArg); setErr != nil {
				return nil, fmt.Errorf("failed to set optional parameter '%s': %w", paramName, setErr)
			}
		}
	}

	if proc.Variadic && proc.VariadicParamName != "" && numProvided > numTotalParams {
		variadicArgsRaw := args[numTotalParams:]
		variadicArgsValue := make([]Value, len(variadicArgsRaw))
		// FIX: Use 'idx' to avoid shadowing the interpreter 'i'.
		for idx, arg := range variadicArgsRaw {
			variadicArgsValue[idx] = i.Wrap(arg)
		}
		if setErr := procInterpreter.SetVariable(proc.VariadicParamName, NewListValue(variadicArgsValue)); setErr != nil {
			return nil, fmt.Errorf("failed to set variadic parameter '%s': %w", proc.VariadicParamName, setErr)
		}
	}

	result, _, _, err = procInterpreter.executeSteps(proc.Steps, false, nil)
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
