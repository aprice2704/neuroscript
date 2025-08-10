// NeuroScript Version: 0.6.2
// File version: 1
// Purpose: Decoder helpers for list-like and map entry node types.
// Filename: pkg/canon/decoder_part2_lists.go
// Risk rating: LOW

package canon

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (r *canonReader) readListLiteral() (*ast.ListLiteralNode, error) {
	n, err := r.readVarint()
	if err != nil {
		return nil, fmt.Errorf("list: count: %w", err)
	}
	l := &ast.ListLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindListLiteral}}
	if n <= 0 {
		l.Elements = []ast.Expression{}
		return l, nil
	}
	l.Elements = make([]ast.Expression, n)
	for i := 0; i < int(n); i++ {
		node, err := r.readNode()
		if err != nil {
			return nil, fmt.Errorf("list: elem[%d]: %w", i, err)
		}
		e, ok := node.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("list: elem[%d]: expected ast.Expression, got %T", i, node)
		}
		l.Elements[i] = e
	}
	return l, nil
}

func (r *canonReader) readElementAccess() (*ast.ElementAccessNode, error) {
	coll, err := r.readNode()
	if err != nil {
		return nil, err
	}
	key, err := r.readNode()
	if err != nil {
		return nil, err
	}
	c, ok := coll.(ast.Expression)
	if !ok {
		return nil, fmt.Errorf("element access: collection must be Expression, got %T", coll)
	}
	k, ok := key.(ast.Expression)
	if !ok {
		return nil, fmt.Errorf("element access: key must be Expression, got %T", key)
	}
	return &ast.ElementAccessNode{
		BaseNode:   ast.BaseNode{NodeKind: types.KindElementAccess},
		Collection: c,
		Accessor:   k,
	}, nil
}

func (r *canonReader) readMapEntry() (*ast.MapEntryNode, error) {
	keyNode, err := r.readNode()
	if err != nil {
		return nil, fmt.Errorf("mapentry: key: %w", err)
	}
	key, ok := keyNode.(*ast.StringLiteralNode)
	if !ok {
		return nil, fmt.Errorf("mapentry: key must be *ast.StringLiteralNode, got %T", keyNode)
	}
	valNode, err := r.readNode()
	if err != nil {
		return nil, fmt.Errorf("mapentry: value: %w", err)
	}
	val, ok := valNode.(ast.Expression)
	if !ok {
		return nil, fmt.Errorf("mapentry: value must be Expression, got %T", valNode)
	}
	return &ast.MapEntryNode{
		BaseNode: ast.BaseNode{NodeKind: types.KindMapEntry},
		Key:      key,
		Value:    val,
	}, nil
}
