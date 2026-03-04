// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 5
// :: description: Comprehensive tests for parsing literals with UI safety overrides.
// :: latestChange: Replaced backticks with ``` in test strings for UI safety.
// :: filename: pkg/parser/ast_builder_literals_test.go
// :: serialization: go

package parser

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestLiteralParsing(t *testing.T) {
	getExpr := func(script string) ast.Expression {
		// Restore backticks before building
		cleanScript := strings.ReplaceAll(script, "```", "```")
		prog := testParseAndBuild(t, cleanScript)
		proc, ok := prog.Procedures["Test"]
		if !ok {
			t.Fatalf("Procedure 'Test' not found in parsed AST")
		}
		if len(proc.Steps) == 0 {
			t.Fatalf("Procedure body empty")
		}
		step := proc.Steps[0]
		if step.Type != "emit" {
			t.Fatalf("Expected emit step, got %s", step.Type)
		}
		return step.Values[0]
	}

	t.Run("Number Literals", func(t *testing.T) {
		expr := getExpr(`func Test() means
			emit 123
		endfunc`)
		if num, ok := expr.(*ast.NumberLiteralNode); !ok || num.Value != 123.0 {
			t.Errorf("Expected NumberLiteralNode with value 123.0, got %T with value %v", expr, num)
		}
	})

	t.Run("Raw Strings - Triple Backtick Interpolation", func(t *testing.T) {
		script := "func Test() means\n\temit ```Hello {{name}}!```\nendfunc"
		expr := getExpr(script)
		if _, ok := expr.(*ast.BinaryOpNode); !ok {
			t.Errorf("Expected BinaryOpNode for interpolated triple-backtick string, got %T", expr)
		}
	})

	t.Run("Raw Strings - Double Bracket", func(t *testing.T) {
		script := "func Test() means\n\temit [[line1\nline2]]\nendfunc"
		expr := getExpr(script)
		strNode, ok := expr.(*ast.StringLiteralNode)
		if !ok {
			t.Fatalf("Expected StringLiteralNode, got %T", expr)
		}
		expected := "line1\nline2"
		if strNode.Value != expected {
			t.Errorf("Double bracket content mismatch.\nExpected:\n%s\nGot:\n%s", expected, strNode.Value)
		}
	})

	t.Run("Double Bracket Interpolation AST", func(t *testing.T) {
		script := "func Test() means\n\temit [[Hello {{@nl}} {{name}}!]]\nendfunc"
		expr := getExpr(script)
		if _, ok := expr.(*ast.BinaryOpNode); !ok {
			t.Errorf("Expected BinaryOpNode for interpolated string, got %T", expr)
		}
	})
}
