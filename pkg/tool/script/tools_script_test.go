// NeuroScript Version: 0.8.0
// File version: 9
// Purpose: Corrects the test to use the internal interpreter and its RunProcedure method directly, ensuring the ScriptHost interface is satisfied.
// filename: pkg/tool/script/tools_script_test.go
package script_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool/script"
	"github.com/google/go-cmp/cmp"
)

// TestScriptTools uses the file-based fixture runner to test the script-loading
// and introspection tools (`LoadScript`, `Script.ListFunctions`).
func TestScriptTools(t *testing.T) {
	root := filepath.Join("testdata")

	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("testdata directory not found, skipping: %s", root)
			return
		}
		t.Fatalf("failed to read testdata directory: %s: %v", root, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		fileName := e.Name()
		if !strings.HasSuffix(fileName, ".ns.txt") {
			continue
		}

		testName := strings.TrimSuffix(fileName, ".ns.txt")

		t.Run(testName, func(t *testing.T) {
			scriptPath := filepath.Join(root, fileName)
			errPath := filepath.Join(root, testName+".expect_err")
			goldenPath := filepath.Join(root, testName+".golden.json")

			scriptBytes, err := os.ReadFile(scriptPath)
			if err != nil {
				t.Fatalf("failed to read script file %s: %v", scriptPath, err)
			}

			hostCtx, err := api.NewHostContextBuilder().
				WithStdin(os.Stdin).
				WithStdout(os.Stdout).
				WithStderr(os.Stderr).
				Build()
			if err != nil {
				t.Fatalf("failed to build host context: %v", err)
			}
			execPolicy := &policy.ExecPolicy{
				Allow: []string{"tool.script.*"},
			}
			interp := interpreter.NewInterpreter(
				interpreter.WithHostContext(hostCtx),
				interpreter.WithExecPolicy(execPolicy),
			)

			for _, toolImpl := range script.ToolsToRegister {
				if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
					t.Fatalf("failed to register tool '%s': %v", toolImpl.Spec.Name, err)
				}
			}

			tree, err := api.Parse(scriptBytes, api.ParseSkipComments)
			if err != nil {
				t.Fatalf("failed to parse test driver script '%s': %v", fileName, err)
			}
			if err := interp.Load(tree); err != nil {
				t.Fatalf("failed to load ast for test driver script '%s': %v", fileName, err)
			}

			finalValue, execErr := interp.RunProcedure("main")

			if _, statErr := os.Stat(errPath); statErr == nil {
				if execErr == nil {
					t.Fatalf("expected an error, but got nil")
				}

				wantErrBytes, readErr := os.ReadFile(errPath)
				if readErr != nil {
					t.Fatalf("failed to read expected error file %s: %v", errPath, readErr)
				}
				expectedCodeStr := strings.TrimSpace(string(wantErrBytes))
				expectedCode, convErr := strconv.Atoi(expectedCodeStr)
				if convErr != nil {
					t.Fatalf("expected error file %s must contain an integer error code, got: %q", errPath, expectedCodeStr)
				}

				var runtimeErr *lang.RuntimeError
				if errors.As(execErr, &runtimeErr) {
					if runtimeErr.Code != lang.ErrorCode(expectedCode) {
						t.Fatalf("wrong error code returned:\n  want: %d\n   got: %d (%s)",
							expectedCode, runtimeErr.Code, runtimeErr.Message)
					}
				} else {
					t.Fatalf("expected a RuntimeError but got a different error type: %T, %v", execErr, execErr)
				}
				return
			}

			if execErr != nil {
				t.Fatalf("unexpected RUNTIME error during test execution: %v", execErr)
			}

			wantJSONBytes, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
			}
			var wantMap map[string]any
			if err := json.Unmarshal(wantJSONBytes, &wantMap); err != nil {
				t.Fatalf("failed to unmarshal golden file %s into map[string]any: %v", goldenPath, err)
			}

			nativeGotVal := lang.Unwrap(finalValue)

			gotMap := map[string]any{"return": nativeGotVal}

			if diff := cmp.Diff(wantMap, gotMap); diff != "" {
				gotJSONBytes, _ := json.MarshalIndent(gotMap, "", "  ")
				t.Fatalf("result mismatch (-want +got):\n%s\n\nGot payload:\n%s", diff, gotJSONBytes)
			}
		})
	}
}
