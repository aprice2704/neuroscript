// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Provides tests for arity checking of procedures with zero or optional arguments. FIX: Changed call to use synchronous ScanDirectory method and removed sleep.
// filename: pkg/nslsp/workspace_optional_args_test.go
// nlines: 107
// risk_rating: HIGH

package nslsp

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	lsp "github.com/sourcegraph/go-lsp"
)

func TestWorkspace_ProcedureOptionalArgs(t *testing.T) {
	// 1. --- Setup a temporary workspace directory ---
	workspaceDir := t.TempDir()

	libFileContent := `
func ProcWithNoArgs means
  set x = 1
endfunc

func ProcWithOptional optional title needs name means
  set y = 2
endfunc
`
	libFilePath := filepath.Join(workspaceDir, "lib_optional.ns")
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
		// Zero-argument function tests
		{"Correct call to zero-arg proc", "func M() means\n call ProcWithNoArgs()\nendfunc", 0, false},
		{"Incorrect call to zero-arg proc", "func M() means\n call ProcWithNoArgs(1)\nendfunc", 1, true},

		// Optional-argument function tests
		{"Correct call with min args", "func M() means\n call ProcWithOptional(\"Bob\")\nendfunc", 0, false},
		{"Correct call with max args", "func M() means\n call ProcWithOptional(\"Dr.\", \"Bob\")\nendfunc", 0, false},
		{"Too few args for optional proc", "func M() means\n call ProcWithOptional()\nendfunc", 1, true},
		{"Too many args for optional proc", "func M() means\n call ProcWithOptional(\"a\", \"b\", \"c\")\nendfunc", 1, true},
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

			var actualErrors []lsp.Diagnostic
			for _, d := range diagnostics {
				if d.Severity != lsp.Information {
					actualErrors = append(actualErrors, d)
				}
			}

			if len(actualErrors) != tc.expectedErrCount {
				prettyDiags := ""
				for _, d := range actualErrors {
					prettyDiags += fmt.Sprintf("\n - %s (Code: %s)", d.Message, d.Code)
				}
				t.Fatalf("Expected %d error(s), but got %d.%s", tc.expectedErrCount, len(actualErrors), prettyDiags)
			}

			if tc.expectArgMismatch {
				if len(actualErrors) > 0 && actualErrors[0].Code != string(DiagCodeArgCountMismatch) {
					t.Errorf("Expected an ArgCountMismatch error code, but got %s", actualErrors[0].Code)
				}
			}
		})
	}
}
