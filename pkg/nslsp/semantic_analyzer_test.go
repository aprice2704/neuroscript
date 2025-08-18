// NeuroScript Version: 0.6.0
// File version: 8
// Purpose: Updated tests to use the server.interpreter.ToolRegistry() method.
// filename: pkg/nslsp/semantic_analyzer_test.go
// nlines: 155
// risk_rating: MEDIUM

package nslsp

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/tool"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // Ensure tools are registered
)

func TestSemanticAnalyzer_BuiltInTools(t *testing.T) {
	// --- Setup: Create a server to get a real tool registry ---
	server := NewServer(log.New(io.Discard, "", 0))
	// THE FIX IS HERE: Access the tool registry via the interpreter and its method.
	registry := server.interpreter.ToolRegistry()
	if registry == nil {
		t.Fatal("Failed to get tool registry from new server instance.")
	}

	parserAPI := parser.NewParserAPI(nil)
	isDebug := os.Getenv("DEBUG_LSP_HOVER_TEST") != ""
	analyzer := NewSemanticAnalyzer(registry, NewExternalToolManager(), isDebug)

	t.Logf("[SemanticAnalyzerTest] Tool registry initialized with %d tools for test run.", registry.NTools())

	testCases := []struct {
		name              string
		script            string
		expectedErrorMsgs []string
	}{
		{
			name:              "Valid: Correct tool and argument count (Mixed Case)",
			script:            "func M() means\n  set x = tool.FS.Read(\"/path/to/file\")\nendfunc",
			expectedErrorMsgs: []string{},
		},
		{
			name:              "Invalid: Undefined tool",
			script:            "func M() means\n  set x = tool.nonexistent.Tool()\nendfunc",
			expectedErrorMsgs: []string{"Tool 'tool.nonexistent.Tool' is not defined."},
		},
		{
			name:              "Invalid: Too few arguments for tool.FS.Read",
			script:            "func M() means\n  set x = tool.FS.Read()\nendfunc",
			expectedErrorMsgs: []string{"Expected 1 argument(s) for tool 'tool.FS.Read', but got 0."},
		},
		{
			name:              "Invalid: Too many arguments for tool.FS.Read",
			script:            "func M() means\n  set x = tool.FS.Read(\"a\", \"b\")\nendfunc",
			expectedErrorMsgs: []string{"Expected 1 argument(s) for tool 'tool.FS.Read', but got 2."},
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

			if len(diagnostics) != len(tc.expectedErrorMsgs) {
				t.Fatalf("Expected %d diagnostic(s), but got %d. Diagnostics: %v", len(tc.expectedErrorMsgs), len(diagnostics), diagnostics)
			}

			for i, expectedMsg := range tc.expectedErrorMsgs {
				if !strings.Contains(diagnostics[i].Message, expectedMsg) {
					t.Errorf("Expected diagnostic message to contain '%s', but got '%s'", expectedMsg, diagnostics[i].Message)
				}
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

	analyzer := NewSemanticAnalyzer(nil, externalManager, false)

	// 3. Define test cases
	testCases := []struct {
		name              string
		script            string
		expectedErrorMsgs []string
	}{
		{
			name:              "Valid: Correct external tool and argument count",
			script:            "func M() means\n  set x = tool.fdm.core.SaveMemory(\"handle1\")\nendfunc",
			expectedErrorMsgs: []string{},
		},
		{
			name:              "Invalid: Too few arguments for external tool",
			script:            "func M() means\n  set x = tool.fdm.core.SaveMemory()\nendfunc",
			expectedErrorMsgs: []string{"Expected 1 argument(s) for tool 'tool.fdm.core.SaveMemory', but got 0."},
		},
		{
			name:              "Invalid: Undefined external tool in known group",
			script:            "func M() means\n  set x = tool.fdm.core.DeleteMemory()\nendfunc",
			expectedErrorMsgs: []string{"Tool 'tool.fdm.core.DeleteMemory' is not defined."},
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

			if len(diagnostics) != len(tc.expectedErrorMsgs) {
				t.Fatalf("Expected %d diagnostic(s), but got %d. Diagnostics: %v", len(tc.expectedErrorMsgs), len(diagnostics), diagnostics)
			}
			for i, expectedMsg := range tc.expectedErrorMsgs {
				if !strings.Contains(diagnostics[i].Message, expectedMsg) {
					t.Errorf("Expected diagnostic message to contain '%s', but got '%s'", expectedMsg, diagnostics[i].Message)
				}
			}
		})
	}
}
