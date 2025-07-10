// filename: pkg/parser/ast_builder_events_test.go
// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Corrected invalid syntax in tests and updated assertions to match expected behavior.
// nlines: 105
// risk_rating: MEDIUM

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestOnEventParsing(t *testing.T) {
	t.Run("Valid On-Event Handler", func(t *testing.T) {
		script := `
			:: title: Test Event Handler
			on event tool.testing.MyEvent("filter_value") as evt do
				emit "Event received: " + evt
			endon
		`
		prog := testParseAndBuild(t, script)
		if len(prog.Events) != 1 {
			t.Fatalf("Expected 1 OnEventDecl, got %d", len(prog.Events))
		}
		eventDecl := prog.Events[0]
		if _, ok := eventDecl.EventNameExpr.(*ast.CallableExprNode); !ok {
			t.Errorf("Expected EventNameExpr to be a CallableExprNode, but got %T", eventDecl.EventNameExpr)
		}
		if eventDecl.EventVarName != "evt" {
			t.Errorf("Expected EventVarName to be 'evt', got '%s'", eventDecl.EventVarName)
		}
		if len(eventDecl.Body) != 1 {
			t.Errorf("Expected 1 step in the event handler body, got %d", len(eventDecl.Body))
		}
	})

	t.Run("Valid On-Event Handler with HandlerName", func(t *testing.T) {
		script := `
			on event tool.testing.MyEvent() named "MyEventHandler" do
				emit "Event received"
			endon
		`
		prog := testParseAndBuild(t, script)
		if len(prog.Events) != 1 {
			t.Fatalf("Expected 1 OnEventDecl, got %d", len(prog.Events))
		}
		eventDecl := prog.Events[0]
		if eventDecl.HandlerName != "MyEventHandler" {
			t.Errorf("Expected HandlerName to be 'MyEventHandler', got '%s'", eventDecl.HandlerName)
		}
	})
}

func TestOnErrorParsing(t *testing.T) {
	t.Run("On-Error handler inside a function", func(t *testing.T) {
		script := `
			func MyFunc() means
				on error do
					emit "An error occurred"
					clear_error
				endon

				fail "trigger error"
			endfunc
		`
		prog := testParseAndBuild(t, script)
		proc, ok := prog.Procedures["MyFunc"]
		if !ok {
			t.Fatal("Procedure 'MyFunc' not found in AST")
		}

		if len(proc.ErrorHandlers) != 1 {
			t.Fatalf("Expected 1 error handler in the procedure, got %d", len(proc.ErrorHandlers))
		}
		handler := proc.ErrorHandlers[0]
		if handler.Type != "on_error" {
			t.Errorf("Expected handler type to be 'on_error', got '%s'", handler.Type)
		}
		if len(handler.Body) != 2 {
			t.Errorf("Expected 2 steps in the error handler body, got %d", len(handler.Body))
		}
	})

	t.Run("Top-level On-Error is a parser error", func(t *testing.T) {
		script := `
			on error do
				emit "This should not parse"
			endon
		`
		testForParserError(t, script)
	})
}
