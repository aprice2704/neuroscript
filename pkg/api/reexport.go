// NeuroScript Version: 0.6.0
// File version: 7
// Purpose: Re-exports core types, tool types, and interpreter options for a clean public API.
// filename: pkg/api/reexport.go
// nlines: 44
// risk_rating: LOW

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Re-exported types for the public API, as per the v0.6 contract.
// These types provide the stable, public-facing surface for all interactions
// with the NeuroScript AST and its components.
type (
	// Foundational types from pkg/types, ensuring a stable AST contract.
	Kind     = types.Kind
	Position = types.Position
	Node     = interfaces.Node
	Tree     = interfaces.Tree

	// SignedAST is the transport wrapper for a canonicalized and signed tree.
	SignedAST struct {
		Blob []byte   // The canonicalized AST, produced by Canonicalise.
		Sum  [32]byte // The BLAKE2b-256 digest of the Blob.
		Sig  []byte   // The Ed25519 signature of the Sum.
	}

	// Value represents the result of an execution.
	Value any

	// Option is a configuration function for an interpreter.
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
// This allows external packages to add functionality to the interpreter.
func WithTool(t ToolImplementation) Option {
	return func(i *interpreter.Interpreter) {
		if err := i.ToolRegistry().RegisterTool(t); err != nil {
			// If the interpreter has a logger, use it.
			if logger := i.GetLogger(); logger != nil {
				logger.Error("failed to register tool via WithTool option", "tool", t.FullName, "error", err)
			}
		}
	}
}
