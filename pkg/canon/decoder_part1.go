// NeuroScript Version: 0.6.2
// File version: 34
// Purpose: Tidy: Removes verbose debug logging now that the canonicalization issues are resolved.
// filename: pkg/canon/decoder_part1.go
// nlines: 170
// risk_rating: HIGH

package canon

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Decode reconstructs an AST Tree from its canonical binary representation.
// It first checks for a 4-byte magic number to ensure data integrity and version compatibility.
func Decode(blob []byte) (*ast.Tree, error) {
	// Re-check the magic number using the latest KindMarker
	magicNumber := []byte{'N', 'S', 'C', byte(types.KindMarker)}
	if len(blob) < len(magicNumber) {
		return nil, fmt.Errorf("cannot decode blob: too short to be valid")
	}

	if !bytes.Equal(blob[:len(magicNumber)], magicNumber) {
		return nil, fmt.Errorf("invalid magic number: blob is not a valid canonical AST or was created with an incompatible version")
	}

	// Start reading *after* the magic number.
	reader := &canonReader{r: bytes.NewReader(blob[len(magicNumber):])}

	root, err := reader.readNode()
	if err != nil {
		return nil, fmt.Errorf("failed to decode root node: %w", err)
	}
	program, ok := root.(*ast.Program)
	if !ok {
		return nil, fmt.Errorf("decoded root node is not a *ast.Program, but %T", root)
	}
	return &ast.Tree{Root: program}, nil
}

type canonReader struct {
	r       *bytes.Reader
	history []string
}

func (r *canonReader) readNode() (ast.Node, error) {
	offset := r.r.Size() - int64(r.r.Len())

	kindVal, err := r.readVarint()
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read node kind: %w", err)
	}
	kind := types.Kind(kindVal)
	r.history = append(r.history, fmt.Sprintf("%v", kind))

	switch kind {
	// Structural Nodes
	case types.KindProgram:
		return r.readProgram()
	case types.KindProcedureDecl:
		return r.readProcedure()
	case types.KindStep:
		return r.readStep()
	case types.KindCommandBlock:
		return r.readCommand()
	case types.KindOnEventDecl:
		return r.readOnEventDecl()

	// Expression Nodes
	case types.KindLValue:
		return r.readLValue()
	case types.KindStringLiteral:
		val, err := r.readString()
		if err != nil {
			return nil, err
		}
		isRaw, err := r.readBool()
		if err != nil {
			return nil, err
		}
		return &ast.StringLiteralNode{BaseNode: ast.BaseNode{NodeKind: kind}, Value: val, IsRaw: isRaw}, nil
	case types.KindNumberLiteral:
		val, err := r.readNumber()
		if err != nil {
			return nil, err
		}
		return &ast.NumberLiteralNode{BaseNode: ast.BaseNode{NodeKind: kind}, Value: val}, nil
	case types.KindBooleanLiteral:
		b, err := r.readBool()
		if err != nil {
			return nil, err
		}
		return &ast.BooleanLiteralNode{BaseNode: ast.BaseNode{NodeKind: kind}, Value: b}, nil
	case types.KindNilLiteral:
		return &ast.NilLiteralNode{BaseNode: ast.BaseNode{NodeKind: kind}}, nil
	case types.KindCallableExpr:
		return r.readCallableExpr()
	case types.KindVariable:
		name, err := r.readString()
		if err != nil {
			return nil, err
		}
		return &ast.VariableNode{BaseNode: ast.BaseNode{NodeKind: kind}, Name: name}, nil
	case types.KindBinaryOp:
		return r.readBinaryOp()
	case types.KindUnaryOp:
		return r.readUnaryOp()
	case types.KindMapLiteral:
		return r.readMapLiteral()
	case types.KindListLiteral:
		return r.readListLiteral()
	case types.KindElementAccess:
		return r.readElementAccess()
	case types.KindSecretRef:
		return r.readSecretRef()
	case types.KindPlaceholder:
		name, err := r.readString()
		if err != nil {
			return nil, err
		}
		return &ast.PlaceholderNode{BaseNode: ast.BaseNode{NodeKind: kind}, Name: name}, nil
	case types.KindLast, types.KindLastResult: // FIX: Handle alias
		return &ast.LastNode{BaseNode: ast.BaseNode{NodeKind: kind}}, nil
	case types.KindEval, types.KindEvalExpr: // FIX: Handle alias
		arg, err := r.readNode()
		if err != nil {
			return nil, err
		}
		return &ast.EvalNode{BaseNode: ast.BaseNode{NodeKind: kind}, Argument: arg.(ast.Expression)}, nil
	case types.KindTypeOf, types.KindTypeOfExpr: // FIX: Handle alias
		arg, err := r.readNode()
		if err != nil {
			return nil, err
		}
		return &ast.TypeOfNode{BaseNode: ast.BaseNode{NodeKind: kind}, Argument: arg.(ast.Expression)}, nil
	case types.KindExpressionStmt:
		expr, err := r.readNode()
		if err != nil {
			return nil, err
		}
		return &ast.ExpressionStatementNode{BaseNode: ast.BaseNode{NodeKind: kind}, Expression: expr.(ast.Expression)}, nil
	case types.KindMapEntry:
		return r.readMapEntry()
	case types.KindMetadataLine: // FIX: Add missing case
		return r.readMetadataLine()
	default:
		return nil, fmt.Errorf("unhandled node kind for decoding: %v (%d) at byte offset %d. History: [%s]", kind, kind, offset, strings.Join(r.history, ", "))
	}
}

// --- Specific reader methods ---

func (r *canonReader) readMapLiteral() (*ast.MapLiteralNode, error) {
	numEntries, err := r.readVarint()
	if err != nil {
		return nil, fmt.Errorf("failed to read map entry count: %w", err)
	}
	m := &ast.MapLiteralNode{
		BaseNode: ast.BaseNode{NodeKind: types.KindMapLiteral},
		Entries:  make([]*ast.MapEntryNode, numEntries),
	}
	for i := 0; i < int(numEntries); i++ {
		entryNode, err := r.readNode()
		if err != nil {
			return nil, fmt.Errorf("failed to read map entry %d: %w", i, err)
		}
		entry, ok := entryNode.(*ast.MapEntryNode)
		if !ok {
			return nil, fmt.Errorf("expected to decode a *ast.MapEntryNode but got %T", entryNode)
		}
		m.Entries[i] = entry
	}
	return m, nil
}

func (r *canonReader) readMetadataLine() (*ast.MetadataLine, error) {
	key, err := r.readString()
	if err != nil {
		return nil, err
	}
	val, err := r.readString()
	if err != nil {
		return nil, err
	}
	return &ast.MetadataLine{
		BaseNode: ast.BaseNode{NodeKind: types.KindMetadataLine},
		Key:      key,
		Value:    val,
	}, nil
}
