// NeuroScript Version: 0.5.2
// File version: 8
// Purpose: Added SandboxDir to the Runtime interface to allow tools to access the interpreter's sandbox.
// filename: pkg/tool/tools_types.go
// nlines: 110
// risk_rating: HIGH

package tool

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Runtime is the minimal surface a tool needs to interact with the VM.
type Runtime interface {
	Println(...any)
	Ask(prompt string) string
	GetVar(name string) (any, bool)
	SetVar(name string, val any)
	CallTool(name string, args []any) (any, error)
	GetLogger() interfaces.Logger
	SandboxDir() string // FIX: Added this method.
}

// ArgType defines the expected data type for a tool argument or return value.
type ArgType string

const (
	ArgTypeAny         ArgType = "any"
	ArgTypeString      ArgType = "string"
	ArgTypeInt         ArgType = "int"
	ArgTypeFloat       ArgType = "float"
	ArgTypeBool        ArgType = "bool"
	ArgTypeMap         ArgType = "map"
	ArgTypeSlice       ArgType = "slice"
	ArgTypeSliceString ArgType = "slice_string"
	ArgTypeSliceInt    ArgType = "slice_int"
	ArgTypeSliceFloat  ArgType = "slice_float"
	ArgTypeSliceBool   ArgType = "slice_bool"
	ArgTypeSliceMap    ArgType = "slice_map"
	ArgTypeSliceAny    ArgType = "slice_any"
	ArgTypeNil         ArgType = "nil"
)

// ToolFunc is the signature for the Go function that implements a tool.
type ToolFunc func(rt Runtime, args []interface{}) (interface{}, error)

// ArgSpec defines the specification for a single tool argument.
type ArgSpec struct {
	Name         string      `json:"name"`
	Type         ArgType     `json:"type"`
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
}

// ToolSpec defines the specification for a tool.
type ToolSpec struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Category        string    `json:"category,omitempty"`
	Args            []ArgSpec `json:"args,omitempty"`
	ReturnType      ArgType   `json:"returnType"`
	ReturnHelp      string    `json:"returnHelp,omitempty"`
	Variadic        bool      `json:"variadic,omitempty"`
	Example         string    `json:"example,omitempty"`
	ErrorConditions string    `json:"errorConditions,omitempty"`
}

// ToolImplementation combines the specification of a tool with its Go function.
type ToolImplementation struct {
	Spec ToolSpec
	Func ToolFunc
}

// IsTool satisfies the lang.Tool interface.
func (t *ToolImplementation) IsTool() {}

// Name returns the name of the tool.
func (t *ToolImplementation) Name() string {
	return t.Spec.Name
}

// ToolRegistrar defines an interface for registering tools.
type ToolRegistrar interface {
	RegisterTool(impl ToolImplementation) error
}

// ToolRegistry defines the interface for a complete tool registry.
type ToolRegistry interface {
	ToolRegistrar
	GetTool(name string) (ToolImplementation, bool)
	ListTools() []ToolSpec
	ExecuteTool(toolName string, args map[string]lang.Value) (lang.Value, error)
}
