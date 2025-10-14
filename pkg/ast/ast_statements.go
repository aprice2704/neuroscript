// NeuroScript Version: 0.8.0
// File version: 29
// Purpose: Adds the TestString() method to LValueNode to satisfy the Expression interface.
// filename: pkg/ast/ast_statements.go
// nlines: 135+
// risk_rating: LOW

package ast

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// AccessorType defines how an element is accessed (e.g., by key or index).
type AccessorType int

const (
	BracketAccess AccessorType = iota
	DotAccess
)

// AccessorNode represents a single part of an element access chain (e.g., `[key]` or `.field`).
type AccessorNode struct {
	BaseNode
	Type       AccessorType
	Key        Expression
	IsOptional bool
}

// LValueNode represents a "left-value" in an assignment, which is a target for a set operation.
type LValueNode struct {
	BaseNode
	Identifier string
	Accessors  []*AccessorNode
}

func (n *LValueNode) String() string {
	// A full string representation would require traversing accessors.
	var sb strings.Builder
	sb.WriteString(n.Identifier)
	for _, acc := range n.Accessors {
		switch acc.Type {
		case DotAccess:
			sb.WriteString(".")
			// Assuming key is a string literal for dot access
			if key, ok := acc.Key.(*StringLiteralNode); ok {
				sb.WriteString(key.Value)
			} else {
				sb.WriteString("<invalid_dot_key>")
			}
		case BracketAccess:
			sb.WriteString("[")
			if acc.Key != nil {
				sb.WriteString(acc.Key.String())
			}
			sb.WriteString("]")
		}
	}
	return sb.String()
}
func (n *LValueNode) TestString() string {
	// For testing, the simple string representation is sufficient and unambiguous.
	return n.String()
}
func (n *LValueNode) expressionNode() {}

// ParamSpec defines a parameter for a procedure.
type ParamSpec struct {
	BaseNode
	Name    string
	Default lang.Value // For optional parameters
}

type Procedure struct {
	BaseNode
	name              string
	Metadata          map[string]string
	Comments          []*Comment
	BlankLinesBefore  int
	RequiredParams    []string
	OptionalParams    []*ParamSpec
	Variadic          bool
	VariadicParamName string
	ReturnVarNames    []string
	ErrorHandlers     []*Step
	Steps             []Step
}

func (p *Procedure) SetName(name string) {
	p.name = name
}

func (p *Procedure) Name() string {
	return p.name
}

func (p *Procedure) IsCallable() {}

// Step represents a single statement or instruction within a procedure's body.
type Step struct {
	BaseNode
	Comments         []*Comment
	BlankLinesBefore int
	Type             string
	LValues          []*LValueNode
	Values           []Expression
	Cond             Expression
	Body             []Step
	ElseBody         []Step
	LoopVarName      string
	IndexVarName     string
	Collection       Expression
	Call             *CallableExprNode
	OnEvent          *OnEventDecl
	AskStmt          *AskStmt
	PromptUserStmt   *PromptUserStmt
	WhisperStmt      *WhisperStmt // Added for 'whisper' statement
	IsFinal          bool
	ErrorName        string
	tool             interfaces.Tool
	ExpressionStmt   *ExpressionStatementNode
}

func (s *Step) String() string {
	if s == nil {
		return "<nil step>"
	}
	// Provides a basic representation. Could be expanded to show more detail.
	return fmt.Sprintf("Step(%s)", s.Type)
}

func (s *Step) expressionNode()    {}
func (s *Step) TestString() string { return s.String() }

// AskStmt represents the structured components of an 'ask' statement.
type AskStmt struct {
	BaseNode
	AgentModelExpr Expression
	PromptExpr     Expression
	WithOptions    Expression
	IntoTarget     *LValueNode
}

// PromptUserStmt represents the structured components of a 'promptuser' statement.
type PromptUserStmt struct {
	BaseNode
	PromptExpr Expression
	IntoTarget *LValueNode
}

// WhisperStmt represents the structured components of a 'whisper' statement.
type WhisperStmt struct {
	BaseNode
	Handle Expression // The handle or channel to whisper to
	Value  Expression // The value to be whispered
}

// ExpressionStatementNode represents a statement that consists of a single expression,
// like a standalone 'must' or a function call. The result of the expression is discarded.
type ExpressionStatementNode struct {
	BaseNode
	Expression Expression
}

func (n *ExpressionStatementNode) isNode()      {}
func (n *ExpressionStatementNode) isStatement() {}
func (n *ExpressionStatementNode) String() string {
	if n.Expression != nil {
		return n.Expression.String()
	}
	return "<nil_expr_stmt>"
}
