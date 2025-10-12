// NeuroScript Version: 0.7.2
// File version: 5
// Purpose: Refactored to use the centralized NewTestHarness, which correctly initializes all components and prevents parser panics.
// filename: pkg/interpreter/append_test.go
// nlines: 95
// risk_rating: LOW

package interpreter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestInterpreter_AppendScript(t *testing.T) {
	script1 := `
		func proc_one() means
			return 1
		endfunc

		on event "event_one" do
			emit "one"
		endfunc
	`
	script2 := `
		func proc_two() means
			return 2
		endfunc

		on event "event_two" do
			emit "two"
		endfunc
	`
	// This script has a conflicting procedure name.
	script3_conflict := `
		func proc_one() means
			return 3
		endfunc
	`

	t.Run("Successfully appends new definitions", func(t *testing.T) {
		h := NewTestHarness(t)
		var capturedEmits []string

		// The harness's HostContext is already configured, so we just add our EmitFunc.
		h.HostContext.EmitFunc = func(v lang.Value) {
			h.T.Logf("[DEBUG] EmitFunc captured: %s", v.String())
			capturedEmits = append(capturedEmits, v.String())
		}
		h.T.Logf("[DEBUG] Test harness created and EmitFunc configured.")

		// Load the first script
		tree1, _ := h.Parser.Parse(script1)
		program1, _, _ := h.ASTBuilder.Build(tree1)
		if err := h.Interpreter.Load(&interfaces.Tree{Root: program1}); err != nil {
			t.Fatalf("Initial load failed: %v", err)
		}
		h.T.Logf("[DEBUG] Loaded script 1.")

		// Append the second script
		tree2, _ := h.Parser.Parse(script2)
		program2, _, _ := h.ASTBuilder.Build(tree2)
		if err := h.Interpreter.AppendScript(&interfaces.Tree{Root: program2}); err != nil {
			t.Fatalf("AppendScript failed unexpectedly: %v", err)
		}
		h.T.Logf("[DEBUG] Appended script 2.")

		// Verify that all definitions are present
		if len(h.Interpreter.KnownProcedures()) != 2 {
			t.Errorf("Expected 2 procedures, got %d", len(h.Interpreter.KnownProcedures()))
		}

		h.T.Logf("[DEBUG] Emitting events to test handlers.")
		h.Interpreter.EmitEvent("event_one", "test", nil)
		h.Interpreter.EmitEvent("event_two", "test", nil)

		output := strings.Join(capturedEmits, "|")
		h.T.Logf("[DEBUG] Final captured output: %s", output)

		if !strings.Contains(output, "one") {
			t.Error("Event handler for 'event_one' did not fire.")
		}
		if !strings.Contains(output, "two") {
			t.Error("Event handler for 'event_two' did not fire.")
		}
	})

	t.Run("Fails on duplicate procedure definition", func(t *testing.T) {
		h := NewTestHarness(t)
		h.T.Logf("[DEBUG] Test harness created for duplicate definition test.")

		// Load the first script
		tree1, _ := h.Parser.Parse(script1)
		program1, _, _ := h.ASTBuilder.Build(tree1)
		if err := h.Interpreter.Load(&interfaces.Tree{Root: program1}); err != nil {
			t.Fatalf("Initial load failed: %v", err)
		}

		// Append the conflicting script
		tree3, _ := h.Parser.Parse(script3_conflict)
		program3, _, _ := h.ASTBuilder.Build(tree3)
		err := h.Interpreter.AppendScript(&interfaces.Tree{Root: program3})
		h.T.Logf("[DEBUG] Appended conflicting script; expecting error.")

		if err == nil {
			t.Fatal("AppendScript should have failed due to duplicate procedure, but it succeeded.")
		}
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodeDuplicate {
			t.Errorf("Expected a Duplicate error, but got: %v", err)
		}
		h.T.Logf("[DEBUG] Correctly received expected error: %v", err)
	})
}
