// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Adds a targeted regression test to prevent panics when parsing event handlers that use 'endfunc'.
// filename: pkg/parser/ast_builder_events_test.go
// nlines: 48
// risk_rating: LOW

package parser_test

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// TestASTBuilder_Regression_EventHandlerWithEndfunc is a regression test specifically
// to prevent a recurrence of a bug where the AST builder would panic when parsing
// an 'on event' block that was terminated with 'endfunc' instead of 'endon'.
//
// The bug was caused by the listener code incorrectly assuming the terminating
// token could only be accessed via a `KW_ENDON()` method on the ANTLR context,
// which would be nil when `endfunc` was used, leading to a nil pointer dereference.
//
// This test ensures that the parsing and AST building process completes without
// panicking for this specific syntax.
func TestASTBuilder_Regression_EventHandlerWithEndfunc(t *testing.T) {
	// The script snippet that specifically caused the panic.
	// Note the use of 'endfunc' instead of 'endon'.
	script := `
		on event "test" do
			emit "bug suck"
		endfunc
	`

	// We create a minimal parser and AST builder. We don't need a full
	// interpreter or test harness for this, as the bug was purely in the
	// parser package.
	logger := logging.NewNoOpLogger()
	p := parser.NewParserAPI(logger)
	builder := parser.NewASTBuilder(logger)

	// We use a recover block to gracefully catch a panic. If a panic
	// occurs during the builder.Build() call, the test will fail with a
	// clear error message instead of crashing the test suite.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ASTBuilder.Build() panicked on valid 'endfunc' syntax: %v", r)
		}
	}()

	// Parse the script. This should succeed.
	tree, err := p.Parse(script)
	if err != nil {
		t.Fatalf("Parser.Parse() failed unexpectedly: %v", err)
	}

	// Build the AST. This is the step that was panicking.
	// The test passes if this line executes without panicking.
	_, _, err = builder.Build(tree)
	if err != nil {
		t.Fatalf("ASTBuilder.Build() returned an unexpected error: %v", err)
	}

	// If we reach here, the test is successful.
	fmt.Println("Regression test passed: 'on event...endfunc' was parsed without panic.")
}
