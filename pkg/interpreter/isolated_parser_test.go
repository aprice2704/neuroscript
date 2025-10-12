// filename: pkg/interpreter/interpreter_isolated_parser_test.go
// File version: 2
// Purpose: Refactored to use the parser components from the centralized TestHarness for a consistent setup.
package interpreter_test

import (
	"testing"
)

// TestIsolatedParser runs a single, isolated parser test on a script
// to ensure the core parser logic is sound.
func TestIsolatedParser(t *testing.T) {
	script := `func TestMustAndErrorHandling(returns result) means
  on error do
    set result = "Caught error: a 'must' condition failed"
    return result
  endon

  set a = 1
  set b = 2

  must a > b

  return "This should not be returned"
endfunc`

	t.Logf("[DEBUG] Turn 1: Starting Isolated Parser Test.")
	h := NewTestHarness(t)
	t.Logf("[DEBUG] Turn 2: Harness created.")

	_, err := h.Parser.Parse(script)
	t.Logf("[DEBUG] Turn 3: Script parsed.")

	if err != nil {
		t.Fatalf("Isolated parser test FAILED. The parser produced an error:\n%v", err)
	} else {
		t.Logf("Isolated parser test PASSED. Script parsed cleanly.")
	}
}
