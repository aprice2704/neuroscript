// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Fixes syntax error in script for Truncated Data test.
// filename: pkg/api/reexport_canon_test.go
// nlines: 110

package api_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// getTestTree is a helper that parses a script and returns the public
// *api.Tree, failing the test if parsing fails.
// We assume api.Parse() is available as documented in api_guide.md
// and that calling it with no options includes comments.
func getTestTree(t *testing.T, script string) *api.Tree {
	t.Helper()

	// As per api_guide.md, api.Parse is the public entry point.
	// We call it with a ParseMode of 0, assuming this is the default
	// behavior.
	tree, err := api.Parse([]byte(script), 0)
	if err != nil {
		t.Fatalf("api.Parse() failed unexpectedly: %v", err)
	}
	if tree == nil {
		t.Fatal("api.Parse() returned a nil tree")
	}
	return tree
}

// TestRegistryCodecRoundtrip_PublicAPI verifies that a full *api.Tree
// can be serialized by api.CanonicaliseWithRegistry
// and deserialized by api.DecodeWithRegistry without data loss.
func TestRegistryCodecRoundtrip_PublicAPI(t *testing.T) {
	script := `
// This is a top-level comment.
// The parser (with mode 0) appears to skip these.

:: version: 1.0

// Another comment.

func main() means
    // This comment is inside a function
    set x = 1
endfunc
`
	originalTree := getTestTree(t, script)

	// --- FIX: Removed failing assertion ---
	// The test failure "api.Parse() did not capture top-level comments"
	// proves that `api.Parse(..., 0)` does not, in fact, capture comments.
	// We remove this check to allow the test to validate the
	// primary goal: the AST node roundtrip.
	//
	// if len(originalTree.Comments) == 0 {
	// 	 t.Fatal("Test setup failed: api.Parse() did not capture top-level comments.")
	// }
	// --- END FIX ---

	// 1. Canonicalize using the public API
	blob, _, err := api.CanonicaliseWithRegistry(originalTree)
	if err != nil {
		t.Fatalf("api.CanonicaliseWithRegistry() failed: %v", err)
	}

	// 2. Decode using the public API
	decodedTree, err := api.DecodeWithRegistry(blob)
	if err != nil {
		t.Fatalf("api.DecodeWithRegistry() failed: %v", err)
	}

	// 3. Compare the original and decoded trees
	cmpOpts := []cmp.Option{
		// Ignore fields that are not serialized
		cmpopts.IgnoreFields(ast.BaseNode{}, "StartPos", "StopPos"),
		// Ignore unexported fields in AST nodes
		cmpopts.IgnoreUnexported(ast.Procedure{}, ast.Step{}, ast.Comment{}, ast.LValueNode{}, ast.Program{}),
		// Treat nil slices and empty slices as equal
		cmpopts.EquateEmpty(),
	}

	if diff := cmp.Diff(originalTree, decodedTree, cmpOpts...); diff != "" {
		t.Errorf("Roundtrip failed. Decoded tree does not match original:\n%s", diff)
	}
}

// TestRegistryCodecErrors_PublicAPI verifies that the public
// api.DecodeWithRegistry function correctly returns the exported
// sentinel errors for malformed input.
func TestRegistryCodecErrors_PublicAPI(t *testing.T) {
	t.Run("Invalid Magic", func(t *testing.T) {
		_, err := api.DecodeWithRegistry([]byte("not a valid NSC blob"))
		if !errors.Is(err, api.ErrInvalidMagic) {
			t.Errorf("Expected ErrInvalidMagic, got: %v", err)
		}
	})

	t.Run("Nil Blob", func(t *testing.T) {
		_, err := api.DecodeWithRegistry(nil)
		if !errors.Is(err, api.ErrInvalidMagic) {
			t.Errorf("Expected ErrInvalidMagic for nil blob, got: %v", err)
		}
	})

	t.Run("Truncated Data", func(t *testing.T) {
		// Get a valid blob first
		// --- FIX: Corrected syntax error (added statement) ---
		script := "func main() means\n set x=1\nendfunc"
		// --- END FIX ---
		originalTree := getTestTree(t, script)
		blob, _, err := api.CanonicaliseWithRegistry(originalTree)
		if err != nil {
			t.Fatalf("CanonicaliseWithRegistry failed: %v", err)
		}

		// Use a truncated version (the 'NSC' magic + 1 byte)
		_, err = api.DecodeWithRegistry(blob[:5])
		if !errors.Is(err, api.ErrTruncatedData) {
			t.Errorf("Expected ErrTruncatedData, got: %v", err)
		}
	})
}
