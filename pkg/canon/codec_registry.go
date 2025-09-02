// NeuroScript Version: 0.6.3
// File version: 5
// Purpose: Corrected undefined KindAccountName to KindSecretRef.
// filename: pkg/canon/codec_registry.go
// nlines: 70
// risk_rating: LOW

package canon

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// NodeCodec defines the symmetrical encoder and decoder functions for a given AST node type.
type NodeCodec struct {
	EncodeFunc func(v *canonVisitor, n ast.Node) error
	DecodeFunc func(r *canonReader) (ast.Node, error)
}

// CodecRegistry maps each AST node Kind to its specific encoder/decoder implementation.
var CodecRegistry map[types.Kind]NodeCodec

func init() {
	CodecRegistry = map[types.Kind]NodeCodec{
		// Top-Level & Declarations
		types.KindProgram:       {EncodeFunc: encodeProgram, DecodeFunc: decodeProgram},
		types.KindCommandBlock:  {EncodeFunc: encodeCommandBlock, DecodeFunc: decodeCommandBlock},
		types.KindProcedureDecl: {EncodeFunc: encodeProcedure, DecodeFunc: decodeProcedure},
		types.KindOnEventDecl:   {EncodeFunc: encodeOnEventDecl, DecodeFunc: decodeOnEventDecl},
		types.KindSecretRef:     {EncodeFunc: encodeSecretRef, DecodeFunc: decodeSecretRef},
		// KindMetadataLine is handled within other nodes, not as a standalone.

		// Statements
		types.KindStep:           {EncodeFunc: encodeStep, DecodeFunc: decodeStep},
		types.KindExpressionStmt: {EncodeFunc: encodeExpressionStmt, DecodeFunc: decodeExpressionStmt},

		// Expressions
		types.KindCallableExpr:  {EncodeFunc: encodeCallableExpr, DecodeFunc: decodeCallableExpr},
		types.KindVariable:      {EncodeFunc: encodeVariable, DecodeFunc: decodeVariable},
		types.KindPlaceholder:   {EncodeFunc: encodePlaceholder, DecodeFunc: decodePlaceholder},
		types.KindLastResult:    {EncodeFunc: encodeLast, DecodeFunc: decodeLast},
		types.KindEvalExpr:      {EncodeFunc: encodeEval, DecodeFunc: decodeEval},
		types.KindTypeOfExpr:    {EncodeFunc: encodeTypeOf, DecodeFunc: decodeTypeOf},
		types.KindBinaryOp:      {EncodeFunc: encodeBinaryOp, DecodeFunc: decodeBinaryOp},
		types.KindUnaryOp:       {EncodeFunc: encodeUnaryOp, DecodeFunc: decodeUnaryOp},
		types.KindElementAccess: {EncodeFunc: encodeElementAccess, DecodeFunc: decodeElementAccess},
		types.KindLValue:        {EncodeFunc: encodeLValue, DecodeFunc: decodeLValue},

		// Literals
		types.KindStringLiteral:  {EncodeFunc: encodeStringLiteral, DecodeFunc: decodeStringLiteral},
		types.KindNumberLiteral:  {EncodeFunc: encodeNumberLiteral, DecodeFunc: decodeNumberLiteral},
		types.KindBooleanLiteral: {EncodeFunc: encodeBooleanLiteral, DecodeFunc: decodeBooleanLiteral},
		types.KindNilLiteral:     {EncodeFunc: encodeNilLiteral, DecodeFunc: decodeNilLiteral},

		// Collections
		types.KindListLiteral: {EncodeFunc: encodeListLiteral, DecodeFunc: decodeListLiteral},
		types.KindMapLiteral:  {EncodeFunc: encodeMapLiteral, DecodeFunc: decodeMapLiteral},
		types.KindMapEntry:    {EncodeFunc: encodeMapEntry, DecodeFunc: decodeMapEntry},

		// Aliases (handled by their main expression type)
		types.KindLast:   {EncodeFunc: encodeLast, DecodeFunc: decodeLast},
		types.KindEval:   {EncodeFunc: encodeEval, DecodeFunc: decodeEval},
		types.KindTypeOf: {EncodeFunc: encodeTypeOf, DecodeFunc: decodeTypeOf},

		// Ask/Prompt statements are not yet implemented in the new codec system.
		// types.KindAskStmt:        {EncodeFunc: encodeAskStmt, DecodeFunc: decodeAskStmt},
		// types.KindPromptUserStmt: {EncodeFunc: encodePromptUserStmt, DecodeFunc: decodePromptUserStmt},
	}
}
