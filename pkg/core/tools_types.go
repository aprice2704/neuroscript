// neuroscript/pkg/core/tools_types.go
package core

// ArgType defines the expected data type for a tool argument or return value.
type ArgType string

const (
	ArgTypeString      ArgType = "string"
	ArgTypeInt         ArgType = "int"   // Represents int64
	ArgTypeFloat       ArgType = "float" // Represents float64
	ArgTypeBool        ArgType = "bool"
	ArgTypeSliceString ArgType = "slice_string" // Specifically []string
	ArgTypeSliceAny    ArgType = "slice_any"    // Represents []interface{} or []string
	// *** ADDED: Definition for list type ***
	ArgTypeList ArgType = "list" // Represents a generic list/slice ([]interface{})
	ArgTypeAny  ArgType = "any"  // Any type is allowed
)

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
	ReturnType  ArgType
}

// ToolFunc is the signature for the Go function that implements a tool.
// It receives the interpreter context and validated/converted arguments.
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
