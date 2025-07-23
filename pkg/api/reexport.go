// NeuroScript Version: 0.6.0
// File version: 10
// Purpose: Adds a public facade for creating fully-qualified tool names.
// filename: pkg/api/reexport.go
// nlines: 66
// risk_rating: LOW

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Re-exported types for the public API, as per the v0.6 contract.
type (
	// Foundational types from pkg/types, ensuring a stable AST contract.
	Kind         = types.Kind
	Position     = types.Position
	Node         = interfaces.Node
	Tree         = interfaces.Tree
	Logger       = interfaces.Logger
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
