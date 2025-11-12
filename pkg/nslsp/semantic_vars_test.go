// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Provides tests for the uninitialized variable check. FIX: Removed unused symbolManager variable declaration.
// filename: pkg/nslsp/semantic_vars_test.go
// nlines: 109

package nslsp

import (
	"io"
	"log"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // Ensure tools are registered
	lsp "github.com/sourcegraph/go-lsp"
)

func TestSemanticAnalyzer_UninitializedVars(t *testing.T) {
	// --- Setup ---
	server := NewServer(log.New(io.Discard, "", 0))
	registry := server.interpreter.ToolRegistry()
	parserAPI := parser.NewParserAPI(nil)
	isDebug := os.Getenv("DEBUG_LSP_HOVER_TEST") != ""
	// FIX: Removed unused 'symbolManager' variable. We pass nil intentionally.
	// Note: We pass nil for symbolManager to isolate this to local-file analysis
	analyzer := NewSemanticAnalyzer(registry, NewExternalToolManager(), nil, isDebug)

	testCases := []struct {
		name             string
		script           string
		expectedCode     interface{} // Can be nil if no error is expected
		expectedSeverity lsp.DiagnosticSeverity
		expectedCount    int
	}{
		{
			name:             "Invalid: Use of uninitialized variable",
			script:           "func M() means\n  set x = my_var + 1\nendfunc",
			expectedCode:     string(DiagCodeUninitializedVar),
			expectedSeverity: lsp.Warning,
			expectedCount:    1,
		},
		{
			name:             "Valid: Variable is initialized",
			script:           "func M() means\n  set my_var = 10\n  set x = my_var + 1\nendfunc",
			expectedCode:     nil,
			expectedSeverity: 0,
			expectedCount:    0,
		},
		{
			name:             "Valid: Variable from 'needs' param",
			script:           "func M(needs p1) means\n  set x = p1\nendfunc",
			expectedCode:     nil,
			expectedSeverity: 0,
			expectedCount:    0,
		},
		{
			name:             "Valid: Variable from 'optional' param",
			script:           "func M(optional p1) means\n  set x = p1\nendfunc",
			expectedCode:     nil,
			expectedSeverity: 0,
			expectedCount:    0,
		},
		{
			name:             "Valid: Variable from 'for' loop",
			script:           "func M() means\n  for each item in [1, 2]\n    set x = item\n  endfor\nendfunc",
			expectedCode:     nil,
			expectedSeverity: 0,
			expectedCount:    0,
		},
		{
			name:             "Valid: Variable from 'ask...into'",
			script:           "func M() means\n  ask \"model\", \"p\" into res\n  set x = res\nendfunc",
			expectedCode:     nil,
			expectedSeverity: 0,
			expectedCount:    0,
		},
		{
			name:             "Valid: Variable from 'promptuser...into'",
			script:           "func M() means\n  promptuser \"p\" into res\n  set x = res\nendfunc",
			expectedCode:     nil,
			expectedSeverity: 0,
			expectedCount:    0,
		},
		{
			name:             "Invalid: Use of variable before 'for' loop init",
			script:           "func M() means\n  set x = item\n  for each item in [1, 2]\n    set y = item\n  endfor\nendfunc",
			expectedCode:     string(DiagCodeUninitializedVar),
			expectedSeverity: lsp.Warning,
			expectedCount:    1,
		},
		{
			name:             "Valid: Event handler 'as' variable",
			script:           "on event \"foo\" as evt do\n  set x = evt\nendon",
			expectedCode:     nil,
			expectedSeverity: 0,
			expectedCount:    0,
		},
		{
			name:             "Invalid: Event handler uninitialized",
			script:           "on event \"foo\" as evt do\n  set x = my_var\nendon",
			expectedCode:     string(DiagCodeUninitializedVar),
			expectedSeverity: lsp.Warning,
			expectedCount:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, syntaxErrors := parserAPI.ParseForLSP("test.ns", tc.script)
			if len(syntaxErrors) > 0 {
				t.Fatalf("Test script has unexpected syntax errors: %v", syntaxErrors)
			}
			if tree == nil {
				t.Fatal("Parser returned a nil tree without errors")
			}

			diagnostics := analyzer.Analyze(tree)

			if tc.expectedCount == 0 {
				if len(diagnostics) != 0 {
					t.Fatalf("Expected 0 diagnostics, but got %d. Diagnostics: %v", len(diagnostics), diagnostics)
				}
				return
			}

			if len(diagnostics) != tc.expectedCount {
				t.Fatalf("Expected %d diagnostic(s), but got %d. Diagnostics: %v", tc.expectedCount, len(diagnostics), diagnostics)
			}

			if diagnostics[0].Code != tc.expectedCode {
				t.Errorf("Expected diagnostic code '%v', but got '%v'", tc.expectedCode, diagnostics[0].Code)
			}
		})
	}
}
