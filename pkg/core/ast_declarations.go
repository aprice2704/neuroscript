// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines top-level declaration nodes for the AST, such as OnEventDecl.
// filename: pkg/core/ast_declarations.go
// nlines: 18
// risk_rating: MEDIUM

package core

// OnEventDecl represents a top-level 'on event ...' declaration,
// as specified in the o3-1 plan.
type OnEventDecl struct {
	Pos           *Position
	EventNameExpr Expression
	EventVarName  string
	Metadata      []*MetadataLine
	Body          []Step
}

// MetadataLine represents a single `:: key: value` line associated with a declaration.
type MetadataLine struct {
	Pos   *Position
	Key   string
	Value string
}
