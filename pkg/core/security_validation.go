// filename: neuroscript/pkg/core/security_validation.go
package core

import (
	// Import errors package
	"fmt"
	"strings"
)

// validateArgumentsAgainstSpec performs detailed validation of raw arguments against the tool's spec.
func (sl *SecurityLayer) validateArgumentsAgainstSpec(toolSpec ToolSpec, rawArgs map[string]interface{}) (map[string]interface{}, error) {
	validatedArgs := make(map[string]interface{})

	for _, specArg := range toolSpec.Args {
		argName := specArg.Name
		rawValue, argProvided := rawArgs[argName]

		sl.logger.Printf("[SEC VALIDATE] Checking arg '%s': Provided=%t, Required=%t, SpecType=%s", argName, argProvided, specArg.Required, specArg.Type)

		if !argProvided {
			if specArg.Required {
				// *** FIXED: Use Sentinel Error + Wrapping ***
				err := fmt.Errorf("required argument %q missing for tool %q: %w", argName, toolSpec.Name, ErrMissingArgument)
				sl.logger.Printf("[SEC VALIDATE] DENIED: %v", err)
				return nil, err
			}
			sl.logger.Printf("[SEC VALIDATE] Optional arg '%s' not provided.", argName)
			continue
		}

		var validatedValue interface{}
		var validationError error // Holds sentinel errors primarily

		// a) Basic Content Checks (e.g., null bytes)
		if strVal, ok := rawValue.(string); ok {
			if strings.Contains(strVal, "\x00") {
				// *** FIXED: Use Sentinel Error ***
				validationError = ErrNullByteInArgument
			}
			// Add other basic checks (length, patterns) here if needed, assigning appropriate sentinel errors
		}
		// Check if basic content validation failed
		if validationError != nil {
			// *** FIXED: Wrap Sentinel Error ***
			err := fmt.Errorf("content validation failed for argument %q: %w", argName, validationError)
			sl.logger.Printf("[SEC VALIDATE] DENIED: %v", err)
			return nil, err
		}

		// b) Type Checking & Coercion
		validatedValue, validationError = sl.validateAndCoerceType(rawValue, specArg.Type, toolSpec.Name, argName)
		if validationError != nil {
			// Error from validateAndCoerceType should already be properly formatted/wrapped
			sl.logger.Printf("[SEC VALIDATE] DENIED (Type Coercion): %v", validationError)
			// *** FIXED: Wrap type validation error with context ***
			// Wrap the already wrapped error coming from validateAndCoerceType
			return nil, fmt.Errorf("type validation failed for argument %q of tool %q: %w", argName, toolSpec.Name, validationError)
		}

		// c) Tool-specific checks (Example: TOOL.Add)
		if toolSpec.Name == "TOOL.Add" && (argName == "num1" || argName == "num2") {
			if _, isNum := ToNumeric(validatedValue); !isNum {
				// *** FIXED: Use Sentinel Error ***
				validationError = ErrValidationArgValue // Or a more specific one if needed
			}
		}
		// Add other tool-specific checks here, assigning sentinel errors

		// Check if tool-specific validation failed
		if validationError != nil {
			// *** FIXED: Wrap Sentinel Error ***
			err := fmt.Errorf("tool-specific validation failed for argument %q of tool %q: %w", argName, toolSpec.Name, validationError)
			sl.logger.Printf("[SEC VALIDATE] DENIED (Tool Specific Check): %v", err)
			return nil, err
		}

		// d) Path Sandboxing
		isPathArg := (toolSpec.Name == "TOOL.ReadFile" && argName == "filepath") ||
			(toolSpec.Name == "TOOL.WriteFile" && argName == "filepath") ||
			(toolSpec.Name == "TOOL.ListDirectory" && argName == "path") ||
			(toolSpec.Name == "TOOL.GitAdd" && argName == "filepath") ||
			(toolSpec.Name == "TOOL.GoCheck" && argName == "target") ||
			(toolSpec.Name == "TOOL.GoBuild" && argName == "target") ||
			(toolSpec.Name == "TOOL.LineCountFile" && argName == "filepath")

		if isPathArg && specArg.Type == ArgTypeString {
			pathStr, _ := validatedValue.(string)
			// SecureFilePath performs sandboxing and returns wrapped sentinel errors (ErrPathViolation, ErrNullByteInArgument)
			_, pathErr := SecureFilePath(pathStr, sl.sandboxRoot)
			if pathErr != nil {
				// Wrap the error from SecureFilePath with context
				// *** FIXED: Wrap returned error ***
				err := fmt.Errorf("sandbox validation failed for path argument %q (%q) relative to root %q: %w", argName, pathStr, sl.sandboxRoot, pathErr)
				sl.logger.Printf("[SEC VALIDATE] DENIED (Sandbox): %v", err)
				return nil, err
			}
			// Store the validated *relative* path string back.
			validatedValue = pathStr
			sl.logger.Printf("[SEC VALIDATE] Path argument '%s' (%q) validated successfully within sandbox %q.", argName, pathStr, sl.sandboxRoot)
		}

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
			// Potentially return an error here if unexpected args are strictly disallowed
			// return nil, fmt.Errorf("unexpected argument %q provided for tool %q: %w", rawArgName, toolSpec.Name, ErrInvalidArgument)
		}
	}

	return validatedArgs, nil
}

// validateAndCoerceType checks if the rawValue matches the expected ArgType and attempts coercion.
// Returns wrapped sentinel errors on failure.
func (sl *SecurityLayer) validateAndCoerceType(rawValue interface{}, expectedType ArgType, toolName, argName string) (interface{}, error) {
	var validatedValue interface{}
	var ok bool
	var typeErr error // Holds the specific sentinel error for type mismatch

	switch expectedType {
	case ArgTypeString:
		validatedValue, ok = rawValue.(string)
		if !ok {
			typeErr = ErrValidationTypeMismatch
		}
	case ArgTypeInt:
		validatedValue, ok = toInt64(rawValue)
		if !ok {
			typeErr = ErrValidationTypeMismatch
		}
	case ArgTypeFloat:
		validatedValue, ok = toFloat64(rawValue)
		if !ok {
			typeErr = ErrValidationTypeMismatch
		}
	case ArgTypeBool:
		validatedValue, ok = ConvertToBool(rawValue)
		if !ok {
			typeErr = ErrValidationTypeMismatch
		}
	case ArgTypeSliceString:
		var convertErr error
		validatedValue, ok, convertErr = ConvertToSliceOfString(rawValue)
		if convertErr != nil {
			// Wrap the underlying conversion error if it exists
			// *** FIXED: Use specific sentinel error ***
			return nil, fmt.Errorf("failed converting to slice of strings: %w", convertErr)
		}
		if !ok { // If conversion didn't error but still failed type check
			typeErr = ErrValidationTypeMismatch
		}
	case ArgTypeSliceAny:
		var convertErr error
		validatedValue, ok, convertErr = convertToSliceOfAny(rawValue)
		if convertErr != nil {
			// *** FIXED: Use specific sentinel error ***
			return nil, fmt.Errorf("failed converting to slice: %w", convertErr)
		}
		if !ok {
			typeErr = ErrValidationTypeMismatch
		}
	case ArgTypeAny:
		validatedValue, ok = rawValue, true // Accept any type
	default:
		// Use internal error for unknown expected type
		typeErr = ErrInternalTool // Or maybe ErrInvalidArgument? Let's stick with InternalTool for now.
		ok = false
	}

	// Check if validation failed within the switch block
	if typeErr != nil || !ok {
		// Wrap the specific sentinel error (typeErr) with context
		// *** FIXED: Wrap Sentinel Error (typeErr) ***
		// Adding %v to show raw value for better debugging
		return nil, fmt.Errorf("expected %s, got %T (%v): %w", expectedType, rawValue, rawValue, typeErr)
	}

	return validatedValue, nil
}
