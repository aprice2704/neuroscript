// filename: pkg/ast/ast_test.go
package ast

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestErrorNode_String(t *testing.T) {
	t.Run("regular error node", func(t *testing.T) {
		pos := &lang.Position{Line: 10, Column: 5, File: "test.ns"}
		node := &ErrorNode{
			Pos:		pos,
			Message:	"something went wrong",
		}
		expected := fmt.Sprintf("Error at %s: %s", pos, node.Message)
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})

	t.Run("nil error node", func(t *testing.T) {
		var node *ErrorNode
		expected := "<nil error node>"
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})
}