// filename: pkg/ast/ast_commands.go
// NeuroScript Version: 0.5.2
// File version: 8
// Purpose: Removed redundant Pos field and GetPos method to unify position handling via BaseNode.
// nlines: 20+
// risk_rating: LOW

package ast

// CommandNode represents a single 'command ... endcommand' block.
// It is a top-level declaration, similar to a Procedure.
type CommandNode struct {
	BaseNode
	BlankLinesBefore int
	Metadata         map[string]string
	Comments         []*Comment
	Body             []Step
	ErrorHandlers    []*Step
}

func (n *CommandNode) isNode()      {}
func (n *CommandNode) isStatement() {}
func (n *CommandNode) String() string {
	return "command ... endcommand"
}
