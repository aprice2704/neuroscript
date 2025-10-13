// NeuroScript Version: 0.8.0
// File version: 16
// Purpose: Corrected invalid 'endfunc' to 'endon' in test script and removed debug logging.
// filename: pkg/interpreter/append_test.go
// nlines: 106
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
		endon
	`
	script2 := `
		func proc_two() means
			return 2
		endfunc

		on event "event_two" do
			emit "two"
		endon
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

		h.HostContext.EmitFunc = func(v lang.Value) {
			capturedEmits = append(capturedEmits, v.String())
		}

		// Load the first script
		tree1, err := h.Parser.Parse(script1)
		if err != nil {
			t.Fatalf("Parser.Parse() for script1 failed: %v", err)
		}
		program1, _, err := h.ASTBuilder.Build(tree1)
		if err != nil {
			t.Fatalf("ASTBuilder.Build() for script1 failed: %v", err)
		}
		if err := h.Interpreter.Load(&interfaces.Tree{Root: program1}); err != nil {
			t.Fatalf("Initial load failed: %v", err)
		}

		// Append the second script
		tree2, err := h.Parser.Parse(script2)
		if err != nil {
			t.Fatalf("Parser.Parse() for script2 failed: %v", err)
		}
		program2, _, err := h.ASTBuilder.Build(tree2)
		if err != nil {
			t.Fatalf("ASTBuilder.Build() for script2 failed: %v", err)
		}
		if err := h.Interpreter.AppendScript(&interfaces.Tree{Root: program2}); err != nil {
			t.Fatalf("AppendScript failed unexpectedly: %v", err)
		}

		// Verify that all definitions are present
		if len(h.Interpreter.KnownProcedures()) != 2 {
			t.Errorf("Expected 2 procedures, got %d", len(h.Interpreter.KnownProcedures()))
		}

		h.Interpreter.EmitEvent("event_one", "test", nil)
		h.Interpreter.EmitEvent("event_two", "test", nil)

		output := strings.Join(capturedEmits, "|")

		if !strings.Contains(output, "one") {
			t.Error("Event handler for 'event_one' did not fire.")
		}
		if !strings.Contains(output, "two") {
			t.Error("Event handler for 'event_two' did not fire.")
		}
	})

	t.Run("Fails on duplicate procedure definition", func(t *testing.T) {
		h := NewTestHarness(t)

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

		if err == nil {
			t.Fatal("AppendScript should have failed due to duplicate procedure, but it succeeded.")
		}
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodeDuplicate {
			t.Errorf("Expected a Duplicate error, but got: %v", err)
		}
	})
}
