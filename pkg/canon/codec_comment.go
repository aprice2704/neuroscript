// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Implements the codec for ast.Comment nodes.
// filename: pkg/canon/codec_comment.go
// nlines: 27

package canon

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func encodeComment(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.Comment)
	v.writeString(node.Text)
	return nil
}

func decodeComment(r *canonReader) (ast.Node, error) {
	node := &ast.Comment{
		BaseNode: ast.BaseNode{NodeKind: types.KindComment},
	}
	var err error
	node.Text, err = r.readString()
	if err != nil {
		return nil, err
	}
	return node, nil
}
