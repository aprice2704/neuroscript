// NeuroScript Version: 0.5.2
// File version: 2.0.0
// Purpose: Corrected call to the renamed test helper function 'NewTestInterpreter'.
// filename: pkg/interpreter/interpreter_return_value_test.go
// nlines: 45
// risk_rating: LOW

package interpreter

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
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
	// 1. Setup the interpreter and parse the script.
	interp, err := NewTestInterpreter(t, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}

	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(script)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: &interfaces.Tree{Root: &interfaces.Tree{Root: &interfaces.Tree{Root: program}}}}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// 2. Run the procedure.
	result, runErr := interp.Run("get_specific_value")
	if runErr != nil {
		t.Fatalf("interp.Run() returned an unexpected error: %v", runErr)
	}

	// 3. Assert the result.
	expected := lang.StringValue{Value: "this is the expected return value"}
	if result != expected {
		t.Errorf("Procedure return value mismatch.\n  Expected: %#v\n  Got:      %#v", expected, result)
	}
}
