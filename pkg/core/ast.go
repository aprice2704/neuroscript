// NeuroScript Version: 0.3.1
// File version: 0.0.7 // Reverted Expression interface, removed non-essential String() methods. Added TypeOfNode, NilLiteralNode.
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
	// String() string // Removed from interface to avoid breaking changes
}

// --- Expression Node Types ---

type ErrorNode struct {
	Pos     *Position
	Message string
}

func (n *ErrorNode) GetPos() *Position { return n.Pos }
func (n *ErrorNode) expressionNode()   {}

// Existing String() method if it was present in your original file, or can be added back if desired for this specific node.
// For now, to minimize changes, I am not re-adding String() methods unless they were in the user-provided base.
// func (n *ErrorNode) String() string {
//	 return fmt.Sprintf("ERROR(%s at %s)", n.Message, n.Pos.String())
// }

type CallTarget struct {
	Pos    *Position
	IsTool bool
	Name   string
}

func (ct *CallTarget) GetPos() *Position { return ct.Pos }

// func (ct *CallTarget) String() string { ... } // Removed if added by me

type CallableExprNode struct {
	Pos       *Position
	Target    CallTarget
	Arguments []Expression
}

func (n *CallableExprNode) GetPos() *Position { return n.Pos }
func (n *CallableExprNode) expressionNode()   {}

// func (n *CallableExprNode) String() string { ... } // Removed if added by me

type VariableNode struct {
	Pos  *Position
	Name string
}

func (n *VariableNode) GetPos() *Position { return n.Pos }
func (n *VariableNode) expressionNode()   {}

// func (n *VariableNode) String() string    { return n.Name } // Removed if added by me

type PlaceholderNode struct {
	Pos  *Position
	Name string
}

func (n *PlaceholderNode) GetPos() *Position { return n.Pos }
func (n *PlaceholderNode) expressionNode()   {}

// func (n *PlaceholderNode) String() string    { return fmt.Sprintf("{{%s}}", n.Name) } // Removed if added by me

type LastNode struct {
	Pos *Position
}

func (n *LastNode) GetPos() *Position { return n.Pos }
func (n *LastNode) expressionNode()   {}

// func (n *LastNode) String() string    { return "LAST" } // Removed if added by me

type EvalNode struct {
	Pos      *Position
	Argument Expression
}

func (n *EvalNode) GetPos() *Position { return n.Pos }
func (n *EvalNode) expressionNode()   {}

// func (n *EvalNode) String() string { ... } // Removed if added by me

type StringLiteralNode struct {
	Pos   *Position
	Value string
	IsRaw bool
}

func (n *StringLiteralNode) GetPos() *Position { return n.Pos }
func (n *StringLiteralNode) expressionNode()   {}

// User had a String() method for this separately, that should be fine if it's not part of the interface.

type NumberLiteralNode struct {
	Pos   *Position
	Value interface{} // int64 or float64
}

func (n *NumberLiteralNode) GetPos() *Position { return n.Pos }
func (n *NumberLiteralNode) expressionNode()   {}

// func (n *NumberLiteralNode) String() string    { return fmt.Sprintf("%v", n.Value) } // Removed if added by me

type BooleanLiteralNode struct {
	Pos   *Position
	Value bool
}

func (n *BooleanLiteralNode) GetPos() *Position { return n.Pos }
func (n *BooleanLiteralNode) expressionNode()   {}

// func (n *BooleanLiteralNode) String() string    { return fmt.Sprintf("%t", n.Value) } // Removed if added by me

// func (n *NilLiteralNode) String() string    { return "nil" } // Removed if added by me

type ListLiteralNode struct {
	Pos      *Position
	Elements []Expression
}

func (n *ListLiteralNode) GetPos() *Position { return n.Pos }
func (n *ListLiteralNode) expressionNode()   {}

// func (n *ListLiteralNode) String() string { ... } // Removed if added by me

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

// func (n *MapEntryNode) String() string { ... } // Removed if added by me

type MapLiteralNode struct {
	Pos     *Position
	Entries []*MapEntryNode
}

func (n *MapLiteralNode) GetPos() *Position { return n.Pos }
func (n *MapLiteralNode) expressionNode()   {}

// func (n *MapLiteralNode) String() string { ... } // Removed if added by me

type ElementAccessNode struct {
	Pos        *Position
	Collection Expression
	Accessor   Expression
}

func (n *ElementAccessNode) GetPos() *Position { return n.Pos }
func (n *ElementAccessNode) expressionNode()   {}

// func (n *ElementAccessNode) String() string { ... } // Removed if added by me

type UnaryOpNode struct {
	Pos      *Position
	Operator string
	Operand  Expression
}

func (n *UnaryOpNode) GetPos() *Position { return n.Pos }
func (n *UnaryOpNode) expressionNode()   {}

// func (n *UnaryOpNode) String() string { ... } // Removed if added by me

type BinaryOpNode struct {
	Pos      *Position
	Left     Expression
	Operator string
	Right    Expression
}

func (n *BinaryOpNode) GetPos() *Position { return n.Pos }
func (n *BinaryOpNode) expressionNode()   {}

// func (n *BinaryOpNode) String() string { ... } // Removed if added by me

// TypeOfNode represents the 'typeof' operator.
type TypeOfNode struct {
	Pos      *Position
	Argument Expression
}

func (n *TypeOfNode) GetPos() *Position { return n.Pos }
func (n *TypeOfNode) expressionNode()   {}

// func (n *TypeOfNode) String() string { ... } // Removed if added by me

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
}

func (s *Step) GetPos() *Position { return s.Pos }

// String() methods for debugging (Example, can be expanded)
// These were in your original AST file, so keeping them if they are separate from the Expression interface requirement
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
	// Temporarily remove stringJoin or ensure it's defined if needed by original String methods
	// For now, simplifying to avoid dependency on the removed stringJoin
	argSummary := fmt.Sprintf("%d args", len(c.Arguments))
	// argStrings := make([]string, len(c.Arguments))
	// for i, arg := range c.Arguments {
	// 	if arg == nil {
	// 		argStrings[i] = "<nil_arg>"
	// 	} else {
	// 		argStrings[i] = arg.String() // This would still require arg to have String()
	// 	}
	// }
	// return fmt.Sprintf("%s(%s)", targetStr, stringJoin(argStrings, ", "))
	return fmt.Sprintf("%s(%s)", targetStr, argSummary)

}

// Helper to try and get position from common AST node types (used for error reporting)
func getExpressionPosition(val interface{}) *Position {
	if expr, ok := val.(Expression); ok {
		return expr.GetPos()
	}
	return nil
}
