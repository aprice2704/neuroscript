// filename: pkg/canon/program_walk_test.go
// NeuroScript Version: 0.6.2
// File version: 1
// Purpose: Proves the canonicalization tree walk correctly visits a 'for' loop's collection.
// nlines: 50
// risk_rating: HIGH

package canon

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// TestProgramWalkForLoopCollection proves that the canonicalizer's main visit
// function correctly traverses into a for-loop's collection. Previous tests
// have failed to isolate this, suggesting a bug in the recursive tree walk.
//
// This test will fail if the final byte blob does not contain the serialized
// representation of the collection variable name ("items").
func TestProgramWalkForLoopCollection(t *testing.T) {
	// 1. Create a full Program AST with a for loop.
	// We use the parser to ensure all parent/child links are correct.
	script := `
func main() means
    for each item in items
        emit item
    endfor
endfunc
`
	parserAPI := parser.NewParserAPI(nil)
	antlrTree, _ := parserAPI.Parse(script)
	builder := parser.NewASTBuilder(nil)
	program, _, _ := builder.Build(antlrTree)
	tree := &ast.Tree{Root: program}

	// 2. Canonicalize the entire program.
	blob, _, err := Canonicalise(tree)
	if err != nil {
		t.Fatalf("Canonicalise() failed unexpectedly: %v", err)
	}

	// 3. To prove the collection was visited, we can decode the blob
	// and check the resulting AST structure.
	decodedTree, err := Decode(blob)
	if err != nil {
		t.Fatalf("Decode() failed unexpectedly: %v", err)
	}

	decodedProgram := decodedTree.Root.(*ast.Program)
	mainProc := decodedProgram.Procedures["main"]
	forStep := mainProc.Steps[0]

	// 4. This assertion will now fail if the collection was not serialized.
	if forStep.Collection == nil {
		t.Fatal("BUG REPRODUCED: The 'Collection' field is nil after a full program decode, proving a tree walk error.")
	}

	collectionVar, ok := forStep.Collection.(*ast.VariableNode)
	if !ok {
		t.Fatalf("Expected collection to be a VariableNode, but got %T", forStep.Collection)
	}

	if collectionVar.Name != "items" {
		t.Fatalf("Expected collection variable to be 'items', but got '%s'", collectionVar.Name)
	}
}
