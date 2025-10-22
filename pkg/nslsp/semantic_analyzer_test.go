// NeuroScript Version: 0.7.0
// File version: 11
// Purpose: Updated tests to provide the new SymbolManager dependency when creating a SemanticAnalyzer. FIX: Added tests for new informational diagnostic on missing optional tool arguments.
// filename: pkg/nslsp/semantic_analyzer_test.go
// nlines: 201
// risk_rating: MEDIUM

package nslsp

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/tool"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // Ensure tools are registered
	lsp "github.com/sourcegraph/go-lsp"
)

func TestSemanticAnalyzer_BuiltInTools(t *testing.T) {
	// --- Setup: Create a server to get a real tool registry ---
	server := NewServer(log.New(io.Discard, "", 0))
	registry := server.interpreter.ToolRegistry()
	if registry == nil {
		t.Fatal("Failed to get tool registry from new server instance.")
	}

	parserAPI := parser.NewParserAPI(nil)
	isDebug := os.Getenv("DEBUG_LSP_HOVER_TEST") != ""
	// THE FIX IS HERE
	symbolManager := NewSymbolManager(log.New(io.Discard, "", 0))
	analyzer := NewSemanticAnalyzer(registry, NewExternalToolManager(), symbolManager, isDebug)

	t.Logf("[SemanticAnalyzerTest] Tool registry initialized with %d tools for test run.", registry.NTools())

	testCases := []struct {
		name             string
		script           string
		expectedCode     interface{} // Can be nil if no error is expected
		expectedSeverity lsp.DiagnosticSeverity
	}{
		{
			name:             "Valid: Correct tool and argument count (tool.FS.Read)",
			script:           "func M() means\n  set x = tool.FS.Read(\"/path/to/file\")\nendfunc",
			expectedCode:     nil,
			expectedSeverity: 0, // No diagnostic
		},
		{
			name:             "Invalid: Undefined tool",
			script:           "func M() means\n  set x = tool.nonexistent.Tool()\nendfunc",
			expectedCode:     string(DiagCodeToolNotFound),
			expectedSeverity: lsp.Error,
		},
		{
			name:             "Invalid: Too few arguments for tool.FS.Read",
			script:           "func M() means\n  set x = tool.FS.Read()\nendfunc",
			expectedCode:     string(DiagCodeArgCountMismatch),
			expectedSeverity: lsp.Error,
		},
		{
			name:             "Invalid: Too many arguments for tool.FS.Read",
			script:           "func M() means\n  set x = tool.FS.Read(\"a\", \"b\")\nendfunc",
			expectedCode:     string(DiagCodeArgCountMismatch),
			expectedSeverity: lsp.Error,
		},
		{
			name:             "Valid: Missing optional arg for tool.FS.List",
			script:           "func M() means\n  set x = tool.FS.List(\"/path\")\nendfunc",
			expectedCode:     string(DiagCodeOptionalArgMissing),
			expectedSeverity: lsp.Information,
		},
		{
			name:             "Valid: All args for tool.FS.List",
			script:           "func M() means\n  set x = tool.FS.List(\"/path\", true)\nendfunc",
			expectedCode:     nil,
			expectedSeverity: 0, // No diagnostic
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

			if tc.expectedCode == nil {
				if len(diagnostics) != 0 {
					t.Fatalf("Expected 0 diagnostics, but got %d. Diagnostics: %v", len(diagnostics), diagnostics)
				}
				return
			}

			if len(diagnostics) != 1 {
				t.Fatalf("Expected 1 diagnostic, but got %d. Diagnostics: %v", len(diagnostics), diagnostics)
			}

			if diagnostics[0].Code != tc.expectedCode {
				t.Errorf("Expected diagnostic code '%v', but got '%v'", tc.expectedCode, diagnostics[0].Code)
			}
			if diagnostics[0].Severity != tc.expectedSeverity {
				t.Errorf("Expected diagnostic severity '%v', but got '%v'", tc.expectedSeverity, diagnostics[0].Severity)
			}
		})
	}
}

func TestSemanticAnalyzer_WithExternalTools(t *testing.T) {
	// 1. Setup: Create a temporary directory and the tools.json file.
	tempDir := t.TempDir()
	toolsJSONPath := filepath.Join(tempDir, "external-tools.json")

	externalToolImpls := []tool.ToolImplementation{
		{
			Spec: tool.ToolSpec{
				Name:        "SaveMemory",
				Group:       "fdm.core",
				Description: "Saves an FDM node.",
				Args: []tool.ArgSpec{
					{Name: "node_handle", Type: "string", Required: true},
					{Name: "as_new", Type: "bool", Required: false}, // Added optional arg
				},
				ReturnType: "bool",
			},
			RequiresTrust: false,
		},
	}
	toolsJSONContent, err := json.MarshalIndent(externalToolImpls, "", "  ")
	if err != nil {
		t.Fatalf("Failed to create tools.json content: %v", err)
	}
	if err := os.WriteFile(toolsJSONPath, toolsJSONContent, 0644); err != nil {
		t.Fatalf("Failed to write tools.json: %v", err)
	}

	// 2. Setup the analyzer with the external tool manager.
	parserAPI := parser.NewParserAPI(nil)
	externalManager := NewExternalToolManager()
	externalManager.LoadFromPaths(log.New(io.Discard, "", 0), tempDir, []string{"external-tools.json"})

	// THE FIX IS HERE
	symbolManager := NewSymbolManager(log.New(io.Discard, "", 0))
	analyzer := NewSemanticAnalyzer(nil, externalManager, symbolManager, false)

	// 3. Define test cases
	testCases := []struct {
		name             string
		script           string
		expectedCode     interface{}
		expectedSeverity lsp.DiagnosticSeverity
	}{
		{
			name:             "Valid: External tool with missing optional arg",
			script:           "func M() means\n  set x = tool.fdm.core.SaveMemory(\"handle1\")\nendfunc",
			expectedCode:     string(DiagCodeOptionalArgMissing),
			expectedSeverity: lsp.Information,
		},
		{
			name:             "Valid: External tool with all args",
			script:           "func M() means\n  set x = tool.fdm.core.SaveMemory(\"handle1\", true)\nendfunc",
			expectedCode:     nil,
			expectedSeverity: 0, // No diagnostic
		},
		{
			name:             "Invalid: Too few arguments for external tool",
			script:           "func M() means\n  set x = tool.fdm.core.SaveMemory()\nendfunc",
			expectedCode:     string(DiagCodeArgCountMismatch),
			expectedSeverity: lsp.Error,
		},
		{
			name:             "Invalid: Undefined external tool in known group",
			script:           "func M() means\n  set x = tool.fdm.core.DeleteMemory()\nendfunc",
			expectedCode:     string(DiagCodeToolNotFound),
			expectedSeverity: lsp.Error,
		},
	}

	// 4. Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, _ := parserAPI.ParseForLSP("test.ns", tc.script)
			if tree == nil {
				t.Fatal("Parser returned a nil tree")
			}
			diagnostics := analyzer.Analyze(tree)

			if tc.expectedCode == nil {
				if len(diagnostics) != 0 {
					t.Fatalf("Expected 0 diagnostics, but got %d.", len(diagnostics))
				}
				return
			}
			if len(diagnostics) != 1 {
				t.Fatalf("Expected 1 diagnostic, but got %d.", len(diagnostics))
			}
			if diagnostics[0].Code != tc.expectedCode {
				t.Errorf("Expected diagnostic code '%v', but got '%v'", tc.expectedCode, diagnostics[0].Code)
			}
			if diagnostics[0].Severity != tc.expectedSeverity {
				t.Errorf("Expected diagnostic severity '%v', but got '%v'", tc.expectedSeverity, diagnostics[0].Severity)
			}
		})
	}
}
