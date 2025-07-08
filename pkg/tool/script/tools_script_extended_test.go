// NeuroScript Version: 0.5.4
// File version: 2
// Purpose: Extended test harness for edge cases in script tools. Fails on missing testdata.
// filename: pkg/tool/script/tools_script_extended_test.go
// nlines: 145
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

// TestScriptToolsExtended uses the file-based fixture runner to test edge cases
// for the script-loading and introspection tools.
func TestScriptToolsExtended(t *testing.T) {
	root := filepath.Join("testdata", "extended")

	entries, err := os.ReadDir(root)
	if err != nil {
		// FAIL instead of skipping if the directory doesn't exist.
		if os.IsNotExist(err) {
			t.Fatalf("extended testdata directory not found: %s", root)
		}
		t.Fatalf("failed to read testdata directory: %s: %v", root, err)
	}

	// FAIL if the directory was found but is empty.
	if len(entries) == 0 {
		t.Fatalf("extended testdata directory is empty: %s", root)
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
			scriptContent := string(scriptBytes)

			// --- Test Setup ---
			interp := interpreter.NewInterpreter()
			for _, toolImpl := range scriptToolsToRegister {
				if err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
					t.Fatalf("failed to register tool '%s': %v", toolImpl.Spec.Name, err)
				}
			}

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
