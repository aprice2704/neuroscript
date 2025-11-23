// NeuroScript Version: 1
// File version: 11
// Purpose: The primary interface for the NeuroScript execution engine.
// Latest change: Added HandleRegistry() method to expose the new Handle system.
// filename: pkg/interfaces/interpreter.go
// nlines: 48
// risk_rating: COSMIC

package interfaces

import (
	"io"

	"github.com/aprice2704/neuroscript/pkg/types"
)

// Interpreter is the primary interface for the NeuroScript execution engine.
type Interpreter interface {
	Load(tree *Tree) error
	ExecuteCommands() (any, error)
	Run(procName string, args ...any) (any, error)
	EmitEvent(eventName string, source string, payload any)
	ToolRegistry() ToolRegistry
	AgentModelsAdmin() AgentModelAdmin
	AccountAdmin() AccountAdmin
	SetSandboxDir(path string)
	SetStdout(w io.Writer)
	SetStderr(w io.Writer)
	SetEmitFunc(f func(any))
	SetAITranscript(w io.Writer)
	GetLogger() Logger
	// RegisterProvider(name string, p any) // This is now handled by ProviderRegistry
	HandleRegistry() HandleRegistry // Added per specification
}

// Node represents a node in the AST.
type Node interface {
	Kind() types.Kind
	GetPos() *types.Position
}

// Tree represents the entire parsed AST.
type Tree struct {
	Root     Node
	Comments []any
}

func (t *Tree) Kind() types.Kind {
	if t.Root == nil {
		return types.KindUnknown
	}
	return t.Root.Kind()
}

func (t *Tree) GetPos() *types.Position {
	if t.Root == nil {
		return nil
	}
	return t.Root.GetPos()
}
