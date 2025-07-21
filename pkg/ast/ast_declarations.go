// filename: pkg/ast/ast_declarations.go
// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Removed redundant Pos fields and GetPos methods to unify position handling via BaseNode.
// nlines: 20+
// risk_rating: MEDIUM

package ast

// OnEventDecl represents a top-level 'on event ...' declaration,
// as specified in the o3-1 plan.
type OnEventDecl struct {
	BaseNode
	BlankLinesBefore int
	EventNameExpr    Expression
	HandlerName      string
	EventVarName     string
	Body             []Step
}

// MetadataLine represents a single `:: key: value` line associated with a declaration.
type MetadataLine struct {
	BaseNode
	Key   string
	Value string
}
