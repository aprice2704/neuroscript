// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines the core, foundational Kind enum for all AST nodes.
// filename: pkg/types/kind.go
// nlines: 30
// risk_rating: LOW

package types

// Kind represents the type of an AST node.
// It's a stable enum; add new kinds at the end only.
type Kind uint8

const (
	KindUnknown Kind = iota // Represents an uninitialized or error node

	// Top-Level & Declarations
	KindProgram
	KindCommandBlock
	KindProcedureDecl
	KindOnEventDecl
	KindMetadataLine
	KindSecretRef

	// Statements
	KindStep
	KindExpressionStmt

	// Expressions
	KindCallableExpr
	KindVariable
	KindPlaceholder
	KindLastResult
	KindEvalExpr
	KindStringLiteral
	KindNumberLiteral
	KindBooleanLiteral
	KindNilLiteral
	KindListLiteral
	KindMapLiteral
	KindElementAccess
	KindUnaryOp
	KindBinaryOp
	KindTypeOfExpr
	KindLValue
)
