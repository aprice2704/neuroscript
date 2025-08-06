// NeuroScript Version: 0.6.2
// File version: 24.0
// Purpose: Tidy: Removes verbose debug logging now that the canonicalization issues are resolved.
// filename: pkg/canon/canonicalize_part1.go
// nlines: 160
// risk_rating: HIGH

package canon

import (
	"bytes"
	"fmt"
	"hash"
	"sort"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
	"golang.org/x/crypto/blake2b"
)

// magicNumber is a dynamic fingerprint ("NSC" + version) to identify a valid canonical blob.
var magicNumber = []byte{'N', 'S', 'C', byte(types.KindMarker)}

// Canonicalise traverses an AST and produces a deterministic, platform-independent
// binary representation. It also returns a BLAKE2b-256 hash of the resulting bytes.
func Canonicalise(tree *ast.Tree) ([]byte, [32]byte, error) {
	if tree == nil || tree.Root == nil {
		return nil, [32]byte{}, fmt.Errorf("cannot canonicalize a nil tree or a tree with a nil root")
	}

	var buf bytes.Buffer
	hasher, _ := blake2b.New256(nil)

	visitor := &canonVisitor{
		w:      &buf,
		hasher: hasher,
	}

	// Write the magic number header first.
	visitor.write(magicNumber)

	err := visitor.visit(tree.Root)
	if err != nil {
		return nil, [32]byte{}, err
	}

	var sum [32]byte
	hasher.Sum(sum[:0])

	return buf.Bytes(), sum, nil
}

// canonVisitor walks the AST and writes its canonical representation.
type canonVisitor struct {
	w      *bytes.Buffer
	hasher hash.Hash
}

// visit is the dispatcher for visiting any node type.
func (v *canonVisitor) visit(node ast.Node) error {
	if node == nil {
		v.writeVarint(int64(types.KindNilLiteral))
		return nil
	}

	var kindToWrite types.Kind
	if _, ok := node.(*ast.MapEntryNode); ok {
		kindToWrite = types.KindMapEntry
	} else {
		kindToWrite = node.Kind()
	}
	v.writeVarint(int64(kindToWrite))

	switch n := node.(type) {
	// Structural Nodes
	case *ast.Program:
		return v.visitProgram(n)
	case *ast.Procedure:
		return v.visitProcedure(n)
	case *ast.Step:
		return v.visitStep(n)
	case *ast.CommandNode:
		return v.visitCommand(n)
	case *ast.OnEventDecl:
		return v.visitOnEventDecl(n)

	// Expression Nodes
	case *ast.LValueNode:
		return v.visitLValue(n)
	case *ast.StringLiteralNode:
		v.writeString(n.Value)
		v.writeBool(n.IsRaw)
		return nil
	case *ast.NumberLiteralNode:
		v.writeNumber(n.Value)
		return nil
	case *ast.BooleanLiteralNode:
		v.writeBool(n.Value)
		return nil
	case *ast.NilLiteralNode:
		return nil
	case *ast.CallableExprNode:
		return v.visitCallableExpr(n)
	case *ast.VariableNode:
		v.writeString(n.Name)
		return nil
	case *ast.BinaryOpNode:
		return v.visitBinaryOp(n)
	case *ast.UnaryOpNode:
		return v.visitUnaryOp(n)
	case *ast.MapLiteralNode:
		return v.visitMapLiteral(n)
	case *ast.ListLiteralNode:
		return v.visitListLiteral(n)
	case *ast.ElementAccessNode:
		return v.visitElementAccess(n)
	case *ast.SecretRef:
		return v.visitSecretRef(n)
	case *ast.PlaceholderNode:
		v.writeString(n.Name)
		return nil
	case *ast.LastNode:
		return nil
	case *ast.EvalNode:
		return v.visit(n.Argument)
	case *ast.TypeOfNode:
		return v.visit(n.Argument)
	case *ast.ExpressionStatementNode:
		return v.visit(n.Expression)
	case *ast.MapEntryNode:
		return v.visitMapEntry(n)

	default:
		return fmt.Errorf("unhandled node type in canonicalizer: %T", n)
	}
}

// --- Specific visitor methods ---

func (v *canonVisitor) visitMapLiteral(m *ast.MapLiteralNode) error {
	v.writeVarint(int64(len(m.Entries)))

	sortedEntries := make([]*ast.MapEntryNode, len(m.Entries))
	copy(sortedEntries, m.Entries)
	sort.Slice(sortedEntries, func(i, j int) bool {
		return sortedEntries[i].Key.Value < sortedEntries[j].Key.Value
	})

	for _, entry := range sortedEntries {
		if err := v.visit(entry); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitMapEntry(e *ast.MapEntryNode) error {
	if err := v.visit(e.Key); err != nil {
		return err
	}
	return v.visit(e.Value)
}
