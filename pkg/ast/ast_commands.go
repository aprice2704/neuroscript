// filename: pkg/core/ast_commands.go
// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Defines the AST node for a single command block.
// nlines: 20
// risk_rating: LOW

package ast

import "github.com/aprice2704/neuroscript/pkg/lang"

// CommandNode represents a single 'command ... endcommand' block.
// It is a top-level declaration, similar to a Procedure.
type CommandNode struct {
	Pos           *lang.Position
	Metadata      map[string]string
	Body          []Step
	ErrorHandlers []*Step
}

func (n *CommandNode) isNode()      {}
func (n *CommandNode) isStatement() {}
func (n *CommandNode) String() string {
	return "command ... endcommand"
}
