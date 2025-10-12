// filename: pkg/interpreter/interpreter_scoping_test.go
// Neuroscript version: 0.5.2
// File version: 19
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
package interpreter_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestErrorScoping_HandlersDoNotLeak verifies that 'on error' handlers are strictly
// function-scoped and that control flow can be correctly managed after an error.
func TestErrorScoping_HandlersDoNotLeak(t *testing.T) {
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
	t.Logf("[DEBUG] Turn 1: Starting TestErrorScoping_HandlersDoNotLeak.")
	h := NewTestHarness(t)
	var outputBuffer bytes.Buffer
	h.HostContext.EmitFunc = func(v lang.Value) {
		fmt.Fprintln(&outputBuffer, v.String())
	}

	tree, antlrParseErr := h.Parser.Parse(scriptContent)
	if antlrParseErr != nil {
		t.Fatalf("Failed to parse script: %v", antlrParseErr)
	}
	programAST, _, buildErr := h.ASTBuilder.Build(tree)
	if buildErr != nil {
		t.Fatalf("Failed to build AST from parsed script: %v", buildErr)
	}
	if err := h.Interpreter.Load(&interfaces.Tree{Root: programAST}); err != nil {
		t.Fatalf("Failed to load program into interpreter: %v", err)
	}
	t.Logf("[DEBUG] Turn 2: Script parsed and loaded.")

	_, execErr := h.Interpreter.Run("main")
	if execErr != nil {
		t.Fatalf("ExecuteProc returned an unexpected error: %v\nOutput:\n%s", execErr, outputBuffer.String())
	}
	t.Logf("[DEBUG] Turn 3: 'main' procedure executed.")

	output := outputBuffer.String()
	t.Logf("--- SCRIPT OUTPUT (Success Test) ---\n%s\n---------------------", output)

	expectedPassMessage := "[PASS] 'on error' handlers are correctly scoped."
	if !strings.Contains(output, expectedPassMessage) {
		t.Errorf("Test failed: output did not contain the expected PASS message.\nExpected to find: '%s'", expectedPassMessage)
	}
	t.Logf("[DEBUG] Turn 4: Assertion passed.")
}
