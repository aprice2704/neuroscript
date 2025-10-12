// NeuroScript Version: 0.5.2
// File version: 3.0.0
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
// filename: pkg/interpreter/interpreter_return_value_test.go
// nlines: 50
// risk_rating: LOW

package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestProcedureReturnValue verifies that the final value returned by an explicit
// 'return' statement in a procedure is the same value received by the Go code
// that calls the 'interp.Run()' method. This prevents regressions on the return path.
func TestProcedureReturnValue(t *testing.T) {
	script := `
func get_specific_value(returns val) means
  return "this is the expected return value"
endfunc
`
	t.Logf("[DEBUG] Turn 1: Starting TestProcedureReturnValue.")
	h := NewTestHarness(t)
	interp := h.Interpreter

	tree, pErr := h.Parser.Parse(script)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}
	t.Logf("[DEBUG] Turn 2: Script parsed and loaded.")

	result, runErr := interp.Run("get_specific_value")
	if runErr != nil {
		t.Fatalf("interp.Run() returned an unexpected error: %v", runErr)
	}
	t.Logf("[DEBUG] Turn 3: 'get_specific_value' procedure executed.")

	expected := lang.StringValue{Value: "this is the expected return value"}
	if result != expected {
		t.Errorf("Procedure return value mismatch.\n  Expected: %#v\n  Got:      %#v", expected, result)
	}
	t.Logf("[DEBUG] Turn 4: Assertion passed.")
}
