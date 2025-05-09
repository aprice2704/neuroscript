// NeuroScript Version: 0.3.1
// File version: 0.1.2
// Add handling for ArgTypeMap in validation.
// nlines: 170
// risk_rating: HIGH
// filename: pkg/core/tools_validation.go
package core

import (
	// Import errors package
	"fmt"
	"io"
	"log"
	// "strings"
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

	// Create slice based on the number of args expected by the spec
	convertedArgs := make([]interface{}, len(spec.Args))
	var firstValidationError error // Keep track of the first specific validation error

	for i := 0; i < len(spec.Args); i++ { // Iterate based on spec length
		argSpec := spec.Args[i]
		var argValue interface{}
		var currentArgValidationError error // Error specific to the current argument iteration

		if i < len(rawArgs) { // Check if a raw arg exists for this spec index
			argValue = rawArgs[i]
			logger.Printf("Processing Arg %d ('%s'): Value=%#v (%T), SpecType=%s, Required=%t", i, argSpec.Name, argValue, argValue, argSpec.Type, argSpec.Required)

			// Handle nil for REQUIRED arguments explicitly first
			if argValue == nil {
				if argSpec.Required {
					// Use the more specific ErrValidationRequiredArgNil
					currentArgValidationError = fmt.Errorf("%w: argument '%s' (index %d) for tool '%s'", ErrValidationRequiredArgNil, argSpec.Name, i, spec.Name)
					logger.Printf("Required arg is nil: %v", currentArgValidationError)
				} else {
					logger.Printf("Optional arg '%s' is nil, accepting.", argSpec.Name)
					convertedArgs[i] = nil // Explicitly set nil for optional nil arg
					continue               // Skip further validation/coercion for nil optional arg
				}
			}
		} else {
			// Argument is missing in the call
			if argSpec.Required {
				// Use the specific ErrValidationRequiredArgMissing
				currentArgValidationError = fmt.Errorf("%w: argument '%s' (index %d) for tool '%s' is required but was not provided", ErrValidationRequiredArgMissing, argSpec.Name, i, spec.Name)
				logger.Printf("Required arg missing: %v", currentArgValidationError)
			} else {
				// Optional argument is missing, set to nil and continue
				logger.Printf("Optional arg '%s' missing, setting to nil.", argSpec.Name)
				convertedArgs[i] = nil
				continue
			}
		}

		// If an error occurred finding/checking the required/nil status, record it and potentially stop
		if currentArgValidationError != nil {
			if firstValidationError == nil {
				firstValidationError = currentArgValidationError
			}
			convertedArgs[i] = nil // Ensure slot is nil if error occurred before coercion
			continue
		}

		// If we get here, argValue is not nil and was provided. Proceed with type validation/coercion.
		coercedValue, coerceErr := validateAndCoerceType(argValue, argSpec.Type, spec.Name, argSpec.Name)
		if coerceErr != nil {
			currentArgValidationError = coerceErr // validateAndCoerceType already wraps ErrValidationTypeMismatch
			logger.Printf("Type Coercion Failed for arg '%s': %v", argSpec.Name, currentArgValidationError)
		} else {
			convertedArgs[i] = coercedValue // Assign coerced value
			logger.Printf("Type Coercion OK for arg '%s': Value=%#v (%T)", argSpec.Name, convertedArgs[i], convertedArgs[i])
		}

		// Handle final error check for this argument *before assigning*
		if currentArgValidationError != nil {
			if firstValidationError == nil {
				firstValidationError = currentArgValidationError
			}
		}

	} // End loop over spec args

	// If a specific validation error occurred during the loop, return it now.
	if firstValidationError != nil {
		logger.Printf("Returning first validation error encountered: %v", firstValidationError)
		return nil, firstValidationError
	}

	// --- Argument count check *after* specific checks ---
	// Only perform count check if no specific errors were found above.
	if numRawArgs < minRequiredArgs || numRawArgs > maxExpectedArgs {
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
		// Use ErrValidationArgCount sentinel error
		return nil, fmt.Errorf("%w: %w", ErrValidationArgCount, err)
	}

	logger.Printf("Validation successful for all args. Returning: %#v", convertedArgs)
	return convertedArgs, nil
}

// validateAndCoerceType - Helper within this file. Returns defined errors on failure.
// Ensures that ErrValidationTypeMismatch is wrapped correctly.
func validateAndCoerceType(rawValue interface{}, expectedType ArgType, toolName, argName string) (interface{}, error) {
	if rawValue == nil {
		return nil, nil
	}

	var finalValue interface{}
	var err error
	ok := true // Assume ok unless conversion fails

	switch expectedType {
	case ArgTypeString:
		finalValue, ok = rawValue.(string)
		if !ok {
			err = fmt.Errorf("expected string, got %T", rawValue)
		}
	case ArgTypeInt:
		var intVal int64
		intVal, ok = toInt64(rawValue)
		if ok {
			finalValue = intVal
		} else {
			err = fmt.Errorf("value %v (%T) cannot be converted to int (int64)", rawValue, rawValue)
		}
	case ArgTypeFloat:
		var floatVal float64
		floatVal, ok = toFloat64(rawValue)
		if ok {
			finalValue = floatVal
		} else {
			err = fmt.Errorf("value %v (%T) cannot be converted to float (float64)", rawValue, rawValue)
		}
	case ArgTypeBool:
		var boolVal bool
		boolVal, ok = ConvertToBool(rawValue)
		if ok {
			finalValue = boolVal
		} else {
			err = fmt.Errorf("value %v (%T) cannot be converted to bool", rawValue, rawValue)
		}
	case ArgTypeSliceString:
		var sliceVal []string
		sliceVal, ok, err = ConvertToSliceOfString(rawValue)
		if ok {
			finalValue = sliceVal
		} else if err == nil {
			err = fmt.Errorf("expected slice of strings, got %T", rawValue)
		}
	case ArgTypeSliceAny:
		var sliceVal []interface{}
		sliceVal, ok, err = convertToSliceOfAny(rawValue)
		if ok {
			finalValue = sliceVal
		} else if err == nil {
			err = fmt.Errorf("expected a slice (list), got %T", rawValue)
		}
	// --- ADDED Case for ArgTypeMap ---
	case ArgTypeMap:
		finalValue, ok = rawValue.(map[string]interface{})
		if !ok {
			err = fmt.Errorf("expected map[string]interface{}, got %T", rawValue)
		}
	// --- END ADD ---
	case ArgTypeAny:
		finalValue, ok = rawValue, true
	default:
		err = fmt.Errorf("%w: unknown expected type '%s' for tool '%s' arg '%s'", ErrInternalTool, expectedType, toolName, argName)
		ok = false
	}

	// Check for errors from conversion helpers or failed assertions
	if err != nil || !ok {
		finalErrMsg := fmt.Sprintf("argument '%s' of tool '%s'", argName, toolName)
		// Ensure ErrValidationTypeMismatch is wrapped consistently
		if err != nil {
			// If an error was returned by a conversion helper, wrap it
			return nil, fmt.Errorf("%w: %s: %w", ErrValidationTypeMismatch, finalErrMsg, err)
		} else {
			// If only 'ok' is false, create the standard mismatch message
			return nil, fmt.Errorf("%w: %s: expected %s, got %T", ErrValidationTypeMismatch, finalErrMsg, expectedType, rawValue)
		}
	}

	return finalValue, nil
}
