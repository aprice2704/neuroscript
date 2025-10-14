// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Inlines the logic for testing a parser error to resolve the undefined function call from a different package.
// filename: pkg/parser/ast_builder_events2_test.go
// nlines: 33
// risk_rating: LOW

package parser_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// TestASTBuilder_Regression_EventHandlerWithEndfunc validates that the parser
// correctly identifies the use of 'endfunc' to terminate an 'on event' block
// as a syntax error. The grammar strictly requires 'endon' for these blocks.
func TestASTBuilder_Regression_EventHandlerWithEndfunc(t *testing.T) {
	// The script snippet that is syntactically invalid.
	script := `
		on event "test" do
			emit "bug suck"
		endfunc
	`

	// Inline the logic from testForParserError to fix the undefined function call.
	t.Helper()
	logger := logging.NewNoOpLogger()
	parserAPI := parser.NewParserAPI(logger)
	_, err := parserAPI.Parse(script)
	if err == nil {
		t.Fatalf("Expected a parser error, but parsing succeeded.")
	}
}
