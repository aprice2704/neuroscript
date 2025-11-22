// NeuroScript/FDM Major Version: 1
// File version: 11
// Purpose: FIX: Disabled the workspace-aware symbol manager for this test by passing nil. FIX: Filter out Information-level diagnostics (like missing optional args) which are not failures.
// filename: pkg/nslsp/grammar_test.go
// nlines: 95

package nslsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/testutil"
	lsp "github.com/sourcegraph/go-lsp"
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

				// THE FIX IS HERE: Filter out Information diagnostics (like missing optional args)
				var errorDiagnostics []lsp.Diagnostic
				for _, d := range diagnostics {
					if d.Severity == lsp.Error || d.Severity == lsp.Warning {
						errorDiagnostics = append(errorDiagnostics, d)
					}
				}

				if len(errorDiagnostics) > 0 {
					t.Errorf("FAIL: Expected 0 semantic errors/warnings, but got %d:", len(errorDiagnostics))
					for _, diag := range errorDiagnostics {
						t.Errorf("  - %s (Severity: %d, Source: %s)", diag.Message, diag.Severity, diag.Source)
					}
				} else {
					// Optional: Log info diagnostics just for visibility
					for _, d := range diagnostics {
						if d.Severity == lsp.Information {
							t.Logf("INFO: Diagnostic ignored: %s", d.Message)
						}
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
