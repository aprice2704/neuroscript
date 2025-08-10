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
func Decode(blob []byte) (*ast.Tree, error) {
	magic := []byte{'N', 'S', 'C', byte(types.KindMarker)}
	if len(blob) < len(magic) {
		return nil, fmt.Errorf("cannot decode blob: too short to be valid")
	}
	if !bytes.Equal(blob[:len(magic)], magic) {
		return nil, fmt.Errorf("invalid magic number: blob is not a valid canonical AST or incompatible version")
	}

	r := &canonReader{r: bytes.NewReader(blob[len(magic):])}
	root, err := r.readNode()
	if err != nil {
		return nil, fmt.Errorf("failed to decode root node: %w", err)
	}
	prog, ok := root.(*ast.Program)
	if !ok {
		return nil, fmt.Errorf("decoded root node is not *ast.Program but %T", root)
	}

	restoreCallTargetKinds(prog)
	return &ast.Tree{Root: prog}, nil
}

type canonReader struct {
	r       *bytes.Reader
	history []string
}

func (r *canonReader) readNode() (ast.Node, error) {
	offset := r.r.Size() - int64(r.r.Len())

	kindVal, err := r.readVarint() // implemented in other decoder parts
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read node kind: %w", err)
	}
	kind := types.Kind(kindVal)
	r.history = append(r.history, fmt.Sprintf("%v", kind))

	switch kind {
	// Structural
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

	// Expressions / literals
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
		num, err := r.readNumber()
		if err != nil {
			return nil, err
		}
		return &ast.NumberLiteralNode{BaseNode: ast.BaseNode{NodeKind: kind}, Value: num}, nil
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

	// Statement-as-expression
	case types.KindExpressionStmt:
		ex, err := r.readNode()
		if err != nil {
			return nil, err
		}
		e, _ := ex.(ast.Expression)
		return &ast.ExpressionStatementNode{BaseNode: ast.BaseNode{NodeKind: kind}, Expression: e}, nil

	// Metadata / misc
	case types.KindMapEntry:
		return r.readMapEntry()
	case types.KindMetadataLine:
		return r.readMetadataLine()

	// ask/promptuser
	case types.KindAskStmt:
		return r.readAskStmt()
	case types.KindPromptUserStmt:
		return r.readPromptUserStmt()

	// Safe aliases
	case types.KindLast, types.KindLastResult:
		return &ast.LastNode{BaseNode: ast.BaseNode{NodeKind: kind}}, nil
	case types.KindEval, types.KindEvalExpr:
		arg, err := r.readNode()
		if err != nil {
			return nil, err
		}
		ex, _ := arg.(ast.Expression)
		return &ast.EvalNode{BaseNode: ast.BaseNode{NodeKind: kind}, Argument: ex}, nil
	case types.KindTypeOf, types.KindTypeOfExpr:
		arg, err := r.readNode()
		if err != nil {
			return nil, err
		}
		ex, _ := arg.(ast.Expression)
		return &ast.TypeOfNode{BaseNode: ast.BaseNode{NodeKind: kind}, Argument: ex}, nil

	default:
		return nil, fmt.Errorf("unhandled node kind for decoding: %v (%d) at byte offset %d. History: [%s]", kind, kind, offset, strings.Join(r.history, ", "))
	}
}
