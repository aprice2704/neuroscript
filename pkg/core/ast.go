// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Consolidated event handler representation to use a single 'Events' slice of OnEventDecl.
// filename: pkg/core/ast.go
// nlines: 48
// risk_rating: HIGH

// Package core ast*.go defines the abstract syntax tree (AST) for NeuroScript programs.
//
// Stack invariants used by the AST builder
// ---------------------------------------
// The builder maintains two slices that act as stacks while walking the ANTLR
// parse tree:
//
//   • valueStack []interface{}
//       Holds every “value‑producing” construct (expressions, []Step blocks,
//       literals, etc.) in last‑in‑first‑out order.
//
//   • blockStepStack [][]Step
//       Mirrors nested statement_list contexts so the builder can switch the
//       receiver slice for step‑building.
//
// The rules for interacting with these stacks are:
//
//   1. enterBlockContext(label) MUST be paired with exactly one
//      exitBlockContext() on every code path.  The helper pushes the current
//      *[]Step on blockStepStack, sets currentSteps = new([]Step), and exit…
//      pops and restores.
//
//   2. Top‑level constructs that own a statement_list (`func`, `on event`,
//      `loop`, etc.) NEVER push/pop directly; they rely entirely on the
//      Statement_list enter/exit callbacks.
//
//   3. Helpers (pushValue/popValue/popNValues) always leave valueStack
//      balanced.  After a successful Build() run, len(valueStack) == 0 and
//      len(blockStepStack) == 0.
//
//   4. When building nodes that consume N operands, push them first then pop
//      in reverse order (LIFO).  Violating this order manifests as “value
//      stack size is X at end of program” errors.
//
// Keep these invariants in mind before changing stack‑related code; most hard‑
// to‑trace bugs stem from breaking one of them.

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
	String() string  // For debugging and string representation
}

// Program represents the entire parsed NeuroScript program.
type Program struct {
	Pos        *Position
	Metadata   map[string]string
	Procedures map[string]*Procedure
	Events     []*OnEventDecl // This is the single, correct field for event handlers.
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
	if step, ok := val.(Step); ok { // Step is in ast_statements.go
		return step.GetPos()
	}
	return nil
}
