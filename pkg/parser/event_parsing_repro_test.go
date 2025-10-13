// NeuroScript Version: 0.7.2
// File version: 5
// Purpose: Corrected the test script to be syntactically valid by adding required newlines, resolving the parser error.
// filename: pkg/parser/event_parsing_repro_test.go
// nlines: 44
// risk_rating: LOW

package parser_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// TestEventHandlerParsing_Success validates that a properly configured
// ASTBuilder can successfully parse a script containing an 'on event' block
// without panicking.
func TestEventHandlerParsing_Success(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting test to validate correct event handler parsing.")
	// FIX: Added newlines to the script to make it syntactically valid.
	script := `
on event "test.event" do
	emit "hello"
endon
`

	logger := logging.NewTestLogger(t)
	parserAPI := parser.NewParserAPI(logger)
	astBuilder := parser.NewASTBuilder(logger)
	t.Logf("[DEBUG] Turn 2: Parser and Builder initialized.")

	var eventHandlerWasRegistered bool
	astBuilder.SetEventHandlerCallback(func(event *ast.OnEventDecl) {
		eventHandlerWasRegistered = true
		t.Logf("[DEBUG] Turn X: Event handler callback was invoked for event: %s", event.EventNameExpr.String())
	})
	t.Logf("[DEBUG] Turn 3: ASTBuilder configured with event handler callback.")

	tree, _, pErr := parserAPI.ParseAndGetStream("test.ns", script)
	if pErr != nil {
		t.Fatalf("ParserAPI.Parse() failed unexpectedly: %v", pErr)
	}
	t.Logf("[DEBUG] Turn 4: Script parsed.")

	_, _, buildErr := astBuilder.Build(tree)
	if buildErr != nil {
		t.Fatalf("ASTBuilder.Build() failed unexpectedly: %v", buildErr)
	}
	t.Logf("[DEBUG] Turn 5: AST build completed successfully.")

	if !eventHandlerWasRegistered {
		t.Error("FAIL: The event handler was parsed, but the registration callback was not invoked.")
	} else {
		t.Logf("PASS: The builder successfully parsed the event handler and invoked the callback.")
	}
}
