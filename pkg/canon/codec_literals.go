// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 3
// :: description: Implements encoders/decoders for literal and simple AST nodes. Added InterpolatedString.
// :: latestChange: Added encodeInterpolatedString and decodeInterpolatedString.
// :: filename: pkg/canon/codec_literals.go
// :: serialization: go

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

func encodeInterpolatedString(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.InterpolatedStringNode)
	v.writeString(node.Delimiter)
	v.writeVarint(int64(len(node.Parts)))
	for _, part := range node.Parts {
		if err := v.visitor(part); err != nil {
			return err
		}
	}
	return nil
}

func decodeInterpolatedString(r *canonReader) (ast.Node, error) {
	node := &ast.InterpolatedStringNode{BaseNode: ast.BaseNode{NodeKind: types.KindInterpolatedString}}
	var err error
	node.Delimiter, err = r.readString()
	if err != nil {
		return nil, err
	}
	count, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	node.Parts = make([]ast.Expression, count)
	for i := 0; i < int(count); i++ {
		part, err := r.visitor()
		if err != nil {
			return nil, err
		}
		node.Parts[i] = part.(ast.Expression)
	}
	return node, nil
}
