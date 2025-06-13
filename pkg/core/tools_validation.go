// NeuroScript Version: 0.3.1
// File version: 0.1.3
// Purpose: Unwraps Value types before validation to handle interpreter-native values.
// nlines: 175
// risk_rating: HIGH
// filename: pkg/core/tools_validation.go

package core

import (
	"fmt"
	"io"
	"log"
)

// ValidateAndConvertArgs checks arguments and returns defined errors on failure.
func ValidateAndConvertArgs(spec ToolSpec, rawArgs []interface{}) ([]interface{}, error) {
	logger := log.New(io.Discard, "[VALIDATE DEBUG] ", log.Ltime|log.Lshortfile)
	// logger = log.New(os.Stderr, "[VALIDATE DEBUG] ", log.Ltime|log.Lshortfile) // Uncomment for debugging

	minRequiredArgs := 0
	for _, argSpec := range spec.Args {
		if argSpec.Required {
			minRequiredArgs++
		}
	}
	maxExpectedArgs := len(spec.Args)
	numRawArgs := len(rawArgs)

	convertedArgs := make([]interface{}, len(spec.Args))
	var firstValidationError error

	for i := 0; i < len(spec.Args); i++ {
		argSpec := spec.Args[i]
		var rawValue interface{} // This is the value from the interpreter, possibly a Value type
		var currentArgValidationError error

		if i < len(rawArgs) {
			rawValue = rawArgs[i]

			// FIX: Unwrap the Value type to its native Go equivalent before validation.
			unwrappedValue := unwrapValue(rawValue)

			logger.Printf("Processing Arg %d ('%s'): RawValue=%#v (%T), UnwrappedValue=%#v (%T), SpecType=%s, Required=%t",
				i, argSpec.Name, rawValue, rawValue, unwrappedValue, unwrappedValue, argSpec.Type, argSpec.Required)

			if unwrappedValue == nil {
				if argSpec.Required {
					currentArgValidationError = fmt.Errorf("%w: argument '%s' (index %d) for tool '%s'", ErrValidationRequiredArgNil, argSpec.Name, i, spec.Name)
					logger.Printf("Required arg is nil: %v", currentArgValidationError)
				} else {
					logger.Printf("Optional arg '%s' is nil, accepting.", argSpec.Name)
					convertedArgs[i] = nil
					continue
				}
			} else {
				// If we have a non-nil value, proceed with type validation/coercion.
				coercedValue, coerceErr := validateAndCoerceType(unwrappedValue, argSpec.Type, spec.Name, argSpec.Name)
				if coerceErr != nil {
					currentArgValidationError = coerceErr
					logger.Printf("Type Coercion Failed for arg '%s': %v", argSpec.Name, currentArgValidationError)
				} else {
					convertedArgs[i] = coercedValue
					logger.Printf("Type Coercion OK for arg '%s': Value=%#v (%T)", argSpec.Name, convertedArgs[i], convertedArgs[i])
				}
			}
		} else {
			// Argument is missing in the call
			if argSpec.Required {
				currentArgValidationError = fmt.Errorf("%w: argument '%s' (index %d) for tool '%s' is required but was not provided", ErrValidationRequiredArgMissing, argSpec.Name, i, spec.Name)
				logger.Printf("Required arg missing: %v", currentArgValidationError)
			} else {
				logger.Printf("Optional arg '%s' missing, setting to nil.", argSpec.Name)
				convertedArgs[i] = nil
				continue
			}
		}

		if currentArgValidationError != nil {
			if firstValidationError == nil {
				firstValidationError = currentArgValidationError
			}
			convertedArgs[i] = nil
			continue
		}
	} // End loop over spec args

	if firstValidationError != nil {
		logger.Printf("Returning first validation error encountered: %v", firstValidationError)
		return nil, firstValidationError
	}

	if numRawArgs < minRequiredArgs || (!spec.Variadic && numRawArgs > maxExpectedArgs) {
		expectedArgsStr := ""
		if minRequiredArgs == maxExpectedArgs {
			expectedArgsStr = fmt.Sprintf("exactly %d", minRequiredArgs)
		} else if maxExpectedArgs > 0 {
			expectedArgsStr = fmt.Sprintf("between %d and %d", minRequiredArgs, maxExpectedArgs)
		} else {
			expectedArgsStr = "exactly 0"
		}
		err := fmt.Errorf("tool '%s' expected %s arguments, but received %d", spec.Name, expectedArgsStr, numRawArgs)
		logger.Printf("Arg count error: %v", err)
		return nil, fmt.Errorf("%w: %w", ErrValidationArgCount, err)
	}

	logger.Printf("Validation successful for all args. Returning: %#v", convertedArgs)
	return convertedArgs, nil
}

// validateAndCoerceType works with native Go types because the calling function now unwraps Values.
func validateAndCoerceType(nativeValue interface{}, expectedType ArgType, toolName, argName string) (interface{}, error) {
	if nativeValue == nil {
		return nil, nil // Should have been handled by the required/optional logic already.
	}

	var finalValue interface{}
	var err error
	ok := true

	switch expectedType {
	case ArgTypeString:
		finalValue, ok = nativeValue.(string)
		if !ok {
			err = fmt.Errorf("expected string, got %T", nativeValue)
		}
	case ArgTypeInt:
		var intVal int64
		intVal, ok = toInt64(nativeValue)
		if ok {
			finalValue = intVal
		} else {
			err = fmt.Errorf("value %v (%T) cannot be converted to int (int64)", nativeValue, nativeValue)
		}
	case ArgTypeFloat:
		var floatVal float64
		floatVal, ok = toFloat64(nativeValue)
		if ok {
			finalValue = floatVal
		} else {
			err = fmt.Errorf("value %v (%T) cannot be converted to float (float64)", nativeValue, nativeValue)
		}
	case ArgTypeBool:
		var boolVal bool
		boolVal, ok = ConvertToBool(nativeValue)
		if ok {
			finalValue = boolVal
		} else {
			err = fmt.Errorf("value %v (%T) cannot be converted to bool", nativeValue, nativeValue)
		}
	case ArgTypeSliceString:
		var sliceVal []string
		sliceVal, ok, err = ConvertToSliceOfString(nativeValue)
		if ok {
			finalValue = sliceVal
		} else if err == nil {
			err = fmt.Errorf("expected slice of strings, got %T", nativeValue)
		}
	case ArgTypeSliceAny:
		var sliceVal []interface{}
		sliceVal, ok, err = convertToSliceOfAny(nativeValue)
		if ok {
			finalValue = sliceVal
		} else if err == nil {
			err = fmt.Errorf("expected a slice (list), got %T", nativeValue)
		}
	case ArgTypeMap:
		finalValue, ok = nativeValue.(map[string]interface{})
		if !ok {
			err = fmt.Errorf("expected map[string]interface{}, got %T", nativeValue)
		}
	case ArgTypeAny:
		finalValue, ok = nativeValue, true
	default:
		err = fmt.Errorf("%w: unknown expected type '%s' for tool '%s' arg '%s'", ErrInternalTool, expectedType, toolName, argName)
		ok = false
	}

	if err != nil || !ok {
		finalErrMsg := fmt.Sprintf("argument '%s' of tool '%s'", argName, toolName)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %w", ErrValidationTypeMismatch, finalErrMsg, err)
		}
		return nil, fmt.Errorf("%w: %s: expected %s, got %T", ErrValidationTypeMismatch, finalErrMsg, expectedType, nativeValue)
	}

	return finalValue, nil
}
