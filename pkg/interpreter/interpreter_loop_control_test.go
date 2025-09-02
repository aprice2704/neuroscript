// NeuroScript Version: 0.5.2
// File version: 3.0.0
// Purpose: Corrected calls to the renamed test helper function 'NewTestInterpreter'.
// filename: pkg/interpreter/interpreter_loop_control_test.go
// nlines: 115
// risk_rating: LOW

package interpreter

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// runControlFlowTest now returns the final value from the run, and the error.
func runLoopControlFlowTest(t *testing.T, script string) (lang.Value, error) {
	t.Helper()
	interp, err := NewTestInterpreter(t, nil, nil, false)
	if err != nil {
		return nil, err
	}

	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(script)
	if pErr != nil {
		return nil, pErr
	}

	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		return nil, bErr
	}

	if err := interp.Load(&interfaces.Tree{Root: &interfaces.Tree{Root: &interfaces.Tree{Root: &interfaces.Tree{Root: program}}}}); err != nil {
		return nil, err
	}

	// Run main and return its result directly.
	return interp.Run("main")
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
		result, err := runControlFlowTest(t, script)
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
		result, err := runControlFlowTest(t, script)
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
		// FIX: The error should come from the build process, not the run process.
		_, err := runControlFlowTest(t, script)
		if err == nil {
			t.Fatal("Expected script to fail during build, but it succeeded.")
		}

		// Check for the specific build-time error message.
		expectedError := "'break' statement found outside of a loop"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error message to contain '%s', but got '%s'", expectedError, err.Error())
		}
	})
}
