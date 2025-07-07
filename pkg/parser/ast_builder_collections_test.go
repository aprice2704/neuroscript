// filename: pkg/parser/ast_builder_collections_test.go
// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Adds test coverage for list and map literal parsing.
// nlines: 85
// risk_rating: LOW

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestListLiteralParsing(t *testing.T) {
	t.Run("Valid list with mixed types", func(t *testing.T) {
		script := `[1, "hello", true, [], {"a":1}]`
		expr := parseExpression(t, script)
		listNode, ok := expr.(*ast.ListLiteralNode)
		if !ok {
			t.Fatalf("Expected a ListLiteralNode, got %T", expr)
		}
		if len(listNode.Elements) != 5 {
			t.Errorf("Expected 5 elements in the list, got %d", len(listNode.Elements))
		}
	})

	t.Run("Empty list", func(t *testing.T) {
		script := `[]`
		expr := parseExpression(t, script)
		listNode, ok := expr.(*ast.ListLiteralNode)
		if !ok {
			t.Fatalf("Expected a ListLiteralNode, got %T", expr)
		}
		if len(listNode.Elements) != 0 {
			t.Errorf("Expected 0 elements in the empty list, got %d", len(listNode.Elements))
		}
	})
}

func TestMapLiteralParsing(t *testing.T) {
	t.Run("Valid map with mixed value types", func(t *testing.T) {
		script := `{"a": 1, "b": "world", "c": [1,2]}`
		expr := parseExpression(t, script)
		mapNode, ok := expr.(*ast.MapLiteralNode)
		if !ok {
			t.Fatalf("Expected a MapLiteralNode, got %T", expr)
		}
		if len(mapNode.Entries) != 3 {
			t.Errorf("Expected 3 entries in the map, got %d", len(mapNode.Entries))
		}
	})

	t.Run("Empty map", func(t *testing.T) {
		script := `{}`
		expr := parseExpression(t, script)
		mapNode, ok := expr.(*ast.MapLiteralNode)
		if !ok {
			t.Fatalf("Expected a MapLiteralNode, got %T", expr)
		}
		if len(mapNode.Entries) != 0 {
			t.Errorf("Expected 0 entries in the empty map, got %d", len(mapNode.Entries))
		}
	})

	t.Run("Map with non-string key is parser error", func(t *testing.T) {
		script := `func t() means\n set x = {a: 1}\nendfunc`
		testForParserError(t, script)
	})
}
