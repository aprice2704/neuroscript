// NeuroScript Version: 0.6.3
// File version: 5
// Purpose: WIDENED SCOPE. This test now decodes the AST and asserts its structure, reproducing the data loss bug that occurs during canonicalization. Updated to use new registry-based canon functions.
// filename: pkg/parser/e2e_repro_canon_test.go
// nlines: 80
// risk_rating: HIGH

package parser_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/canon"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// TestE2ECanonReproFailure reproduces the data loss from the e2e test by
// fully encoding and then decoding the AST. The test will fail on the final
// assertion because the `canon` package does not correctly handle the `Values`
// of a `return` statement, causing them to be dropped during serialization.
func TestE2ECanonReproFailure(t *testing.T) {
	src := `
func main(returns msg) means
  set msg = "hello world"
  return msg
endfunc
`
	logger := logging.NewTestLogger(t)
	parserAPI := parser.NewParserAPI(logger)
	antlrTree, _, err := parserAPI.ParseAndGetStream("source.ns", src)
	if err != nil {
		t.Fatalf("Parse() failed unexpectedly: %v", err)
	}

	builder := parser.NewASTBuilder(logger)
	program, _, err := builder.Build(antlrTree)
	if err != nil {
		t.Fatalf("Build() failed unexpectedly: %v", err)
	}
	tree := &ast.Tree{Root: program}

	// 1. Canonicalize the AST into a binary blob.
	t.Log("Attempting to canonicalize the AST...")
	encoded, _, err := canon.CanonicaliseWithRegistry(tree)
	if err != nil {
		t.Fatalf("canon.CanonicaliseWithRegistry failed: %v", err)
	}
	t.Log("Canonicalize successful.")

	// 2. Decode the blob back into a new AST structure.
	t.Log("Attempting to decode the blob...")
	decodedTree, err := canon.DecodeWithRegistry(encoded)
	if err != nil {
		t.Fatalf("canon.DecodeWithRegistry returned an error: %v", err)
	}
	if decodedTree == nil {
		t.Fatal("canon.DecodeWithRegistry returned a nil tree without error.")
	}
	t.Log("Decode successful. Now inspecting the decoded AST.")

	// 3. Inspect the DECODED AST to prove the data was lost.
	decodedProgram, ok := decodedTree.Root.(*ast.Program)
	if !ok {
		t.Fatalf("Decoded root is not *ast.Program, but %T", decodedTree.Root)
	}

	proc, ok := decodedProgram.Procedures["main"]
	if !ok {
		t.Fatal("Procedure 'main' not found in decoded AST")
	}

	if len(proc.Steps) != 2 {
		t.Fatalf("Expected 2 steps in decoded procedure, got %d", len(proc.Steps))
	}

	returnStep := proc.Steps[1]
	if returnStep.Type != "return" {
		t.Fatalf("Expected the second step to be 'return', but got '%s'", returnStep.Type)
	}

	// THIS IS THE ASSERTION THAT WILL NOW FAIL
	if len(returnStep.Values) != 1 {
		t.Fatalf("BUG REPRODUCED: Expected the decoded return step's 'Values' slice to have 1 element, but it has %d", len(returnStep.Values))
	}

	returnValue := returnStep.Values[0]
	if _, ok := returnValue.(*ast.VariableNode); !ok {
		t.Errorf("Expected the return value to be of type *ast.VariableNode, but got %T", returnValue)
	}
}
