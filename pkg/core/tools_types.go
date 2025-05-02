// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 19:49:00 PDT
// filename: pkg/core/tools_types.go
package core

import (
	"fmt" // Needed for error formatting
	// "context" // Context not needed in ToolFunc signature

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
	ArgTypeMap    ArgType = "map"   // Represents map[string]interface{} -> genai.TypeObject
	ArgTypeAny    ArgType = "any"   // Any type allowed -> Defaulting to String for schema

	// Specific slice types (may consolidate with ArgTypeList/ArgTypeMap later)
	ArgTypeSliceString ArgType = "slice_string" // -> genai.TypeArray (items: string)
	ArgTypeSliceAny    ArgType = "slice_any"    // -> genai.TypeArray (items: any/string?)
	ArgTypeSliceMap    ArgType = "slice_map"    // +++ ADDED +++ -> genai.TypeArray (items: object)
)

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
	// +++ ADDED ArgTypeSliceMap to this case +++
	case ArgTypeList, ArgTypeSliceString, ArgTypeSliceAny, ArgTypeSliceMap: // Treat all list types as Array for now
		// TODO: Potentially specify item types (e.g., genai.Items(genai.TypeObject)) for more accurate schemas
		return genai.TypeArray, nil
	case ArgTypeMap:
		return genai.TypeObject, nil
	case ArgTypeAny:
		return genai.TypeString, nil // Defaulting 'any' to String for schema
	default:
		// Use fmt.Errorf for consistent error formatting
		return genai.TypeUnspecified, fmt.Errorf("unsupported ArgType: %q", at)
	}
}

// ArgSpec defines the specification for a single argument to a tool.
type ArgSpec struct {
	Name        string
	Type        ArgType
	Description string
	Required    bool
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

// ToolPrefix Constant
const ToolPrefix = "TOOL." // Standard prefix for tool calls in NeuroScript
