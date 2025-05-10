// NeuroScript Version: 0.3.1
// File version: 0.0.5 // Correct Step struct, clarify Expression interface, add ErrorNode
// filename: pkg/core/ast.go
package core

import (
	"fmt"
)

// --- Position Information ---

// Position represents a location in the source file.
type Position struct {
	Line   int
	Column int
	File   string
}

func (p *Position) GetPos() *Position { return p }

func (p *Position) String() string {
	if p == nil {
		return "<nil position>"
	}
	filePart := ""
	if p.File != "" && p.File != "<unknown>" && p.File != "<input stream>" {
		filePart = fmt.Sprintf("%s:", p.File)
	}
	return fmt.Sprintf("%s%d:%d", filePart, p.Line, p.Column)
}

// --- Program Root Node ---

type Program struct {
	Metadata   map[string]string
	Procedures map[string]*Procedure
	Pos        *Position
}

func (p *Program) GetPos() *Position { return p.Pos }

// --- Expression Interface ---

// Expression is implemented by all AST nodes that evaluate to a value.
type Expression interface {
	GetPos() *Position
	expressionNode() // Marker method
}

// --- AST Node Base (Optional embedding for common fields) ---
// Not strictly necessary but can be useful. For now, GetPos() is in each.

// --- Expression Node Types ---

type ErrorNode struct {
	Pos     *Position
	Message string
}

func (n *ErrorNode) GetPos() *Position { return n.Pos }
func (n *ErrorNode) expressionNode()   {}

type CallTarget struct {
	Pos    *Position
	IsTool bool
	Name   string
}

func (ct *CallTarget) GetPos() *Position { return ct.Pos }

type CallableExprNode struct {
	Pos       *Position
	Target    CallTarget
	Arguments []Expression
}

func (n *CallableExprNode) GetPos() *Position { return n.Pos }
func (n *CallableExprNode) expressionNode()   {}

type VariableNode struct {
	Pos  *Position
	Name string
}

func (n *VariableNode) GetPos() *Position { return n.Pos }
func (n *VariableNode) expressionNode()   {}

type PlaceholderNode struct {
	Pos  *Position
	Name string
}

func (n *PlaceholderNode) GetPos() *Position { return n.Pos }
func (n *PlaceholderNode) expressionNode()   {}

type LastNode struct {
	Pos *Position
}

func (n *LastNode) GetPos() *Position { return n.Pos }
func (n *LastNode) expressionNode()   {}

type EvalNode struct {
	Pos      *Position
	Argument Expression
}

func (n *EvalNode) GetPos() *Position { return n.Pos }
func (n *EvalNode) expressionNode()   {}

type StringLiteralNode struct {
	Pos   *Position
	Value string
	IsRaw bool
}

func (n *StringLiteralNode) GetPos() *Position { return n.Pos }
func (n *StringLiteralNode) expressionNode()   {}

type NumberLiteralNode struct {
	Pos   *Position
	Value interface{} // int64 or float64
}

func (n *NumberLiteralNode) GetPos() *Position { return n.Pos }
func (n *NumberLiteralNode) expressionNode()   {}

type BooleanLiteralNode struct {
	Pos   *Position
	Value bool
}

func (n *BooleanLiteralNode) GetPos() *Position { return n.Pos }
func (n *BooleanLiteralNode) expressionNode()   {}

type ListLiteralNode struct {
	Pos      *Position
	Elements []Expression
}

func (n *ListLiteralNode) GetPos() *Position { return n.Pos }
func (n *ListLiteralNode) expressionNode()   {}

type MapEntryNode struct {
	Pos   *Position // Position of the key
	Key   *StringLiteralNode
	Value Expression
}

func (n *MapEntryNode) GetPos() *Position {
	if n.Key != nil {
		return n.Key.Pos
	}
	return n.Pos // Fallback
}

type MapLiteralNode struct {
	Pos     *Position
	Entries []*MapEntryNode // Changed to slice of pointers
}

func (n *MapLiteralNode) GetPos() *Position { return n.Pos }
func (n *MapLiteralNode) expressionNode()   {}

type ElementAccessNode struct {
	Pos        *Position
	Collection Expression
	Accessor   Expression
}

func (n *ElementAccessNode) GetPos() *Position { return n.Pos }
func (n *ElementAccessNode) expressionNode()   {}

type UnaryOpNode struct {
	Pos      *Position
	Operator string
	Operand  Expression
}

func (n *UnaryOpNode) GetPos() *Position { return n.Pos }
func (n *UnaryOpNode) expressionNode()   {}

type BinaryOpNode struct {
	Pos      *Position
	Left     Expression
	Operator string
	Right    Expression
}

func (n *BinaryOpNode) GetPos() *Position { return n.Pos }
func (n *BinaryOpNode) expressionNode()   {}

// --- Procedure & Step Structures ---

type ParamSpec struct {
	Name         string
	DefaultValue interface{}
}

type Procedure struct {
	Pos               *Position
	Name              string
	RequiredParams    []string
	OptionalParams    []ParamSpec
	Variadic          bool
	VariadicParamName string
	ReturnVarNames    []string
	Steps             []Step // Use a slice of Step values
	OriginalSignature string
	Metadata          map[string]string
}

func (p *Procedure) GetPos() *Position { return p.Pos }

// Step represents a single operation.
// Specific fields are used based on 'Type'.
type Step struct {
	Pos    *Position
	Type   string            // "set", "call", "return", "emit", "if", "while", "for", "must", "mustbe", "fail", "clear_error", "ask", "break", "continue", "on_error"
	Target string            // Variable name (set, for_each iterVar, ask intoVar)
	Cond   Expression        // Condition (if, while), Collection (for_each)
	Value  Expression        // RHS (set), Emit value (emit), Must condition (must), Fail message (fail), Prompt (ask)
	Values []Expression      // For 'return' with multiple values
	Body   []Step            // For blocks like if, else, while, for_each, on_error
	Else   []Step            // For 'else' part of 'if'
	Call   *CallableExprNode // For 'call' statement
	// Metadata  map[string]string // If steps can have metadata directly
	// StepError *ErrorNode     // If we want to embed errors for malformed steps (alternative to listener.errors)
}

func (s *Step) GetPos() *Position { return s.Pos }

// String() methods for debugging (Example, can be expanded)
func (s StringLiteralNode) String() string {
	if s.IsRaw {
		return fmt.Sprintf("```%s```", s.Value)
	}
	return fmt.Sprintf("%q", s.Value)
}

func (c CallableExprNode) String() string {
	targetStr := c.Target.Name
	if c.Target.IsTool {
		targetStr = "tool." + targetStr
	}
	argSummary := fmt.Sprintf("%d args", len(c.Arguments))
	return fmt.Sprintf("%s(%s)", targetStr, argSummary)
}

// Helper to try and get position from common AST node types (used for error reporting)
// This is a simplified version; a more robust way would be to ensure all relevant
// AST components have GetPos or use the antlr token directly in the builder.
func getExpressionPosition(val interface{}) *Position {
	if expr, ok := val.(Expression); ok {
		return expr.GetPos()
	}
	// Add more specific types if needed, or rely on caller to provide token position
	return nil
}
