// NeuroScript Version: 0.6.0
// File version: 23
// Purpose: Added GetExecPolicy to the Runtime interface to satisfy the policygate.Runtime interface.
// filename: pkg/tool/tool_types.go
// nlines: 156
// risk_rating: HIGH

package tool

import (
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Runtime is the minimal surface a tool needs to interact with the VM.
type Runtime interface {
	Println(...any)
	PromptUser(prompt string) (string, error)
	GetVar(name string) (any, bool)
	SetVar(name string, val any)
	CallTool(name types.FullName, args []any) (any, error)
	GetLogger() interfaces.Logger
	SandboxDir() string
	ToolRegistry() ToolRegistry
	LLM() interfaces.LLMClient
	RegisterHandle(obj interface{}, typePrefix string) (string, error)
	GetHandleValue(handle string, expectedTypePrefix string) (interface{}, error)

	AgentModels() interfaces.AgentModelReader
	AgentModelsAdmin() interfaces.AgentModelAdmin
	// GetGrantSet returns the currently active capability grant set for policy enforcement.
	GetGrantSet() *capability.GrantSet
	// GetExecPolicy returns the currently active execution policy.
	GetExecPolicy() *policy.ExecPolicy
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
	Name            types.ToolName  `json:"name"`
	Group           types.ToolGroup `json:"groupname"`
	FullName        types.FullName  `json:"fullname"`
	Description     string          `json:"description"`
	Category        string          `json:"category,omitempty"`
	Args            []ArgSpec       `json:"args,omitempty"`
	ReturnType      ArgType         `json:"returnType"`
	ReturnHelp      string          `json:"returnHelp,omitempty"`
	Variadic        bool            `json:"variadic,omitempty"`
	Example         string          `json:"example,omitempty"`
	ErrorConditions string          `json:"errorConditions,omitempty"`
}

// Methods to satisfy policy.ToolSpecProvider interface
func (ts ToolSpec) FullNameForChecksum() types.FullName { return ts.FullName }
func (ts ToolSpec) ReturnTypeForChecksum() string       { return string(ts.ReturnType) }
func (ts ToolSpec) ArgCountForChecksum() int            { return len(ts.Args) }

// ToolImplementation combines the specification of a tool with its Go function
// and its policy requirements.
type ToolImplementation struct {
	FullName          types.FullName          `json:"-"`
	Spec              ToolSpec                `json:"spec"`
	Func              ToolFunc                `json:"-"`
	RequiresTrust     bool                    `json:"requiresTrust"`
	RequiredCaps      []capability.Capability `json:"requiredCaps,omitempty"`
	Effects           []string                `json:"effects,omitempty"`
	SignatureChecksum string                  `json:"signatureChecksum"`
}

// IsTool satisfies the lang.Tool interface.
func (t ToolImplementation) IsTool() {}

// Name returns the fully qualified name of the tool.
func (t *ToolImplementation) Name() types.FullName {
	return t.FullName
}

// ToolRegistrar defines an interface for registering tools.
type ToolRegistrar interface {
	RegisterTool(impl ToolImplementation) (ToolImplementation, error)
}

// ToolRegistry defines the interface for a complete tool registry.
type ToolRegistry interface {
	ToolRegistrar
	GetTool(name types.FullName) (ToolImplementation, bool)
	GetToolShort(group types.ToolGroup, name types.ToolName) (ToolImplementation, bool)
	ListTools() []ToolImplementation
	NTools() int
	ExecuteTool(toolName types.FullName, args map[string]lang.Value) (lang.Value, error)
	CallFromInterpreter(interp Runtime, fullname types.FullName, args []lang.Value) (lang.Value, error)
	// NewViewForInterpreter creates a new registry that shares the tool definitions
	// of the parent but is bound to a different interpreter runtime.
	NewViewForInterpreter(interpreter Runtime) ToolRegistry
}
