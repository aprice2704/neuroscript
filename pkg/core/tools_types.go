// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 15:33:57 PDT // Restore ToolPrefix, ArgTypeList, ToGenaiType
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
	ArgTypeAny         ArgType = "any"
	ArgTypeString      ArgType = "string"
	ArgTypeInt         ArgType = "int"          // Represents int64 -> genai.TypeInteger
	ArgTypeFloat       ArgType = "float"        // Represents float64 -> genai.TypeNumber
	ArgTypeBool        ArgType = "bool"         // Represents bool -> genai.TypeBoolean
	ArgTypeMap         ArgType = "map"          // Represents map[string]interface{} -> genai.TypeObject
	ArgTypeSlice       ArgType = "slice"        // Generic slice, prefer more specific below
	ArgTypeSliceString ArgType = "slice_string" // -> genai.TypeArray (items: string)
	ArgTypeSliceInt    ArgType = "slice_int"    // -> genai.TypeArray (items: integer) - Added for clarity
	ArgTypeSliceFloat  ArgType = "slice_float"  // -> genai.TypeArray (items: number) - Added for clarity
	ArgTypeSliceBool   ArgType = "slice_bool"   // -> genai.TypeArray (items: boolean) - Added for clarity
	ArgTypeSliceMap    ArgType = "slice_map"    // -> genai.TypeArray (items: object)
	ArgTypeSliceAny    ArgType = "slice_any"    // -> genai.TypeArray (items: any/string?)
	ArgTypeNil         ArgType = "nil"          // Represents no meaningful return value
)

// ArgTypeList provides a slice of all defined ArgType constants.
// Useful for validation or generating documentation.
var ArgTypeList = []ArgType{
	ArgTypeAny,
	ArgTypeString,
	ArgTypeInt,
	ArgTypeFloat,
	ArgTypeBool,
	ArgTypeMap,
	ArgTypeSlice,
	ArgTypeSliceString,
	ArgTypeSliceInt,
	ArgTypeSliceFloat,
	ArgTypeSliceBool,
	ArgTypeSliceMap,
	ArgTypeSliceAny,
	ArgTypeNil,
}

// ToolPrefix Constant - Standard prefix for tool calls in NeuroScript.
const ToolPrefix = "TOOL." // <<< ADDED BACK

// ToGenaiType converts the internal ArgType to the corresponding genai.Type.
// Used for generating function declaration schemas for the LLM.
func (at ArgType) ToGenaiType() (genai.Type, error) {
	switch at {
	case ArgTypeString, ArgTypeAny: // Treat 'any' as string for schema generation
		return genai.TypeString, nil
	case ArgTypeInt:
		return genai.TypeInteger, nil
	case ArgTypeFloat:
		return genai.TypeNumber, nil
	case ArgTypeBool:
		return genai.TypeBoolean, nil
	case ArgTypeMap:
		return genai.TypeObject, nil
	case ArgTypeSlice, ArgTypeSliceString, ArgTypeSliceInt, ArgTypeSliceFloat, ArgTypeSliceBool, ArgTypeSliceMap, ArgTypeSliceAny:
		// TODO: Specify item types for arrays more accurately if needed by genai API later
		return genai.TypeArray, nil
	case ArgTypeNil:
		// There's no direct equivalent for 'nil' return in genai schema.
		// Option 1: Use Unspecified (might be ignored or cause issues)
		// Option 2: Return an error, forcing tools with nil return to be excluded from LLM schema?
		// Option 3: Map to something benign like String? (Less accurate)
		// Let's return an error for now, assuming tools returning nil shouldn't be directly callable *by* the LLM expecting a value.
		return genai.TypeUnspecified, fmt.Errorf("cannot convert ArgTypeNil to a genai.Type for LLM function declaration")
	default:
		return genai.TypeUnspecified, fmt.Errorf("unsupported ArgType '%s' cannot be converted to genai.Type", at)
	}
}

// ToolFunc is the signature for the Go function that implements a tool.
// ... (rest of ToolFunc, ToolSpec, ArgSpec, ToolImplementation unchanged) ...
type ToolFunc func(interpreter *Interpreter, args []interface{}) (interface{}, error)
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
type ToolImplementation struct {
	Spec ToolSpec
	Func ToolFunc
}

// ToolRegistrar defines an interface for registering tools, typically implemented by the ToolRegistry.
// ... (ToolRegistrar interface unchanged) ...
type ToolRegistrar interface {
	RegisterTool(impl ToolImplementation) error
}
