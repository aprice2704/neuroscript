// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Corrected package declaration and struct literals to resolve build errors.
// filename: pkg/ast/ast_expressions_test.go
// nlines: 200
// risk_rating: LOW

package ast

import (
	"strconv"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestExpressionStringers(t *testing.T) {
	pos := &types.Position{Line: 1, Column: 1, File: "test.ns"}

	t.Run("CallTarget", func(t *testing.T) {
		ctFunc := &CallTarget{Name: "myFunc"}
		if ctFunc.String() != "myFunc" {
			t.Errorf("Expected 'myFunc', got '%s'", ctFunc.String())
		}

		ctTool := &CallTarget{Name: "fs.read", IsTool: true}
		if ctTool.String() != "tool.fs.read" {
			t.Errorf("Expected 'tool.fs.read', got '%s'", ctTool.String())
		}
	})

	t.Run("CallableExprNode", func(t *testing.T) {
		node := &CallableExprNode{
			Target: CallTarget{Name: "myFunc"},
			Arguments: []Expression{
				&NumberLiteralNode{Value: 123},
				&StringLiteralNode{Value: "hello"},
			},
		}
		expected := "myFunc(123, \"hello\")"
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})

	t.Run("VariableNode", func(t *testing.T) {
		node := &VariableNode{Name: "myVar"}
		if node.String() != "myVar" {
			t.Errorf("Expected 'myVar', got '%s'", node.String())
		}
	})

	t.Run("PlaceholderNode", func(t *testing.T) {
		node := &PlaceholderNode{Name: "myPlaceholder"}
		expected := "{{myPlaceholder}}"
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})

	t.Run("LastNode", func(t *testing.T) {
		node := &LastNode{}
		if node.String() != "last" {
			t.Errorf("Expected 'last', got '%s'", node.String())
		}
	})

	t.Run("EvalNode", func(t *testing.T) {
		node := &EvalNode{Argument: &StringLiteralNode{Value: "x + 1"}}
		expected := "eval(\"x + 1\")"
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})

	t.Run("StringLiteralNode", func(t *testing.T) {
		t.Run("regular string", func(t *testing.T) {
			node := &StringLiteralNode{Value: "hello world"}
			expected := strconv.Quote("hello world")
			if node.String() != expected {
				t.Errorf("Expected '%s', got '%s'", expected, node.String())
			}
		})
		t.Run("raw string", func(t *testing.T) {
			node := &StringLiteralNode{Value: "raw content", IsRaw: true}
			expected := "```raw content```"
			if node.String() != expected {
				t.Errorf("Expected '%s', got '%s'", expected, node.String())
			}
		})
	})

	t.Run("NumberLiteralNode", func(t *testing.T) {
		intNode := &NumberLiteralNode{Value: int64(42)}
		if intNode.String() != "42" {
			t.Errorf("Expected '42', got '%s'", intNode.String())
		}

		floatNode := &NumberLiteralNode{Value: 3.14}
		if floatNode.String() != "3.14" {
			t.Errorf("Expected '3.14', got '%s'", floatNode.String())
		}
	})

	t.Run("BooleanLiteralNode", func(t *testing.T) {
		trueNode := &BooleanLiteralNode{Value: true}
		if trueNode.String() != "true" {
			t.Errorf("Expected 'true', got '%s'", trueNode.String())
		}
		falseNode := &BooleanLiteralNode{Value: false}
		if falseNode.String() != "false" {
			t.Errorf("Expected 'false', got '%s'", falseNode.String())
		}
	})

	t.Run("ListLiteralNode", func(t *testing.T) {
		node := &ListLiteralNode{
			Elements: []Expression{
				&NumberLiteralNode{Value: 1},
				nil, // Test nil element
			},
		}
		expected := "[1, <nil_expr>]"
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})

	t.Run("MapLiteralNode", func(t *testing.T) {
		node := &MapLiteralNode{
			BaseNode: BaseNode{StartPos: pos},
			Entries: []*MapEntryNode{
				{
					Key:   &StringLiteralNode{Value: "a"},
					Value: &NumberLiteralNode{Value: 1},
				},
				nil, // Test nil entry
			},
		}
		expected := "{\"a\": 1, <nil_entry>}"
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})

	t.Run("ElementAccessNode", func(t *testing.T) {
		node := &ElementAccessNode{
			Collection: &VariableNode{Name: "myList"},
			Accessor:   &NumberLiteralNode{Value: 0},
		}
		expected := "myList[0]"
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})

	t.Run("UnaryOpNode", func(t *testing.T) {
		node := &UnaryOpNode{
			Operator: "-",
			Operand:  &VariableNode{Name: "x"},
		}
		expected := "-x"
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})

	t.Run("BinaryOpNode", func(t *testing.T) {
		node := &BinaryOpNode{
			Left:     &VariableNode{Name: "a"},
			Operator: "+",
			Right:    &VariableNode{Name: "b"},
		}
		expected := "(a + b)"
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})

	t.Run("TypeOfNode", func(t *testing.T) {
		node := &TypeOfNode{Argument: &VariableNode{Name: "myVar"}}
		expected := "typeof(myVar)"
		if node.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, node.String())
		}
	})

	t.Run("NilLiteralNode", func(t *testing.T) {
		node := &NilLiteralNode{}
		if node.String() != "nil" {
			t.Errorf("Expected 'nil', got '%s'", node.String())
		}
	})
}
