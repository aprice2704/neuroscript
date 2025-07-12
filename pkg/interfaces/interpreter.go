// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Updated to import Kind from the foundational types package.
// filename: pkg/interfaces/interpreter.go
// nlines: 30
// risk_rating: HIGH

package interfaces

import (
	"context"

	"github.com/aprice2704/neuroscript/pkg/types" // CORRECTED: Import types package
)

// Node is the interface required by the new integration contract.
type Node interface {
	GetPos() *types.Position
	End() *types.Position
	Kind() types.Kind // CORRECTED: Use types.Kind
}

// Tree represents the entire parsed NeuroScript program.
type Tree struct {
	Root     Node
	Comments []interface{}
}

// ExecResult represents the outcome of a script execution.
type ExecResult struct {
	Output string
	Error  error
	Value  interface{}
}

// SecretResolver is a function type that resolves a secret reference.
type SecretResolver func(ref Node) (string, error)

// InterpreterConfig holds the configuration for an interpreter run.
type InterpreterConfig struct {
	ResolveSecret SecretResolver
}

// Interpreter defines the public contract for executing a NeuroScript AST.
type Interpreter interface {
	ExecCommand(ctx context.Context, tree *Tree, cfg InterpreterConfig) (*ExecResult, error)
}
