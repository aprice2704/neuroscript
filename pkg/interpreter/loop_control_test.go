// NeuroScript Version: 0.5.2
// File version: 4.0.0
// Purpose: Refactored to use the centralized TestHarness, ensuring proper initialization and alignment with the modern API.
// filename: pkg/interpreter/interpreter_loop_control_test.go
// nlines: 120
// risk_rating: LOW

package interpreter_test

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func runLoopControlFlowTest(t *testing.T, script string) (lang.Value, error) {
	t.Helper()
	h := NewTestHarness(t)
	h.T.Logf("[DEBUG] Turn 1: Harness created for script:\n%s", script)

	tree, pErr := h.Parser.Parse(script)
	if pErr != nil {
		h.T.Logf("[DEBUG] Turn 2: Parser failed: %v", pErr)
		return nil, pErr
	}
	program, _, bErr := h.ASTBuilder.Build(tree)
	if bErr != nil {
		h.T.Logf("[DEBUG] Turn 3: AST Builder failed: %v", bErr)
		return nil, bErr
	}
	if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
		h.T.Logf("[DEBUG] Turn 4: Load failed: %v", err)
		return nil, err
	}
	h.T.Logf("[DEBUG] Turn 5: Calling Run('main').")
	result, runErr := h.Interpreter.Run("main")
	h.T.Logf("[DEBUG] Turn 6: Run('main') completed. Result: %#v, Error: %v", result, runErr)
	return result, runErr
}

func TestLoopControlStatements(t *testing.T) {
	t.Run("break_exits_only_inner_loop", func(t *testing.T) {
		script := `
			func main() means
				set outer_count = 0
				set inner_count = 0
				set i = 0
				while i < 3
					set outer_count = outer_count + 1
					set j = 0
					while j < 5
						if j == 2
							break
						endif
						set inner_count = inner_count + 1
						set j = j + 1
					endwhile
					set i = i + 1
				endwhile
				return outer_count, inner_count
			endfunc
		`
		result, err := runLoopControlFlowTest(t, script)
		if err != nil {
			t.Fatalf("script failed: %v", err)
		}

		resSlice, ok := result.(lang.ListValue)
		if !ok {
			t.Fatalf("Expected list result, got %T", result)
		}
		if len(resSlice.Value) != 2 {
			t.Fatalf("Expected 2 return values, got %d", len(resSlice.Value))
		}

		outerCount, _ := lang.ToFloat64(resSlice.Value[0])
		innerCount, _ := lang.ToFloat64(resSlice.Value[1])

		if outerCount != 3 {
			t.Errorf("Expected outer_count to be 3, got %v", outerCount)
		}
		if innerCount != 6 {
			t.Errorf("Expected inner_count to be 6 (2 per outer loop), got %v", innerCount)
		}
	})

	t.Run("continue_skips_only_inner_loop", func(t *testing.T) {
		script := `
			func main() means
				set total = 0
				set i = 0
				while i < 3
					set i = i + 1
					if i == 2
						continue
					endif
					set total = total + 1
				endwhile
				return total
			endfunc
		`
		result, err := runLoopControlFlowTest(t, script)
		if err != nil {
			t.Fatalf("script failed: %v", err)
		}
		total, _ := lang.ToFloat64(result)
		if total != 2 {
			t.Errorf("Expected total to be 2, got %v", total)
		}
	})

	t.Run("break_outside_loop_is_a_builder_error", func(t *testing.T) {
		script := `
			func main() means
				break
			endfunc
		`
		_, err := runLoopControlFlowTest(t, script)
		if err == nil {
			t.Fatal("Expected script to fail during build, but it succeeded.")
		}

		expectedError := "'break' statement found outside of a loop"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error message to contain '%s', but got '%s'", expectedError, err.Error())
		}
	})
}
