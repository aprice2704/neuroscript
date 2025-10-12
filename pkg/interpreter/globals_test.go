// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Refactored to use the centralized TestHarness and WithGlobals option for a robust and modern test setup.
// filename: pkg/interpreter/interpreter_globals_test.go
// nlines: 60
// risk_rating: LOW

package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestInterpreter_WithGlobals(t *testing.T) {
	script := `
func main(returns string) means
    return my_global_var
endfunc
`
	t.Logf("[DEBUG] Turn 1: Starting TestInterpreter_WithGlobals.")
	globals := map[string]interface{}{
		"my_global_var": "hello from globals",
	}

	h := NewTestHarness(t)
	// We need a new interpreter with our specific globals, so we create one
	// but reuse the harness's HostContext to ensure proper initialization.
	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(h.HostContext),
		interpreter.WithGlobals(globals),
	)
	t.Logf("[DEBUG] Turn 2: Test harness and new interpreter with globals created.")

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
	t.Logf("[DEBUG] Turn 3: Script parsed and loaded.")

	result, runErr := interp.Run("main")
	if runErr != nil {
		t.Fatalf("interp.Run() returned an unexpected error: %v", runErr)
	}
	t.Logf("[DEBUG] Turn 4: 'main' procedure executed.")

	expected := lang.StringValue{Value: "hello from globals"}
	if result != expected {
		t.Errorf("Procedure return value mismatch.\n  Expected: %#v\n  Got:      %#v", expected, result)
	}
	t.Logf("[DEBUG] Turn 5: Assertion passed.")
}
