// filename: pkg/core/tools_validation.go
package core

import (
	"fmt"
	"io" // Import io for discard logger
	"log"
	// No longer need math, strconv, reflect, strings here
)

// ValidateAndConvertArgs checks if the provided raw arguments match the tool's
// specification and attempts to convert them to the expected Go types.
// Relies on helpers from evaluation_helpers.go for type conversion/checking.
func ValidateAndConvertArgs(spec ToolSpec, rawArgs []interface{}) ([]interface{}, error) {
	// Use a discard logger by default, can be replaced if needed
	logger := log.New(io.Discard, "[VALIDATE DEBUG] ", log.Ltime|log.Lshortfile)
	// logger = log.New(os.Stderr, "[VALIDATE DEBUG] ", log.Ltime|log.Lshortfile) // Uncomment for debug

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
		return nil, err
	}

	convertedArgs := make([]interface{}, numRawArgs)
	for i := 0; i < numRawArgs; i++ {
		// --- Check bounds before accessing spec.Args ---
		if i >= len(spec.Args) {
			logger.Printf("Ignoring extra raw argument at index %d (no corresponding ArgSpec)", i)
			continue
		}
		// --- End bounds check ---

		argSpec := spec.Args[i]
		argValue := rawArgs[i]
		var validationErr error = nil
		var validatedValue interface{} // To hold the value after type coercion

		logger.Printf("Processing Arg %d ('%s'): Value=%#v (%T), SpecType=%s, Required=%t", i, argSpec.Name, argValue, argValue, argSpec.Type, argSpec.Required)

		// Handle nil
		if argValue == nil {
			if argSpec.Required {
				validationErr = fmt.Errorf("tool '%s' argument '%s' (index %d): is required, but received nil", spec.Name, argSpec.Name, i)
				logger.Printf("Required arg is nil: %v", validationErr)
			} else {
				logger.Printf("Optional arg '%s' is nil, accepting.", argSpec.Name)
				convertedArgs[i] = nil // Explicitly set nil for optional nil arg
				continue               // Skip further validation for nil optional arg
			}
		}

		// Perform type validation and coercion *if* not already handled by nil check
		if validationErr == nil {
			// Call the main type validation/coercion helper
			coercedValue, coerceErr := validateAndCoerceType(argValue, argSpec.Type, spec.Name, argSpec.Name)
			if coerceErr != nil {
				validationErr = coerceErr // Assign error from coercion
				logger.Printf("Initial Type Coercion Failed for arg '%s': %v", argSpec.Name, validationErr)
			} else {
				validatedValue = coercedValue // Store successfully coerced/validated value
				logger.Printf("Initial Type Coercion OK for arg '%s': Value=%#v (%T)", argSpec.Name, validatedValue, validatedValue)
			}
		}

		// Perform Tool-Specific Validation *if* initial validation/coercion passed
		if validationErr == nil {
			// Check if validatedValue is actually set (it might not be if coercion failed but didn't set validationErr somehow)
			if validatedValue == nil && argValue != nil {
				// If the original value wasn't nil, but validatedValue is, coercion likely failed implicitly.
				validationErr = fmt.Errorf("internal error: type coercion resulted in nil for non-nil input %T for argument '%s'", argValue, argSpec.Name)
				logger.Printf("Error: %v", validationErr)
			} else {
				// --- REMOVED TOOL.Add Specific Check ---
				// Now rely on ArgSpec.Type = ArgTypeFloat for TOOL.Add
				// --- END REMOVAL ---

				// --- Add other tool-specific checks here if needed ---
				// Example: Check if string arg matches a specific pattern for another tool
				// if spec.Name == "SomeOtherTool" && argSpec.Name == "specific_format_string" {
				//    strVal, _ := validatedValue.(string)
				//    if !regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, strVal) {
				// 	      validationErr = fmt.Errorf("argument '%s' requires YYYY-MM-DD format", argSpec.Name)
				//    }
				// }
			}
		}

		// Handle final error check
		if validationErr != nil {
			return nil, validationErr // Return the first error encountered
		}

		// Store the final validated (and potentially coerced/checked) value
		convertedArgs[i] = validatedValue
		logger.Printf("Successfully validated Arg %d ('%s'). Stored: %#v (%T)", i, argSpec.Name, convertedArgs[i], convertedArgs[i])

	} // End loop

	logger.Printf("Validation successful for all args. Returning: %#v", convertedArgs)
	return convertedArgs, nil
}

// validateAndCoerceType - Primary type validation/coercion logic.
// Relies on helpers from evaluation_helpers.go.
// This function ensures the value conforms to the expected ArgType.
func validateAndCoerceType(rawValue interface{}, expectedType ArgType, toolName, argName string) (interface{}, error) {

	if rawValue == nil {
		return nil, nil // Allow nil to pass through; required check is done earlier
	}

	var finalValue interface{}
	var err error
	ok := false // Used to track successful conversion

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
		finalValue, ok = ConvertToBool(rawValue) // Use helper from evaluation_helpers.go
		if !ok {
			err = fmt.Errorf("value %v (%T) cannot be converted to bool", rawValue, rawValue)
		}
	case ArgTypeSliceString:
		finalValue, ok, err = convertToSliceOfString(rawValue) // Use helper (returns error directly)
		if !ok && err == nil {                                 // If helper returned !ok but no specific error
			err = fmt.Errorf("expected slice of strings, got %T", rawValue)
		}
	case ArgTypeSliceAny:
		finalValue, ok, err = convertToSliceOfAny(rawValue) // Use helper (returns error directly)
		if !ok && err == nil {                              // If helper returned !ok but no specific error
			err = fmt.Errorf("expected a slice, got %T", rawValue)
		}
	case ArgTypeAny:
		// For ArgTypeAny, perform basic sanity checks but generally accept.
		finalValue, ok = rawValue, true
	default:
		err = fmt.Errorf("internal error: unknown expected type '%s'", expectedType)
		ok = false
	}

	// Check for errors or failed conversions
	if err != nil || !ok {
		// Construct a more user-friendly error message consistent with test output
		finalErrMsg := fmt.Sprintf("type validation failed for argument '%s' of tool '%s': ", argName, toolName)
		if err != nil {
			// Append the specific error from the helper if it exists
			finalErrMsg += err.Error()
		} else {
			// Default message if helper didn't provide specific error
			finalErrMsg += fmt.Sprintf("expected %s, got %T", expectedType, rawValue)
		}
		return nil, fmt.Errorf(finalErrMsg)
	}

	return finalValue, nil
}
