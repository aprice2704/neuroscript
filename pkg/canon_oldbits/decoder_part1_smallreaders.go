package canon

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (r *canonReader) readMapLiteral() (*ast.MapLiteralNode, error) {
	n, err := r.readVarint()
	if err != nil {
		return nil, fmt.Errorf("map entry count: %w", err)
	}
	m := &ast.MapLiteralNode{
		BaseNode: ast.BaseNode{NodeKind: types.KindMapLiteral},
		Entries:  make([]*ast.MapEntryNode, n),
	}
	for i := 0; i < int(n); i++ {
		node, err := r.readNode()
		if err != nil {
			return nil, fmt.Errorf("map entry[%d]: %w", i, err)
		}
		me, ok := node.(*ast.MapEntryNode)
		if !ok {
			return nil, fmt.Errorf("expected *ast.MapEntryNode, got %T", node)
		}
		m.Entries[i] = me
	}
	return m, nil
}

func (r *canonReader) readMetadataLine() (*ast.MetadataLine, error) {
	k, err := r.readString()
	if err != nil {
		return nil, err
	}
	v, err := r.readString()
	if err != nil {
		return nil, err
	}
	return &ast.MetadataLine{
		BaseNode: ast.BaseNode{NodeKind: types.KindMetadataLine},
		Key:      k,
		Value:    v,
	}, nil
}
