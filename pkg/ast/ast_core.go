// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines the core interfaces, enums, and base structs for the AST.
// filename: pkg/ast/ast_core.go
// nlines: 85
// risk_rating: HIGH

package ast

import "fmt"

// --- Core AST Contract ---

// Position defines a 1-based location in a source file.
type Position struct {
	Line int
	Col  int
}

// String returns a human-readable representation of the position.
// Note: This does not include the filename, which is stored on the Tree.
func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

// Kind is a stable, unsigned integer identifying the type of an AST node.
type Kind uint8

// Node is the interface that all nodes in the AST must implement.
type Node interface {
	// Pos returns the starting position of the node.
	Pos() Position
	// End returns the ending position of the node.
	End() Position
	// Kind returns the specific type of the node.
	Kind() Kind
}

// Tree represents a fully parsed source file.
type Tree struct {
	Root     Node
	Comments []*Comment
	Filepath string // The full path to the source file.
}

// Comment represents a single comment in the source code.
type Comment struct {
	BaseNode
	Text string
}

// Kind returns the specific type for this node.
func (n *Comment) Kind() Kind { return KindComment }

// BaseNode provides the common fields and methods for all AST nodes.
// It is intended to be embedded in specific node structs.
type BaseNode struct {
	StartPos Position
	StopPos  Position
}

// Pos returns the starting position of the node.
func (n *BaseNode) Pos() Position { return n.StartPos }

// End returns the ending position of the node.
func (n *BaseNode) End() Position { return n.StopPos }

// --- Node Kind Enum ---

const (
	KindUnknown Kind = iota

	// Declarations
	KindProgram
	KindCommandNode
	KindProcedure
	KindOnEventDecl

	// Statements
	KindStep
	KindExpressionStmt

	// Expressions
	KindCallableExpr
	KindVariable
	KindPlaceholder
	KindLast
	KindEval
	KindStringLiteral
	KindNumberLiteral
	KindBooleanLiteral
	KindListLiteral
	KindMapLiteral
	KindElementAccess
	KindUnaryOp
	KindBinaryOp
	KindTypeOf
	KindNilLiteral
	KindLValue
	KindSecretRef // As per plan

	// Misc
	KindComment
	KindError
	KindMapEntry
)
