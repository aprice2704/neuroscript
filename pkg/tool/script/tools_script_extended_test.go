// NeuroScript Version: 0.8.0
// File version: 11
// Purpose: Fixes test failures by using the concrete interpreter which now implements the necessary interfaces, and corrects a syntax error in a test case.
// filename: pkg/tool/script/tools_script_extended_test.go
package script_test

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
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
			set _ = tool.script.LoadScript("")
		endfunc
		`,
		wantExecErrCode: lang.ErrorCodeToolExecutionFailed,
	},
	{
		name: "load_script_with_syntax_error",
		script: `
		:: title: Test that loading a script with a syntax error fails gracefully.
		func main() means
			set _ = tool.script.LoadScript("FUNKY CHICKEN")
		endfunc
		`,
		wantExecErrCode: lang.ErrorCodeSyntax,
	},
	{
		name: "list_after_failed_load",
		script: `
		:: title: Test that a failed load does not alter the function list
		func first_func() means
			return 1
		endfunc
		func main() means
			on error do
				clear_error
			endon
			set _ = tool.script.LoadScript("FUNK bad() means endfunc")
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			logger := logging.NewTestLogger(t)
			hostCtx, err := api.NewHostContextBuilder().
				WithLogger(logger).
				WithStdin(os.Stdin).
				WithStdout(os.Stdout).
				WithStderr(os.Stderr).
				Build()
			if err != nil {
				t.Fatalf("failed to build host context: %v", err)
			}

			execPolicy := &policy.ExecPolicy{Allow: []string{"tool.script.*"}}
			interp := interpreter.NewInterpreter(
				interpreter.WithHostContext(hostCtx),
				interpreter.WithExecPolicy(execPolicy),
			)

			// for _, toolImpl := range script.ToolsToRegister {
			// 	if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
			// 		t.Fatalf("failed to register tool '%s': %v", toolImpl.Spec.Name, err)
			// 	}
			// }

			tree, err := api.Parse([]byte(tc.script), api.ParseSkipComments)
			if err != nil {
				t.Fatalf("failed to parse test driver script '%s': %v", tc.name, err)
			}

			if err := interp.Load(tree); err != nil {
				t.Fatalf("failed to load ast for test driver script '%s': %v", tc.name, err)
			}

			finalValue, err := interp.RunProcedure("main")

			if tc.wantExecErrCode != 0 {
				if err == nil {
					t.Fatalf("Expected an error with code %d, but got nil", tc.wantExecErrCode)
				}
				var runtimeErr *lang.RuntimeError
				if !errors.As(err, &runtimeErr) {
					t.Fatalf("Expected a *lang.RuntimeError, but got %T: %v", err, err)
				}
				if runtimeErr.Code != tc.wantExecErrCode {
					t.Fatalf("Expected error code %d, but got %d. Message: %s", tc.wantExecErrCode, runtimeErr.Code, runtimeErr.Message)
				}
				return
			}

			if err != nil {
				t.Fatalf("Test script execution failed unexpectedly for '%s': %v", tc.name, err)
			}

			result := lang.Unwrap(finalValue)

			if tc.checkFunc != nil {
				tc.checkFunc(t, result, err)
				return
			}

			if !reflect.DeepEqual(tc.wantResult, result) {
				t.Errorf("Unexpected result for test '%s'.\n- want: %#v (%T)\n-  got: %#v (%T)",
					tc.name, tc.wantResult, tc.wantResult, result, result)
			}
		})
	}
}
