// filename: pkg/api/canon_for_each_repro_test.go
// NeuroScript Version: 0.6.2
// File version: 1
// Purpose: Provides a definitive, failing test case to prove that the canonicalization process drops the 'Collection' field from a 'for each' loop's AST node.
// nlines: 60
// risk_rating: HIGH

package api_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/ast"
)

// TestForEachCanonicalizationFailure reproduces the exact data loss scenario seen
// in the zadeh integration test. It proves that the public API's Canonicalise/Decode
// cycle fails to preserve the 'Collection' field of a 'for each' step, resulting
// in a malformed AST that causes a runtime panic in the interpreter.
//
// This test is expected to FAIL until the underlying bug in the internal `canon`
// package is fixed.
func TestForEachCanonicalizationFailure(t *testing.T) {
	// 1. A minimal script that mirrors the failing structure.
	src := `
func main() means
    set my_items = [1, 2, 3]
    for each item in my_items
        emit item
    endfor
endfunc
`
	// 2. Parse the script into an AST.
	tree, err := api.Parse([]byte(src), api.ParsePreserveComments)
	if err != nil {
		t.Fatalf("api.Parse() failed unexpectedly: %v", err)
	}

	// 3. Run the AST through the full canonicalization and decoding cycle.
	// This simulates the exact process used when signing and loading scripts.
	blob, _, err := api.Canonicalise(tree)
	if err != nil {
		t.Fatalf("api.Canonicalise() failed unexpectedly: %v", err)
	}

	decodedTree, _, err := api.Decode(blob)
	if err != nil {
		t.Fatalf("api.Decode() failed unexpectedly: %v", err)
	}

	// 4. Inspect the DECODED AST to prove data was lost.
	program, ok := decodedTree.Root.(*ast.Program)
	if !ok {
		t.Fatalf("Decoded tree root is not *ast.Program, but %T", decodedTree.Root)
	}
	proc, ok := program.Procedures["main"]
	if !ok {
		t.Fatal("Procedure 'main' not found in decoded AST")
	}
	if len(proc.Steps) != 2 {
		t.Fatalf("Expected 2 steps in decoded procedure, got %d", len(proc.Steps))
	}

	forEachStep := proc.Steps[1]
	if forEachStep.Type != "for" {
		t.Fatalf("Expected second step to be 'for', got '%s'", forEachStep.Type)
	}

	// 5. This is the assertion that MUST fail.
	// It proves the 'Collection' field was lost during the Canonicalise/Decode round-trip.
	if forEachStep.Collection == nil {
		t.Fatal("BUG REPRODUCED: The 'Collection' field of the 'for each' step is nil after decoding.")
	}
}
