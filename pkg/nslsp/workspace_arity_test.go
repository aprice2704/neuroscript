// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Provides tests for workspace-aware arity checking of procedure calls in both `set` and `call` statements. FIX: Changed call to use synchronous ScanDirectory method and removed sleep.
// filename: pkg/nslsp/workspace_arity_test.go
// nlines: 97
// risk_rating: HIGH

package nslsp

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	lsp "github.com/sourcegraph/go-lsp"
)

func TestWorkspace_ProcedureArityDiagnostics(t *testing.T) {
	// 1. --- Setup a temporary workspace directory ---
	workspaceDir := t.TempDir()

	libFileContent := `
func ProcWithTwoArgs needs a, b means
  set x = 1
endfunc
`
	libFilePath := filepath.Join(workspaceDir, "lib.ns")
	if err := os.WriteFile(libFilePath, []byte(libFileContent), 0644); err != nil {
		t.Fatalf("Failed to write lib file: %v", err)
	}

	// 2. --- Scan the workspace ---
	logger := log.New(io.Discard, "", 0)
	symbolManager := NewSymbolManager(logger)
	// THE FIX IS HERE: Call the new synchronous method.
	symbolManager.ScanDirectory(workspaceDir)

	// 3. --- Define test cases ---
	testCases := []struct {
		name              string
		script            string
		expectedErrCount  int
		expectArgMismatch bool
	}{
		{"Correct arity in call", "func M() means\n call ProcWithTwoArgs(1, 2)\nendfunc", 0, false},
		{"Too few args in call", "func M() means\n call ProcWithTwoArgs(1)\nendfunc", 1, true},
		{"Too many args in call", "func M() means\n call ProcWithTwoArgs(1, 2, 3)\nendfunc", 1, true},
		{"Correct arity in set", "func M() means\n set x = ProcWithTwoArgs(1, 2)\nendfunc", 0, false},
		{"Too few args in set", "func M() means\n set x = ProcWithTwoArgs(1)\nendfunc", 1, true},
		{"Too many args in set", "func M() means\n set x = ProcWithTwoArgs(1, 2, 3)\nendfunc", 1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parserAPI := parser.NewParserAPI(nil)
			analyzer := NewSemanticAnalyzer(nil, nil, symbolManager, false)

			tree, syntaxErrors := parserAPI.ParseForLSP("test.ns", tc.script)
			if len(syntaxErrors) > 0 {
				t.Fatalf("Test script has unexpected syntax errors: %v", syntaxErrors)
			}

			diagnostics := analyzer.Analyze(tree)

			// Filter out the "defined in another file" info messages
			var actualErrors []lsp.Diagnostic
			for _, d := range diagnostics {
				if d.Severity != lsp.Information {
					actualErrors = append(actualErrors, d)
				}
			}

			if len(actualErrors) != tc.expectedErrCount {
				t.Fatalf("Expected %d error(s), but got %d. Diagnostics: %+v", tc.expectedErrCount, len(actualErrors), actualErrors)
			}

			if tc.expectArgMismatch {
				if len(actualErrors) > 0 && actualErrors[0].Code != string(DiagCodeArgCountMismatch) {
					t.Errorf("Expected an ArgCountMismatch error code, but got %s", actualErrors[0].Code)
				}
			}
		})
	}
}
