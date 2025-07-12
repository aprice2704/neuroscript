// filename: pkg/canon/canonicalize_extra_test.go
// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Corrected invalid test case and added more robust checks.
// nlines: 135
// risk_rating: LOW

package canon

import (
	"bytes"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestCanonicalize_EmptyAndNil(t *testing.T) {
	t.Run("empty program", func(t *testing.T) {
		tree := &ast.Tree{Root: ast.NewProgram()}
		_, _, err := Canonicalise(tree)
		if err != nil {
			t.Errorf("Canonicalise() with empty program failed: %v", err)
		}
	})

	t.Run("nil root", func(t *testing.T) {
		tree := &ast.Tree{Root: nil}
		_, _, err := Canonicalise(tree)
		if err == nil {
			t.Error("Canonicalise() with nil root should have failed, but did not")
		}
	})

	t.Run("procedure with one step is valid", func(t *testing.T) {
		// FIX: The grammar requires a non-empty statement list. This test now
		// uses a valid script and verifies it can be canonicalized.
		script := `func empty() means
			emit "ok"
		endfunc`
		parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
		antlrTree, pErr := parserAPI.Parse(script)
		if pErr != nil {
			t.Fatalf("Parser failed unexpectedly: %v", pErr)
		}
		builder := parser.NewASTBuilder(logging.NewNoOpLogger())
		program, _, bErr := builder.Build(antlrTree)
		if bErr != nil {
			t.Fatalf("AST builder failed unexpectedly for a valid script: %v", bErr)
		}

		// Now, canonicalize the valid tree
		_, _, cErr := Canonicalise(&ast.Tree{Root: program})
		if cErr != nil {
			t.Errorf("Canonicalization of a valid simple function failed: %v", cErr)
		}
	})
}

func TestCanonicalize_OrderDeterminism(t *testing.T) {
	script1 := `
		func b() means
			emit "b"
		endfunc
		on event "a" do
			emit "a"
		endon
		func a() means
			emit "a"
		endfunc
	`
	script2 := `
		func a() means
			emit "a"
		endfunc
		on event "a" do
			emit "a"
		endon
		func b() means
			emit "b"
		endfunc
	`

	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	builder := parser.NewASTBuilder(logging.NewNoOpLogger())

	// Canonicalize first script
	antlrTree1, _ := parserAPI.Parse(script1)
	program1, _, _ := builder.Build(antlrTree1)
	tree1 := &ast.Tree{Root: program1}
	bytes1, sum1, _ := Canonicalise(tree1)

	// Canonicalize second script
	antlrTree2, _ := parserAPI.Parse(script2)
	program2, _, _ := builder.Build(antlrTree2)
	tree2 := &ast.Tree{Root: program2}
	bytes2, sum2, _ := Canonicalise(tree2)

	if !bytes.Equal(bytes1, bytes2) {
		t.Error("Canonicalization of differently ordered but semantically identical scripts produced different byte slices")
		t.Logf("Bytes 1: %x", bytes1)
		t.Logf("Bytes 2: %x", bytes2)
	}
	if sum1 != sum2 {
		t.Error("Canonicalization of differently ordered but semantically identical scripts produced different hashes")
	}
}

func TestCanonicalize_NumberLiterals(t *testing.T) {
	script1 := `func a() means
		set x = -0.0
	endfunc`
	script2 := `func a() means
		set x = 0.0
	endfunc`

	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	builder := parser.NewASTBuilder(logging.NewNoOpLogger())

	// Canonicalize first script
	antlrTree1, _ := parserAPI.Parse(script1)
	program1, _, _ := builder.Build(antlrTree1)
	tree1 := &ast.Tree{Root: program1}
	bytes1, _, _ := Canonicalise(tree1)

	// Canonicalize second script
	antlrTree2, _ := parserAPI.Parse(script2)
	program2, _, _ := builder.Build(antlrTree2)
	tree2 := &ast.Tree{Root: program2}
	bytes2, _, _ := Canonicalise(tree2)

	if !bytes.Equal(bytes1, bytes2) {
		t.Error("Canonicalization of -0.0 and 0.0 produced different byte slices, which is undesirable.")
		t.Logf("Bytes for -0.0: %x", bytes1)
		t.Logf("Bytes for  0.0: %x", bytes2)
	}
}
