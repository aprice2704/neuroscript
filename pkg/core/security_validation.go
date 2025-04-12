// filename: neuroscript/pkg/core/security_validation.go
package core

import (
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
				return nil, fmt.Errorf("required argument '%s' missing for tool '%s'", argName, toolSpec.Name)
			}
			sl.logger.Printf("[SEC VALIDATE] Optional arg '%s' not provided.", argName)
			continue
		}

		var validatedValue interface{}
		var validationError error

		// a) Basic Content Checks (Example)
		if strVal, ok := rawValue.(string); ok {
			if strings.Contains(strVal, "\x00") {
				validationError = fmt.Errorf("argument '%s' contains null byte", argName)
			}
			// Add length checks or other pattern checks if needed
		}
		if validationError != nil {
			return nil, fmt.Errorf("content validation failed for argument '%s': %w", argName, validationError)
		}

		// b) Type Checking & Coercion
		validatedValue, validationError = sl.validateAndCoerceType(rawValue, specArg.Type, toolSpec.Name, argName)
		if validationError != nil {
			return nil, fmt.Errorf("type validation failed for argument '%s': %w", argName, validationError)
		}

		// c) Tool-specific checks (e.g., numeric check for TOOL.Add)
		if toolSpec.Name == "TOOL.Add" && (argName == "num1" || argName == "num2") {
			if _, isNum := ToNumeric(validatedValue); !isNum {
				validationError = fmt.Errorf("argument '%s' for TOOL.Add must be numeric, got %T", argName, validatedValue)
			}
		}
		// Add other tool-specific checks here if needed

		if validationError != nil { // Check again after tool-specific validation
			sl.logger.Printf("[SEC] DENIED (Tool Specific Check): %v", validationError)
			return nil, validationError
		}

		// d) Path Sandboxing
		isPathArg := (toolSpec.Name == "TOOL.ReadFile" && argName == "filepath") ||
			(toolSpec.Name == "TOOL.WriteFile" && argName == "filepath") ||
			(toolSpec.Name == "TOOL.ListDirectory" && argName == "path") ||
			(toolSpec.Name == "TOOL.GitAdd" && argName == "filepath") ||
			(toolSpec.Name == "TOOL.GoCheck" && argName == "target") || // Added GoCheck
			(toolSpec.Name == "TOOL.GoBuild" && argName == "target") || // Added GoBuild
			(toolSpec.Name == "TOOL.LineCountFile" && argName == "filepath") // Added LineCountFile

		if isPathArg && specArg.Type == ArgTypeString {
			pathStr, _ := validatedValue.(string)
			// Use the sandboxRoot configured in the SecurityLayer
			_, pathErr := SecureFilePath(pathStr, sl.sandboxRoot)
			if pathErr != nil {
				errMsg := fmt.Sprintf("sandbox validation failed for path argument '%s' (%q) relative to root %q: %v", argName, pathStr, sl.sandboxRoot, pathErr)
				sl.logger.Printf("[SEC] DENIED (Sandbox): %s", errMsg)
				return nil, fmt.Errorf(errMsg) // Return the detailed error
			}
			// Store the validated *relative* path string back.
			// The actual tool function will resolve it again against CWD (which should be sandbox root in agent mode).
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
		}
	}

	return validatedArgs, nil
}

// validateAndCoerceType checks if the rawValue matches the expected ArgType and attempts coercion.
func (sl *SecurityLayer) validateAndCoerceType(rawValue interface{}, expectedType ArgType, toolName, argName string) (interface{}, error) {
	var validatedValue interface{}
	var ok bool
	var err error
	switch expectedType {
	case ArgTypeString:
		validatedValue, ok = rawValue.(string)
		if !ok {
			err = fmt.Errorf("expected string, got %T", rawValue)
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
		// *** FIXED: Use exported function ***
		validatedValue, ok, err = ConvertToSliceOfString(rawValue) // Use exported helper
		if !ok && err == nil {
			err = fmt.Errorf("expected slice of strings, got %T", rawValue)
		}
	case ArgTypeSliceAny:
		validatedValue, ok, err = convertToSliceOfAny(rawValue) // Can remain unexported if only used here
		if !ok && err == nil {
			err = fmt.Errorf("expected a slice, got %T", rawValue)
		}
	case ArgTypeAny:
		validatedValue, ok = rawValue, true
	default:
		err = fmt.Errorf("internal error: unknown expected type '%s'", expectedType)
		ok = false
	}
	if err != nil || !ok {
		finalErrMsg := fmt.Sprintf("type validation failed for argument '%s' of tool '%s'", argName, toolName)
		if err != nil {
			finalErrMsg = fmt.Sprintf("%s: %v", finalErrMsg, err)
		} else {
			finalErrMsg = fmt.Sprintf("%s: expected %s, got %T", finalErrMsg, expectedType, rawValue)
		}
		return nil, fmt.Errorf(finalErrMsg)
	}
	return validatedValue, nil
}
