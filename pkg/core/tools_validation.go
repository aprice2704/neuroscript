// filename: pkg/core/tools_validation.go
package core

import (
	"fmt"
	"io"
	"log"
	// "strings"
)

// ValidateAndConvertArgs checks arguments and returns defined errors on failure.
func ValidateAndConvertArgs(spec ToolSpec, rawArgs []interface{}) ([]interface{}, error) {
	logger := log.New(io.Discard, "[VALIDATE DEBUG] ", log.Ltime|log.Lshortfile)
	// logger = log.New(os.Stderr, "[VALIDATE DEBUG] ", log.Ltime|log.Lshortfile)

	minRequiredArgs := 0
	for _, argSpec := range spec.Args {
		if argSpec.Required {
			minRequiredArgs++
		}
	}
	maxExpectedArgs := len(spec.Args)
	numRawArgs := len(rawArgs)

	// Check argument count
	if numRawArgs < minRequiredArgs || numRawArgs > maxExpectedArgs {
		expectedArgsStr := ""
		if minRequiredArgs == maxExpectedArgs {
			expectedArgsStr = fmt.Sprintf("exactly %d", minRequiredArgs)
		} else if maxExpectedArgs > 0 {
			expectedArgsStr = fmt.Sprintf("between %d and %d", minRequiredArgs, maxExpectedArgs)
		} else {
			expectedArgsStr = "exactly 0"
		}
		if numRawArgs == 0 && minRequiredArgs == 0 && maxExpectedArgs > 0 {
			expectedArgsStr = fmt.Sprintf("up to %d optional", maxExpectedArgs)
		}
		err := fmt.Errorf("tool '%s' expected %s arguments, but received %d", spec.Name, expectedArgsStr, numRawArgs)
		logger.Printf("Arg count error: %v", err)
		return nil, fmt.Errorf("%w: %w", ErrValidationArgCount, err)
	}

	// Create slice based on the number of args expected by the spec
	convertedArgs := make([]interface{}, len(spec.Args))

	for i := 0; i < len(spec.Args); i++ { // Iterate based on spec length
		argSpec := spec.Args[i]
		var argValue interface{}
		var validationErr error = nil // Initialize error for this arg scope
		var validatedValue interface{}

		if i < len(rawArgs) { // Check if a raw arg exists for this spec index
			argValue = rawArgs[i]
		} else {
			// Argument is missing in the call
			if argSpec.Required {
				validationErr = fmt.Errorf("%w: argument '%s' (index %d) for tool '%s' is required but was not provided", ErrValidationRequiredArgNil, argSpec.Name, i, spec.Name)
				logger.Printf("Required arg missing: %v", validationErr)
				return nil, validationErr // Return immediately
			} else {
				// Optional argument is missing, set to nil and continue
				logger.Printf("Optional arg '%s' missing, setting to nil.", argSpec.Name)
				convertedArgs[i] = nil
				continue
			}
		}

		logger.Printf("Processing Arg %d ('%s'): Value=%#v (%T), SpecType=%s, Required=%t", i, argSpec.Name, argValue, argValue, argSpec.Type, argSpec.Required)

		// Handle nil for REQUIRED arguments explicitly first
		if argValue == nil {
			if argSpec.Required {
				validationErr = fmt.Errorf("%w: argument '%s' (index %d) for tool '%s'", ErrValidationRequiredArgNil, argSpec.Name, i, spec.Name)
				logger.Printf("Required arg is nil: %v", validationErr)
				// *** ENSURE EARLY RETURN is definitely here ***
				return nil, validationErr
			} else {
				logger.Printf("Optional arg '%s' is nil, accepting.", argSpec.Name)
				convertedArgs[i] = nil // Explicitly set nil for optional nil arg
				continue               // Skip further validation/coercion for nil optional arg
			}
		}

		// If we get here, argValue is not nil. Proceed with type validation/coercion.
		coercedValue, coerceErr := validateAndCoerceType(argValue, argSpec.Type, spec.Name, argSpec.Name) // Use local helper
		if coerceErr != nil {
			// *** FIXED: Do not wrap ErrValidationTypeMismatch again ***
			// Wrap ErrValidationTypeMismatch around the specific error
			// validationErr = fmt.Errorf("%w: %w", ErrValidationTypeMismatch, coerceErr)
			validationErr = coerceErr // Assign directly, as validateAndCoerceType already wraps
			// *** End Fix ***
			logger.Printf("Type Coercion Failed for arg '%s': %v", argSpec.Name, validationErr)
		} else {
			validatedValue = coercedValue
			logger.Printf("Type Coercion OK for arg '%s': Value=%#v (%T)", argSpec.Name, validatedValue, validatedValue)
		}

		// Perform Tool-Specific Validation *if* type coercion passed
		if validationErr == nil {
			// Add tool-specific checks here if needed, potentially setting validationErr
		}

		// Handle final error check for this argument *before assigning*
		if validationErr != nil {
			return nil, validationErr // Return the first error encountered
		}

		convertedArgs[i] = validatedValue
		logger.Printf("Successfully validated Arg %d ('%s'). Stored: %#v (%T)", i, argSpec.Name, convertedArgs[i], convertedArgs[i])

	} // End loop

	// Log any extra args provided beyond the spec
	if numRawArgs > len(spec.Args) {
		logger.Printf("Ignoring %d extra raw argument(s) provided.", numRawArgs-len(spec.Args))
	}

	logger.Printf("Validation successful for all args. Returning: %#v", convertedArgs)
	return convertedArgs, nil
}

// validateAndCoerceType - Helper within this file. Returns defined errors on failure.
func validateAndCoerceType(rawValue interface{}, expectedType ArgType, toolName, argName string) (interface{}, error) {
	if rawValue == nil {
		// Should have been caught earlier if required, okay if optional.
		return nil, nil
	}
	var finalValue interface{}
	var err error
	ok := false
	switch expectedType {
	case ArgTypeString:
		finalValue, ok = rawValue.(string)
		if !ok {
			err = fmt.Errorf("expected string, got %T", rawValue)
		}
	case ArgTypeInt:
		finalValue, ok = toInt64(rawValue) // Use helper from evaluation_helpers.go
		if !ok {
			err = fmt.Errorf("value %v (%T) cannot be converted to int (int64)", rawValue, rawValue)
		}
	case ArgTypeFloat:
		finalValue, ok = toFloat64(rawValue) // Use helper from evaluation_helpers.go
		if !ok {
			err = fmt.Errorf("value %v (%T) cannot be converted to float (float64)", rawValue, rawValue)
		}
	case ArgTypeBool:
		finalValue, ok = ConvertToBool(rawValue) // Use helper from helpers.go
		if !ok {
			err = fmt.Errorf("value %v (%T) cannot be converted to bool", rawValue, rawValue)
		}
	case ArgTypeSliceString:
		// *** FIXED: Use exported function ***
		finalValue, ok, err = ConvertToSliceOfString(rawValue) // Use exported helper from helpers.go
		if !ok && err == nil {
			err = fmt.Errorf("expected slice of strings, got %T", rawValue)
		}
	case ArgTypeSliceAny:
		finalValue, ok, err = convertToSliceOfAny(rawValue) // Use helper from helpers.go
		if !ok && err == nil {
			err = fmt.Errorf("expected a slice (list), got %T", rawValue)
		}
	case ArgTypeAny:
		finalValue, ok = rawValue, true
	default:
		err = fmt.Errorf("%w: unknown expected type '%s' for tool '%s' arg '%s'", ErrInternalTool, expectedType, toolName, argName)
		ok = false
	}
	// Check for errors or failed conversions
	if err != nil || !ok {
		finalErrMsg := fmt.Sprintf("argument '%s' of tool '%s'", argName, toolName)
		if err != nil {
			// *** Wrap ErrValidationTypeMismatch here ***
			return nil, fmt.Errorf("%w: %s: %w", ErrValidationTypeMismatch, finalErrMsg, err)
		} else {
			// *** Wrap ErrValidationTypeMismatch here ***
			return nil, fmt.Errorf("%w: %s: expected %s, got %T", ErrValidationTypeMismatch, finalErrMsg, expectedType, rawValue)
		}
	}
	return finalValue, nil
}
