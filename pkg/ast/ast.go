// NeuroScript Version: 0.8.0
// File version: 10
// Purpose: Added a TestString() method to the Expression interface for unambiguous AST structure validation in tests.
// filename: pkg/ast/ast.go
// nlines: 75+
// risk_rating: MEDIUM

package ast

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Node is an alias for the foundational Node interface.
type Node = interfaces.Node

// Tree is an alias for the foundational Tree struct.
type Tree = interfaces.Tree

// Expression is an interface for all expression AST nodes. It embeds Node.
type Expression interface {
	Node
	expressionNode() // Marker method
	String() string
	TestString() string // For unambiguous test output
}

// BaseNode provides the common fields for all AST nodes, fulfilling the Node interface.
type BaseNode struct {
	StartPos *types.Position
	StopPos  *types.Position
	NodeKind types.Kind // CORRECTED
}

// GetPos provides the implementation for the Node interface's GetPos() method.
func (n *BaseNode) GetPos() *types.Position { return n.StartPos }

// End returns the ending position of the node.
func (n *BaseNode) End() *types.Position { return n.StopPos }

// Kind returns the kind of the node.
func (n *BaseNode) Kind() types.Kind { return n.NodeKind } // CORRECTED

// Comment represents a source code comment.
type Comment struct {
	BaseNode
	Text string
}

// Program represents the entire parsed NeuroScript program.
type Program struct {
	BaseNode
	Metadata    map[string]string
	Procedures  map[string]*Procedure
	Events      []*OnEventDecl
	Expressions []Expression
	Commands    []*CommandNode
	Comments    []*Comment
}

// NewProgram creates and initializes a new Program node.
func NewProgram() *Program {
	return &Program{
		BaseNode:    BaseNode{NodeKind: types.KindProgram}, // CORRECTED
		Metadata:    make(map[string]string),
		Procedures:  make(map[string]*Procedure),
		Events:      make([]*OnEventDecl, 0),
		Commands:    make([]*CommandNode, 0),
		Expressions: make([]Expression, 0),
		Comments:    make([]*Comment, 0),
	}
}

// SecretRef represents a reference to a secret (e.g., secret "path").
type SecretRef struct {
	BaseNode
	Path string
	Enc  string
	Raw  []byte
}

func (n *SecretRef) expressionNode() {}
func (n *SecretRef) String() string {
	return fmt.Sprintf("secret %q", n.Path)
}
func (n *SecretRef) TestString() string { return n.String() }

// ErrorNode captures a parsing or semantic error encountered during AST construction.
type ErrorNode struct {
	BaseNode
	Message string
}

func (n *ErrorNode) expressionNode() {}
func (n *ErrorNode) String() string {
	if n == nil {
		return "<nil error node>"
	}
	return fmt.Sprintf("Error at %s: %s", n.GetPos(), n.Message)
}
func (n *ErrorNode) TestString() string { return n.String() }

// --- Helper for getting types.Position from nodes that implement Expression or Step ---
func getExpressionPosition(val interface{}) *types.Position {
	if expr, ok := val.(Expression); ok {
		return expr.GetPos()
	}
	if step, ok := val.(Step); ok {
		return step.GetPos()
	}
	return nil
}
