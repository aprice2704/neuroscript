// filename: pkg/core/security_validation.go
package core

import (
	"fmt"
	"strings"
)

// validateArgumentsAgainstSpec performs detailed validation of raw arguments against the tool's spec.
// This is called by SecurityLayer.ValidateToolCall.
func (sl *SecurityLayer) validateArgumentsAgainstSpec(toolSpec ToolSpec, rawArgs map[string]interface{}) (map[string]interface{}, error) {
	validatedArgs := make(map[string]interface{})

	// Loop through arguments DEFINED in the tool's specification
	for _, specArg := range toolSpec.Args {
		argName := specArg.Name
		rawValue, argProvided := rawArgs[argName]

		sl.logger.Printf("[SEC VALIDATE] Checking arg '%s': Provided=%t, Required=%t, SpecType=%s", argName, argProvided, specArg.Required, specArg.Type)

		if !argProvided {
			if specArg.Required {
				return nil, fmt.Errorf("required argument '%s' missing for tool '%s'", argName, toolSpec.Name)
			}
			sl.logger.Printf("[SEC VALIDATE] Optional arg '%s' not provided.", argName)
			continue
		}

		// --- Argument Validation & Sanitization Logic ---
		var validatedValue interface{}
		var validationError error

		// a) Basic Content Checks
		if strVal, ok := rawValue.(string); ok {
			if strings.Contains(strVal, "\x00") {
				validationError = fmt.Errorf("argument '%s' contains null byte", argName)
			}
			if len(strVal) > 8192 {
				validationError = fmt.Errorf("argument '%s' exceeds maximum length limit (8192)", argName)
			}
			// TODO: Add more content checks
		}
		if validationError != nil {
			return nil, fmt.Errorf("content validation failed for argument '%s': %w", argName, validationError)
		}

		// b) Type Checking & Coercion
		validatedValue, validationError = sl.validateAndCoerceType(rawValue, specArg.Type, toolSpec.Name, argName)
		if validationError != nil {
			return nil, fmt.Errorf("type validation failed for argument '%s': %w", argName, validationError)
		}

		// c) Path Sandboxing (Applied specifically based on Tool and Arg Name)
		//    Refine this logic - maybe add metadata to ArgSpec?
		isPathArg := (toolSpec.Name == "TOOL.ReadFile" && argName == "filepath") ||
			(toolSpec.Name == "TOOL.WriteFile" && argName == "filepath") ||
			(toolSpec.Name == "TOOL.ListDirectory" && argName == "path") ||
			(toolSpec.Name == "TOOL.GitAdd" && argName == "filepath") // Add other path args here

		if isPathArg {
			pathStr, ok := validatedValue.(string) // Use the type-coerced value if applicable
			if !ok {
				// Should have been caught by type checking, but double-check
				return nil, fmt.Errorf("internal validation error: path argument '%s' for tool '%s' is not a string after type check (%T)", argName, toolSpec.Name, validatedValue)
			}

			// Apply SecureFilePath check using the layer's sandboxRoot
			_, pathErr := SecureFilePath(pathStr, sl.sandboxRoot)
			if pathErr != nil {
				errMsg := fmt.Sprintf("sandbox validation failed for path argument '%s' (%q): %v", argName, pathStr, pathErr)
				sl.logger.Printf("[SEC] DENIED (Sandbox): %s", errMsg)
				return nil, fmt.Errorf(errMsg) // Return specific security error
			}
			// If validation passes, validatedValue already holds the pathStr
			sl.logger.Printf("[SEC VALIDATE] Path argument '%s' (%q) validated successfully within sandbox %q.", argName, pathStr, sl.sandboxRoot)
		}

		// Store the validated (and potentially coerced/sanitized) value
		validatedArgs[argName] = validatedValue
		sl.logger.Printf("[SEC VALIDATE] Arg '%s' validated successfully. Value: %v (%T)", argName, validatedValue, validatedValue)

	} // End loop through spec args

	// Check for unexpected arguments
	for rawArgName := range rawArgs {
		foundInSpec := false
		for _, specArg := range toolSpec.Args {
			if rawArgName == specArg.Name {
				foundInSpec = true
				break
			}
		}
		if !foundInSpec {
			sl.logger.Printf("[WARN SEC VALIDATE] Tool '%s' called with unexpected argument '%s'. Ignoring.", toolSpec.Name, rawArgName)
		}
	}

	return validatedArgs, nil
}

// validateAndCoerceType checks if the rawValue matches the expected ArgType and attempts coercion.
func (sl *SecurityLayer) validateAndCoerceType(rawValue interface{}, expectedType ArgType, toolName, argName string) (interface{}, error) {
	var validatedValue interface{}
	var ok bool
	var err error // Use a separate error variable for type conversion issues

	// Use helpers for consistent conversion/checking logic
	switch expectedType {
	case ArgTypeString:
		strVal, isString := rawValue.(string)
		if !isString {
			err = fmt.Errorf("expected string, got %T", rawValue)
		} else {
			validatedValue, ok = strVal, true
		}
	case ArgTypeInt:
		validatedValue, ok = toInt64(rawValue)
		if !ok {
			err = fmt.Errorf("expected integer, got %T (%v)", rawValue, rawValue)
		}
	case ArgTypeFloat:
		validatedValue, ok = toFloat64(rawValue)
		if !ok {
			err = fmt.Errorf("expected number, got %T (%v)", rawValue, rawValue)
		}
	case ArgTypeBool:
		validatedValue, ok = ConvertToBool(rawValue)
		if !ok {
			err = fmt.Errorf("expected boolean, got %T (%v)", rawValue, rawValue)
		}
	case ArgTypeSliceString:
		validatedValue, ok, err = convertToSliceOfString(rawValue) // Helper returns potential conversion error
		// ok reflects if *any* slice was returned, err indicates specific conversion issues
		if !ok && err == nil {
			err = fmt.Errorf("expected slice of strings, got %T", rawValue)
		} // Assign generic error if helper didn't set one
	case ArgTypeSliceAny:
		validatedValue, ok, err = convertToSliceOfAny(rawValue)
		if !ok && err == nil {
			err = fmt.Errorf("expected a slice, got %T", rawValue)
		}
	case ArgTypeAny:
		validatedValue, ok = rawValue, true // Accept anything
	default:
		err = fmt.Errorf("internal error: unknown expected type '%s'", expectedType)
		ok = false
	}

	// Combine potential errors from helpers or direct checks
	if err != nil || !ok {
		finalErrMsg := fmt.Sprintf("type validation failed for argument '%s' of tool '%s'", argName, toolName)
		if err != nil {
			finalErrMsg = fmt.Sprintf("%s: %v", finalErrMsg, err) // Include specific conversion error
		} else {
			// Generic message if ok is false but no specific error was set
			finalErrMsg = fmt.Sprintf("%s: expected %s, got %T", finalErrMsg, expectedType, rawValue)
		}
		return nil, fmt.Errorf(finalErrMsg)
	}

	return validatedValue, nil
}
