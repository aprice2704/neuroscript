// filename: pkg/parser/ast_builder_literals_test.go
// NeuroScript Version: 0.9.6
// File version: 2
// Purpose: Comprehensive tests for parsing literals. Fixed NumberLiteral comparison to use float syntax to avoid type mismatch.

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestLiteralParsing(t *testing.T) {
	// Helper to extract the first expression from an emit statement
	// Uses the shared testParseAndBuild helper assumed to be in the package test scope
	getExpr := func(script string) ast.Expression {
		prog := testParseAndBuild(t, script)
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
		if len(step.Values) == 0 {
			t.Fatalf("Emit step has no values")
		}
		return step.Values[0]
	}

	t.Run("Number Literals", func(t *testing.T) {
		// Integer
		expr := getExpr(`func Test() means
			emit 123
		endfunc`)
		// FIX: Use 123.0 to match the float64 type stored in the AST node
		if num, ok := expr.(*ast.NumberLiteralNode); !ok || num.Value != 123.0 {
			t.Errorf("Expected NumberLiteralNode with value 123.0, got %T with value %v", expr, num)
		}

		// Float
		expr = getExpr(`func Test() means
			emit 45.67
		endfunc`)
		if num, ok := expr.(*ast.NumberLiteralNode); !ok || num.Value != 45.67 {
			t.Errorf("Expected NumberLiteralNode with value 45.67, got %T with value %v", expr, num)
		}
	})

	t.Run("String Literals - Standard", func(t *testing.T) {
		// Double Quoted
		expr := getExpr(`func Test() means
			emit "hello world"
		endfunc`)
		strNode, ok := expr.(*ast.StringLiteralNode)
		if !ok {
			t.Fatalf("Expected StringLiteralNode, got %T", expr)
		}
		if strNode.Value != "hello world" {
			t.Errorf("Expected 'hello world', got %q", strNode.Value)
		}
		if strNode.IsRaw {
			t.Error("Expected IsRaw=false for double quoted string")
		}

		// Single Quoted
		expr = getExpr(`func Test() means
			emit 'single quoted'
		endfunc`)
		strNode, ok = expr.(*ast.StringLiteralNode)
		if !ok {
			t.Fatalf("Expected StringLiteralNode, got %T", expr)
		}
		if strNode.Value != "single quoted" {
			t.Errorf("Expected 'single quoted', got %q", strNode.Value)
		}
	})

	t.Run("String Literals - Escaping", func(t *testing.T) {
		// Double quote escaping
		expr := getExpr(`func Test() means
			emit "say \"hello\""
		endfunc`)
		strNode, ok := expr.(*ast.StringLiteralNode)
		if !ok {
			t.Fatalf("Expected StringLiteralNode, got %T", expr)
		}
		if strNode.Value != `say "hello"` {
			t.Errorf("Expected 'say \"hello\"', got %q", strNode.Value)
		}

		// Single quote escaping (Manual handling check)
		expr = getExpr(`func Test() means
			emit 'it\'s ok'
		endfunc`)
		strNode, ok = expr.(*ast.StringLiteralNode)
		if !ok {
			t.Fatalf("Expected StringLiteralNode, got %T", expr)
		}
		if strNode.Value != "it's ok" {
			t.Errorf("Expected \"it's ok\", got %q", strNode.Value)
		}
	})

	t.Run("Raw Strings - Triple Backtick", func(t *testing.T) {
		script := "func Test() means\n\temit ```line1\nline2```\nendfunc"
		expr := getExpr(script)
		strNode, ok := expr.(*ast.StringLiteralNode)
		if !ok {
			t.Fatalf("Expected StringLiteralNode, got %T", expr)
		}
		expected := "line1\nline2"
		if strNode.Value != expected {
			t.Errorf("Triple backtick content mismatch.\nExpected:\n%s\nGot:\n%s", expected, strNode.Value)
		}
		if !strNode.IsRaw {
			t.Error("Expected IsRaw=true for triple backtick string")
		}
	})

	t.Run("Raw Strings - Triple Single Quote", func(t *testing.T) {
		// Basic Triple Single Quote
		script := "func Test() means\n\temit '''raw content'''\nendfunc"
		expr := getExpr(script)
		strNode, ok := expr.(*ast.StringLiteralNode)
		if !ok {
			t.Fatalf("Expected StringLiteralNode, got %T", expr)
		}
		if strNode.Value != "raw content" {
			t.Errorf("Expected 'raw content', got %q", strNode.Value)
		}
		if !strNode.IsRaw {
			t.Error("Expected IsRaw=true for triple single quote string")
		}

		// Nested Backticks (The critical fix case)
		// We expect this to parse cleanly now that TRIPLE_SQ_STRING is supported.
		script = "func Test() means\n\temit '''contains ```backticks``` inside'''\nendfunc"
		expr = getExpr(script)
		strNode, ok = expr.(*ast.StringLiteralNode)
		if !ok {
			t.Fatalf("Expected StringLiteralNode, got %T", expr)
		}
		expected := "contains ```backticks``` inside"
		if strNode.Value != expected {
			t.Errorf("Nested backtick content mismatch.\nExpected: %q\nGot:      %q", expected, strNode.Value)
		}
	})

	t.Run("Booleans and Nil", func(t *testing.T) {
		// True
		expr := getExpr(`func Test() means
			emit true
		endfunc`)
		if boolNode, ok := expr.(*ast.BooleanLiteralNode); !ok || !boolNode.Value {
			t.Errorf("Expected BooleanLiteralNode(true), got %T(%v)", expr, boolNode)
		}

		// False
		expr = getExpr(`func Test() means
			emit false
		endfunc`)
		if boolNode, ok := expr.(*ast.BooleanLiteralNode); !ok || boolNode.Value {
			t.Errorf("Expected BooleanLiteralNode(false), got %T(%v)", expr, boolNode)
		}

		// Nil
		expr = getExpr(`func Test() means
			emit nil
		endfunc`)
		if _, ok := expr.(*ast.NilLiteralNode); !ok {
			t.Errorf("Expected NilLiteralNode, got %T", expr)
		}
	})
}
