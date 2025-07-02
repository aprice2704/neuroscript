// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Added an Expressions slice to the Program struct to handle top-level expressions.
// filename: pkg/ast/ast.go
// nlines: 50
// risk_rating: HIGH

package ast

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Expression is an interface for all expression AST nodes.
type Expression interface {
	GetPos() *lang.Position
	expressionNode() // Marker method
	String() string
}

// Program represents the entire parsed NeuroScript program.
type Program struct {
	Pos         *lang.Position
	Metadata    map[string]string
	Procedures  map[string]*Procedure
	Events      []*OnEventDecl
	Expressions []Expression   // FIX: Added to hold top-level expressions
	Commands    []*CommandNode // ADDED
}

// NewProgram creates and initializes a new Program node.
func NewProgram() *Program {
	return &Program{
		Metadata:    make(map[string]string),
		Procedures:  make(map[string]*Procedure),
		Events:      make([]*OnEventDecl, 0),
		Commands:    make([]*CommandNode, 0),
		Expressions: make([]Expression, 0),
	}
}

func (p *Program) GetPos() *lang.Position { return p.Pos }

// ErrorNode captures a parsing or semantic error encountered during AST construction.
type ErrorNode struct {
	Pos     *lang.Position
	Message string
}

func (n *ErrorNode) GetPos() *lang.Position { return n.Pos }
func (n *ErrorNode) expressionNode()        {}
func (n *ErrorNode) String() string {
	if n == nil {
		return "<nil error node>"
	}
	return fmt.Sprintf("Error at %s: %s", n.Pos, n.Message)
}

// --- Helper for getting lang.Position from nodes that implement Expression or Step ---
func getExpressionPosition(val interface{}) *lang.Position {
	if expr, ok := val.(Expression); ok {
		return expr.GetPos()
	}
	if step, ok := val.(Step); ok {
		return &step.Position
	}
	return nil
}
