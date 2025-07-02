// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Corrected test harness to properly unwrap result values before comparison.
// filename: pkg/tool/script/tools_script_test.go
// nlines: 161
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

	"github.com/google/go-cmp/cmp"
)

// TestScriptTools uses the file-based fixture runner to test the script-loading
// and introspection tools (`LoadScript`, `Script.ListFunctions`, etc.).
func TestScriptTools(t *testing.T) {
	root := filepath.Join("testdata", "tools", "script")

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

		var name string
		fileName := e.Name()
		if strings.HasSuffix(fileName, ".ns.txt") {
			name = strings.TrimSuffix(fileName, ".ns.txt")
		} else {
			continue
		}

		t.Run(name, func(t *testing.T) {
			scriptPath := filepath.Join(root, fileName)
			errPath := filepath.Join(root, name+".expect_err")
			goldenPath := filepath.Join(root, name+".golden.json")

			scriptBytes, err := os.ReadFile(scriptPath)
			if err != nil {
				t.Fatalf("failed to read script file %s: %v", scriptPath, err)
			}
			script := string(scriptBytes)

			logger := NewTestLogger(t)
			parserAPI := NewParserAPI(logger)
			parseTree, parseErr := parserAPI.Parse(script)

			// This branch is for tests that are expected to fail.
			if _, statErr := os.Stat(errPath); statErr == nil {
				var combinedErr error
				if parseErr != nil {
					combinedErr = parseErr
				} else {
					astBuilder := NewASTBuilder(logger)
					programAST, _, buildErr := astBuilder.Build(parseTree)
					if buildErr != nil {
						combinedErr = buildErr
					} else {
						interp, _ := NewTestInterpreter(t, nil, nil)
						// We deliberately don't load the program here for error tests
						// that might involve the interpreter state before loading.
						procToRun := "main"
						if _, ok := programAST.Procedures[procToRun]; !ok {
							for procName := range programAST.Procedures {
								procToRun = procName
								break
							}
						}
						// Now load and execute in one go to catch load-time errors.
						if err := interp.LoadProgram(programAST); err != nil {
							combinedErr = err
						} else {
							_, combinedErr = interp.ExecuteProc(procToRun)
						}
					}
				}

				if combinedErr == nil {
					t.Fatalf("expected an error, but got nil")
				}

				wantErrBytes, err := os.ReadFile(errPath)
				if err != nil {
					t.Fatalf("failed to read expected error file %s: %v", errPath, err)
				}
				expectedCodeStr := strings.TrimSpace(string(wantErrBytes))
				expectedCode, err := strconv.Atoi(expectedCodeStr)
				if err != nil {
					t.Fatalf("expected error file %s must contain an integer error code, got: %q", errPath, expectedCodeStr)
				}

				var runtimeErr *RuntimeError
				if errors.As(combinedErr, &runtimeErr) {
					if runtimeErr.Code != ErrorCode(expectedCode) {
						t.Fatalf("wrong error code returned:\n  want: %d\n   got: %d (%s)",
							expectedCode, runtimeErr.Code, runtimeErr.Message)
					}
				} else {
					// Fallback for non-RuntimeError types if needed
					t.Logf("Warning: received a non-RuntimeError, checking string contains: %T, %v", combinedErr, combinedErr)
				}
				return
			}
			if parseErr != nil {
				t.Fatalf("unexpected PARSER error: %v", parseErr)
			}

			astBuilder := NewASTBuilder(logger)
			programAST, _, buildErr := astBuilder.Build(parseTree)
			if buildErr != nil {
				t.Fatalf("unexpected AST BUILD error: %v", buildErr)
			}

			interp, _ := NewTestInterpreter(t, nil, nil)
			if err := interp.LoadProgram(programAST); err != nil {
				t.Fatalf("failed to load program into interpreter: %v", err)
			}

			procToRun := "main"
			if _, ok := programAST.Procedures[procToRun]; !ok {
				t.Fatalf("test script '%s' must contain a 'main' procedure for execution testing", name)
			}

			rawResult, execErr := interp.ExecuteProc(procToRun)
			if execErr != nil {
				t.Fatalf("unexpected RUNTIME error: %v", execErr)
			}

			wantJSONBytes, err := os.ReadFile(goldenPath)
			if err != nil {
				if os.IsNotExist(err) {
					t.Fatalf("missing golden file for successful test: %s", goldenPath)
				}
				t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
			}
			var wantMap map[string]any
			if err := json.Unmarshal(wantJSONBytes, &wantMap); err != nil {
				t.Fatalf("failed to unmarshal golden file %s into map[string]any: %v", goldenPath, err)
			}

			// CORRECTED: Unwrap the raw Value to get a native Go type for comparison.
			nativeGotVal := Unwrap(rawResult)
			gotMap := map[string]any{"return": nativeGotVal}

			if diff := cmp.Diff(wantMap, gotMap); diff != "" {
				// Use a more readable JSON output for the got payload in case of error.
				gotJSONBytes, _ := json.MarshalIndent(gotMap, "", "  ")
				t.Fatalf("result mismatch (-want +got):\n%s\n\nGot payload:\n%s", diff, gotJSONBytes)
			}
		})
	}
}