// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Provides tests for the AST decoder.
// filename: pkg/canon/decoder_test.go
// nlines: 65
// risk_rating: HIGH

package canon

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestDecodeGolden(t *testing.T) {
	script := `
func main() means
    set message = "hello"
endfunc
`
	// 1. Create the initial AST.
	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, _ := parserAPI.Parse(script)
	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, _ := builder.Build(antlrTree)
	tree := &ast.Tree{Root: program}

	// 2. Canonicalize it.
	blob, _, err := Canonicalise(tree)
	if err != nil {
		t.Fatalf("Canonicalise() failed: %v", err)
	}

	// 3. Decode it.
	decodedTree, err := Decode(blob)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	// 4. Verify the decoded structure.
	if decodedTree == nil {
		t.Fatal("Decoded tree is nil")
	}
	decodedProgram, ok := decodedTree.Root.(*ast.Program)
	if !ok {
		t.Fatalf("Decoded root is not a *ast.Program, but %T", decodedTree.Root)
	}
	if len(decodedProgram.Procedures) != 1 {
		t.Fatalf("Expected 1 procedure in decoded program, got %d", len(decodedProgram.Procedures))
	}
	if _, ok := decodedProgram.Procedures["main"]; !ok {
		t.Error("Decoded program is missing 'main' procedure")
	}
}

func TestDecodeErrors(t *testing.T) {
	t.Run("decode empty blob", func(t *testing.T) {
		_, err := Decode([]byte{})
		if err == nil {
			t.Error("Expected error when decoding empty blob, but got nil")
		}
	})

	t.Run("decode nil blob", func(t *testing.T) {
		_, err := Decode(nil)
		if err == nil {
			t.Error("Expected error when decoding nil blob, but got nil")
		}
	})

	t.Run("decode corrupted blob", func(t *testing.T) {
		// A single byte is not a valid canonicalized stream.
		corruptedBlob := []byte{0x01}
		_, err := Decode(corruptedBlob)
		if err == nil {
			t.Error("Expected error when decoding corrupted blob, but got nil")
		}
	})
}
