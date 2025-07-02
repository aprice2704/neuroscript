// NeuroScript Version: 0.3.1
// File version: 0.0.5 // Added Category, Example, ReturnHelp, Variadic, ErrorConditions to ToolSpec. Added DefaultValue to ArgSpec.
// nlines: 90 // Approximate, will increase
// risk_rating: HIGH
// filename: pkg/tool/tools_types.go

package tool

import (
	"fmt"	// Needed for error formatting

	"github.com/google/generative-ai-go/genai"	// Needed for genai.Type
)

// ArgType defines the expected data type for a tool argument or return value.
type ArgType string

// NOTE: Keep string values lowercase for consistency in specs/parsing.
const (
	ArgTypeAny		ArgType	= "any"
	ArgTypeString		ArgType	= "string"
	ArgTypeInt		ArgType	= "int"			// Represents int64 -> genai.TypeInteger
	ArgTypeFloat		ArgType	= "float"		// Represents float64 -> genai.TypeNumber
	ArgTypeBool		ArgType	= "bool"		// Represents bool -> genai.TypeBoolean
	ArgTypeMap		ArgType	= "map"			// Represents map[string]interface{} -> genai.TypeObject
	ArgTypeSlice		ArgType	= "slice"		// Generic slice, prefer more specific below
	ArgTypeSliceString	ArgType	= "slice_string"	// -> genai.TypeArray (items: string)
	ArgTypeSliceInt		ArgType	= "slice_int"		// -> genai.TypeArray (items: integer)
	ArgTypeSliceFloat	ArgType	= "slice_float"		// -> genai.TypeArray (items: number)
	ArgTypeSliceBool	ArgType	= "slice_bool"		// -> genai.TypeArray (items: boolean)
	ArgTypeSliceMap		ArgType	= "slice_map"		// -> genai.TypeArray (items: object)
	ArgTypeSliceAny		ArgType	= "slice_any"		// -> genai.TypeArray (items: any/string?)
	ArgTypeNil		ArgType	= "nil"			// Represents no meaningful return value
)

// ArgTypeList provides a slice of all defined ArgType constants.
var ArgTypeList = []ArgType{
	ArgTypeAny, ArgTypeString, ArgTypeInt, ArgTypeFloat, ArgTypeBool, ArgTypeMap,
	ArgTypeSlice, ArgTypeSliceString, ArgTypeSliceInt, ArgTypeSliceFloat,
	ArgTypeSliceBool, ArgTypeSliceMap, ArgTypeSliceAny, ArgTypeNil,
}

// ToolPrefix Constant - Standard prefix for tool calls in NeuroScript.
const ToolPrefix = "TOOL."

// ToGenaiType converts the internal ArgType to the corresponding genai.Type.
func (at ArgType) ToGenaiType() (genai.Type, error) {
	switch at {
	case ArgTypeString, ArgTypeAny:
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
		return genai.TypeArray, nil
	case ArgTypeNil:
		return genai.TypeUnspecified, fmt.Errorf("cannot convert ArgTypeNil to a genai.Type for LLM function declaration expecting a specific type")
	default:
		return genai.TypeUnspecified, fmt.Errorf("unsupported ArgType '%s' cannot be converted to genai.Type", at)
	}
}

// ToolFunc is the signature for the Go function that implements a tool.
type ToolFunc func(interpreter *Interpreter, args []interface{}) (interface{}, error)

// ArgSpec defines the specification for a single tool argument.
type ArgSpec struct {
	Name		string		`json:"name"`
	Type		ArgType		`json:"type"`
	Description	string		`json:"description"`
	Required	bool		`json:"required"`
	DefaultValue	interface{}	`json:"defaultValue,omitempty"`	// Default value if not required and not provided.
}

// ToolSpec defines the specification for a tool.
type ToolSpec struct {
	Name		string		`json:"name"`
	Description	string		`json:"description"`
	Category	string		`json:"category,omitempty"`	// Tool category for grouping/filtering.
	Args		[]ArgSpec	`json:"args,omitempty"`
	ReturnType	ArgType		`json:"returnType"`
	ReturnHelp	string		`json:"returnHelp,omitempty"`		// Detailed explanation of what is returned.
	Variadic	bool		`json:"variadic,omitempty"`		// Does the tool accept variable args for the last parameter?
	Example		string		`json:"example,omitempty"`		// A short NeuroScript example of how to call the tool.
	ErrorConditions	string		`json:"errorConditions,omitempty"`	// Description of common error conditions or types.
}

// ToolImplementation combines the specification of a tool with its Go function.
type ToolImplementation struct {
	Spec	ToolSpec
	Func	ToolFunc
}

// ToolRegistrar defines an interface for registering tools.
// The ToolRegistry struct in tools_registry.go implements this.
type ToolRegistrar interface {
	RegisterTool(impl ToolImplementation) error
}

// ToolRegistry defines the interface for a complete tool registry.
// *Interpreter is expected to implement this interface.
// The methods here should align with what *core.ToolRegistry (struct in tools_registry.go) provides.
type ToolRegistry interface {
	ToolRegistrar								// Embeds RegisterTool(impl ToolImplementation) error
	GetTool(name string) (ToolImplementation, bool)				// Returns the full implementation
	ListTools() []ToolSpec							// Returns a list of specs
	ExecuteTool(toolName string, args map[string]Value) (Value, error)	// <- must use value here
}