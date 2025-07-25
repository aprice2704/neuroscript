// NeuroScript Version: 0.6.0
// File version: 12
// Purpose: Corrects the return type of InstantiateAllStandardTools to []ToolImplementation to simplify tool registration for consumers.
// filename: pkg/api/reexport.go
// nlines: 80
// risk_rating: LOW

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Re-exported types for the public API, as per the v0.7 contract.
type (
	// Foundational types from pkg/types, ensuring a stable AST contract.
	Kind     = types.Kind
	Position = types.Position
	Node     = interfaces.Node
	Tree     = interfaces.Tree

	// Logging and error types.
	Logger       = interfaces.Logger
	LogLevel     = interfaces.LogLevel
	RuntimeError = lang.RuntimeError

	// SignedAST is the transport wrapper for a canonicalized and signed tree.
	SignedAST struct {
		Blob []byte
		Sum  [32]byte
		Sig  []byte
	}

	Value  any
	Option = interpreter.InterpreterOption

	// Tool-related types needed to define custom tools.
	ToolImplementation = tool.ToolImplementation
	ArgSpec            = tool.ToolSpec
	Runtime            = tool.Runtime
	ToolFunc           = tool.ToolFunc
	ToolSpec           = tool.ToolSpec
	FullName           = types.FullName
)

// WithTool creates an interpreter option to register a custom tool.
func WithTool(t ToolImplementation) Option {
	return func(i *interpreter.Interpreter) {
		if err := i.ToolRegistry().RegisterTool(t); err != nil {
			if logger := i.GetLogger(); logger != nil {
				logger.Error("failed to register tool via WithTool option", "tool", t.FullName, "error", err)
			}
		}
	}
}

// InstantiateAllStandardTools returns a slice of ToolImplementation structs for all
// standard tools that come with NeuroScript.
// NOTE: This is a placeholder. You will need to add the actual tool implementations here.
func InstantiateAllStandardTools() []ToolImplementation {
	impls := []ToolImplementation{
		// Example:
		// file.NewFileTool(),
		// http.NewHTTPTool(),
	}
	return impls
}

// RegisterCriticalErrorHandler allows the host application to override the default
// panic behavior for critical errors.
func RegisterCriticalErrorHandler(h func(*lang.RuntimeError)) {
	lang.RegisterCriticalHandler(h)
}

// MakeToolFullName creates a correctly formatted, fully-qualified tool name.
// It acts as a public facade for the internal types.MakeFullName function.
func MakeToolFullName(group, name string) types.FullName {
	return types.MakeFullName(group, name)
}
