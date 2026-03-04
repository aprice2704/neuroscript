// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 7
// :: description: Comprehensive tests for parsing literals with UI safety overrides. Added severe edge-case tests.
// :: latestChange: Added tests for adjacent placeholders, unclosed tags, and non-interpolated raw strings.
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
		cleanScript := strings.ReplaceAll(script, "```", "`"+"`"+"`")
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
		if _, ok := expr.(*ast.InterpolatedStringNode); !ok {
			t.Errorf("Expected InterpolatedStringNode for interpolated triple-backtick string, got %T", expr)
		}
	})

	t.Run("Raw Strings - Double Bracket", func(t *testing.T) {
		script := "func Test() means\n\temit [[line1\nline2]]\nendfunc"
		expr := getExpr(script)
		strNode, ok := expr.(*ast.InterpolatedStringNode)
		if !ok {
			t.Fatalf("Expected InterpolatedStringNode, got %T", expr)
		}
		// In an InterpolatedStringNode without templates, Parts contains the string.
		if len(strNode.Parts) != 1 {
			t.Fatalf("Expected exactly 1 part in double bracket literal")
		}
		expected := "line1\nline2"
		lit, ok := strNode.Parts[0].(*ast.StringLiteralNode)
		if !ok || lit.Value != expected {
			t.Errorf("Double bracket content mismatch.\nExpected:\n%s\nGot:\n%s", expected, lit.Value)
		}
	})

	t.Run("Double Bracket Interpolation AST", func(t *testing.T) {
		script := "func Test() means\n\temit [[Hello {{@nl}} {{name}}!]]\nendfunc"
		expr := getExpr(script)
		if _, ok := expr.(*ast.InterpolatedStringNode); !ok {
			t.Errorf("Expected InterpolatedStringNode for interpolated string, got %T", expr)
		}
	})

	t.Run("Triple Single Quotes No Interpolation", func(t *testing.T) {
		script := "func Test() means\n\temit '''Hello {{name}}!'''\nendfunc"
		expr := getExpr(script)

		// Triple single quotes should NOT trigger interpolation parsing
		strNode, ok := expr.(*ast.StringLiteralNode)
		if !ok {
			t.Fatalf("Expected StringLiteralNode for ''', got %T", expr)
		}
		if strNode.Value != "Hello {{name}}!" {
			t.Errorf("Content mismatch, expected raw 'Hello {{name}}!', got: %s", strNode.Value)
		}
	})

	t.Run("Adjacent Placeholders", func(t *testing.T) {
		script := "func Test() means\n\temit [[{{a}}{{b}}]]\nendfunc"
		expr := getExpr(script)
		strNode, ok := expr.(*ast.InterpolatedStringNode)
		if !ok {
			t.Fatalf("Expected InterpolatedStringNode, got %T", expr)
		}
		if len(strNode.Parts) != 2 {
			t.Fatalf("Expected exactly 2 parts, got %d", len(strNode.Parts))
		}
		// Part 1 should be variable 'a'
		if v, ok := strNode.Parts[0].(*ast.VariableNode); !ok || v.Name != "a" {
			t.Errorf("Expected VariableNode 'a', got: %T", strNode.Parts[0])
		}
		// Part 2 should be variable 'b'
		if v, ok := strNode.Parts[1].(*ast.VariableNode); !ok || v.Name != "b" {
			t.Errorf("Expected VariableNode 'b', got: %T", strNode.Parts[1])
		}
	})

	t.Run("Unclosed Placeholder Tag", func(t *testing.T) {
		script := "func Test() means\n\temit [[Hello {{name]]\nendfunc"
		expr := getExpr(script)
		strNode, ok := expr.(*ast.InterpolatedStringNode)
		if !ok {
			t.Fatalf("Expected InterpolatedStringNode, got %T", expr)
		}
		if len(strNode.Parts) != 2 {
			t.Fatalf("Expected exactly 2 parts, got %d", len(strNode.Parts))
		}
		// Part 1: "Hello "
		if sl, ok := strNode.Parts[0].(*ast.StringLiteralNode); !ok || sl.Value != "Hello " {
			t.Errorf("Expected 'Hello ', got %v", strNode.Parts[0])
		}
		// Part 2: "{{name" (Because it didn't find a closing }})
		if sl, ok := strNode.Parts[1].(*ast.StringLiteralNode); !ok || sl.Value != "{{name" {
			t.Errorf("Expected literal '{{name', got %v", strNode.Parts[1])
		}
	})
}
