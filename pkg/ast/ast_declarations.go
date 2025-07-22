// filename: pkg/ast/ast_declarations.go
// NeuroScript Version: 0.6.0
// File version: 5
// Purpose: Added Metadata, Comments, and BlankLinesBefore fields to OnEventDecl.
// nlines: 25+
// risk_rating: MEDIUM

package ast

// OnEventDecl represents a top-level 'on event ...' declaration.
type OnEventDecl struct {
	BaseNode
	BlankLinesBefore int
	Metadata         map[string]string // Added for consistency
	Comments         []*Comment        // Added for consistency
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
