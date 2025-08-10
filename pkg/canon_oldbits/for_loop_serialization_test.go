// filename: pkg/canon/for_loop_serialization_test.go
// NeuroScript Version: 0.6.2
// File version: 1
// Purpose: Proves that the canonicalization encoder/decoder for ast.Step omits all fields for a 'for' loop.
// nlines: 65
// risk_rating: HIGH

package canon_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/ast"
)

// TestForLoopSerialization proves that the canonicalization process is lossy for
// 'for each' loops. The canon.visitStep and canon.readStep functions are missing
// the "for" case, meaning the loop's collection, variable, and body are never
// serialized or deserialized.
//
// This test will fail until the encoder and decoder are updated to handle the
// specific fields of a for loop (`Collection`, `LoopVarName`, `Body`).
func TestForLoopSerialization(t *testing.T) {
	// A standard 'for each' loop script.
	src := `
func main() means
    set items = [1, 2]
    for each item in items
        emit item
    endfor
endfunc
`
	// 1. Parse the script to get a known-good AST.
	tree, err := api.Parse([]byte(src), api.ParsePreserveComments)
	if err != nil {
		t.Fatalf("api.Parse() failed unexpectedly: %v", err)
	}

	// 2. Run the AST through the full Canonicalise/Decode cycle.
	blob, _, err := api.Canonicalise(tree)
	if err != nil {
		t.Fatalf("api.Canonicalise() failed unexpectedly: %v", err)
	}
	decodedTree, _, err := api.Decode(blob)
	if err != nil {
		t.Fatalf("api.Decode() failed unexpectedly: %v", err)
	}

	// 3. Inspect the decoded AST.
	program := decodedTree.Root.(*ast.Program)
	proc := program.Procedures["main"]
	if len(proc.Steps) != 2 {
		t.Fatalf("Expected 2 steps in decoded procedure, got %d", len(proc.Steps))
	}

	forEachStep := proc.Steps[1]
	if forEachStep.Type != "for" {
		t.Fatalf("Expected second step to be 'for', got '%s'", forEachStep.Type)
	}

	// 4. These assertions will fail because the fields were never encoded or decoded.
	if forEachStep.Collection == nil {
		t.Fatal("BUG REPRODUCED: The 'Collection' field is nil after decoding.")
	}
	if forEachStep.LoopVarName == "" {
		t.Fatal("BUG REPRODUCED: The 'LoopVarName' field is empty after decoding.")
	}
	if len(forEachStep.Body) == 0 {
		t.Fatal("BUG REPRODUCED: The 'Body' slice is empty after decoding.")
	}
}
