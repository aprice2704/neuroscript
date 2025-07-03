// filename: pkg/ast/ast_commands_test.go
package ast_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestCommandNode_String(t *testing.T) {
	node := &ast.CommandNode{}
	expected := "command ... endcommand"
	if node.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, node.String())
	}
}
