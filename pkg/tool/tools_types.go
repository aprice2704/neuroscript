// NeuroScript Version: 0.3.1
// File version: 0.0.5 // Added Category, Example, ReturnHelp, Variadic, ErrorConditions to ToolSpec. Added DefaultValue to ArgSpec.
// nlines: 90 // Approximate, will increase
// risk_rating: HIGH
// filename: pkg/tool/tools_types.go

package tool

import (
	// Needed for error formatting

	"github.com/aprice2704/neuroscript/pkg/lang"
	// Needed for genai.Type
)

// Runtime is the minimal surface a tool needs to interact with the VM.
// Add methods only when you find a tool that actually needs one.
type RunTime interface {
	// User-visible I/O ----------------------------------
	Println(...any)           // e.g. for shell tool
	Ask(prompt string) string // used by ai tools

	// Variable access -----------------------------------
	GetVar(name string) (any, bool)
	SetVar(name string, val any)

	// Tool dispatch (optional) --------------------------
	CallTool(name string, args []any) (any, error)
}

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
	ArgTypeSliceInt    ArgType = "slice_int"    // -> genai.TypeArray (items: integer)
	ArgTypeSliceFloat  ArgType = "slice_float"  // -> genai.TypeArray (items: number)
	ArgTypeSliceBool   ArgType = "slice_bool"   // -> genai.TypeArray (items: boolean)
	ArgTypeSliceMap    ArgType = "slice_map"    // -> genai.TypeArray (items: object)
	ArgTypeSliceAny    ArgType = "slice_any"    // -> genai.TypeArray (items: any/string?)
	ArgTypeNil         ArgType = "nil"          // Represents no meaningful return value
)

// ArgTypeList provides a slice of all defined ArgType constants.
var ArgTypeList = []ArgType{
	ArgTypeAny, ArgTypeString, ArgTypeInt, ArgTypeFloat, ArgTypeBool, ArgTypeMap,
	ArgTypeSlice, ArgTypeSliceString, ArgTypeSliceInt, ArgTypeSliceFloat,
	ArgTypeSliceBool, ArgTypeSliceMap, ArgTypeSliceAny, ArgTypeNil,
}

// ToolPrefix Constant - Standard prefix for tool calls in NeuroScript.
const ToolPrefix = "TOOL."

// ToolFunc is the signature for the Go function that implements a tool.
type ToolFunc func(rt RunTime, args []interface{}) (interface{}, error)

// ArgSpec defines the specification for a single tool argument.
type ArgSpec struct {
	Name         string      `json:"name"`
	Type         ArgType     `json:"type"`
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"defaultValue,omitempty"` // Default value if not required and not provided.
}

// ToolSpec defines the specification for a tool.
type ToolSpec struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Category        string    `json:"category,omitempty"` // Tool category for grouping/filtering.
	Args            []ArgSpec `json:"args,omitempty"`
	ReturnType      ArgType   `json:"returnType"`
	ReturnHelp      string    `json:"returnHelp,omitempty"`      // Detailed explanation of what is returned.
	Variadic        bool      `json:"variadic,omitempty"`        // Does the tool accept variable args for the last parameter?
	Example         string    `json:"example,omitempty"`         // A short NeuroScript example of how to call the tool.
	ErrorConditions string    `json:"errorConditions,omitempty"` // Description of common error conditions or types.
}

// ToolImplementation combines the specification of a tool with its Go function.
type ToolImplementation struct {
	Spec ToolSpec
	Func ToolFunc
}

// ToolRegistrar defines an interface for registering tools.
// The ToolRegistry struct in tools_registry.go implements this.
type ToolRegistrar interface {
	RegisterTool(impl ToolImplementation) error
}

// ToolRegistry defines the interface for a complete tool registry.
// *Interpreter is expected to implement this interface.
// The methods here should align with what * ToolRegistry (struct in tools_registry.go) provides.
type ToolRegistry interface {
	ToolRegistrar                                                                // Embeds RegisterTool(impl ToolImplementation) error
	GetTool(name string) (ToolImplementation, bool)                              // Returns the full implementation
	ListTools() []ToolSpec                                                       // Returns a list of specs
	ExecuteTool(toolName string, args map[string]lang.Value) (lang.Value, error) // <- must use value here
}
