// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Added an Expressions slice to the Program struct to handle top-level expressions.
// filename: pkg/core/ast.go
// nlines: 50
// risk_rating: HIGH

package core

import (
	"fmt"
)

// Position represents a location in the source code.
type Position struct {
	Line   int
	Column int
	File   string
}

func (p *Position) String() string {
	if p == nil {
		return "<nil position>"
	}
	return fmt.Sprintf("line %d, col %d", p.Line, p.Column)
}

// Expression is an interface for all expression AST nodes.
type Expression interface {
	GetPos() *Position
	expressionNode() // Marker method
	String() string
}

// Program represents the entire parsed NeuroScript program.
type Program struct {
	Pos         *Position
	Metadata    map[string]string
	Procedures  map[string]*Procedure
	Events      []*OnEventDecl
	Expressions []Expression // FIX: Added to hold top-level expressions
}

func (p *Program) GetPos() *Position { return p.Pos }

// ErrorNode captures a parsing or semantic error encountered during AST construction.
type ErrorNode struct {
	Pos     *Position
	Message string
}

func (n *ErrorNode) GetPos() *Position { return n.Pos }
func (n *ErrorNode) expressionNode()   {}
func (n *ErrorNode) String() string {
	if n == nil {
		return "<nil error node>"
	}
	return fmt.Sprintf("Error at %s: %s", n.Pos, n.Message)
}

// --- Helper for getting position from nodes that implement Expression or Step ---
func getExpressionPosition(val interface{}) *Position {
	if expr, ok := val.(Expression); ok {
		return expr.GetPos()
	}
	if step, ok := val.(Step); ok {
		return step.Pos
	}
	return nil
}
