// NeuroScript Version: 0.5.2
// File version: 5.0.0
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// filename: pkg/interpreter/interpreter_scoping_extended_test.go
// nlines: 130
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// runScopeTestScript is a helper to execute a script and return the final interpreter state.
func runScopeTestScript(t *testing.T, scriptContent string, initialVars map[string]lang.Value) (*interpreter.Interpreter, error) {
	t.Helper()
	h := NewTestHarness(t)
	interp := h.Interpreter

	if initialVars != nil {
		for k, v := range initialVars {
			interp.SetVariable(k, v)
		}
	}
	h.T.Logf("[DEBUG] Turn 1: Harness created, initial vars set.")

	scriptName := "scoping_test"
	fullScript := fmt.Sprintf("func %s() means\n%s\nendfunc", scriptName, scriptContent)

	tree, pErr := h.Parser.Parse(fullScript)
	if pErr != nil {
		return interp, pErr
	}
	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		return interp, bErr
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		return interp, err
	}
	h.T.Logf("[DEBUG] Turn 2: Script parsed and loaded.")

	_, execErr := interp.Run(scriptName)
	h.T.Logf("[DEBUG] Turn 3: Script executed.")
	return interp, execErr
}

func TestVariableScopingExtended(t *testing.T) {
	t.Run("Variable Shadowing in If Block", func(t *testing.T) {
		script := `
			set x = 10
			if true
				set x = 20
				set y = 5
			endif
		`
		interp, err := runScopeTestScript(t, script, nil)
		if err != nil {
			t.Fatal(err)
		}

		finalX, _ := interp.GetVariable("x")
		expectedX := lang.NumberValue{Value: 20}
		if finalX != expectedX {
			t.Errorf("Variable assignment failed. Expected x to be %#v, got %#v", expectedX, finalX)
		}

		finalY, _ := interp.GetVariable("y")
		expectedY := lang.NumberValue{Value: 5}
		if finalY != expectedY {
			t.Errorf("Expected y to be set to %#v, got %#v", expectedY, finalY)
		}
	})

	t.Run("For Loop Variable Has Correct Final Value", func(t *testing.T) {
		script := `
			for each loop_var in my_list
				set z = loop_var
			endfor
		`
		initialVars := map[string]lang.Value{
			"my_list": lang.NewListValue([]lang.Value{
				lang.NumberValue{Value: 1},
				lang.NumberValue{Value: 2},
			}),
		}
		interp, err := runScopeTestScript(t, script, initialVars)
		if err != nil {
			t.Fatal(err)
		}

		finalLoopVar, exists := interp.GetVariable("loop_var")
		if !exists {
			t.Error("Expected loop variable 'loop_var' to exist, but it did not.")
		} else {
			expectedVar := lang.NumberValue{Value: 2}
			if finalLoopVar != expectedVar {
				t.Errorf("Expected loop variable to have final value %#v, got %#v", expectedVar, finalLoopVar)
			}
		}
	})

	t.Run("Procedure Scope Isolation", func(t *testing.T) {
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
		h.T.Logf("[DEBUG] Turn 1: Harness created, var 'a' set to 'original'.")

		_, err := interp.Run("main")
		if err != nil {
			t.Fatalf("script execution failed: %v", err)
		}
		h.T.Logf("[DEBUG] Turn 2: 'main' procedure executed.")

		finalA, _ := interp.GetVariable("a")
		expectedA := lang.StringValue{Value: "original"}
		if finalA != expectedA {
			t.Errorf("Procedure call modified a variable in the parent scope. Expected 'a' to be %#v, got %#v", expectedA, finalA)
		}
		h.T.Logf("[DEBUG] Turn 3: Assertion passed.")
	})
}
