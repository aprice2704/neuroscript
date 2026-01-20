// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 2
// :: description: Fixed MapEntryNode key handling to support Expression type (sorting by string rep, casting correctly).
// :: latestChange: Updated encode/decode to handle Expression keys.
// :: filename: pkg/canon/codec_collections.go
// :: serialization: go

package canon

import (
	"sort"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func encodeListLiteral(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.ListLiteralNode)
	v.writeVarint(int64(len(node.Elements)))
	for _, elem := range node.Elements {
		if err := v.visitor(elem); err != nil {
			return err
		}
	}
	return nil
}

func decodeListLiteral(r *canonReader) (ast.Node, error) {
	node := &ast.ListLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindListLiteral}}
	count, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	node.Elements = make([]ast.Expression, count)
	for i := 0; i < int(count); i++ {
		elem, err := r.visitor()
		if err != nil {
			return nil, err
		}
		node.Elements[i] = elem.(ast.Expression)
	}
	return node, nil
}

func encodeMapLiteral(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.MapLiteralNode)
	v.writeVarint(int64(len(node.Entries)))
	// Sort entries by key string representation for deterministic output.
	// Since Key is now an Expression, we use TestString() for stable sorting.
	sort.Slice(node.Entries, func(i, j int) bool {
		return node.Entries[i].Key.TestString() < node.Entries[j].Key.TestString()
	})
	for _, entry := range node.Entries {
		if err := v.visitor(entry); err != nil {
			return err
		}
	}
	return nil
}

func decodeMapLiteral(r *canonReader) (ast.Node, error) {
	node := &ast.MapLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindMapLiteral}}
	count, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	node.Entries = make([]*ast.MapEntryNode, count)
	for i := 0; i < int(count); i++ {
		entry, err := r.visitor()
		if err != nil {
			return nil, err
		}
		node.Entries[i] = entry.(*ast.MapEntryNode)
	}
	return node, nil
}

func encodeMapEntry(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.MapEntryNode)
	if err := v.visitor(node.Key); err != nil {
		return err
	}
	return v.visitor(node.Value)
}

func decodeMapEntry(r *canonReader) (ast.Node, error) {
	node := &ast.MapEntryNode{BaseNode: ast.BaseNode{NodeKind: types.KindMapEntry}}
	key, err := r.visitor()
	if err != nil {
		return nil, err
	}
	// FIX: Key is now an ast.Expression, not *ast.StringLiteralNode
	node.Key = key.(ast.Expression)
	value, err := r.visitor()
	if err != nil {
		return nil, err
	}
	node.Value = value.(ast.Expression)
	return node, nil
}

func encodeElementAccess(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.ElementAccessNode)
	if err := v.visitor(node.Collection); err != nil {
		return err
	}
	return v.visitor(node.Accessor)
}

func decodeElementAccess(r *canonReader) (ast.Node, error) {
	node := &ast.ElementAccessNode{BaseNode: ast.BaseNode{NodeKind: types.KindElementAccess}}
	collection, err := r.visitor()
	if err != nil {
		return nil, err
	}
	node.Collection = collection.(ast.Expression)
	accessor, err := r.visitor()
	if err != nil {
		return nil, err
	}
	node.Accessor = accessor.(ast.Expression)
	return node, nil
}
