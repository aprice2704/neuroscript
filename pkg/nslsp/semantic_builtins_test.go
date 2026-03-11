// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 5
// :: description: Provides semantic tests for built-in function arity validation and predefined variable recognition.
// :: latestChange: Removed DBG traces after fixing parsing extraction for keyword-based built-ins.
// :: filename: pkg/nslsp/semantic_builtins_test.go
// :: serialization: go

package nslsp

import (
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all"
	lsp "github.com/sourcegraph/go-lsp"
)

func TestSemanticAnalyzer_BuiltInArity(t *testing.T) {
	isDebug := os.Getenv("DEBUG_LSP_HOVER_TEST") != ""

	testCases := []struct {
		name             string
		script           string
		expectedCode     interface{}
		expectedSeverity lsp.DiagnosticSeverity
		expectedCount    int
	}{
		// --- Existing Arity Tests ---
		{
			name:          "Valid: len(list)",
			script:        "func M() means\n  set x = len([1, 2, 3])\nendfunc",
			expectedCount: 0,
		},
		{
			name:             "Invalid: len() with no args",
			script:           "func M() means\n  set x = len()\nendfunc",
			expectedCode:     string(DiagCodeArgCountMismatch),
			expectedSeverity: lsp.Warning,
			expectedCount:    1,
		},
		{
			name:             "Invalid: len(a, b) with too many args",
			script:           "func M() means\n  set x = len(1, 2)\nendfunc",
			expectedCode:     string(DiagCodeArgCountMismatch),
			expectedSeverity: lsp.Warning,
			expectedCount:    1,
		},
		{
			name:          "Valid: sin(number)",
			script:        "func M() means\n  set x = sin(1.0)\nendfunc",
			expectedCount: 0,
		},
		{
			name:             "Invalid: sin(a, b)",
			script:           "func M() means\n  set x = sin(1, 2)\nendfunc",
			expectedCode:     string(DiagCodeArgCountMismatch),
			expectedSeverity: lsp.Warning,
			expectedCount:    1,
		},
		// --- Type Check Arity Tests ---
		{
			name:          "Valid: is_string(val)",
			script:        "func M() means\n  set x = is_string(\"test\")\nendfunc",
			expectedCount: 0,
		},
		{
			name:             "Invalid: is_string()",
			script:           "func M() means\n  set x = is_string()\nendfunc",
			expectedCode:     string(DiagCodeArgCountMismatch),
			expectedSeverity: lsp.Warning,
			expectedCount:    1,
		},
		{
			name:          "Valid: is_error(e)",
			script:        "func M() means\n  set x = is_error(nil)\nendfunc",
			expectedCount: 0,
		},
		{
			name:          "Valid: is_fuzzy(f)",
			script:        "func M() means\n  set x = is_fuzzy(1)\nendfunc",
			expectedCount: 0,
		},
		// --- Built-in Keywords vs Identifiers ---
		{
			name:          "Valid: typeof(val)",
			script:        "func M() means\n  set x = typeof(123)\nendfunc",
			expectedCount: 0,
		},
		{
			name:          "Valid: eval(string)",
			script:        "func M() means\n  set x = eval(\"set y = 1\")\nendfunc",
			expectedCount: 0,
		},
		{
			name:          "Valid: char(num)",
			script:        "func M() means\n  set x = char(65)\nendfunc",
			expectedCount: 0,
		},
		{
			name:          "Valid: ord(str)",
			script:        "func M() means\n  set x = ord(\"A\")\nendfunc",
			expectedCount: 0,
		},
		// --- Predefined Variable Recognition ---
		{
			name:          "Valid: Use self in whisper",
			script:        "func M() means\n  whisper self, \"msg\"\nendfunc",
			expectedCount: 0,
		},
		{
			name:          "Valid: Use system_error_message in handler",
			script:        "func M() means\n  on error do\n    emit system_error_message\n  endon\n  fail \"err\"\nendfunc",
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- Setup INSIDE t.Run to ensure complete isolation ---
			server := NewServer(log.New(io.Discard, "", 0))
			registry := server.interpreter.ToolRegistry()
			parserAPI := parser.NewParserAPI(nil)

			// No workspace symbols needed for built-in tests
			analyzer := NewSemanticAnalyzer(registry, NewExternalToolManager(), nil, isDebug)

			// Use unique filename per test to avoid any potential parser caching issues
			safeName := strings.ReplaceAll(tc.name, " ", "_")
			safeName = strings.ReplaceAll(safeName, ":", "")
			fileName := "test_" + safeName + ".ns"

			tree, syntaxErrors := parserAPI.ParseForLSP(fileName, tc.script)
			if len(syntaxErrors) > 0 {
				t.Fatalf("Test script has unexpected syntax errors: %v", syntaxErrors)
			}

			diagnostics := analyzer.Analyze(tree)

			// Filter to just relevant diagnostics (Arity mismatches and Uninitialized vars)
			var relevantDiags []lsp.Diagnostic
			for _, d := range diagnostics {
				if d.Code == string(DiagCodeArgCountMismatch) || d.Code == string(DiagCodeUninitializedVar) {
					relevantDiags = append(relevantDiags, d)
				}
			}

			if len(relevantDiags) != tc.expectedCount {
				t.Fatalf("Expected %d diagnostic(s), but got %d. All Diagnostics: %v", tc.expectedCount, len(relevantDiags), diagnostics)
			}

			if tc.expectedCount > 0 {
				if relevantDiags[0].Code != tc.expectedCode {
					t.Errorf("Expected diagnostic code '%v', but got '%v'", tc.expectedCode, relevantDiags[0].Code)
				}
				if relevantDiags[0].Severity != tc.expectedSeverity {
					t.Errorf("Expected diagnostic severity '%v', but got '%v'", tc.expectedSeverity, relevantDiags[0].Severity)
				}
			}
		})
	}
}
