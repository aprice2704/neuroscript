// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 4
// :: description: Provides tests for the uninitialized variable check.
// :: latestChange: Moved test environment setup inside t.Run to prevent state leakage and Heisenfails.
// :: filename: pkg/nslsp/semantic_vars_test.go
// :: serialization: go

package nslsp

import (
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // Ensure tools are registered
	lsp "github.com/sourcegraph/go-lsp"
)

func TestSemanticAnalyzer_UninitializedVars(t *testing.T) {
	isDebug := os.Getenv("DEBUG_LSP_HOVER_TEST") != ""

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
		// --- Predefined Variables ---
		{
			name:             "Valid: Use of 'self'",
			script:           "func M() means\n  whisper self, \"hello\"\nendfunc",
			expectedCode:     nil,
			expectedSeverity: 0,
			expectedCount:    0,
		},
		{
			name:             "Valid: Use of 'system_error_message'",
			script:           "command\n  on error do\n    emit system_error_message\n  endon\n  fail \"oops\"\nendcommand",
			expectedCode:     nil,
			expectedSeverity: 0,
			expectedCount:    0,
		},
		{
			name:             "Valid: Use of 'stdout' and 'stderr'",
			script:           "func M() means\n  set x = stdout\n  set y = stderr\nendfunc",
			expectedCode:     nil,
			expectedSeverity: 0,
			expectedCount:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- Setup INSIDE t.Run to ensure complete isolation ---
			server := NewServer(log.New(io.Discard, "", 0))
			registry := server.interpreter.ToolRegistry()
			parserAPI := parser.NewParserAPI(nil)

			// Note: We pass nil for symbolManager to isolate this to local-file analysis
			analyzer := NewSemanticAnalyzer(registry, NewExternalToolManager(), nil, isDebug)

			safeName := strings.ReplaceAll(tc.name, " ", "_")
			safeName = strings.ReplaceAll(safeName, ":", "")
			fileName := "test_" + safeName + ".ns"

			tree, syntaxErrors := parserAPI.ParseForLSP(fileName, tc.script)
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

			// Filter to just the UninitializedVar diagnostics
			var relevantDiags []lsp.Diagnostic
			for _, d := range diagnostics {
				if d.Code == string(DiagCodeUninitializedVar) {
					relevantDiags = append(relevantDiags, d)
				}
			}

			if len(relevantDiags) != tc.expectedCount {
				t.Fatalf("Expected %d UninitializedVar diagnostic(s), but got %d. All Diagnostics: %v", tc.expectedCount, len(relevantDiags), diagnostics)
			}

			if relevantDiags[0].Code != tc.expectedCode {
				t.Errorf("Expected diagnostic code '%v', but got '%v'", tc.expectedCode, relevantDiags[0].Code)
			}
		})
	}
}
