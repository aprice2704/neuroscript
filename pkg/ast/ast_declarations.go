// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines top-level declaration nodes for the AST, such as OnEventDecl.
// filename: pkg/core/ast_declarations.go
// nlines: 18
// risk_rating: MEDIUM

package ast

import "github.com/aprice2704/neuroscript/pkg/lang"

// OnEventDecl represents a top-level 'on event ...' declaration,
// as specified in the o3-1 plan.
// This is a hypothetical example of the change needed
type OnEventDecl struct {
	Pos           *lang.Position
	EventNameExpr Expression
	HandlerName   string // <--- ADD THIS FIELD
	EventVarName  string
	Body          []Step
}

// MetadataLine represents a single `:: key: value` line associated with a declaration.
type MetadataLine struct {
	Pos   *lang.Position
	Key   string
	Value string
}
