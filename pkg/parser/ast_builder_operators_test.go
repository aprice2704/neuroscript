// filename: pkg/parser/ast_builder_operators_test.go
package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestOperatorPrecedence(t *testing.T) {
	// Test that multiplication binds tighter than addition
	script := "2 + 3 * 4"
	expr := parseExpression(t, script)

	binaryOp, ok := expr.(*ast.BinaryOpNode)
	if !ok {
		t.Fatalf("Expected top-level expression to be a BinaryOpNode, got %T", expr)
	}

	if binaryOp.Operator != "+" {
		t.Errorf("Expected top-level operator to be '+', got '%s'", binaryOp.Operator)
	}

	// Check that the right-hand side is the multiplication
	rhs, ok := binaryOp.Right.(*ast.BinaryOpNode)
	if !ok {
		t.Fatalf("Expected RHS of addition to be a BinaryOpNode for multiplication, got %T", binaryOp.Right)
	}

	if rhs.Operator != "*" {
		t.Errorf("Expected RHS operator to be '*', got '%s'", rhs.Operator)
	}
}
