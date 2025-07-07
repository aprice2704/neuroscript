// NeuroScript Version: 0.5.2
// File version: 1.2.0
// Purpose: Corrected syntax in the infinite loop test by adding a statement to the loop body, fixing the parser error.
// filename: pkg/interpreter/interpreter_resource_usage_test.go
// nlines: 80
// risk_rating: MEDIUM

package interpreter

import (
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
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
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		parserAPI := parser.NewParserAPI(interp.GetLogger())
		ast, _ := parserAPI.Parse(script)
		prog, _, _ := parser.NewASTBuilder(interp.GetLogger()).Build(ast)
		interp.Load(prog)

		_, err := interp.Run("main")

		if err == nil {
			t.Fatal("Expected an error for exceeding max recursion depth, but got nil")
		}
		if !errors.Is(err, lang.ErrMaxCallDepthExceeded) {
			t.Errorf("Expected error to be ErrMaxCallDepthExceeded, but got: %v", err)
		}
	})

	t.Run("Maximum Loop Iterations", func(t *testing.T) {
		// FIX: Added 'set a = 1' to the loop body to make the script
		// syntactically valid for the parser.
		script := `
			while true
				set a = 1
			endwhile
		`
		interp, _ := newLocalTestInterpreter(t, nil, nil)
		// Lower the limit for a faster test
		interp.maxLoopIterations = 500

		_, err := interp.ExecuteScriptString("infinite_loop_test", script, nil)

		if err == nil {
			t.Fatal("Expected an error for exceeding max loop iterations, but got nil")
		}
		if !strings.Contains(err.Error(), "exceeded max iterations") {
			t.Errorf("Expected error message to contain 'exceeded max iterations', but got: %s", err.Error())
		}
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
		_, err := runScopeTestScript(t, script, nil)
		if !errors.Is(err, lang.ErrResourceExhaustion) {
			t.Errorf("Expected ErrResourceExhaustion, got %v", err)
		}
	})
}
