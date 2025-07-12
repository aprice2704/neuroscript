// filename: pkg/ast/ast_declarations.go
// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Augmented declaration nodes with BaseNode to conform to the Node interface.
// nlines: 25
// risk_rating: MEDIUM

package ast

import "github.com/aprice2704/neuroscript/pkg/types"

// OnEventDecl represents a top-level 'on event ...' declaration,
// as specified in the o3-1 plan.
type OnEventDecl struct {
	BaseNode
	Pos           *types.Position
	EventNameExpr Expression
	HandlerName   string
	EventVarName  string
	Body          []Step
}

// GetPos returns the legacy position field. It also satisfies the Node interface.
func (n *OnEventDecl) GetPos() *types.Position { return n.Pos }

// MetadataLine represents a single `:: key: value` line associated with a declaration.
type MetadataLine struct {
	BaseNode
	Pos   *types.Position
	Key   string
	Value string
}

// GetPos returns the legacy position field. It also satisfies the Node interface.
func (n *MetadataLine) GetPos() *types.Position { return n.Pos }
