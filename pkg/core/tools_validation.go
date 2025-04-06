// neuroscript/pkg/core/tools_validation.go
package core

import (
	"fmt"
	"strings"
)

// ValidateAndConvertArgs checks if the provided raw arguments match the tool's
// specification and attempts to convert them to the expected Go types (int64, float64, bool, string, []string, []interface{}).
// It uses the ArgSpec defined for the tool.
func ValidateAndConvertArgs(spec ToolSpec, rawArgs []interface{}) ([]interface{}, error) {
	minRequiredArgs := 0
	for _, argSpec := range spec.Args {
		if argSpec.Required {
			minRequiredArgs++
		}
	}
	maxExpectedArgs := len(spec.Args)
	numRawArgs := len(rawArgs)

	// Check argument count against required and maximum allowed
	if numRawArgs < minRequiredArgs || numRawArgs > maxExpectedArgs {
		expectedArgsStr := ""
		if minRequiredArgs == maxExpectedArgs {
			expectedArgsStr = fmt.Sprintf("exactly %d", minRequiredArgs)
		} else if maxExpectedArgs > 0 { // Handle cases with only optional args
			expectedArgsStr = fmt.Sprintf("between %d and %d", minRequiredArgs, maxExpectedArgs)
		} else {
			expectedArgsStr = "exactly 0" // If no args are defined
		}
		// More specific message if only optional args are missing
		if numRawArgs == 0 && minRequiredArgs == 0 && maxExpectedArgs > 0 {
			expectedArgsStr = fmt.Sprintf("up to %d optional", maxExpectedArgs)
		}
		return nil, fmt.Errorf("tool '%s' expected %s arguments, but received %d", spec.Name, expectedArgsStr, numRawArgs)
	}

	convertedArgs := make([]interface{}, numRawArgs)
	for i := 0; i < numRawArgs; i++ {
		argSpec := spec.Args[i]
		argValue := rawArgs[i]
		var conversionErr error = nil

		// Handle nil gracefully - if arg is not required, nil is okay unless a specific type conversion fails below.
		if argValue == nil {
			if argSpec.Required {
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): is required, but received nil", spec.Name, argSpec.Name, i)
			} else {
				convertedArgs[i] = nil // Explicitly set nil for optional arg
				continue               // Skip further conversion for nil optional args
			}
		}

		if conversionErr == nil { // Only proceed if arg is non-nil or required nil check passed
			switch argSpec.Type {
			case ArgTypeString:
				// *** Lenient conversion using fmt.Sprintf ***
				convertedArgs[i] = fmt.Sprintf("%v", argValue)
			case ArgTypeInt:
				intVal, converted := toInt64(argValue) // Use helper
				if converted {
					convertedArgs[i] = intVal
				} else {
					conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): value %v (%T) cannot be converted to %s (int64)", spec.Name, argSpec.Name, i, argValue, argValue, argSpec.Type)
				}
			case ArgTypeFloat:
				floatVal, converted := toFloat64(argValue) // Use helper
				if converted {
					convertedArgs[i] = floatVal
				} else {
					conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): value %v (%T) cannot be converted to %s (float64)", spec.Name, argSpec.Name, i, argValue, argValue, argSpec.Type)
				}
			case ArgTypeBool:
				boolVal, converted := argValue.(bool)
				if !converted {
					// Try conversion if not already bool
					strVal, isStr := argValue.(string)
					if isStr {
						lowerV := strings.ToLower(strVal)
						if lowerV == "true" || lowerV == "1" {
							boolVal = true
							converted = true
						} else if lowerV == "false" || lowerV == "0" {
							boolVal = false
							converted = true
						}
					} else if intVal, isInt := toInt64(argValue); isInt { // Allow 0/1 int
						if intVal == 1 {
							boolVal = true
							converted = true
						} else if intVal == 0 {
							boolVal = false
							converted = true
						}
					} else if floatVal, isFloat := toFloat64(argValue); isFloat { // Allow 0.0/1.0 float
						if floatVal == 1.0 {
							boolVal = true
							converted = true
						} else if floatVal == 0.0 {
							boolVal = false
							converted = true
						}
					}
				}

				if converted {
					convertedArgs[i] = boolVal
				} else {
					conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): value %v (%T) cannot be converted to %s", spec.Name, argSpec.Name, i, argValue, argValue, argSpec.Type)
				}
			case ArgTypeSliceString:
				// Expect []string directly OR []interface{} containing only strings/convertible
				valStr, isSliceStr := argValue.([]string)
				valAny, isSliceAny := argValue.([]interface{})

				if isSliceStr {
					convertedArgs[i] = valStr // Already correct type
				} else if isSliceAny {
					// Attempt conversion from []interface{} to []string (lenient)
					strSlice := make([]string, len(valAny))
					for j, item := range valAny {
						if item == nil { // Handle nil elements
							strSlice[j] = "" // Convert nil to empty string
						} else {
							// Convert non-strings using fmt.Sprintf
							strSlice[j] = fmt.Sprintf("%v", item)
						}
					}
					convertedArgs[i] = strSlice
				} else {
					conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
				}
			case ArgTypeSliceAny:
				// Accept []interface{} or []string
				valAny, isSliceAny := argValue.([]interface{})
				valStr, isSliceStr := argValue.([]string)

				if isSliceAny {
					convertedArgs[i] = valAny
				} else if isSliceStr {
					// Convert []string to []interface{}
					anySlice := make([]interface{}, len(valStr))
					for j, s := range valStr {
						anySlice[j] = s
					}
					convertedArgs[i] = anySlice
				} else {
					conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s (e.g., list literal or []string), but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
				}
			case ArgTypeAny:
				convertedArgs[i] = argValue // Accept any type without conversion
			default:
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): internal validation error - unknown expected type %s", spec.Name, argSpec.Name, i, argSpec.Type)
			} // End switch
		} // End if conversionErr == nil

		// Check for errors accumulated during conversion attempts
		if conversionErr != nil {
			return nil, conversionErr
		}

		// Defensive check: Ensure something was assigned if input wasn't nil
		if convertedArgs[i] == nil && argValue != nil && argSpec.Type != ArgTypeAny {
			// This might indicate a logic error in the conversion above
			return nil, fmt.Errorf("internal validation error: tool '%s' argument '%s' (index %d) - conversion logic failed for type %s from %T resulting in nil", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
		}
	} // End for loop

	return convertedArgs, nil // Return success if loop completes
}
