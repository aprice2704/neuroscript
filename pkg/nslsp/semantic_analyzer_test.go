// NeuroScript Version: 0.3.1
// File version: 3
// Purpose: CRITICAL FIX - Manually register all tools for the test's isolated server instance.
// filename: pkg/nslsp/semantic_analyzer_test.go
// nlines: 90
// risk_rating: LOW

package nslsp

import (
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/tool"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // Ensure tools are registered
)

func TestSemanticAnalyzer(t *testing.T) {
	// --- Setup: Create a real tool registry and parser API instance ---
	registry := tool.NewToolRegistry(nil)
	// ** THE FIX IS HERE: Explicitly register tools for the test **
	if err := tool.RegisterCoreTools(registry); err != nil {
		t.Fatalf("Failed to register core tools: %v", err)
	}
	if err := tool.RegisterExtendedTools(registry); err != nil {
		t.Fatalf("Failed to register extended tools: %v", err)
	}

	parserAPI := parser.NewParserAPI(nil)
	isDebug := os.Getenv("DEBUG_LSP_HOVER_TEST") != ""
	analyzer := NewSemanticAnalyzer(registry, isDebug)

	// Add debug log to confirm tool loading in this specific test
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
		{
			name:              "Valid: tool.Meta.ListTools with zero arguments (Mixed Case)",
			script:            "func M() means\n  set x = tool.Meta.ListTools()\nendfunc",
			expectedErrorMsgs: []string{},
		},
		{
			name:              "Invalid: tool.Meta.ListTools with one argument (Mixed Case)",
			script:            "func M() means\n  set x = tool.Meta.ListTools(1)\nendfunc",
			expectedErrorMsgs: []string{"Expected 0 argument(s) for tool 'tool.Meta.ListTools', but got 1."},
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
