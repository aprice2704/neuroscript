// neuroscript/pkg/core/tools_validation.go
package core

import (
	"fmt"
	"strings"
)

// ValidateAndConvertArgs checks if the provided raw arguments match the tool's
// specification and attempts to convert them to the expected Go types.
// Made type checking stricter.
func ValidateAndConvertArgs(spec ToolSpec, rawArgs []interface{}) ([]interface{}, error) {
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
		return nil, fmt.Errorf("tool '%s' expected %s arguments, but received %d", spec.Name, expectedArgsStr, numRawArgs)
	}

	convertedArgs := make([]interface{}, numRawArgs)
	for i := 0; i < numRawArgs; i++ {
		argSpec := spec.Args[i]
		argValue := rawArgs[i]
		var conversionErr error = nil

		// Handle nil
		if argValue == nil {
			if argSpec.Required {
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): is required, but received nil", spec.Name, argSpec.Name, i)
			} else {
				convertedArgs[i] = nil
				continue
			}
		}

		if conversionErr == nil {
			switch argSpec.Type {
			case ArgTypeString:
				// --- Stricter String Check ---
				strVal, ok := argValue.(string)
				if ok {
					convertedArgs[i] = strVal
				} else {
					// Optionally allow conversion from number/bool? For now, require string.
					conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected string, but received type %T", spec.Name, argSpec.Name, i, argValue)
				}
				// --- End Stricter Check ---
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
				// (Bool conversion logic remains relatively lenient as before)
				boolVal, converted := argValue.(bool)
				if !converted {
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
					} else if intVal, isInt := toInt64(argValue); isInt {
						if intVal == 1 {
							boolVal = true
							converted = true
						} else if intVal == 0 {
							boolVal = false
							converted = true
						}
					} else if floatVal, isFloat := toFloat64(argValue); isFloat {
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
				// --- Stricter SliceString Check ---
				valStr, isSliceStr := argValue.([]string)
				valAny, isSliceAny := argValue.([]interface{})

				if isSliceStr {
					convertedArgs[i] = valStr
				} else if isSliceAny {
					// Convert []interface{} ONLY IF all elements are strings
					strSlice := make([]string, len(valAny))
					canConvert := true
					for j, item := range valAny {
						if itemStr, ok := item.(string); ok {
							strSlice[j] = itemStr
						} else {
							canConvert = false
							conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but element %d has incompatible type %T (expected string)", spec.Name, argSpec.Name, i, argSpec.Type, j, item)
							break // Exit inner loop on first non-string element
						}
					}
					if canConvert {
						convertedArgs[i] = strSlice
					}
					// If canConvert is false, conversionErr is already set
				} else {
					conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
				}
				// --- End Stricter Check ---
			case ArgTypeSliceAny:
				// --- SliceAny Check ---
				valAny, isSliceAny := argValue.([]interface{})
				valStr, isSliceStr := argValue.([]string) // Also accept []string

				if isSliceAny {
					convertedArgs[i] = valAny // Already correct interface{} slice
				} else if isSliceStr {
					// Convert []string to []interface{} for consistency
					anySlice := make([]interface{}, len(valStr))
					for j, s := range valStr {
						anySlice[j] = s
					}
					convertedArgs[i] = anySlice
				} else {
					conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s (e.g., list literal or []string), but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
				}
				// --- End SliceAny Check ---
			case ArgTypeAny:
				convertedArgs[i] = argValue // Accept any type
			default:
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): internal validation error - unknown expected type %s", spec.Name, argSpec.Name, i, argSpec.Type)
			}
		}

		if conversionErr != nil {
			return nil, conversionErr
		}
		if convertedArgs[i] == nil && argValue != nil && argSpec.Type != ArgTypeAny {
			return nil, fmt.Errorf("internal validation error: tool '%s' argument '%s' (index %d) - conversion logic failed for type %s from %T resulting in nil", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
		}
	}

	return convertedArgs, nil
}
