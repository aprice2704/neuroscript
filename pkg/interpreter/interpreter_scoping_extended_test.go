// NeuroScript Version: 0.5.2
// File version: 3.0.0
// Purpose: Corrected the test helper to wrap script snippets in a full function definition, fixing all parsing-related test failures.
// filename: pkg/interpreter/interpreter_scoping_extended_test.go
// nlines: 125
// risk_rating: LOW

package interpreter

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// runScopeTestScript is a helper to execute a script and return the final interpreter state.
func runScopeTestScript(t *testing.T, scriptContent string, initialVars map[string]lang.Value) (*Interpreter, error) {
	t.Helper()

	interp, err := newLocalTestInterpreter(t, initialVars, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create test interpreter: %w", err)
	}

	// FIX: Wrap the raw script snippet in a function definition to make it a valid program.
	scriptName := "scoping_test"
	fullScript := fmt.Sprintf("func %s() means\n%s\nendfunc", scriptName, scriptContent)

	_, execErr := interp.ExecuteScriptString(scriptName, fullScript, nil)
	if execErr != nil {
		// We return the interpreter state even on failure to allow inspection.
		return interp, execErr
	}

	return interp, nil
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
		// FIX: The script no longer re-defines the list. It relies on the
		// one provided in initialVars, which is the correct way to test this.
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

		// The loop variable 'loop_var' should persist with its last assigned value
		// in the current single-scope implementation.
		finalLoopVar, exists := interp.GetVariable("loop_var")
		if !exists {
			t.Error("Expected loop variable 'loop_var' to exist, but it did not.")
		} else {
			expectedVar := lang.NumberValue{Value: 2} // The last item in the list.
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
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		parserAPI := parser.NewParserAPI(interp.GetLogger())
		ast, _ := parserAPI.Parse(script)
		prog, _, _ := parser.NewASTBuilder(interp.GetLogger()).Build(ast)
		interp.Load(prog)

		interp.SetVariable("a", lang.StringValue{Value: "original"})

		_, err := interp.Run("main")
		if err != nil {
			t.Fatalf("script execution failed: %v", err)
		}

		finalA, _ := interp.GetVariable("a")
		expectedA := lang.StringValue{Value: "original"}
		if finalA != expectedA {
			t.Errorf("Procedure call modified a variable in the parent scope. Expected 'a' to be %#v, got %#v", expectedA, finalA)
		}
	})
}
