// NeuroScript Version: 0.5.2
// File version: 17
// Purpose: Corrected GetPos method signatures to return *lang.Position, satisfying the Expression interface.
// filename: pkg/ast/ast_statements.go
// nlines: 85
// risk_rating: MEDIUM

package ast

import "github.com/aprice2704/neuroscript/pkg/lang"

// AccessorType defines how an element is accessed (e.g., by key or index).
type AccessorType int

const (
	BracketAccess AccessorType = iota
	DotAccess
)

// AccessorNode represents a single part of an element access chain (e.g., `[key]` or `.field`).
type AccessorNode struct {
	Pos        *lang.Position
	Type       AccessorType
	Key        Expression
	IsOptional bool
}

// LValueNode represents a "left-value" in an assignment, which is a target for a set operation.
type LValueNode struct {
	Position   lang.Position
	Identifier string
	Accessors  []*AccessorNode
}

// FIX: Return a pointer to the position to satisfy the Expression interface.
func (n *LValueNode) GetPos() *lang.Position { return &n.Position }
func (n *LValueNode) String() string {
	// A full string representation would require traversing accessors.
	return n.Identifier
}
func (n *LValueNode) expressionNode() {}

// ParamSpec defines a parameter for a procedure.
type ParamSpec struct {
	Name    string
	Default lang.Value // For optional parameters
}

type Procedure struct {
	Position          lang.Position
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

// FIX: Return a pointer to the position.
func (p *Procedure) GetPos() *lang.Position {
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
	Position     lang.Position
	Type         string
	LValues      []*LValueNode
	Values       []Expression
	Cond         Expression
	Body         []Step
	ElseBody     []Step
	LoopVarName  string
	IndexVarName string
	Collection   Expression
	Call         *CallableExprNode
	OnEvent      *OnEventDecl
	AskIntoVar   string
	IsFinal      bool
	ErrorName    string
	tool         lang.Tool
}

// FIX: Return a pointer to the position.
func (s *Step) GetPos() *lang.Position {
	return &s.Position
}
