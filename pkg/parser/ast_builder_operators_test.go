package parser

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// Helper to parse an expression string and return the top-level AST node
func parseExpression(t *testing.T, exprStr string) ast.Expression {
	t.Helper()
	// Create a dummy script that uses the expression in a valid statement.
	script := fmt.Sprintf("func t() means\n set x = %s\nendfunc", exprStr)

	// Use the existing test helper to parse the script and get the procedure body.
	bodyNodes := parseStringToProcedureBodyNodes(t, script, "t")
	if len(bodyNodes) < 1 {
		t.Fatalf("Expected at least one statement in the parsed test script, but got 0")
	}

	// Get the 'set' statement from the parsed body.
	setStep := bodyNodes[0]
	if setStep.Type != "set" {
		t.Fatalf("Expected the first statement to be 'set', but got '%s'", setStep.Type)
	}
	if len(setStep.Values) < 1 {
		t.Fatalf("Expected the 'set' statement to have at least one value, but got 0")
	}

	// The first value of the 'set' statement is our expression.
	return setStep.Values[0]
}

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
