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

// --- Tool Specification ---

type ToolSpec struct {
	Name        string
	Description string
	Args        []ArgSpec
	ReturnType  ArgType // Consider adding validation for return type too
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
	// Check arity first
	expectedNumArgs := len(spec.Args)
	requiredArgs := 0
	optionalArgs := 0
	for _, argSpec := range spec.Args {
		if argSpec.Required {
			requiredArgs++
		} else {
			optionalArgs++
		}
	}

	if len(rawArgs) < requiredArgs || len(rawArgs) > expectedNumArgs {
		arityMsg := ""
		if optionalArgs == 0 {
			// Use "exactly" only if min and max required args are the same
			arityMsg = fmt.Sprintf("exactly %d", requiredArgs)
		} else {
			arityMsg = fmt.Sprintf("between %d and %d", requiredArgs, expectedNumArgs)
		}
		// Ensure the error message format matches the test expectation exactly
		return nil, fmt.Errorf("tool '%s' expected %s arguments, but received %d", spec.Name, arityMsg, len(rawArgs))
	}

	// Create slice for converted args, matching the number provided
	convertedArgs := make([]interface{}, len(rawArgs))

	// Iterate through the expected args based on the spec
	for i, argSpec := range spec.Args {
		// If this expected arg wasn't provided (must be optional)
		if i >= len(rawArgs) {
			continue // Skip conversion for unprovided optional args
		}

		argValue := rawArgs[i] // The raw evaluated value from the interpreter
		var conversionErr error
		conversionSuccessful := false // Flag to track if conversion succeeded for this arg

		switch argSpec.Type {
		case ArgTypeString:
			// --- MODIFIED: Stricter Check ---
			strVal, ok := argValue.(string)
			if ok {
				convertedArgs[i] = strVal // Pass the original string
				conversionSuccessful = true
			} else {
				// Error if the evaluated value isn't actually a string
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but received type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
			// --- End Modification ---

		case ArgTypeInt:
			var intVal int64 // Use int64 internally
			converted := false
			switch v := argValue.(type) {
			case int64:
				intVal = v
				converted = true // Already int64
			case int:
				intVal = int64(v)
				converted = true
			case int32:
				intVal = int64(v)
				converted = true
			// Allow conversion from float if precision is not lost
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
			// Allow conversion from string
			case string:
				parsedVal, err := strconv.ParseInt(v, 10, 64)
				if err == nil {
					intVal = parsedVal
					converted = true
				}
			}
			if converted {
				convertedArgs[i] = intVal // Store as int64
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
				converted = true // Already float64
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
				convertedArgs[i] = floatVal // Store as float64
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
			// Allow 0/1 conversion?
			case int, int32, int64:
				if fmt.Sprintf("%v", v) == "1" {
					boolVal = true
					converted = true
				} else if fmt.Sprintf("%v", v) == "0" {
					boolVal = false
					converted = true
				}
			case float32, float64:
				if fmt.Sprintf("%v", v) == "1.0" || fmt.Sprintf("%v", v) == "1" {
					boolVal = true
					converted = true
				} else if fmt.Sprintf("%v", v) == "0.0" || fmt.Sprintf("%v", v) == "0" {
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
			case []string: // Already the correct type
				convertedArgs[i] = v
				converted = true
			case []interface{}: // Convert from []interface{}
				strSlice := make([]string, len(v))
				for idx, item := range v {
					strSlice[idx] = fmt.Sprintf("%v", item)
				} // Convert each element to string
				convertedArgs[i] = strSlice
				converted = true
			}
			if converted {
				conversionSuccessful = true
			} else {
				// Error message should clearly state expectation vs reality
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}

		case ArgTypeSliceAny:
			// Check if it's a slice type we can reasonably pass
			if _, ok := argValue.([]interface{}); ok {
				convertedArgs[i] = argValue
				conversionSuccessful = true
			} else if _, ok := argValue.([]string); ok { // Also allow []string?
				convertedArgs[i] = argValue
				conversionSuccessful = true
			} // Add other specific slice types if needed
			if !conversionSuccessful {
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): expected %s, but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}

		case ArgTypeAny: // Accept anything without conversion
			convertedArgs[i] = argValue // No conversion needed
			conversionSuccessful = true

		default: // Unknown target type in spec
			conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): internal validation error - unknown expected type %s", spec.Name, argSpec.Name, i, argSpec.Type)
		} // End switch argSpec.Type

		// Check if an error occurred during this argument's processing
		if conversionErr != nil {
			return nil, conversionErr // Return the specific error
		}

		// Defensive check: If no error occurred, conversionSuccessful should be true.
		if !conversionSuccessful && conversionErr == nil {
			// This path indicates a logic error in the switch statement above
			return nil, fmt.Errorf("internal validation error: tool '%s' argument '%s' (index %d) - no conversion success or error for type %s from %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
		}

	} // End for loop over spec.Args

	// Return only the converted args slice, which matches the length of rawArgs
	return convertedArgs, nil
}
