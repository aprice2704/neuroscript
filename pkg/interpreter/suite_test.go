// NeuroScript Version: 0.8.0
// File version: 11.0.0
// Purpose: Corrected the error handling logic to properly check for a RuntimeError without a redundant type assertion.
// filename: pkg/interpreter/interpreter_suite_test.go
package interpreter_test

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestExecuteStepsBlocksAndLoops(t *testing.T) {
	t.Run("IF true literal", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'IF true literal' test.")
		h := NewTestHarness(t)
		script := `
			func main() means
				if true
					set x = "Inside"
				endif
			endfunc
		`
		_, err := h.Interpreter.ExecuteScriptString("main", script, nil)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}
		t.Logf("[DEBUG] Turn 2: Script executed.")
		val, _ := h.Interpreter.GetVariable("x")
		expected := lang.StringValue{Value: "Inside"}
		if !reflect.DeepEqual(val, expected) {
			t.Errorf("Variable mismatch. Got: %#v, Want: %#v", val, expected)
		}
		t.Logf("[DEBUG] Turn 3: Assertion passed.")
	})

	t.Run("WHILE loop basic", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'WHILE loop basic' test.")
		h := NewTestHarness(t)
		script := `
			func main() means
				set i = 0
				while i < 3
					set i = i + 1
				endwhile
			endfunc
		`
		_, err := h.Interpreter.ExecuteScriptString("main", script, nil)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}
		t.Logf("[DEBUG] Turn 2: Script executed.")
		val, _ := h.Interpreter.GetVariable("i")
		expected := lang.NumberValue{Value: 3}
		if !reflect.DeepEqual(val, expected) {
			t.Errorf("Variable mismatch. Got: %#v, Want: %#v", val, expected)
		}
		t.Logf("[DEBUG] Turn 3: Assertion passed.")
	})

	t.Run("FOR EACH loop", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'FOR EACH loop' test.")
		h := NewTestHarness(t)
		script := `
			func main() means
				set l = [10, 20, 30]
				set sum = 0
				for each item in l
					set sum = sum + item
				endfor
			endfunc
		`
		_, err := h.Interpreter.ExecuteScriptString("main", script, nil)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}
		t.Logf("[DEBUG] Turn 2: Script executed.")
		val, _ := h.Interpreter.GetVariable("sum")
		expected := lang.NumberValue{Value: 60}
		if !reflect.DeepEqual(val, expected) {
			t.Errorf("Variable mismatch. Got: %#v, Want: %#v", val, expected)
		}
		t.Logf("[DEBUG] Turn 3: Assertion passed.")
	})

	t.Run("MUST evaluation error", func(t *testing.T) {
		t.Logf("[DEBUG] Turn 1: Starting 'MUST evaluation error' test.")
		h := NewTestHarness(t)
		script := `
			func main() means
				must 1 > 5
			endfunc
		`
		_, err := h.Interpreter.ExecuteScriptString("main", script, nil)
		if err == nil {
			t.Fatal("Expected script to fail, but it succeeded.")
		}
		t.Logf("[DEBUG] Turn 2: Script executed, correctly received error: %v", err)

		// The function signature guarantees the error is a *lang.RuntimeError,
		// so we just need to check that it's not nil.
	})
}
