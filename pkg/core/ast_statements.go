// filename: pkg/core/ast_statements.go
// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Added the 'Commands' field to the Program struct to support command blocks.
// nlines: 107
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
	ErrorHandlers     []*Step
}

func (p *Procedure) GetPos() *Position { return p.Pos }

// Step represents a single statement or control flow structure in a procedure.
type Step struct {
	Pos  *Position
	Type string

	LValues []Expression
	Value   Expression

	Call *CallableExprNode

	Cond Expression

	Body []Step

	Else []Step

	LoopVarName string
	Collection  Expression

	PromptExpr Expression
	AskIntoVar string

	Values []Expression
}

func (s *Step) GetPos() *Position { return s.Pos }
