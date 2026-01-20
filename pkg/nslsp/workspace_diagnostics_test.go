// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 8
// :: description: Fixed compilation error (arg mismatch) and added extensive debug logging to diagnose missing Info diagnostic.
// :: latestChange: Corrected NewSemanticAnalyzer args and added %+v logging.
// :: filename: pkg/nslsp/workspace_diagnostics_test.go
// :: serialization: go

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
	// Use a discard logger to avoid noise, or os.Stdout for debugging if needed
	logger := log.New(io.Discard, "[nslsp-test] ", log.LstdFlags)
	symbolManager := NewSymbolManager(logger)

	// Synchronous scan
	symbolManager.ScanDirectory(workspaceDir)

	// Verify symbol presence
	sym, found := symbolManager.GetSymbolInfo("HelperProc")
	if !found {
		t.Fatal("Synchronization failed: HelperProc not found in symbol manager after scan.")
	}
	// Debug: Inspect symbol to ensure it's valid (using %+v to avoid field guessing)
	t.Logf("Debug: Found symbol in manager: %+v", sym)

	// 3. --- Analyze the main file ---
	parserAPI := parser.NewParserAPI(nil)

	// FIX: NewSemanticAnalyzer(registry, externalTools, symbolManager, isDebug)
	// Passing nil for registry/externalTools is valid for this test case.
	// Setting isDebug=true should enable Information diagnostics.
	analyzer := NewSemanticAnalyzer(nil, nil, symbolManager, true)

	tree, syntaxErrors := parserAPI.ParseForLSP(mainFilePath, mainFileContent)
	if len(syntaxErrors) > 0 {
		t.Fatalf("Main file has unexpected syntax errors: %v", syntaxErrors)
	}
	if tree == nil {
		t.Fatal("Parser returned a nil tree for the main file")
	}

	diagnostics := analyzer.Analyze(tree)

	// 4. --- Verify the diagnostics ---
	// Log ALL diagnostics to help debug why we aren't seeing the Info one.
	t.Logf("Received %d diagnostics:", len(diagnostics))
	for i, d := range diagnostics {
		t.Logf("  [%d] Severity: %d | Code: %s | Source: %s | Message: %s", i, d.Severity, d.Code, d.Source, d.Message)
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

	if !foundWarning {
		t.Error("Expected a 'Warning' diagnostic for 'DoesNotExist', but didn't find one.")
	}

	if !foundInfo {
		t.Error("Expected an 'Information' diagnostic for 'HelperProc' (defined in lib), but didn't find one.")
	}
}
