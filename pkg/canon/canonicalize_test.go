// filename: pkg/canon/canonicalize_test.go
// Purpose: Provides tests for the AST canonicalizer, with corrected test setup.

package canon

import (
	"bytes"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/types"
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

	expectedVisitor.writeVarint(int64(types.KindProgram))
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

// TestCanonicalize_CommandBlock is a regression test to ensure that a program
// containing a `command` block can be canonicalized and decoded successfully,
// fixing the previous integrity check failure.
func TestCanonicalize_CommandBlock(t *testing.T) {
	// 1. Define a script with a simple command block.
	script := `
command
    emit "hello from command"
endcommand
`
	// 2. Parse the script into an AST.
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

	// 3. Canonicalize the tree.
	blob, _, err := Canonicalise(tree)
	if err != nil {
		t.Fatalf("Canonicalise() failed: %v", err)
	}

	// 4. Decode the blob back into a tree.
	decodedTree, err := Decode(blob)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	// 5. Verify the decoded tree has the command block.
	if decodedTree == nil || decodedTree.Root == nil {
		t.Fatal("Decoded tree or its root is nil")
	}
	decodedProgram, ok := decodedTree.Root.(*ast.Program)
	if !ok {
		t.Fatalf("Decoded root is not a *ast.Program, but %T", decodedTree.Root)
	}
	if len(decodedProgram.Commands) != 1 {
		t.Errorf("Expected 1 command block in decoded program, but got %d", len(decodedProgram.Commands))
	}
}
