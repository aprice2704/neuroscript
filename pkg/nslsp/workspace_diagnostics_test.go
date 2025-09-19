// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: FIX: Updated the test to use the renamed GetSymbolInfo method. FIX: Changed call to use synchronous ScanDirectory method and removed sleep.
// filename: pkg/nslsp/workspace_diagnostics_test.go
// nlines: 112
// risk_rating: HIGH

package nslsp

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	lsp "github.com/sourcegraph/go-lsp"
)

func TestWorkspace_UndefinedProcedureDiagnostics(t *testing.T) {
	// 1. --- Setup a temporary workspace directory ---
	workspaceDir := t.TempDir()

	// Create a library file with a helper procedure
	libDir := filepath.Join(workspaceDir, "lib")
	if err := os.Mkdir(libDir, 0755); err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}
	libFileContent := "func HelperProc() means\n  set x = 1\nendfunc"
	libFilePath := filepath.Join(libDir, "helpers.ns")
	if err := os.WriteFile(libFilePath, []byte(libFileContent), 0644); err != nil {
		t.Fatalf("Failed to write lib file: %v", err)
	}

	// Create a main file that calls various procedures
	mainFileContent := `
func LocalProc() means
  set y = 2
endfunc

func Main() means
  call LocalProc()
  call HelperProc()
  call DoesNotExist()
endfunc
`
	mainFilePath := filepath.Join(workspaceDir, "main.ns.txt")
	if err := os.WriteFile(mainFilePath, []byte(mainFileContent), 0644); err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// 2. --- Scan the workspace ---
	logger := log.New(io.Discard, "", 0)
	symbolManager := NewSymbolManager(logger)
	// THE FIX IS HERE: Call the new synchronous method for the whole temp directory.
	symbolManager.ScanDirectory(workspaceDir)

	// Wait for the scan to complete by trying to get a known symbol.
	_, found := symbolManager.GetSymbolInfo("HelperProc")
	if !found {
		t.Fatal("Synchronization failed: HelperProc not found in symbol manager after scan.")
	}

	// 3. --- Analyze the main file ---
	parserAPI := parser.NewParserAPI(nil)
	analyzer := NewSemanticAnalyzer(nil, nil, symbolManager, false)

	tree, syntaxErrors := parserAPI.ParseForLSP(mainFilePath, mainFileContent)
	if len(syntaxErrors) > 0 {
		t.Fatalf("Main file has unexpected syntax errors: %v", syntaxErrors)
	}
	if tree == nil {
		t.Fatal("Parser returned a nil tree for the main file")
	}

	diagnostics := analyzer.Analyze(tree)

	// 4. --- Verify the diagnostics ---
	if len(diagnostics) != 2 {
		t.Fatalf("Expected 2 diagnostics, but got %d. Diagnostics: %+v", len(diagnostics), diagnostics)
	}

	foundInfo := false
	foundWarning := false

	for _, diag := range diagnostics {
		// Check for the "defined elsewhere" info message
		if diag.Severity == lsp.Information && strings.Contains(diag.Message, "HelperProc") {
			foundInfo = true
		}
		// Check for the "not defined" warning message
		if diag.Severity == lsp.Warning && diag.Code == string(DiagCodeProcNotFound) && strings.Contains(diag.Message, "DoesNotExist") {
			foundWarning = true
		}
	}

	if !foundInfo {
		t.Error("Expected an 'Information' diagnostic for the procedure defined in another file, but didn't find one.")
	}

	if !foundWarning {
		t.Error("Expected a 'Warning' diagnostic for the completely undefined procedure, but didn't find one.")
	}
}
