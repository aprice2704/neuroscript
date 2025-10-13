// NeuroScript Version: 0.8.0
// File version: 8.0.0
// Purpose: Rewrote tests to assert that command blocks are correctly sandboxed and do not leak variables, per new design rules.
// filename: pkg/interpreter/scoping_extended_test.go
// nlines: 100
// risk_rating: LOW

package interpreter_test

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestVariableScopingExtended(t *testing.T) {
	t.Run("If Block Does Not Leak Variables", func(t *testing.T) {
		h := NewTestHarness(t)
		interp := h.Interpreter
		// Set a variable in the parent scope that will be shadowed.
		interp.SetVariable("x", lang.NumberValue{Value: 10})

		script := `
			command
				# This 'x' should shadow the parent's 'x'.
				set x = 20
				# This 'y' should only exist within this command block.
				set y = 5
			endcommand
		`
		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		interp.Load(&interfaces.Tree{Root: program})
		result, err := interp.Execute(program)
		if err != nil {
			t.Fatalf("script execution failed: %v", err)
		}

		// Check for success code
		if code, _ := result.(lang.NumberValue); code.Value != 0 {
			t.Errorf("Expected command to return 0 for success, but got %v", result)
		}

		// Assert that the parent scope was NOT modified.
		finalX, _ := interp.GetVariable("x")
		expectedX := lang.NumberValue{Value: 10}
		if !reflect.DeepEqual(finalX, expectedX) {
			t.Errorf("Command block modified parent variable 'x'. Expected %#v, got %#v", expectedX, finalX)
		}

		_, yExists := interp.GetVariable("y")
		if yExists {
			t.Error("Variable 'y' from command block leaked into parent scope.")
		}
	})

	t.Run("For Loop Variable Does Not Leak", func(t *testing.T) {
		h := NewTestHarness(t)
		interp := h.Interpreter
		interp.SetVariable("my_list", lang.NewListValue([]lang.Value{
			lang.NumberValue{Value: 1},
			lang.NumberValue{Value: 2},
		}))

		script := `
			command
				for each loop_var in my_list
					set z = loop_var
				endfor
			endcommand
		`
		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		interp.Load(&interfaces.Tree{Root: program})
		_, err := interp.Execute(program)
		if err != nil {
			t.Fatalf("script execution failed: %v", err)
		}

		// Assert that loop variables did NOT leak into the parent scope.
		_, loopVarExists := interp.GetVariable("loop_var")
		if loopVarExists {
			t.Error("Loop variable 'loop_var' from command block leaked into parent scope.")
		}
		_, zExists := interp.GetVariable("z")
		if zExists {
			t.Error("Variable 'z' from command block leaked into parent scope.")
		}
	})

	t.Run("Procedure Scope Isolation", func(t *testing.T) {
		// This test is already correct as it validates sandboxing for funcs.
		script := `
			func modify_var() means
				set a = "modified_in_proc"
			endfunc

			func main() means
				call modify_var()
			endfunc
		`
		h := NewTestHarness(t)
		interp := h.Interpreter
		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		interp.Load(&interfaces.Tree{Root: program})

		interp.SetVariable("a", lang.StringValue{Value: "original"})

		_, err := interp.Run("main")
		if err != nil {
			t.Fatalf("script execution failed: %v", err)
		}

		finalA, _ := interp.GetVariable("a")
		expectedA := lang.StringValue{Value: "original"}
		if !reflect.DeepEqual(finalA, expectedA) {
			t.Errorf("Procedure call modified a variable in the parent scope. Expected 'a' to be %#v, got %#v", expectedA, finalA)
		}
	})
}
