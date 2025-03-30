package core

import (
	"fmt"
	"reflect" // Might be useful for advanced type checking
	"strconv" // For potential type conversions later
	"strings"
)

// --- Tool Argument Specification ---

// ArgType defines the expected type of a tool argument.
// Using constants for type safety and clarity.
type ArgType string

const (
	ArgTypeString      ArgType = "string"
	ArgTypeInt         ArgType = "int"          // Interpreter would need to parse string to int
	ArgTypeFloat       ArgType = "float"        // Interpreter would need to parse string to float
	ArgTypeBool        ArgType = "bool"         // Interpreter would need to parse "true"/"false"
	ArgTypeSliceString ArgType = "slice_string" // Represents []string
	ArgTypeSliceAny    ArgType = "slice_any"    // Represents []interface{}
	ArgTypeAny         ArgType = "any"          // Allows any type, tool must handle
	// Add other types as needed (e.g., map)
)

// ArgSpec defines the specification for a single tool argument.
type ArgSpec struct {
	Name        string  // Name of the argument (for documentation/error messages)
	Type        ArgType // Expected type of the argument
	Description string  // Help text explaining the argument
	Required    bool    // Is this argument mandatory? (Future use, for now assume all are)
}

// --- Tool Specification ---

// ToolSpec defines the specification for a tool, including its arguments and return type.
type ToolSpec struct {
	Name        string    // The name used after TOOL., e.g., "StringLength"
	Description string    // Help text explaining what the tool does
	Args        []ArgSpec // Ordered list of argument specifications
	ReturnType  ArgType   // Expected type of the return value (for documentation/future checks)
}

// --- Tool Function Implementation ---

// ToolFunc defines the expected signature for any Go function implementing a tool.
// It receives evaluated arguments from the interpreter as []interface{}
// and should return the result as interface{} or an error.
type ToolFunc func(interpreter *Interpreter, args []interface{}) (interface{}, error)

// --- Tool Implementation Registry ---

// ToolImplementation holds the specification and the Go function for a tool.
type ToolImplementation struct {
	Spec ToolSpec
	Func ToolFunc
}

// ToolRegistry manages the registration and lookup of available tools.
// It can be a field within the Interpreter struct or a separate package variable.
type ToolRegistry struct {
	tools map[string]ToolImplementation // Map from tool name (e.g., "StringLength") to its implementation
}

// NewToolRegistry creates an empty tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ToolImplementation),
	}
}

// RegisterTool adds a tool's specification and implementation function to the registry.
func (tr *ToolRegistry) RegisterTool(impl ToolImplementation) error {
	if _, exists := tr.tools[impl.Spec.Name]; exists {
		return fmt.Errorf("tool '%s' already registered", impl.Spec.Name)
	}
	if impl.Func == nil {
		return fmt.Errorf("tool '%s' registration is missing implementation function", impl.Spec.Name)
	}
	tr.tools[impl.Spec.Name] = impl
	// fmt.Printf("[Registry] Registered TOOL.%s\n", impl.Spec.Name) // --- COMMENTED OUT ---
	return nil
}

// GetTool retrieves a tool's implementation from the registry.
func (tr *ToolRegistry) GetTool(name string) (ToolImplementation, bool) {
	impl, found := tr.tools[name]
	return impl, found
}

// --- Argument Validation/Conversion Helper (Example - Needs refinement) ---

// ValidateAndConvertArgs checks provided arguments against the spec and attempts basic conversions.
// This would be called by the interpreter *before* calling the ToolFunc.
func ValidateAndConvertArgs(spec ToolSpec, rawArgs []interface{}) ([]interface{}, error) {
	// 1. Check argument count
	// TODO: Handle optional args based on ArgSpec.Required later
	if len(rawArgs) != len(spec.Args) {
		return nil, fmt.Errorf("tool '%s' expected %d arguments, but received %d", spec.Name, len(spec.Args), len(rawArgs))
	}

	convertedArgs := make([]interface{}, len(rawArgs))

	// 2. Check types and attempt basic conversions
	for i, argSpec := range spec.Args {
		argValue := rawArgs[i]
		argType := reflect.TypeOf(argValue)
		argKind := argType.Kind()

		// fmt.Printf("Validating arg %d (%s): expected %s, got %T (%v)\n", i, argSpec.Name, argSpec.Type, argValue, argValue) // Debug

		match := false
		switch argSpec.Type {
		case ArgTypeString:
			// Most things can be reasonably converted to string via Sprintf
			convertedArgs[i] = fmt.Sprintf("%v", argValue)
			match = true // Assume conversion is always possible
		case ArgTypeInt:
			strVal := fmt.Sprintf("%v", argValue) // Convert to string first
			intVal, err := strconv.Atoi(strVal)
			if err == nil {
				convertedArgs[i] = intVal
				match = true
			} else {
				return nil, fmt.Errorf("tool '%s' argument '%s' (index %d) expected %s, but received '%v' which cannot be converted to int", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
		case ArgTypeBool:
			strVal := strings.ToLower(fmt.Sprintf("%v", argValue))
			if strVal == "true" {
				convertedArgs[i] = true
				match = true
			} else if strVal == "false" {
				convertedArgs[i] = false
				match = true
			} else {
				return nil, fmt.Errorf("tool '%s' argument '%s' (index %d) expected %s, but received '%v' which cannot be converted to bool", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
		case ArgTypeSliceString:
			// Check if it's already []string or []interface{} convertible to []string
			switch v := argValue.(type) {
			case []string:
				convertedArgs[i] = v
				match = true
			case []interface{}:
				strSlice := make([]string, len(v))
				canConvert := true
				for idx, item := range v {
					if s, ok := item.(string); ok {
						strSlice[idx] = s
					} else {
						canConvert = false
						break
					}
				}
				if canConvert {
					convertedArgs[i] = strSlice
					match = true
				}
			}
			if !match {
				return nil, fmt.Errorf("tool '%s' argument '%s' (index %d) expected %s, but received incompatible type %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
		case ArgTypeAny:
			convertedArgs[i] = argValue // Pass through directly
			match = true
		default:
			// Basic kind check for other types (can be expanded)
			// This is a simplistic check, might need refinement for float, slice_any etc.
			if (argSpec.Type == ArgTypeFloat && (argKind == reflect.Float32 || argKind == reflect.Float64)) ||
				(argSpec.Type == ArgTypeSliceAny && argKind == reflect.Slice) {
				convertedArgs[i] = argValue
				match = true
			} else {
				return nil, fmt.Errorf("tool '%s' argument '%s' (index %d): type check failed for expected type %s (received %T)", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
			}
		}
		// if !match { // Should be handled by specific type checks now
		// return nil, fmt.Errorf("tool '%s' argument '%s' (index %d): type mismatch, expected %s, got %T", spec.Name, argSpec.Name, i, argSpec.Type, argValue)
		// }
	}

	return convertedArgs, nil
}
