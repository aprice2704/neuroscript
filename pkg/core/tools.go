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
	ArgTypeInt         ArgType = "int"
	ArgTypeFloat       ArgType = "float"
	ArgTypeBool        ArgType = "bool"
	ArgTypeSliceString ArgType = "slice_string"
	ArgTypeSliceAny    ArgType = "slice_any"
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
	if len(rawArgs) != len(spec.Args) {
		return nil, fmt.Errorf("tool '%s' expected %d arguments, but received %d", spec.Name, len(spec.Args), len(rawArgs))
	}

	convertedArgs := make([]interface{}, len(rawArgs))

	for i, argSpec := range spec.Args {
		argValue := rawArgs[i]
		argValueStr := fmt.Sprintf("%v", argValue) // String representation for errors/conversions

		match := false
		conversionErr := error(nil)

		switch argSpec.Type {
		case ArgTypeString:
			convertedArgs[i] = argValueStr // Use string representation
			match = true
		case ArgTypeInt:
			intVal, err := strconv.Atoi(argValueStr)
			if err == nil {
				convertedArgs[i] = intVal
				match = true
			} else {
				// *** FIX: Use %q for argValueStr in error to match test ***
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d) expected %s, but received %q which cannot be converted to int", spec.Name, argSpec.Name, i, argSpec.Type, argValueStr)
			}
		case ArgTypeBool:
			strValLower := strings.ToLower(argValueStr)
			if strValLower == "true" {
				convertedArgs[i] = true
				match = true
			} else if strValLower == "false" {
				convertedArgs[i] = false
				match = true
			} else {
				// *** FIX: Use %q for argValueStr in error ***
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d) expected %s, but received %q which cannot be converted to bool", spec.Name, argSpec.Name, i, argSpec.Type, argValueStr)
			}
		case ArgTypeSliceString:
			switch v := argValue.(type) {
			case []string:
				convertedArgs[i] = v
				match = true
			case []interface{}:
				strSlice := make([]string, len(v))
				canConvert := true
				for idx, item := range v {
					// Use Sprintf for robust conversion from interface{} items
					strSlice[idx] = fmt.Sprintf("%v", item)
				}
				if canConvert { // Always true now with Sprintf
					convertedArgs[i] = strSlice
					match = true
				}
			}
			if !match {
				conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d) expected %s, but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
		case ArgTypeAny:
			convertedArgs[i] = argValue
			match = true
		default:
			// Simplified check for other specific types if needed, otherwise error
			conversionErr = fmt.Errorf("tool '%s' argument '%s' (index %d): type check failed for expected type %s (received %T)", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
		}

		if conversionErr != nil {
			return nil, conversionErr // Return specific conversion error
		}
		if !match { // Should ideally not happen if all types covered or error set
			return nil, fmt.Errorf("tool '%s' argument '%s' (index %d): internal validation error for type %s (received %T)", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
		}
	}

	return convertedArgs, nil
}
