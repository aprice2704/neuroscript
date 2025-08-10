// NeuroScript Version: 0.6.3
// File version: 1
// Purpose: Provides an end-to-end test for the new registry-based codec system.
// filename: pkg/canon/codec_test.go
// nlines: 60
// risk_rating: HIGH

package canon

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestRegistryCodecRoundtrip verifies that a simple AST can be successfully
// encoded and decoded using the new registry-based system, ensuring the
// architecture is sound.
func TestRegistryCodecRoundtrip(t *testing.T) {
	script := `
func main() means
    set x = "hello"
endfunc
`
	// 1. Parse the script to get the original, known-good AST.
	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, pErr := parserAPI.Parse(script)
	if pErr != nil {
		t.Fatalf("parser.Parse() failed unexpectedly: %v", pErr)
	}
	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, bErr := builder.Build(antlrTree)
	if bErr != nil {
		t.Fatalf("ast.Build() failed unexpectedly: %v", bErr)
	}
	originalTree := &ast.Tree{Root: program}

	// 2. Run the AST through the new registry-based Canonicalise/Decode cycle.
	blob, _, err := CanonicaliseWithRegistry(originalTree)
	if err != nil {
		t.Fatalf("CanonicaliseWithRegistry() failed unexpectedly: %v", err)
	}
	decodedTree, err := DecodeWithRegistry(blob)
	if err != nil {
		t.Fatalf("DecodeWithRegistry() failed unexpectedly: %v", err)
	}

	// 3. Perform a deep comparison and fail if there are any differences.
	cmpOpts := []cmp.Option{
		cmpopts.IgnoreFields(ast.BaseNode{}, "StartPos", "StopPos"),
		cmpopts.IgnoreUnexported(ast.Procedure{}, ast.Step{}, ast.LValueNode{}),
		cmpopts.EquateEmpty(),
	}

	if diff := cmp.Diff(originalTree, decodedTree, cmpOpts...); diff != "" {
		t.Errorf("FAIL: The decoded AST does not match the original using the registry codec.\n%s", diff)
	}
}
