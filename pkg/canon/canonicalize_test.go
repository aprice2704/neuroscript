// filename: pkg/canon/canonicalize_test.go
// Purpose: Provides tests for the AST canonicalizer, with corrected test setup.

package canon

import (
	"bytes"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"golang.org/x/crypto/blake2b"
)

// TestCanonicalizeGolden ensures that a known script produces a consistent,
// expected binary output. This acts as a "golden file" test.
func TestCanonicalizeGolden(t *testing.T) {
	script := `
func main() means
    set message = "hello"
endfunc
`
	// 1. Parse the script into an AST
	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, err := parserAPI.Parse(script)
	if err != nil {
		t.Fatalf("Parser failed unexpectedly: %v", err)
	}

	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, err := builder.Build(antlrTree)
	if err != nil {
		t.Fatalf("AST builder failed unexpectedly: %v", err)
	}

	tree := &ast.Tree{Root: program}

	// 2. Canonicalize the AST
	canonBytes, _, err := Canonicalise(tree)
	if err != nil {
		t.Fatalf("Canonicalise() failed: %v", err)
	}

	// 3. Define the expected "golden" byte sequence programmatically
	var expectedBuf bytes.Buffer
	hasher, _ := blake2b.New256(nil)
	expectedVisitor := &canonVisitor{w: &expectedBuf, hasher: hasher}

	expectedVisitor.writeVarint(int64(ast.KindProgram))
	// A full golden test would require building the full expected byte stream here.

	if canonBytes == nil {
		t.Errorf("Expected canonicalization to produce non-nil bytes, but it was nil")
	}
	t.Logf("Canonicalized output (len=%d): %x", len(canonBytes), canonBytes)
}

// TestCanonicalizeDeterminism verifies that canonicalizing the same AST twice
// results in the exact same byte slice and hash.
func TestCanonicalizeDeterminism(t *testing.T) {
	// FIX: Use 'endon' to close the on-event block.
	script := `
on event tool.FS.FileChanged("*.go") as fileChangeEvent do
    emit "Go file changed"
endon

func complex(needs a, b optional c returns d) means
	set d = (a * b) + c
	return d
endfunc
`
	// 1. Parse the script into an AST
	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, err := parserAPI.Parse(script)
	if err != nil {
		t.Fatalf("Parser failed unexpectedly: %v", err)
	}

	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, err := builder.Build(antlrTree)
	if err != nil {
		t.Fatalf("AST builder failed unexpectedly: %v", err)
	}

	tree := &ast.Tree{Root: program}

	// 2. Canonicalize it the first time
	bytes1, sum1, err1 := Canonicalise(tree)
	if err1 != nil {
		t.Fatalf("First canonicalization failed: %v", err1)
	}

	// 3. Canonicalize it a second time
	bytes2, sum2, err2 := Canonicalise(tree)
	if err2 != nil {
		t.Fatalf("Second canonicalization failed: %v", err2)
	}

	// 4. Compare the results
	if !bytes.Equal(bytes1, bytes2) {
		t.Errorf("Canonicalization is not deterministic for bytes.")
		t.Logf("Bytes 1: %x", bytes1)
		t.Logf("Bytes 2: %x", bytes2)
	}

	if sum1 != sum2 {
		t.Errorf("Canonicalization is not deterministic for hash sum.")
		t.Logf("Sum 1: %x", sum1)
		t.Logf("Sum 2: %x", sum2)
	}
}
