// NeuroScript Version: 0.8.0
// File version: 10
// Purpose: Removed RegisterProvider method to enforce use of a separate ProviderRegistry.
// filename: pkg/interfaces/interpreter.go
// nlines: 44
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
