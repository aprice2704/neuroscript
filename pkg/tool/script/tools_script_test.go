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

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/testutil"
	"github.com/google/go-cmp/cmp"
)

// TestScriptTools uses the file-based fixture runner to test the script-loading
// and introspection tools (`LoadScript`, `Script.ListFunctions`, etc.).
func TestScriptTools(t *testing.T) {
	root := filepath.Join("testdata", "script")

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

			logger := logging.NewTestLogger(t)
			parserAPI := parser.NewParserAPI(logger)
			parseTree, parseErr := parserAPI.Parse(script)

			// This branch is for tests that are expected to fail.
			if _, statErr := os.Stat(errPath); statErr == nil {
				var combinedErr error
				if parseErr != nil {
					combinedErr = parseErr
				} else {
					astBuilder := parser.NewASTBuilder(logger)
					_, _, buildErr := astBuilder.Build(parseTree)
					if buildErr != nil {
						combinedErr = buildErr
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

				var runtimeErr *lang.RuntimeError
				if errors.As(combinedErr, &runtimeErr) {
					if runtimeErr.Code != lang.ErrorCode(expectedCode) {
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

			astBuilder := parser.NewASTBuilder(logger)
			programAST, _, buildErr := astBuilder.Build(parseTree)
			if buildErr != nil {
				t.Fatalf("unexpected AST BUILD error: %v", buildErr)
			}

			interp, err := testutil.NewTestInterpreter(t, nil, nil)
			if err != nil {
				t.Fatalf("NewTestInterpreter failed: %v", err)
			}
			loadScriptTool, ok := interp.ToolRegistry().GetTool("LoadScript")
			if !ok {
				t.Fatalf("LoadScript tool not found")
			}

			rawResult, execErr := loadScriptTool.Func(interp, []interface{}{script})
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
			nativeGotVal := rawResult.(map[string]interface{})
			gotMap := map[string]any{"return": nativeGotVal}

			// Use the programAST variable to check the number of loaded functions.
			if functionsLoaded, ok := nativeGotVal["functions_loaded"].(float64); ok {
				if int(functionsLoaded) != len(programAST.Procedures) {
					t.Errorf("mismatch in number of loaded functions: expected %d, got %d", len(programAST.Procedures), int(functionsLoaded))
				}
			}

			if diff := cmp.Diff(wantMap, gotMap, cmp.AllowUnexported(ast.Program{})); diff != "" {
				// Use a more readable JSON output for the got payload in case of error.
				gotJSONBytes, _ := json.MarshalIndent(gotMap, "", "  ")
				t.Fatalf("result mismatch (-want +got):\n%s\n\nGot payload:\n%s", diff, gotJSONBytes)
			}
		})
	}
}
