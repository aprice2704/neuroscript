package ast

import "testing"

func TestCommandNode_String(t *testing.T) {
	node := &CommandNode{}
	expected := "command ... endcommand"
	if node.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, node.String())
	}
}
