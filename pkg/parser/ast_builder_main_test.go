// filename: pkg/parser/ast_builder_main_test.go
package parser

import (
	"reflect"
	"sort"
	"testing"
)

func TestASTBuilder_Build_NilTree(t *testing.T) {
	logger := logging.NewNoLogger()
	astBuilder := NewASTBuilder(logger)

	_, _, err := astBuilder.Build(nil)
	if err == nil {
		t.Fatal("Expected an error when building from a nil tree, but got nil")
	}
	if err.Error() != "cannot build AST from nil parse tree" {
		t.Errorf("Expected error message 'cannot build AST from nil parse tree', but got '%s'", err.Error())
	}
}

func TestMapKeys(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		if MapKeys(nil) != nil {
			t.Error("Expected nil for a nil map")
		}
	})

	t.Run("empty map", func(t *testing.T) {
		if len(MapKeys(make(map[string]string))) != 0 {
			t.Error("Expected an empty slice for an empty map")
		}
	})

	t.Run("map with keys", func(t *testing.T) {
		m := map[string]string{
			"b": "2",
			"a": "1",
			"c": "3",
		}
		expected := []string{"a", "b", "c"}
		actual := MapKeys(m)

		sort.Strings(actual)

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected keys %v, but got %v", expected, actual)
		}
	})
}
