// NeuroScript Version: 0.6.2
// File version: 9
// Purpose: FIX: Updates test paths to reflect the new additional_features.ns and additional_command_block.ns files.
// filename: pkg/canon/comprehensive_e2e_test.go
// nlines: 90+
// risk_rating: HIGH

package canon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// runRoundtripComparison is a helper to perform the full parse -> canonicalize -> decode -> compare cycle.
func runRoundtripComparison(t *testing.T, scriptBytes []byte) {
	t.Helper()

	// 1. Parse the script to get the original, known-good AST.
	parserAPI := parser.NewParserAPI(logging.NewNoOpLogger())
	antlrTree, pErr := parserAPI.Parse(string(scriptBytes))
	if pErr != nil {
		t.Fatalf("parser.Parse() failed unexpectedly: %v", pErr)
	}
	builder := parser.NewASTBuilder(logging.NewNoOpLogger())
	program, _, bErr := builder.Build(antlrTree)
	if bErr != nil {
		t.Fatalf("ast.Build() failed unexpectedly: %v", bErr)
	}
	originalTree := &ast.Tree{Root: program}

	// 2. Run the AST through the full Canonicalise/Decode cycle.
	blob, _, err := Canonicalise(originalTree)
	if err != nil {
		t.Fatalf("Canonicalise() failed unexpectedly: %v", err)
	}
	decodedTree, err := Decode(blob)
	if err != nil {
		t.Fatalf("Decode() failed unexpectedly: %v", err)
	}

	// 3. Perform a deep comparison and fail if there are any differences.
	cmpOpts := []cmp.Option{
		cmpopts.IgnoreFields(ast.BaseNode{}, "StartPos", "StopPos"),
		cmpopts.IgnoreUnexported(ast.Procedure{}, ast.Step{}),
		cmpopts.EquateEmpty(),
		cmpopts.SortSlices(func(a, b *ast.MapEntryNode) bool { return a.Key.Value < b.Key.Value }),
	}

	if diff := cmp.Diff(originalTree, decodedTree, cmpOpts...); diff != "" {
		t.Errorf("FAIL: The decoded AST does not match the original. The following fields were lost or altered during canonicalization:\n%s", diff)
	}
}

// TestComprehensiveGrammarRoundtrip is the ultimate regression test for the canonicalization process.
func TestComprehensiveGrammarRoundtrip(t *testing.T) {
	t.Run("Library Script", func(t *testing.T) {
		scriptPath := filepath.Join("..", "antlr", "comprehensive_grammar.ns")
		src, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatalf("Failed to read comprehensive_grammar.ns: %v", err)
		}
		runRoundtripComparison(t, src)
	})

	t.Run("Command Script", func(t *testing.T) {
		scriptPath := filepath.Join("..", "antlr", "command_block.ns")
		src, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatalf("Failed to read command_block.ns: %v", err)
		}
		runRoundtripComparison(t, src)
	})

	t.Run("Additional Features Library Script", func(t *testing.T) {
		scriptPath := filepath.Join("..", "antlr", "additional_features.ns")
		src, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatalf("Failed to read additional_features.ns: %v", err)
		}
		runRoundtripComparison(t, src)
	})

	t.Run("Additional Features Command Script", func(t *testing.T) {
		scriptPath := filepath.Join("..", "antlr", "additional_command_block.ns")
		src, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatalf("Failed to read additional_command_block.ns: %v", err)
		}
		runRoundtripComparison(t, src)
	})
}
