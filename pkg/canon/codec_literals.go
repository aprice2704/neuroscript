// NeuroScript Version: 0.6.3
// File version: 2
// Purpose: Implements encoders/decoders for literal and simple AST nodes. Removed debugging print statements.
// filename: pkg/canon/codec_literals.go
// nlines: 120
// risk_rating: LOW

package canon

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func encodeStringLiteral(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.StringLiteralNode)
	v.writeString(node.Value)
	v.writeBool(node.IsRaw)
	return nil
}

func decodeStringLiteral(r *canonReader) (ast.Node, error) {
	node := &ast.StringLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindStringLiteral}}
	var err error
	node.Value, err = r.readString()
	if err != nil {
		return nil, err
	}
	node.IsRaw, err = r.readBool()
	if err != nil {
		return nil, err
	}
	return node, nil
}

func encodeNumberLiteral(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.NumberLiteralNode)
	v.writeNumber(node.Value)
	return nil
}

func decodeNumberLiteral(r *canonReader) (ast.Node, error) {
	node := &ast.NumberLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindNumberLiteral}}
	var err error
	node.Value, err = r.readNumber()
	if err != nil {
		return nil, err
	}
	return node, nil
}

func encodeBooleanLiteral(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.BooleanLiteralNode)
	v.writeBool(node.Value)
	return nil
}

func decodeBooleanLiteral(r *canonReader) (ast.Node, error) {
	node := &ast.BooleanLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindBooleanLiteral}}
	var err error
	node.Value, err = r.readBool()
	if err != nil {
		return nil, err
	}
	return node, nil
}

func encodeNilLiteral(v *canonVisitor, n ast.Node) error {
	return nil // No payload
}

func decodeNilLiteral(r *canonReader) (ast.Node, error) {
	return &ast.NilLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindNilLiteral}}, nil
}

func encodeVariable(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.VariableNode)
	v.writeString(node.Name)
	return nil
}

func decodeVariable(r *canonReader) (ast.Node, error) {
	node := &ast.VariableNode{BaseNode: ast.BaseNode{NodeKind: types.KindVariable}}
	var err error
	node.Name, err = r.readString()
	if err != nil {
		return nil, err
	}
	return node, nil
}
