// pkg/tools/tools.go
package core

import (
	"fmt"
	"strconv"
	"strings"
)

// --- Tool Argument Specification ---
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
func ValidateAndConvertArgs(spec ToolSpec, rawArgs []interface{}) ([]interface{}, error) {
	minRequiredArgs := 0
	for _, argSpec := range spec.Args {
		if argSpec.Required {
			minRequiredArgs++
		}
	}
	maxExpectedArgs := len(spec.Args)
	numRawArgs := len(rawArgs)

	if numRawArgs < minRequiredArgs {
		expectedArgsStr := fmt.Sprintf("at least %d", minRequiredArgs)
		if maxExpectedArgs == minRequiredArgs {
			expectedArgsStr = fmt.Sprintf("exactly %d", minRequiredArgs)
		}
		return nil, fmt.Errorf("tool '%s' expected %s arguments, but received %d", spec.Name, expectedArgsStr, numRawArgs)
	}
	if numRawArgs > maxExpectedArgs {
		expectedArgsStr := fmt.Sprintf("at most %d", maxExpectedArgs)
		if maxExpectedArgs == minRequiredArgs {
			expectedArgsStr = fmt.Sprintf("exactly %d", maxExpectedArgs)
		}
		return nil, fmt.Errorf("tool '%s' expected %s arguments, but received %d", spec.Name, expectedArgsStr, numRawArgs)
	}

	convertedArgs := make([]interface{}, numRawArgs)
	for i := 0; i < numRawArgs; i++ {
		argSpec := spec.Args[i]
		argValue := rawArgs[i]
		var conversionErr error
		conversionSuccessful := false

		switch argSpec.Type {
		// *** REVERTED: Stricter check for string type ***
		case ArgTypeString:
			strVal, ok := argValue.(string)
			if ok {
				convertedArgs[i] = strVal
				conversionSuccessful = true
			} else {
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but received type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
		// *** END REVERT ***

		case ArgTypeInt:
			var intVal int64
			converted := false
			switch v := argValue.(type) {
			case int64:
				intVal = v
				converted = true
			case int: // Allow Go int type
				intVal = int64(v)
				converted = true
			case int32:
				intVal = int64(v)
				converted = true
			case float64: // Allow float if it represents a whole number
				if v == float64(int64(v)) {
					intVal = int64(v)
					converted = true
				}
			case float32: // Allow float if it represents a whole number
				if v == float32(int64(v)) {
					intVal = int64(v)
					converted = true
				}
			case string: // Allow numeric strings
				parsedVal, err := strconv.ParseInt(v, 10, 64)
				if err == nil {
					intVal = parsedVal
					converted = true
				} else {
					fVal, errF := strconv.ParseFloat(v, 64)
					if errF == nil && fVal == float64(int64(fVal)) {
						intVal = int64(fVal)
						converted = true
					}
				}
			case bool: // Convert bool to 0 or 1
				if v {
					intVal = 1
				} else {
					intVal = 0
				}
				converted = true
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
			case bool: // Convert bool to 0.0 or 1.0
				if v {
					floatVal = 1.0
				} else {
					floatVal = 0.0
				}
				converted = true
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
				if lowerV == "true" || lowerV == "1" {
					boolVal = true
					converted = true
				} else if lowerV == "false" || lowerV == "0" {
					boolVal = false
					converted = true
				}
			case int64:
				if v == 1 {
					boolVal = true
					converted = true
				} else if v == 0 {
					boolVal = false
					converted = true
				}
			case float64:
				if v == 1.0 {
					boolVal = true
					converted = true
				} else if v == 0.0 {
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
		case ArgTypeSliceString: // Still requires []string specifically
			_, ok := argValue.([]string)
			if ok {
				convertedArgs[i] = argValue
				conversionSuccessful = true
			} else {
				// Check if it's []interface{} and *all* elements are strings? (More complex, maybe not needed)
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
		case ArgTypeSliceAny: // Accepts []interface{} or []string
			_, isSliceAny := argValue.([]interface{})
			_, isSliceStr := argValue.([]string)
			if isSliceAny || isSliceStr {
				convertedArgs[i] = argValue
				conversionSuccessful = true
			} else {
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s (e.g., list literal), but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
		case ArgTypeAny:
			convertedArgs[i] = argValue
			conversionSuccessful = true
		default:
			conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): internal validation error - unknown expected type %s", spec.Name, argSpec.Name, i, argSpec.Type)
		}

		if conversionErr != nil {
			return nil, conversionErr
		}
		if !conversionSuccessful {
			return nil, fmt.Errorf("internal validation error: tool '%s' argument '%s' (index %d) - conversion failed unexpectedly for type %s from %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
		}
	}
	return convertedArgs, nil
}
