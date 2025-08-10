// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Defines the core, foundational Kind enum for all AST nodes and adds a String() method for debugging.
// filename: pkg/types/kind.go
// nlines: 70
// risk_rating: LOW

package types

import "strconv"

// Kind represents the type of an AST node.
// It's a stable enum; add new kinds just before KindMarker only.
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

	KindLast
	KindEval
	KindTypeOf
	KindMapEntry
	KindAskStmt
	KindPromptUserStmt

	// ^^^^^^ add new kinds above this ^^^^^^^
	// KindMarker is not a real kind. It is a sentinel value used in tests to
	// ensure all actual kinds are handled in switch statements.
	KindMarker
)

func (k Kind) String() string {
	switch k {
	case KindUnknown:
		return "Unknown"
	case KindProgram:
		return "Program"
	case KindCommandBlock:
		return "CommandBlock"
	case KindProcedureDecl:
		return "ProcedureDecl"
	case KindOnEventDecl:
		return "OnEventDecl"
	case KindMetadataLine:
		return "MetadataLine"
	case KindSecretRef:
		return "SecretRef"
	case KindStep:
		return "Step"
	case KindExpressionStmt:
		return "ExpressionStmt"
	case KindCallableExpr:
		return "CallableExpr"
	case KindVariable:
		return "Variable"
	case KindPlaceholder:
		return "Placeholder"
	case KindLastResult:
		return "LastResult"
	case KindEvalExpr:
		return "EvalExpr"
	case KindStringLiteral:
		return "StringLiteral"
	case KindNumberLiteral:
		return "NumberLiteral"
	case KindBooleanLiteral:
		return "BooleanLiteral"
	case KindNilLiteral:
		return "NilLiteral"
	case KindListLiteral:
		return "ListLiteral"
	case KindMapLiteral:
		return "MapLiteral"
	case KindElementAccess:
		return "ElementAccess"
	case KindUnaryOp:
		return "UnaryOp"
	case KindBinaryOp:
		return "BinaryOp"
	case KindTypeOfExpr:
		return "TypeOfExpr"
	case KindLValue:
		return "LValue"
	case KindLast:
		return "Last"
	case KindEval:
		return "Eval"
	case KindTypeOf:
		return "TypeOf"
	case KindMapEntry:
		return "MapEntry"
	case KindAskStmt:
		return "AskStmt"
	case KindPromptUserStmt:
		return "PromptUserStmt"
	case KindMarker:
		return "Marker"
	default:
		return "Kind(" + strconv.Itoa(int(k)) + ")"
	}
}
