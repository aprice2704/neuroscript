// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides a workspace-aware test for the uninitialized variable check, ensuring global procedures and built-in functions are not flagged as errors.
// filename: pkg/nslsp/semantic_vars_workspace_test.go
// nlines: 83

package nslsp

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all"
	lsp "github.com/sourcegraph/go-lsp"
)

func TestSemanticAnalyzer_WorkspaceVars(t *testing.T) {
	// 1. --- Setup a temporary workspace directory ---
	workspaceDir := t.TempDir()

	// Create a library file with a helper procedure
	libFileContent := "func GlobalHelper() means\n  set x = 1\nendfunc"
	libFilePath := filepath.Join(workspaceDir, "lib.ns")
	if err := os.WriteFile(libFilePath, []byte(libFileContent), 0644); err != nil {
		t.Fatalf("Failed to write lib file: %v", err)
	}

	// 2. --- Scan the workspace ---
	logger := log.New(io.Discard, "", 0)
	symbolManager := NewSymbolManager(logger)
	symbolManager.ScanDirectory(workspaceDir)

	// 3. --- Setup Analyzer ---
	server := NewServer(log.New(io.Discard, "", 0))
	registry := server.interpreter.ToolRegistry()
	parserAPI := parser.NewParserAPI(nil)
	analyzer := NewSemanticAnalyzer(registry, NewExternalToolManager(), symbolManager, false)

	// 4. --- Define script to be analyzed ---
	mainFileContent := `
func Main() means
  # This one is truly uninitialized
  set a = my_uninitialized_var
  
  # This is a built-in function
  set b = sin(1)
  
  # This is a procedure defined in lib.ns
  set c = GlobalHelper
endfunc
`

	// 5. --- Analyze the main file ---
	tree, syntaxErrors := parserAPI.ParseForLSP("main.ns", mainFileContent)
	if len(syntaxErrors) > 0 {
		t.Fatalf("Main file has unexpected syntax errors: %v", syntaxErrors)
	}
	if tree == nil {
		t.Fatal("Parser returned a nil tree for the main file")
	}

	diagnostics := analyzer.Analyze(tree)

	// 6. --- Verify the diagnostics ---
	// We filter out the "defined elsewhere" info message for GlobalHelper
	var warningDiagnostics []lsp.Diagnostic
	for _, d := range diagnostics {
		if d.Severity == lsp.Warning {
			warningDiagnostics = append(warningDiagnostics, d)
		}
	}

	if len(warningDiagnostics) != 1 {
		var msgs []string
		for _, d := range diagnostics {
			msgs = append(msgs, d.Message)
		}
		t.Fatalf("Expected 1 warning diagnostic, but got %d. Diagnostics: %v", len(warningDiagnostics), msgs)
	}

	diag := warningDiagnostics[0]
	if diag.Code != string(DiagCodeUninitializedVar) {
		t.Errorf("Expected diagnostic code to be UninitializedVar, got %s", diag.Code)
	}
	if !strings.Contains(diag.Message, "my_uninitialized_var") {
		t.Errorf("Expected diagnostic message to be about 'my_uninitialized_var', got: %s", diag.Message)
	}
}
