// filename: pkg/parser/expression_statement_test.go
package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestExpressionStatement(t *testing.T) {
	t.Run("Valid inside a function", func(t *testing.T) {
		script := `
			func MyFunc() means
				tool.MyTool("some_arg")
			endfunc
		`
		prog := testParseAndBuild(t, script)
		proc := prog.Procedures["MyFunc"]
		if len(proc.Steps) != 1 {
			t.Fatalf("Expected 1 step, got %d", len(proc.Steps))
		}
		step := proc.Steps[0]
		if step.Type != "expression_statement" {
			t.Errorf("Expected step type 'expression_statement', got '%s'", step.Type)
		}
		if step.ExpressionStmt == nil {
			t.Fatal("Expected ExpressionStmt to be non-nil")
		}

		call, ok := step.ExpressionStmt.Expression.(*ast.CallableExprNode)
		if !ok {
			t.Fatalf("Expected expression to be a CallableExprNode, got %T", step.ExpressionStmt.Expression)
		}
		if call.Target.Name != "MyTool" {
			t.Errorf("Expected tool call to 'MyTool', got '%s'", call.Target.Name)
		}
	})

	t.Run("Invalid outside a block", func(t *testing.T) {
		script := `
			# This should not be allowed at the top level of a script.
			tool.MyTool("this should cause a parser error")
		`
		testForParserError(t, script)
	})
}
