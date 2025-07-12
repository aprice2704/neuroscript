// filename: pkg/ast/ast_statements.go
// NeuroScript Version: 0.5.2
// File version: 19
// Purpose: Augmented all statement and declaration nodes with BaseNode.
// nlines: 110+
// risk_rating: MEDIUM

package ast

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
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
	Pos        *types.Position
	Type       AccessorType
	Key        Expression
	IsOptional bool
}

// LValueNode represents a "left-value" in an assignment, which is a target for a set operation.
type LValueNode struct {
	BaseNode
	Position   types.Position
	Identifier string
	Accessors  []*AccessorNode
}

// GetPos satisfies the old contract and the new Node interface.
func (n *LValueNode) GetPos() *types.Position { return &n.Position }
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
	Position          types.Position
	name              string
	Metadata          map[string]string
	RequiredParams    []string
	OptionalParams    []*ParamSpec
	Variadic          bool
	VariadicParamName string
	ReturnVarNames    []string
	ErrorHandlers     []*Step
	Steps             []Step
}

// GetPos satisfies the old contract and the new Node interface.
func (p *Procedure) GetPos() *types.Position {
	return &p.Position
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
	Position       types.Position
	Type           string
	LValues        []*LValueNode
	Values         []Expression
	Cond           Expression
	Body           []Step
	ElseBody       []Step
	LoopVarName    string
	IndexVarName   string
	Collection     Expression
	Call           *CallableExprNode
	OnEvent        *OnEventDecl
	AskIntoVar     string
	IsFinal        bool
	ErrorName      string
	tool           interfaces.Tool
	ExpressionStmt *ExpressionStatementNode
}

// GetPos satisfies the old contract and the new Node interface.
func (s *Step) GetPos() *types.Position {
	return &s.Position
}

// ExpressionStatementNode represents a statement that consists of a single expression,
// like a standalone 'must' or a function call. The result of the expression is discarded.
type ExpressionStatementNode struct {
	BaseNode
	Pos        *types.Position
	Expression Expression
}

func (n *ExpressionStatementNode) GetPos() *types.Position { return n.Pos }
func (n *ExpressionStatementNode) isNode()                 {}
func (n *ExpressionStatementNode) isStatement()            {}
func (n *ExpressionStatementNode) String() string {
	if n.Expression != nil {
		return n.Expression.String()
	}
	return "<nil_expr_stmt>"
}
