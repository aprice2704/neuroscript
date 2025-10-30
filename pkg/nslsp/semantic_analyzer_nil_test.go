// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides a "nil-hardness" test for the semantic analyzer to ensure it does not panic on malformed ASTs (e.g., from syntax errors) or nil inputs.
// filename: pkg/nslsp/semantic_analyzer_nil_test.go
// nlines: 66

package nslsp

import (
	"io"
	"log"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/parser"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all" // Ensure tools are registered
)

func TestSemanticAnalyzer_NilHardness(t *testing.T) {
	// --- Setup ---
	server := NewServer(log.New(io.Discard, "", 0))
	registry := server.interpreter.ToolRegistry()
	parserAPI := parser.NewParserAPI(nil)
	symbolManager := NewSymbolManager(log.New(io.Discard, "", 0))
	analyzer := NewSemanticAnalyzer(registry, NewExternalToolManager(), symbolManager, false)

	testCases := []struct {
		name   string
		script string
	}{
		{
			name:   "Malformed func with no name (panic vector)",
			script: "func () means\n  set x = 1\nendfunc",
		},
		{
			name:   "Just the func keyword (panic vector)",
			script: "func",
		},
		{
			name:   "Empty script",
			script: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We still parse the script to get a (potentially malformed) tree
			tree, _ := parserAPI.ParseForLSP("test.ns", tc.script)

			// The actual test: Call Analyze and ensure it does not panic.
			// We wrap this in a defer/recover in case the test fails,
			// so we can report it gracefully.
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("PANIC: SemanticAnalyzer.Analyze paniced with: %v", r)
				}
			}()

			// This is the call that was panicking.
			diagnostics := analyzer.Analyze(tree)

			// If we got here, it didn't panic.
			t.Logf("SUCCESS: Analyze() did not panic. (Diagnostics returned: %d)", len(diagnostics))
		})
	}

	// Test nil tree input directly
	t.Run("Nil tree input", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("PANIC: SemanticAnalyzer.Analyze paniced with nil tree: %v", r)
			}
		}()
		diagnostics := analyzer.Analyze(nil)
		if diagnostics != nil {
			t.Errorf("Expected nil diagnostics for a nil tree, but got %d diagnostics", len(diagnostics))
		}
		t.Log("SUCCESS: Analyze(nil) did not panic and returned nil.")
	})
}
