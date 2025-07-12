// filename: pkg/canon/canonicalize.go
// NeuroScript Version: 0.5.2
// File version: 6
// Purpose: Implemented constant folding for unary minus to ensure semantic canonicalization.
// nlines: 180+
// risk_rating: MEDIUM

package canon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"sort"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"golang.org/x/crypto/blake2b"
)

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
		v.writeVarint(int64(ast.KindNilLiteral))
		return nil
	}

	// Do not write the kind for UnaryOp '-' on a number, as it will be folded.
	if un, ok := node.(*ast.UnaryOpNode); ok && un.Operator == "-" {
		if _, isNum := un.Operand.(*ast.NumberLiteralNode); isNum {
			return v.visitUnaryOp(un)
		}
	}

	v.writeVarint(int64(node.Kind()))

	switch n := node.(type) {
	case *ast.Program:
		return v.visitProgram(n)
	case *ast.Procedure:
		return v.visitProcedure(n)
	case *ast.Step:
		return v.visitStep(n)
	case *ast.LValueNode:
		return v.visitLValue(n)
	case *ast.StringLiteralNode:
		v.writeString(n.Value)
		return nil
	case *ast.NumberLiteralNode:
		v.writeNumber(n.Value)
		return nil
	case *ast.BooleanLiteralNode:
		v.writeBool(n.Value)
		return nil
	case *ast.OnEventDecl:
		return v.visitOnEventDecl(n)
	case *ast.CallableExprNode:
		return v.visitCallableExpr(n)
	case *ast.VariableNode:
		v.writeString(n.Name)
		return nil
	case *ast.BinaryOpNode:
		return v.visitBinaryOp(n)
	case *ast.UnaryOpNode:
		return v.visitUnaryOp(n)
	case *ast.NilLiteralNode:
		return nil
	default:
		return fmt.Errorf("unhandled node type in canonicalizer: %T", n)
	}
}

// --- Specific visitor methods ---
func (v *canonVisitor) visitUnaryOp(u *ast.UnaryOpNode) error {
	// Check for constant folding opportunity: '-' on a number literal
	if u.Operator == "-" {
		if num, ok := u.Operand.(*ast.NumberLiteralNode); ok {
			// Fold the constant. Instead of writing UnaryOp, write a new NumberLiteral.
			v.writeVarint(int64(ast.KindNumberLiteral))
			switch val := num.Value.(type) {
			case int64:
				v.writeNumber(-val)
			case float64:
				v.writeNumber(-val)
			}
			return nil
		}
	}

	// Default behavior for other unary ops
	v.writeString(u.Operator)
	return v.visit(u.Operand)
}

func (v *canonVisitor) visitProgram(p *ast.Program) error {
	procNames := make([]string, 0, len(p.Procedures))
	for name := range p.Procedures {
		procNames = append(procNames, name)
	}
	sort.Strings(procNames)

	v.writeVarint(int64(len(procNames)))
	for _, name := range procNames {
		if err := v.visit(p.Procedures[name]); err != nil {
			return err
		}
	}

	v.writeVarint(int64(len(p.Events)))
	for _, event := range p.Events {
		if err := v.visit(event); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitProcedure(p *ast.Procedure) error {
	v.writeString(p.Name())
	v.writeVarint(int64(len(p.Steps)))
	for _, step := range p.Steps {
		if err := v.visit(&step); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitStep(s *ast.Step) error {
	v.writeString(s.Type)
	switch s.Type {
	case "set":
		v.writeVarint(int64(len(s.LValues)))
		for _, lval := range s.LValues {
			if err := v.visit(lval); err != nil {
				return err
			}
		}
		v.writeVarint(int64(len(s.Values)))
		for _, rval := range s.Values {
			if err := v.visit(rval); err != nil {
				return err
			}
		}
	case "emit":
		return v.visit(s.Values[0])
	case "call":
		return v.visit(s.Call)
	}
	return nil
}

func (v *canonVisitor) visitOnEventDecl(e *ast.OnEventDecl) error {
	if err := v.visit(e.EventNameExpr); err != nil {
		return err
	}
	v.writeVarint(int64(len(e.Body)))
	for _, step := range e.Body {
		if err := v.visit(&step); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitCallableExpr(c *ast.CallableExprNode) error {
	v.writeString(c.Target.Name)
	v.writeBool(c.Target.IsTool)
	v.writeVarint(int64(len(c.Arguments)))
	for _, arg := range c.Arguments {
		if err := v.visit(arg); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitLValue(lval *ast.LValueNode) error {
	v.writeString(lval.Identifier)
	return nil
}

func (v *canonVisitor) visitBinaryOp(b *ast.BinaryOpNode) error {
	v.writeString(b.Operator)
	if err := v.visit(b.Left); err != nil {
		return err
	}
	return v.visit(b.Right)
}

// --- Primitive Writers ---

func (v *canonVisitor) write(p []byte) {
	v.w.Write(p)
	v.hasher.Write(p)
}

func (v *canonVisitor) writeVarint(x int64) {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, x)
	v.write(buf[:n])
}

func (v *canonVisitor) writeString(s string) {
	v.writeVarint(int64(len(s)))
	if len(s) > 0 {
		v.write([]byte(s))
	}
}

func (v *canonVisitor) writeBool(b bool) {
	if b {
		v.write([]byte{1})
	} else {
		v.write([]byte{0})
	}
}

func (v *canonVisitor) writeNumber(val interface{}) {
	// FIX: Use math.Signbit to correctly handle signed zero for deterministic output.
	if f, ok := val.(float64); ok && f == 0 && math.Signbit(f) {
		val = 0.0 // Normalize -0.0 to 0.0
	}
	strVal := fmt.Sprintf("%v", val)
	v.writeString(strVal)
}
