// pkg/tools/tools.go
package core

import (
	"fmt"
	"strconv"
	"strings"
)

// --- Tool Argument Specification --- (No changes below here needed for interface)
type ArgType string

const (
	ArgTypeString      ArgType = "string"
	ArgTypeInt         ArgType = "int"   // Represents Go's int64
	ArgTypeFloat       ArgType = "float" // Represents Go's float64
	ArgTypeBool        ArgType = "bool"
	ArgTypeSliceString ArgType = "slice_string" // Represents Go's []string
	ArgTypeSliceAny    ArgType = "slice_any"    // Represents Go's []interface{}
	ArgTypeAny         ArgType = "any"
)

type ArgSpec struct {
	Name        string
	Type        ArgType
	Description string
	Required    bool
}
type ToolSpec struct {
	Name        string
	Description string
	Args        []ArgSpec
	ReturnType  ArgType
}

// --- Tool Function Implementation ---
// --- CHANGED: Use InterpreterContext interface ---
type ToolFunc func(interpreter *Interpreter, args []interface{}) (interface{}, error)

// --- Tool Implementation Registry ---
type ToolImplementation struct {
	Spec ToolSpec
	Func ToolFunc
}
type ToolRegistry struct {
	tools map[string]ToolImplementation
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ToolImplementation),
	}
}
func (tr *ToolRegistry) RegisterTool(impl ToolImplementation) error {
	if _, exists := tr.tools[impl.Spec.Name]; exists {
		return fmt.Errorf("tool '%s' already registered", impl.Spec.Name)
	}
	if impl.Func == nil {
		return fmt.Errorf("tool '%s' registration is missing implementation function", impl.Spec.Name)
	}
	tr.tools[impl.Spec.Name] = impl
	return nil
}
func (tr *ToolRegistry) GetTool(name string) (ToolImplementation, bool) {
	impl, found := tr.tools[name]
	return impl, found
}

// --- Argument Validation/Conversion Helper ---
// (No changes needed in ValidateAndConvertArgs logic itself)
func ValidateAndConvertArgs(spec ToolSpec, rawArgs []interface{}) ([]interface{}, error) {
	minRequiredArgs := 0
	for _, argSpec := range spec.Args {
		if argSpec.Required {
			minRequiredArgs++
		}
	}
	maxExpectedArgs := len(spec.Args)
	numRawArgs := len(rawArgs)

	// Check Minimum Requirement
	if numRawArgs < minRequiredArgs {
		expectedArgsStr := ""
		if maxExpectedArgs == minRequiredArgs {
			expectedArgsStr = fmt.Sprintf("exactly %d", minRequiredArgs)
		} else {
			expectedArgsStr = fmt.Sprintf("at least %d", minRequiredArgs)
		}
		return nil, fmt.Errorf("tool '%s' expected %s arguments, but received %d", spec.Name, expectedArgsStr, numRawArgs)
	}

	// Check Maximum Requirement
	if numRawArgs > maxExpectedArgs {
		expectedArgsStr := ""
		if maxExpectedArgs == minRequiredArgs {
			expectedArgsStr = fmt.Sprintf("exactly %d", maxExpectedArgs)
		} else {
			expectedArgsStr = fmt.Sprintf("at most %d", maxExpectedArgs)
		}
		return nil, fmt.Errorf("tool '%s' expected %s arguments, but received %d", spec.Name, expectedArgsStr, numRawArgs)
	}

	// Conversion logic
	convertedArgs := make([]interface{}, numRawArgs)
	for i := 0; i < numRawArgs; i++ {
		if i >= len(spec.Args) {
			return nil, fmt.Errorf("internal error: trying to process arg index %d beyond spec length %d for tool '%s'", i, len(spec.Args), spec.Name)
		}
		argSpec := spec.Args[i]
		argValue := rawArgs[i]
		var conversionErr error
		conversionSuccessful := false

		switch argSpec.Type {
		case ArgTypeString:
			strVal, ok := argValue.(string)
			if ok {
				convertedArgs[i] = strVal
				conversionSuccessful = true
			} else {
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but received type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
		case ArgTypeInt:
			var intVal int64
			converted := false
			switch v := argValue.(type) {
			case int64:
				intVal = v
				converted = true
			case int:
				intVal = int64(v)
				converted = true
			case int32:
				intVal = int64(v)
				converted = true
			case float64:
				if v == float64(int64(v)) {
					intVal = int64(v)
					converted = true
				}
			case float32:
				if v == float32(int64(v)) {
					intVal = int64(v)
					converted = true
				}
			case string:
				parsedVal, err := strconv.ParseInt(v, 10, 64)
				if err == nil {
					intVal = parsedVal
					converted = true
				}
			}
			if converted {
				convertedArgs[i] = intVal
				conversionSuccessful = true
			} else {
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but received %T (%v) which cannot be converted to int", spec.Name, argSpec.Name, i, argSpec.Type, argValue, argValue)
			}
		case ArgTypeFloat:
			var floatVal float64
			converted := false
			switch v := argValue.(type) {
			case float64:
				floatVal = v
				converted = true
			case float32:
				floatVal = float64(v)
				converted = true
			case int64:
				floatVal = float64(v)
				converted = true
			case int:
				floatVal = float64(v)
				converted = true
			case int32:
				floatVal = float64(v)
				converted = true
			case string:
				parsedVal, err := strconv.ParseFloat(v, 64)
				if err == nil {
					floatVal = parsedVal
					converted = true
				}
			}
			if converted {
				convertedArgs[i] = floatVal
				conversionSuccessful = true
			} else {
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but received %T (%v) which cannot be converted to float", spec.Name, argSpec.Name, i, argSpec.Type, argValue, argValue)
			}
		case ArgTypeBool:
			var boolVal bool
			converted := false
			switch v := argValue.(type) {
			case bool:
				boolVal = v
				converted = true
			case string:
				lowerV := strings.ToLower(v)
				if lowerV == "true" {
					boolVal = true
					converted = true
				} else if lowerV == "false" {
					boolVal = false
					converted = true
				}
			}
			if converted {
				convertedArgs[i] = boolVal
				conversionSuccessful = true
			} else {
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but received %T (%v) which cannot be converted to bool", spec.Name, argSpec.Name, i, argSpec.Type, argValue, argValue)
			}
		case ArgTypeSliceString:
			converted := false
			switch v := argValue.(type) {
			case []string:
				convertedArgs[i] = v
				converted = true
			case []interface{}:
				strSlice := make([]string, len(v))
				canConvertAll := true
				for idx, item := range v {
					if strItem, ok := item.(string); ok {
						strSlice[idx] = strItem
					} else {
						// If conversion isn't straightforward, maybe use fmt.Sprintf
						// but flag it might not be the desired behavior.
						// For simplicity here, we'll require elements to be strings
						// or handle the conversion error. Let's require strings.
						canConvertAll = false
						conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected slice of strings, but found element of type %T at index %d", spec.Name, argSpec.Name, i, item, idx)
						break
					}
				}
				if canConvertAll {
					convertedArgs[i] = strSlice
					converted = true
				}

			}
			if converted {
				conversionSuccessful = true
			} else if conversionErr == nil { // Only set default error if no specific one occurred
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s (e.g., list literal of strings), but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
		case ArgTypeSliceAny:
			// Check if it's already []interface{} or []string (common cases)
			_, isSliceAny := argValue.([]interface{})
			_, isSliceStr := argValue.([]string)

			if isSliceAny || isSliceStr {
				convertedArgs[i] = argValue // Pass through
				conversionSuccessful = true
			} else {
				// If not directly assignable, reject
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s (e.g., list literal), but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
		case ArgTypeAny:
			convertedArgs[i] = argValue
			conversionSuccessful = true
		default:
			conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): internal validation error - unknown expected type %s", spec.Name, argSpec.Name, i, argSpec.Type)
		}

		// Final check for the loop iteration
		if conversionErr != nil {
			return nil, conversionErr
		}
		if !conversionSuccessful {
			// This case should ideally not be reached if logic above is correct
			return nil, fmt.Errorf("internal validation error: tool '%s' argument '%s' (index %d) - conversion failed unexpectedly for type %s from %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
		}
	}

	return convertedArgs, nil
}
