// NeuroScript Version: 0.8.0
// File version: 5.0.0
// Purpose: Refactored to use the centralized TestHarness and modern API calls for a robust test setup. Removed call to obsolete SetMaxLoopIterations.
// filename: pkg/interpreter/interpreter_resource_usage_test.go
// nlines: 85
// risk_rating: MEDIUM

package interpreter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestResourceUsageLimits(t *testing.T) {
	t.Run("Maximum Recursion Depth", func(t *testing.T) {
		script := `
			func infinite_recursion() means
				call infinite_recursion()
			endfunc

			func main() means
				call infinite_recursion()
			endfunc
		`
		t.Logf("[DEBUG] Turn 1: Starting 'Maximum Recursion Depth' test.")
		h := NewTestHarness(t)
		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		h.Interpreter.Load(&interfaces.Tree{Root: program})
		t.Logf("[DEBUG] Turn 2: Script loaded, running 'main'.")

		_, err := h.Interpreter.Run("main")
		t.Logf("[DEBUG] Turn 3: Execution finished, expecting error.")

		if err == nil {
			t.Fatal("Expected an error for exceeding max recursion depth, but got nil")
		}
		if !errors.Is(err, lang.ErrMaxCallDepthExceeded) {
			t.Errorf("Expected error to be ErrMaxCallDepthExceeded, but got: %v", err)
		}
		t.Logf("[DEBUG] Turn 4: Correctly received expected error: %v", err)
	})

	t.Run("Maximum Loop Iterations", func(t *testing.T) {
		script := `
			func main() means
				while true
					set a = 1
				endwhile
			endfunc
		`
		t.Logf("[DEBUG] Turn 1: Starting 'Maximum Loop Iterations' test.")
		h := NewTestHarness(t)
		// The default limit of 1000 is fine for this test.

		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		h.Interpreter.Load(&interfaces.Tree{Root: program})
		t.Logf("[DEBUG] Turn 2: Script loaded, running 'main'.")

		_, err := h.Interpreter.Run("main")
		t.Logf("[DEBUG] Turn 3: Execution finished, expecting error.")

		if err == nil {
			t.Fatal("Expected an error for exceeding max loop iterations, but got nil")
		}
		if !strings.Contains(err.Error(), "exceeded max iterations") {
			t.Errorf("Expected error message to contain 'exceeded max iterations', but got: %s", err.Error())
		}
		t.Logf("[DEBUG] Turn 4: Correctly received expected error: %v", err)
	})

	t.Run("Large Data Structure Allocation", func(t *testing.T) {
		t.Skip("Skipping large data structure test. Implement when memory limits are in place.")
		script := `
			func main() means
				set big_list = []
				set i = 0
				while i < 1000000 # 1 Million
					set big_list = tool.List.Append(big_list, i)
					set i = i + 1
				endwhile
			endfunc
		`
		h := NewTestHarness(t)
		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		h.Interpreter.Load(&interfaces.Tree{Root: program})
		_, err := h.Interpreter.Run("main")
		if !errors.Is(err, lang.ErrResourceExhaustion) {
			t.Errorf("Expected ErrResourceExhaustion, got %v", err)
		}
	})
}
