// NeuroScript Version: 0.7.0
// File version: 10
// Purpose: FIX: Disabled the workspace-aware symbol manager for this test by passing nil, as it's designed to validate grammar in isolation.
// filename: pkg/nslsp/grammar_test.go
// nlines: 83
// risk_rating: LOW

package nslsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/testutil"
)

// TestGrammarFiles_NoError validates that a suite of known-good .ns files
// parse and analyze with zero syntax or semantic errors.
func TestGrammarFiles_NoError(t *testing.T) {
	// --- Setup ---
	interp := testutil.NewTestInterpreterWithAllTools(t)
	registry := interp.ToolRegistry()
	parserAPI := parser.NewParserAPI(nil)

	// ** THE FIX IS HERE: Pass a nil SymbolManager to disable workspace scanning for this test. **
	analyzer := NewSemanticAnalyzer(registry, NewExternalToolManager(), nil, false)

	t.Logf("[GrammarTest] Tool registry initialized with %d tools for test run.", registry.NTools())

	// --- Find all .ns files in the sibling antlr directory ---
	grammarDir := "../antlr"
	files, err := os.ReadDir(grammarDir)
	if err != nil {
		t.Fatalf("Could not read grammar directory '%s': %v", grammarDir, err)
	}

	foundTestFile := false
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".ns") {
			continue
		}

		foundTestFile = true
		filePath := filepath.Join(grammarDir, file.Name())

		t.Run(file.Name(), func(t *testing.T) {
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read test file '%s': %v", filePath, err)
			}

			// --- 1. Check for Syntax Errors ---
			tree, syntaxErrors := parserAPI.ParseForLSP(filePath, string(content))
			if len(syntaxErrors) > 0 {
				t.Errorf("FAIL: Expected 0 syntax errors, but got %d:", len(syntaxErrors))
				for _, se := range syntaxErrors {
					t.Errorf("  - %s", se.Msg)
				}
			}

			// --- 2. Check for Semantic Errors ---
			if tree != nil {
				diagnostics := analyzer.Analyze(tree)
				if len(diagnostics) > 0 {
					t.Errorf("FAIL: Expected 0 semantic diagnostics, but got %d:", len(diagnostics))
					for _, diag := range diagnostics {
						t.Errorf("  - %s (Source: %s)", diag.Message, diag.Source)
					}
				}
			} else if len(syntaxErrors) == 0 {
				t.Error("FAIL: Parser returned a nil tree but no syntax errors, which is unexpected.")
			}
		})
	}

	if !foundTestFile {
		t.Skipf("Skipping test: No .ns files found in '%s'", grammarDir)
	}
}
