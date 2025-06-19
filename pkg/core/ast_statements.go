// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Updated AST nodes to support multiple assignment targets and treat LValues as Expressions.
// filename: pkg/core/ast_statements.go
// nlines: 104
// risk_rating: HIGH

package core

import (
	"fmt"
	"strings"
)

// AccessorType distinguishes between bracket and dot access in an LValue
type AccessorType int

const (
	BracketAccess AccessorType = iota // e.g., a[expression]
	DotAccess                         // e.g., a.field
)

// AccessorNode represents one part of an lvalue's accessor chain (e.g., "[index]" or ".field")
type AccessorNode struct {
	Pos        *Position
	Type       AccessorType
	IndexOrKey Expression // For BracketAccess (LBRACK expression RBRACK)
	FieldName  string     // For DotAccess (DOT IDENTIFIER)
}

func (an *AccessorNode) String() string {
	if an.Type == BracketAccess {
		return fmt.Sprintf("[%s]", an.IndexOrKey.String())
	}
	return fmt.Sprintf(".%s", an.FieldName)
}

// LValueNode represents the left-hand side of an assignment that can be complex
type LValueNode struct {
	Pos        *Position
	Identifier string         // The base variable name (e.g., 'a' in a["key"])
	Accessors  []AccessorNode // Sequence of bracket or dot accessors
}

func (n *LValueNode) GetPos() *Position { return n.Pos }
func (n *LValueNode) String() string {
	var sb strings.Builder
	sb.WriteString(n.Identifier)
	for _, acc := range n.Accessors {
		sb.WriteString(acc.String())
	}
	return sb.String()
}

// ADDED: This marker method makes LValueNode satisfy the Expression interface.
func (n *LValueNode) expressionNode() {}

// ParamSpec defines a parameter in a procedure signature.
type ParamSpec struct {
	Name         string
	DefaultValue interface{} // For optional parameters
}

// Procedure represents a user-defined function.
type Procedure struct {
	Pos               *Position
	Name              string
	RequiredParams    []string
	OptionalParams    []ParamSpec // Name and default value
	Variadic          bool        // If the last param is variadic
	VariadicParamName string
	ReturnVarNames    []string // Names of variables to return
	Steps             []Step
	OriginalSignature string            // For debugging/LSP
	Metadata          map[string]string // Procedure-level metadata
}

func (p *Procedure) GetPos() *Position { return p.Pos }

// Step represents a single statement or control flow structure in a procedure.
type Step struct {
	Pos  *Position
	Type string // e.g., "set", "call", "if", "return", "emit", "must", "fail", "clear_error", "ask", "on_error", "on_event"

	// For "set":
	// MODIFIED: Changed LValue *LValueNode to LValues []Expression to support multiple assignments.
	LValues []Expression // Target(s) of the assignment, can be complex (e.g., var[index].field)
	Value   Expression   // RHS expression for set

	// For "call":
	Call *CallableExprNode // Details of the function/tool call

	// For "if", "while", "must" (conditional variant):
	Cond Expression // Condition expression

	// For "if", "while", "for", "on_error", "on_event":
	Body []Step // Block of steps

	// For "if":
	Else []Step // Else block for if statements

	// For "for each":
	LoopVarName string     // Variable for each item in the loop (e.g., 'item' in 'for each myList')
	Collection  Expression // The collection expression to iterate over

	// For "ask", "on_event":
	PromptExpr Expression // The prompt expression for 'ask', or the event name expression for 'on_event'
	AskIntoVar string     // Optional variable to store the result of 'ask' or the event payload

	// For "return", "emit", "fail", "must" (unconditional variant):
	Values []Expression // For return statements that might return multiple values
}

func (s *Step) GetPos() *Position { return s.Pos }
