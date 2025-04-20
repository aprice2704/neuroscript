// filename: pkg/core/tools_types.go
package core

import (
	"fmt" // Needed for error formatting

	"github.com/google/generative-ai-go/genai" // Needed for genai.Type
)

// ArgType defines the expected data type for a tool argument or return value.
type ArgType string

// NOTE: Keep string values lowercase for consistency in specs/parsing.
const (
	ArgTypeString ArgType = "string"
	ArgTypeInt    ArgType = "int"   // Represents int64 -> genai.TypeInteger
	ArgTypeFloat  ArgType = "float" // Represents float64 -> genai.TypeNumber
	ArgTypeBool   ArgType = "bool"  // Represents bool -> genai.TypeBoolean
	ArgTypeList   ArgType = "list"  // Represents a generic list/slice -> genai.TypeArray
	ArgTypeMap    ArgType = "map"   // Represents map[string]interface{} -> genai.TypeObject (Added)
	ArgTypeAny    ArgType = "any"   // Any type allowed -> Not directly mappable, maybe skip or default? Let's default to String for now.

	// Deprecated potentially? Use ArgTypeList instead.
	ArgTypeSliceString ArgType = "slice_string" // -> genai.TypeArray (items: string)
	ArgTypeSliceAny    ArgType = "slice_any"    // -> genai.TypeArray (items: any/string?)

)

// +++ ADDED: ToGenaiType method +++
// ToGenaiType converts the internal ArgType to the corresponding genai.Type for function declarations.
func (at ArgType) ToGenaiType() (genai.Type, error) {
	switch at {
	case ArgTypeString:
		return genai.TypeString, nil
	case ArgTypeInt:
		return genai.TypeInteger, nil
	case ArgTypeFloat:
		return genai.TypeNumber, nil // Gemini uses "Number" for floats/ints
	case ArgTypeBool:
		return genai.TypeBoolean, nil
	case ArgTypeList, ArgTypeSliceString, ArgTypeSliceAny: // Treat all list types as Array for now
		// TODO: Future enhancement: Specify item type for arrays if possible in genai.Schema
		return genai.TypeArray, nil
	case ArgTypeMap:
		return genai.TypeObject, nil // Added Map type
	case ArgTypeAny:
		// 'any' doesn't map directly. Defaulting to String might be safest for declarations,
		// but this indicates the tool needs careful handling or a more specific type.
		// Alternatively, could return an error or a specific 'any' type if genai supports it.
		// Let's return String for now and log a warning implicitly via usage.
		return genai.TypeString, nil // Or return an error? fmt.Errorf("ArgTypeAny cannot be directly mapped to genai.Type")
	default:
		return genai.TypeUnspecified, fmt.Errorf("unsupported ArgType: %q", at)
	}
}

// --- END ADDED ---

// ArgSpec defines the specification for a single argument to a tool.
type ArgSpec struct {
	Name        string
	Type        ArgType
	Description string
	Required    bool
	// TODO: Add optional 'DefaultValue' field?
	// TODO: Add optional 'EnumValues' field if Type is ArgTypeEnum? (Or handle via validation)
}

// ToolSpec defines the specification for a callable tool.
type ToolSpec struct {
	Name        string
	Description string
	Args        []ArgSpec
	ReturnType  ArgType // Specifies the expected type of the Go value returned by ToolFunc
}

// ToolFunc is the signature for the Go function that implements a tool.
// It receives the interpreter context and validated/converted arguments.
// Arguments are passed as a slice in the order defined in ToolSpec.Args.
type ToolFunc func(interpreter *Interpreter, args []interface{}) (interface{}, error)

// ToolImplementation holds the specification and the Go function for a tool.
type ToolImplementation struct {
	Spec ToolSpec
	Func ToolFunc
}

// ToolRegistrar defines an interface for registering tools, typically implemented by the ToolRegistry.
type ToolRegistrar interface {
	RegisterTool(impl ToolImplementation) error
}
