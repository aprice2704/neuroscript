// NeuroScript Version: 0.7.2
// File version: 1
// Purpose: Adds a test to verify that the canonicalization process is deterministic, producing identical byte output for the same AST input.
// filename: pkg/canon/determinism_test.go
// nlines: 50
// risk_rating: MEDIUM

package canon_test

import (
	"bytes"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/canon"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/google/go-cmp/cmp"
)

// TestCanonicalizationIsDeterministic ensures that serializing the exact same
// AST twice produces byte-for-byte identical blobs. This is critical for
// any system that uses hashing or caching of the canonical format.
func TestCanonicalizationIsDeterministic(t *testing.T) {
	script := `
:: title: Determinism Test
:: version: 1.0

func main(returns result) means
	set my_map = {"b": 2, "a": 1} // Unordered keys
	set total = 0
	for each item in [3, 1, 2] // Unordered elements
		set total = total + item
	endfor
	return total
endfunc
`
	// 1. Parse the script to create a consistent input AST.
	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	astBuilder := parser.NewASTBuilder(logging.NewNoOpLogger())
	antlrTree, _, _ := parserAPI.ParseAndGetStream("deterministic_test.ns", script)
	program, _, _ := astBuilder.Build(antlrTree)
	tree := &ast.Tree{Root: program}

	// 2. Canonicalize the AST twice.
	blob1, _, err1 := canon.CanonicaliseWithRegistry(tree)
	if err1 != nil {
		t.Fatalf("First canonicalization failed: %v", err1)
	}

	blob2, _, err2 := canon.CanonicaliseWithRegistry(tree)
	if err2 != nil {
		t.Fatalf("Second canonicalization failed: %v", err2)
	}

	// 3. Assert that the two blobs are identical.
	if !bytes.Equal(blob1, blob2) {
		t.Errorf("Canonicalization is not deterministic. The two blobs are not identical.")
		// Use go-cmp's diff for a more readable failure message if they differ.
		if diff := cmp.Diff(blob1, blob2); diff != "" {
			t.Logf("Diff (-blob1 +blob2):\n%s", diff)
		}
	}
}
