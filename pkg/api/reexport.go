// NeuroScript Version: 0.6.0
// File version: 18
// Purpose: Corrected the WithTool implementation to remove flawed logic and rely on the authoritative RegisterTool function.
// filename: pkg/api/reexport.go
// nlines: 75
// risk_rating: MEDIUM

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
	ArgSpec            = tool.ArgSpec
	Runtime            = tool.Runtime
	ToolFunc           = tool.ToolFunc
	ToolSpec           = tool.ToolSpec
	FullName           = types.FullName
)

// WithTool creates an interpreter option to register a custom tool.
func WithTool(t ToolImplementation) Option {
	return func(i *interpreter.Interpreter) {
		// The RegisterTool function is the authority on canonicalizing names.
		// We simply pass the implementation through to it.
		if _, err := i.ToolRegistry().RegisterTool(t); err != nil {
			if logger := i.GetLogger(); logger != nil {
				logger.Error("failed to register tool via WithTool option", "tool", t.Spec.Name, "error", err)
			}
		}
	}
}

// InstantiateAllStandardTools returns a slice of ToolImplementation structs for all
// standard tools that come with NeuroScript.
func InstantiateAllStandardTools() []ToolImplementation {
	// This remains a placeholder as the standard tools are auto-bundled.
	return []ToolImplementation{}
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
