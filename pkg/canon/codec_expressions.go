// NeuroScript Version: 0.6.3
// File version: 2
// Purpose: Implements encoders/decoders for complex expression AST nodes. Added BinaryOp and UnaryOp.
// filename: pkg/canon/codec_expressions.go
// nlines: 120
// risk_rating: HIGH

package canon

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func encodeCallableExpr(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.CallableExprNode)
	tempVisitor := &canonVisitor{w: v.w, hasher: v.hasher}

	tempVisitor.write([]byte{CallMagic1, CallMagic2, CallWireVersion, CallLayoutHeader})
	tempVisitor.writeBool(node.Target.IsTool)
	tempVisitor.writeString(node.Target.Name)
	tempVisitor.writeVarint(int64(len(node.Arguments)))
	for _, arg := range node.Arguments {
		if err := v.visitor(arg); err != nil {
			return err
		}
	}
	return nil
}

func decodeCallableExpr(r *canonReader) (ast.Node, error) {
	node := &ast.CallableExprNode{
		BaseNode: ast.BaseNode{NodeKind: types.KindCallableExpr},
		Target:   ast.CallTarget{BaseNode: ast.BaseNode{NodeKind: types.KindVariable}},
	}

	m1, err := r.readByte()
	if err != nil {
		return nil, fmt.Errorf("callable: read magic[0]: %w", err)
	}
	m2, err := r.readByte()
	if err != nil {
		return nil, fmt.Errorf("callable: read magic[1]: %w", err)
	}
	ver, err := r.readByte()
	if err != nil {
		return nil, fmt.Errorf("callable: read version: %w", err)
	}
	layout, err := r.readByte()
	if err != nil {
		return nil, fmt.Errorf("callable: read layout: %w", err)
	}

	if m1 != CallMagic1 || m2 != CallMagic2 || ver != CallWireVersion {
		return nil, fmt.Errorf("callable: bad header: got [%02X %02X] ver=%02X", m1, m2, ver)
	}

	if layout != CallLayoutHeader {
		return nil, fmt.Errorf("callable: unsupported layout %d", layout)
	}

	node.Target.IsTool, err = r.readBool()
	if err != nil {
		return nil, err
	}
	node.Target.Name, err = r.readString()
	if err != nil {
		return nil, err
	}
	argc, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	node.Arguments = make([]ast.Expression, argc)
	for i := 0; i < int(argc); i++ {
		argNode, err := r.visitor()
		if err != nil {
			return nil, err
		}
		node.Arguments[i] = argNode.(ast.Expression)
	}

	return node, nil
}

func encodeBinaryOp(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.BinaryOpNode)
	v.writeString(node.Operator)
	if err := v.visitor(node.Left); err != nil {
		return err
	}
	return v.visitor(node.Right)
}

func decodeBinaryOp(r *canonReader) (ast.Node, error) {
	node := &ast.BinaryOpNode{BaseNode: ast.BaseNode{NodeKind: types.KindBinaryOp}}
	var err error
	node.Operator, err = r.readString()
	if err != nil {
		return nil, err
	}
	left, err := r.visitor()
	if err != nil {
		return nil, err
	}
	node.Left = left.(ast.Expression)
	right, err := r.visitor()
	if err != nil {
		return nil, err
	}
	node.Right = right.(ast.Expression)
	return node, nil
}

func encodeUnaryOp(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.UnaryOpNode)
	v.writeString(node.Operator)
	return v.visitor(node.Operand)
}

func decodeUnaryOp(r *canonReader) (ast.Node, error) {
	node := &ast.UnaryOpNode{BaseNode: ast.BaseNode{NodeKind: types.KindUnaryOp}}
	var err error
	node.Operator, err = r.readString()
	if err != nil {
		return nil, err
	}
	operand, err := r.visitor()
	if err != nil {
		return nil, err
	}
	node.Operand = operand.(ast.Expression)
	return node, nil
}
