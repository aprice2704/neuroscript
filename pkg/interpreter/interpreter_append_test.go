// NeuroScript Version: 0.7.2
// File version: 2
// Purpose: Corrects test scripts to use 'endfunc' instead of 'endon' for procedure definitions, fixing the parser panic.
// filename: pkg/interpreter/interpreter_append_test.go
// nlines: 100
// risk_rating: LOW

package interpreter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
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
		interp, err := interpreter.NewTestInterpreter(t, nil, nil, false)
		if err != nil {
			t.Fatalf("Failed to create interpreter: %v", err)
		}

		parserAPI := parser.NewParserAPI(nil)

		// Load the first script
		tree1, _ := parserAPI.Parse(script1)
		program1, _, _ := parser.NewASTBuilder(nil).Build(tree1)
		if err := interp.Load(&interfaces.Tree{Root: program1}); err != nil {
			t.Fatalf("Initial load failed: %v", err)
		}

		// Append the second script
		tree2, _ := parserAPI.Parse(script2)
		program2, _, _ := parser.NewASTBuilder(nil).Build(tree2)
		if err := interp.AppendScript(&interfaces.Tree{Root: program2}); err != nil {
			t.Fatalf("AppendScript failed unexpectedly: %v", err)
		}

		// Verify that all definitions are present
		if len(interp.KnownProcedures()) != 2 {
			t.Errorf("Expected 2 procedures, got %d", len(interp.KnownProcedures()))
		}
		if _, ok := interp.KnownProcedures()["proc_one"]; !ok {
			t.Error("proc_one not found after append")
		}
		if _, ok := interp.KnownProcedures()["proc_two"]; !ok {
			t.Error("proc_two not found after append")
		}

		// A bit of a hack to check event handlers since they're not public
		var eventsFound int
		var capturedEmits []string
		interp.SetEmitFunc(func(v lang.Value) {
			capturedEmits = append(capturedEmits, v.String())
		})
		interp.EmitEvent("event_one", "test", nil)
		interp.EmitEvent("event_two", "test", nil)
		if strings.Contains(strings.Join(capturedEmits, ""), "one") {
			eventsFound++
		}
		if strings.Contains(strings.Join(capturedEmits, ""), "two") {
			eventsFound++
		}
		if eventsFound != 2 {
			t.Errorf("Expected 2 event handlers to be active, but only found %d", eventsFound)
		}
	})

	t.Run("Fails on duplicate procedure definition", func(t *testing.T) {
		interp, err := interpreter.NewTestInterpreter(t, nil, nil, false)
		if err != nil {
			t.Fatalf("Failed to create interpreter: %v", err)
		}

		parserAPI := parser.NewParserAPI(nil)

		// Load the first script
		tree1, _ := parserAPI.Parse(script1)
		program1, _, _ := parser.NewASTBuilder(nil).Build(tree1)
		if err := interp.Load(&interfaces.Tree{Root: program1}); err != nil {
			t.Fatalf("Initial load failed: %v", err)
		}

		// Append the conflicting script
		tree3, _ := parserAPI.Parse(script3_conflict)
		program3, _, _ := parser.NewASTBuilder(nil).Build(tree3)
		err = interp.AppendScript(&interfaces.Tree{Root: program3})

		if err == nil {
			t.Fatal("AppendScript should have failed due to duplicate procedure, but it succeeded.")
		}
		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodeDuplicate {
			t.Errorf("Expected a Duplicate error, but got: %v", err)
		}
	})
}
