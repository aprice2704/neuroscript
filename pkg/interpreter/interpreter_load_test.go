// NeuroScript Version: 0.5.2
// File version: 1.0.0
// Purpose: Adds dedicated tests for the interpreter's Load method, ensuring correct state initialization and handling of edge cases.
// filename: pkg/interpreter/interpreter_load_test.go
// nlines: 95
// risk_rating: LOW

package interpreter

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestInterpreterLoad(t *testing.T) {
	t.Run("Successful Load", func(t *testing.T) {
		script := `
			:: name: Test Event
			on event "my_event" do
				emit "hello"
			endon

			func my_proc() means
				return 1
			endfunc
		`
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		parserAPI := parser.NewParserAPI(interp.GetLogger())
		p, pErr := parserAPI.Parse(script)
		if pErr != nil {
			t.Fatalf("Failed to parse script: %v", pErr)
		}
		program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
		if bErr != nil {
			t.Fatalf("Failed to build AST: %v", bErr)
		}

		err := interp.Load(program)
		if err != nil {
			t.Fatalf("Load() returned an unexpected error: %v", err)
		}

		if len(interp.KnownProcedures()) != 1 {
			t.Errorf("Expected 1 procedure to be loaded, got %d", len(interp.KnownProcedures()))
		}
		if _, ok := interp.KnownProcedures()["my_proc"]; !ok {
			t.Error("Procedure 'my_proc' was not found after loading")
		}

		if len(interp.eventManager.eventHandlers) != 1 {
			t.Errorf("Expected 1 event handler to be loaded, got %d", len(interp.eventManager.eventHandlers))
		}
		if _, ok := interp.eventManager.eventHandlers["my_event"]; !ok {
			t.Error("Event handler for 'my_event' was not found after loading")
		}
	})

	t.Run("Reloading a Program Clears Old State", func(t *testing.T) {
		interp, _ := newLocalTestInterpreter(t, nil, nil)

		// Load initial program
		program1 := &ast.Program{
			Procedures: map[string]*ast.Procedure{"proc1": {}},
			Events: []*ast.OnEventDecl{{
				EventNameExpr: &ast.StringLiteralNode{Value: "event1"},
				Body:          []ast.Step{},
			}},
		}
		if err := interp.Load(program1); err != nil {
			t.Fatalf("Initial load failed: %v", err)
		}

		// Verify initial state
		if len(interp.KnownProcedures()) != 1 || len(interp.eventManager.eventHandlers) != 1 {
			t.Fatal("Initial program state was not loaded correctly")
		}

		// Load a new, different program
		program2 := &ast.Program{
			Procedures: map[string]*ast.Procedure{"proc2": {}, "proc3": {}},
			Events:     []*ast.OnEventDecl{},
		}
		if err := interp.Load(program2); err != nil {
			t.Fatalf("Reload failed: %v", err)
		}

		// Verify the old state is gone and the new state is correct
		if len(interp.KnownProcedures()) != 2 {
			t.Errorf("Expected 2 procedures after reload, got %d", len(interp.KnownProcedures()))
		}
		if _, ok := interp.KnownProcedures()["proc1"]; ok {
			t.Error("Procedure 'proc1' from old program was not cleared")
		}
		if len(interp.eventManager.eventHandlers) != 0 {
			t.Errorf("Expected 0 event handlers after reload, got %d", len(interp.eventManager.eventHandlers))
		}
	})

	t.Run("Load Nil Program", func(t *testing.T) {
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		// This should not panic
		err := interp.Load(nil)
		if err != nil {
			t.Fatalf("Load(nil) returned an error: %v", err)
		}
	})

	t.Run("Load Empty Program", func(t *testing.T) {
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		program := &ast.Program{
			Procedures: make(map[string]*ast.Procedure),
			Events:     []*ast.OnEventDecl{},
		}
		err := interp.Load(program)
		if err != nil {
			t.Fatalf("Load(empty) returned an error: %v", err)
		}
		if len(interp.KnownProcedures()) != 0 {
			t.Error("Loaded procedures should be empty")
		}
		if len(interp.eventManager.eventHandlers) != 0 {
			t.Error("Loaded event handlers should be empty")
		}
	})
}
