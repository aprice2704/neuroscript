// NeuroScript Version: 0.5.4
// File version: 1
// Purpose: Test harness for the script-loading and introspection tools.
// filename: pkg/tool/script/tools_script_test.go
// nlines: 140
// risk_rating: LOW
package script

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/google/go-cmp/cmp"
)

// TestScriptTools uses the file-based fixture runner to test the script-loading
// and introspection tools (`LoadScript`, `Script.ListFunctions`).
func TestScriptTools(t *testing.T) {
	// The testdata directory must be relative to the test file.
	// It should contain the .ns.txt scripts and their corresponding
	// .golden.json or .expect_err files.
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
			// Setup paths for the script and its expected output/error
			scriptPath := filepath.Join(root, fileName)
			errPath := filepath.Join(root, testName+".expect_err")
			goldenPath := filepath.Join(root, testName+".golden.json")

			scriptBytes, err := os.ReadFile(scriptPath)
			if err != nil {
				t.Fatalf("failed to read script file %s: %v", scriptPath, err)
			}
			scriptContent := string(scriptBytes)

			// --- Test Setup ---
			// Create a real interpreter instance. This interpreter must be correctly
			// implementing the `scriptHost` interface for the type assertion in
			// the tool functions to succeed.
			interp := interpreter.NewInterpreter()

			// Manually register the script tools with the interpreter's registry.
			for _, toolImpl := range scriptToolsToRegister {
				if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
					t.Fatalf("failed to register tool '%s': %v", toolImpl.Spec.Name, err)
				}
			}

			// Parse the test script itself into an AST
			p := parser.NewParserAPI(interp.GetLogger())
			program, pErr := p.Parse(scriptContent)
			if pErr != nil {
				t.Fatalf("failed to parse test driver script '%s': %v", fileName, pErr)
			}

			astBuilder := parser.NewASTBuilder(interp.GetLogger())
			programAST, _, bErr := astBuilder.Build(program)
			if bErr != nil {
				t.Fatalf("failed to build ast for test driver script '%s': %v", fileName, bErr)
			}

			// --- Execute Test and Check Results ---
			finalValue, execErr := interp.LoadAndRun(programAST, "main")

			// This branch is for tests that are expected to fail.
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
				return // End error test case
			}

			// This branch is for tests that are expected to succeed.
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

			// The result of the NeuroScript execution is a lang.Value, which needs to be unwrapped.
			nativeGotVal := lang.Unwrap(finalValue)

			// The golden files are structured with a "return" key holding the value.
			gotMap := map[string]any{"return": nativeGotVal}

			if diff := cmp.Diff(wantMap, gotMap); diff != "" {
				gotJSONBytes, _ := json.MarshalIndent(gotMap, "", "  ")
				t.Fatalf("result mismatch (-want +got):\n%s\n\nGot payload:\n%s", diff, gotJSONBytes)
			}
		})
	}
}
