// filename: pkg/ast/ast_statements.go
// NeuroScript Version: 0.5.2
// File version: 22
// Purpose: Removed redundant Position/Pos fields and GetPos methods to unify position handling via BaseNode.
// nlines: 90+
// risk_rating: MEDIUM

package ast

import (
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
	return n.Identifier
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
	AskIntoVar       string
	IsFinal          bool
	ErrorName        string
	tool             interfaces.Tool
	ExpressionStmt   *ExpressionStatementNode
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
