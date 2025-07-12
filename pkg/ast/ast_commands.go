// filename: pkg/ast/ast_commands.go
// NeuroScript Version: 0.5.2
// File version: 5
// Purpose: Augmented CommandNode with BaseNode to satisfy the Node interface.
// nlines: 20+
// risk_rating: LOW

package ast

import "github.com/aprice2704/neuroscript/pkg/types"

// CommandNode represents a single 'command ... endcommand' block.
// It is a top-level declaration, similar to a Procedure.
type CommandNode struct {
	BaseNode
	Pos           *types.Position
	Metadata      map[string]string
	Body          []Step
	ErrorHandlers []*Step
}

func (n *CommandNode) GetPos() *types.Position { return n.Pos }
func (n *CommandNode) isNode()                 {}
func (n *CommandNode) isStatement()            {}
func (n *CommandNode) String() string {
	return "command ... endcommand"
}
