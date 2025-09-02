// NeuroScript Version: 0.6.0
// File version: 3
// Purpose: Corrected the call to interp.Load to pass the correct AST structure.
// filename: pkg/interpreter/interpreter_globals_test.go
// nlines: 55
// risk_rating: LOW

package interpreter

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

func TestInterpreter_WithGlobals(t *testing.T) {
	script := `
func main(returns string) means
    return my_global_var
endfunc
`
	// 1. Setup the interpreter with an initial global variable using the WithGlobals option.
	globals := map[string]interface{}{
		"my_global_var": "hello from globals",
	}

	// Use the standard NewInterpreter with the option to correctly simulate the API usage.
	interp := NewInterpreter(
		WithGlobals(globals),
		WithLogger(logging.NewTestLogger(t)),
	)

	// 2. Parse and load the script.
	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(script)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// 3. Run the procedure.
	result, runErr := interp.Run("main")
	if runErr != nil {
		t.Fatalf("interp.Run() returned an unexpected error: %v", runErr)
	}

	// 4. Assert the result.
	expected := lang.StringValue{Value: "hello from globals"}
	if result != expected {
		t.Errorf("Procedure return value mismatch.\n  Expected: %#v\n  Got:      %#v", expected, result)
	}
}
