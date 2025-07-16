// filename: pkg/parser/ast_lvalue_ast_test.go
// Verifies that the AST builder produces the correct accessor chain for the
// complex l‑value a.b[0]["c"].d[1].
//
// The comparison ignores positional and bookkeeping fields so the test is
// resilient to future changes in debug / token‑tracking logic.

package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// TestComplexLValueAST ensures the builder creates the expected accessor chain.
func TestComplexLValueAST(t *testing.T) {
	script := `func t() means
set a.b[0]["c"].d[1] = 42
endfunc`

	got := parseScriptToLValueNode(t, script)

	want := &ast.LValueNode{
		Identifier: "a",
		Accessors: []*ast.AccessorNode{
			{Type: ast.DotAccess, Key: &ast.StringLiteralNode{Value: "b"}},
			{Type: ast.BracketAccess, Key: &ast.NumberLiteralNode{Value: float64(0)}},
			{Type: ast.BracketAccess, Key: &ast.StringLiteralNode{Value: "c"}},
			{Type: ast.DotAccess, Key: &ast.StringLiteralNode{Value: "d"}},
			{Type: ast.BracketAccess, Key: &ast.NumberLiteralNode{Value: float64(1)}},
		},
	}

	cmpOpts := []cmp.Option{
		// Ignore all positional / bookkeeping fields
		cmpopts.IgnoreFields(ast.BaseNode{}, "StartPos", "StopPos", "NodeKind"),
		cmpopts.IgnoreFields(ast.LValueNode{}, "Position"),
		cmpopts.IgnoreFields(ast.AccessorNode{}, "Pos"),
		cmpopts.IgnoreFields(ast.StringLiteralNode{}, "Pos"),
		cmpopts.IgnoreFields(ast.NumberLiteralNode{}, "Pos"),
	}

	if diff := cmp.Diff(want, got, cmpOpts...); diff != "" {
		t.Fatalf("AST mismatch (-want +got):\n%s", diff)
	}
}
