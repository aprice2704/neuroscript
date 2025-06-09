// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Added 'Events' field to Program struct to support top-level event handlers.
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
	File   string // Optional: filename or source identifier
}

func (p *Position) GetPos() *Position { return p }
func (p *Position) String() string {
	if p == nil {
		return "(unknown pos)"
	}
	if p.File != "" {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
	}
	return fmt.Sprintf("line %d, col %d", p.Line, p.Column)
}

// Expression is an interface for all expression AST nodes.
type Expression interface {
	GetPos() *Position
	expressionNode() // Marker method
	String() string  // For debugging and string representation
}

// Program represents the entire parsed NeuroScript program.
type Program struct {
	Pos           *Position
	Metadata      map[string]string
	Procedures    map[string]*Procedure
	Events        []*OnEventDecl
	EventHandlers []*OnEventNode `json:"event_handlers"`
}

func (p *Program) GetPos() *Position { return p.Pos }

// ErrorNode represents a parsing or semantic error encountered during AST construction.
type ErrorNode struct {
	Pos     *Position
	Message string
}

func (n *ErrorNode) GetPos() *Position { return n.Pos }
func (n *ErrorNode) expressionNode()   {}
func (n *ErrorNode) String() string    { return fmt.Sprintf("Error(%s): %s", n.Pos, n.Message) }

// --- Helper for getting position from any node that implements Expression or Step ---
func getExpressionPosition(val interface{}) *Position {
	if expr, ok := val.(Expression); ok {
		return expr.GetPos()
	}
	if step, ok := val.(Step); ok { // Step is in ast_statements.go
		return step.GetPos()
	}
	return nil
}

// OnEventNode represents an 'on event' block, a top-level construct.
type OnEventNode struct {
	Pos             *Position
	EventNameExpr   Expression
	PayloadVariable string
	Steps           []Step
}

func (n *OnEventNode) GetPos() *Position { return n.Pos }
func (n *OnEventNode) GetSteps() []Step  { return n.Steps } // Implement StepContainer
func (n *OnEventNode) node()             {}
func (n *OnEventNode) String() string    { return "OnEventNode" }
