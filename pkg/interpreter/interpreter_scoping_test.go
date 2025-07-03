// filename: pkg/interpreter/interpreter_scoping_test.go
package interpreter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// TestErrorScoping_HandlersDoNotLeak verifies that 'on error' handlers are strictly
// function-scoped and that control flow can be correctly managed after an error.
// This is the "EXPECT TO PASS" test case.
func TestErrorScoping_HandlersDoNotLeak(t *testing.T) {
	// This script is designed to PASS if error handlers are correctly scoped.
	scriptContent := `
:: name: Test On Error Scoping
:: file_version: 4
:: purpose: Verifies 'on error' handlers are function-scoped, using a flag for control flow.
:: target: main

func test_local_handler() means
    on error do
        emit "SUCCESS: Caught expected error inside test_local_handler."
        clear_error
    endon
    emit "--> Testing local 'on error' handler..."
    fail "This is an intentional error to be caught locally."
endfunc

func test_for_leakage() means
    emit "--> Testing for handler leakage..."
    fail "This error should be caught by the main function's handler, not any other."
endfunc

func main() means
    emit "[START] Testing 'on error' handler scoping."
    set test_passed = false
    call test_local_handler()
    emit "--------------------------------------------------"
    on error do
        emit "SUCCESS: Main handler correctly caught propagated error."
        emit "[PASS] 'on error' handlers are correctly scoped."
        clear_error
        set test_passed = true
    endon
    call test_for_leakage()
    if test_passed != true
        emit "[FAIL] The error from test_for_leakage() was not caught by the main handler."
    endif
endfunc
`

	var stdout bytes.Buffer
	// FIX: Use exported NewInterpreter with options.
	interp := NewInterpreter(WithStdout(&stdout), WithLogger(logging.NewTestLogger(t)))

	// FIX: Use the actual parser API.
	parserAPI := parser.NewParserAPI(logging.NewTestLogger(t))
	antlrTree, antlrParseErr := parserAPI.Parse(scriptContent)
	if antlrParseErr != nil {
		t.Fatalf("Failed to parse script: %v", antlrParseErr)
	}

	// FIX: Use the actual AST builder.
	astBuilder := parser.NewASTBuilder(logging.NewTestLogger(t))
	programAST, _, buildErr := astBuilder.Build(antlrTree)
	if buildErr != nil {
		t.Fatalf("Failed to build AST from parsed script: %v", buildErr)
	}

	if err := interp.Load(programAST); err != nil {
		t.Fatalf("Failed to load program into interpreter: %v", err)
	}

	_, execErr := interp.Run("main")
	if execErr != nil {
		t.Fatalf("ExecuteProc returned an unexpected error: %v\nOutput:\n%s", execErr, stdout.String())
	}

	output := stdout.String()
	t.Logf("--- SCRIPT OUTPUT (Success Test) ---\n%s\n---------------------", output)

	expectedPassMessage := "[PASS] 'on error' handlers are correctly scoped."
	if !strings.Contains(output, expectedPassMessage) {
		t.Errorf("Test failed: output did not contain the expected PASS message.\nExpected to find: '%s'", expectedPassMessage)
	}

	unexpectedFailMessage := "[FAIL] The error from test_for_leakage() was not caught by the main handler."
	if strings.Contains(output, unexpectedFailMessage) {
		t.Errorf("Test failed: output contained the FAIL message.\nFound: '%s'", unexpectedFailMessage)
	}
}

// TestErrorScoping_SimulatedLeakageFailure provides a counter-example to the success test.
// It runs a script that is explicitly written to mimic the old, leaky behavior.
// This is the "EXPECT TO FAIL" test case.
func TestErrorScoping_SimulatedLeakageFailure(t *testing.T) {
	// This script simulates a failure. The handler in 'outer_procedure_with_handler'
	// will catch the error from the separate 'inner_procedure_that_fails'.
	scriptContent := `
:: name: Test On Error Scoping Failure Simulation
:: file_version: 1
:: purpose: Demonstrates what an error scope leak would look like for testing purposes.
:: target: main

func inner_procedure_that_fails() means
    emit "--> inner_procedure_that_fails is running and will now fail."
    fail "Error originating from inner_procedure."
endfunc

func outer_procedure_with_handler() means
    on error do
        # This message indicates the 'leak' was successfully simulated.
        emit "[SIMULATED LEAK]: Handler from outer_procedure caught error from inner_procedure."
        clear_error
    endon
    emit "--> outer_procedure_with_handler is running."
    call inner_procedure_that_fails()
endfunc

func main() means
    emit "[START] Simulating 'on error' handler leakage."
    call outer_procedure_with_handler()
    emit "[END] Simulation finished."
endfunc
`

	var stdout bytes.Buffer
	interp := NewInterpreter(WithStdout(&stdout), WithLogger(logging.NewTestLogger(t)))

	parserAPI := parser.NewParserAPI(logging.NewTestLogger(t))
	antlrTree, antlrParseErr := parserAPI.Parse(scriptContent)
	if antlrParseErr != nil {
		t.Fatalf("Failed to parse script: %v", antlrParseErr)
	}

	astBuilder := parser.NewASTBuilder(logging.NewTestLogger(t))
	programAST, _, buildErr := astBuilder.Build(antlrTree)
	if buildErr != nil {
		t.Fatalf("Failed to build AST from parsed script: %v", buildErr)
	}

	if err := interp.Load(programAST); err != nil {
		t.Fatalf("Failed to load program into interpreter: %v", err)
	}

	_, execErr := interp.Run("main")
	if execErr != nil {
		t.Fatalf("ExecuteProc returned an unexpected error: %v\nOutput:\n%s", execErr, stdout.String())
	}

	output := stdout.String()
	t.Logf("--- SCRIPT OUTPUT (Failure Simulation Test) ---\n%s\n---------------------", output)

	// In this test, we EXPECT to find the simulated leak message.
	// This confirms our ability to detect what a failure looks like.
	expectedLeakMessage := "[SIMULATED LEAK]: Handler from outer_procedure caught error from inner_procedure."
	if !strings.Contains(output, expectedLeakMessage) {
		t.Errorf("Failure simulation test failed: output did not contain the simulated leak message.\nExpected to find: '%s'", expectedLeakMessage)
	}
}
