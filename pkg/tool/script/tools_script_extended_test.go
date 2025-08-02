// filename: pkg/tool/script/tools_script_extended_test.go
package script

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

type scriptTestCase struct {
	name            string
	script          string
	wantResult      interface{}
	wantExecErrCode lang.ErrorCode // For checking specific runtime error codes
	checkFunc       func(t *testing.T, result interface{}, err error)
}

// Test cases for extended script tool functionality
var scriptTestCases = []scriptTestCase{
	{
		name: "list_before_load",
		script: `
		:: title: Test listing functions on a fresh interpreter

		func main() means
			// This should return a map containing only the 'main' function itself.
			return tool.script.ListFunctions()
		endfunc
		`,
		wantResult: map[string]interface{}{
			"main": "procedure main()",
		},
	},
	{
		name: "load_and_list",
		script: `
		:: title: Test loading a script and then listing functions

		func main() means
			set script_to_load = "func new_func() means\nreturn 123\nendfunc"
			set _ = tool.script.LoadScript(script_to_load)
			return tool.script.ListFunctions()
		endfunc
		`,
		checkFunc: func(t *testing.T, result interface{}, err error) {
			if err != nil {
				t.Fatalf("checkFunc received unexpected error: %v", err)
			}
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("Expected result to be a map, but got %T", result)
			}
			// We expect 'main' from the test script and 'new_func' from the loaded script.
			if _, ok := resultMap["main"]; !ok {
				t.Error("Expected function list to contain 'main'")
			}
			if _, ok := resultMap["new_func"]; !ok {
				t.Error("Expected function list to contain 'new_func'")
			}
			if len(resultMap) != 2 {
				t.Errorf("Expected 2 functions, but got %d", len(resultMap))
			}
		},
	},
	{
		name: "load_empty_script",
		script: `
		:: title: Test loading an empty script string

		func main() means
			// The tool should return an execution failure wrapping the syntax error.
			set _ = tool.script.LoadScript("")
		endfunc
		`,
		wantExecErrCode: lang.ErrorCodeToolExecutionFailed, // Corrected to 21
	},
	{
		name: "load_script_with_syntax_error",
		script: `
		:: title: Test that loading a script with a syntax error fails gracefully.

		func main() means
			// The tool should return an execution failure wrapping the parse error.
			set _ = tool.script.LoadScript("FUNKY CHICKEN")
		endfunc
		`,
		wantExecErrCode: lang.ErrorCodeToolExecutionFailed, // Corrected to 21
	},
	{
		name: "list_after_failed_load",
		script: `
		:: title: Test that a failed load does not alter the function list

		// Define a function that should persist.
		func first_func() means
			return 1
		endfunc

		func main() means
			// Try to load a script with a syntax error.
			// The error should be caught and cleared, not crash the script.
			on error do
				clear_error
			endon
			set _ = tool.script.LoadScript("FUNK bad() means\nENDFUNK")

			// Now, list the functions. Only 'main' and 'first_func' should exist.
			return tool.script.ListFunctions()
		endfunc
		`,
		checkFunc: func(t *testing.T, result interface{}, err error) {
			if err != nil {
				t.Fatalf("checkFunc received unexpected error: %v", err)
			}
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("Expected result to be a map, but got %T", result)
			}
			if _, ok := resultMap["main"]; !ok {
				t.Error("Function list missing 'main'")
			}
			if _, ok := resultMap["first_func"]; !ok {
				t.Error("Function list missing 'first_func'")
			}
			if len(resultMap) != 2 {
				t.Errorf("Expected 2 functions after failed load, but found %d", len(resultMap))
			}
		},
	},
	{
		name: "load_metadata_only",
		script: `
		:: title: Test loading a script with only metadata

		func main() means
			// This should succeed and return a map describing the load operation.
			set script_body = ":: title: A file with no code"
			return tool.script.LoadScript(script_body)
		endfunc
		`,
		wantResult: map[string]interface{}{
			"functions_loaded":      float64(0),
			"event_handlers_loaded": float64(0),
			"metadata": map[string]interface{}{
				"title": "A file with no code",
			},
		},
	},
}

func TestScriptToolsExtended(t *testing.T) {
	for _, tc := range scriptTestCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Setup interpreter
			logger := logging.NewTestLogger(t)
			interp := interpreter.NewInterpreter(
				interpreter.WithLogger(logger),
			)

			// Manually register the script tools with the interpreter's registry.
			for _, toolImpl := range scriptToolsToRegister {
				if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
					t.Fatalf("failed to register tool '%s': %v", toolImpl.Spec.Name, err)
				}
			}

			// --- PARSE AND EXECUTE THE TEST SCRIPT ---
			p := parser.NewParserAPI(logger)
			program, pErr := p.Parse(tc.script)
			if pErr != nil {
				t.Fatalf("failed to parse test driver script '%s': %v", tc.name, pErr)
			}

			astBuilder := parser.NewASTBuilder(logger)
			programAST, _, bErr := astBuilder.Build(program)
			if bErr != nil {
				t.Fatalf("failed to build ast for test driver script '%s': %v", tc.name, bErr)
			}

			// Execute the script by loading the AST and running the 'main' procedure.
			finalValue, err := interp.LoadAndRun(programAST, "main")

			// --- CHECK RESULTS ---
			// Check for expected execution errors by code
			if tc.wantExecErrCode != 0 {
				if err == nil {
					t.Fatalf("Expected an error with code %d, but got nil", tc.wantExecErrCode)
				}
				var runtimeErr *lang.RuntimeError
				if !errors.As(err, &runtimeErr) {
					t.Fatalf("Expected a *lang.RuntimeError, but got %T: %v", err, err)
				}
				if runtimeErr.Code != tc.wantExecErrCode {
					t.Fatalf("Expected error code %d, but got %d", tc.wantExecErrCode, runtimeErr.Code)
				}
				// Execution failed as expected, test is done.
				return
			}

			// Check for unexpected execution errors
			if err != nil {
				t.Fatalf("Test script execution failed unexpectedly for '%s': %v", tc.name, err)
			}

			// Unwrap the NeuroScript value to its native Go type for comparison.
			result := lang.Unwrap(finalValue)

			// If a custom check function is provided, use it.
			if tc.checkFunc != nil {
				tc.checkFunc(t, result, err)
				return
			}

			// Otherwise, do a deep equal comparison on the result.
			if !reflect.DeepEqual(tc.wantResult, result) {
				t.Errorf("Unexpected result for test '%s'.\n- want: %#v (%T)\n-  got: %#v (%T)",
					tc.name, tc.wantResult, tc.wantResult, result, result)
			}
		})
	}
}
