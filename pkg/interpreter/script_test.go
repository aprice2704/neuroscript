// filename: pkg/interpreter/interpreter_script_test.go
// Neuroscript version: 0.5.2
// File version: 19
// Purpose: Refactored to use the centralized TestHarness for robust and consistent interpreter initialization.
package interpreter_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/google/go-cmp/cmp"
)

func TestInterpreterFixtures(t *testing.T) {
	root := filepath.Join("testdata", "interpreter")

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
		} else if strings.HasSuffix(fileName, ".ns") {
			name = strings.TrimSuffix(fileName, ".ns")
		} else {
			continue
		}

		t.Run(name, func(t *testing.T) {
			t.Logf("[DEBUG] Turn 1: Starting fixture test for '%s'.", name)
			scriptPath := filepath.Join(root, fileName)
			errPath := filepath.Join(root, name+".expect_err")
			goldenPath := filepath.Join(root, name+".golden.json")

			scriptBytes, err := os.ReadFile(scriptPath)
			if err != nil {
				t.Fatalf("failed to read script file %s: %v", scriptPath, err)
			}
			script := string(scriptBytes)

			h := NewTestHarness(t)
			parseTree, parseErr := h.Parser.Parse(script)
			t.Logf("[DEBUG] Turn 2: Script parsed.")

			if _, statErr := os.Stat(errPath); statErr == nil {
				if parseErr == nil {
					t.Fatalf("expected a parser error, but got nil")
				}
				wantErrBytes, err := os.ReadFile(errPath)
				if err != nil {
					t.Fatalf("failed to read expected error file %s: %v", errPath, err)
				}
				wantErrMsg := strings.TrimSpace(string(wantErrBytes))
				if !strings.Contains(parseErr.Error(), wantErrMsg) {
					t.Fatalf("error mismatch:\n  want: contains %q\n   got: %q", wantErrMsg, parseErr.Error())
				}
				t.Logf("[DEBUG] Turn 3: Correctly received expected parser error.")
				return
			}
			if parseErr != nil {
				t.Fatalf("unexpected PARSER error: %v", parseErr)
			}

			programAST, _, buildErr := h.ASTBuilder.Build(parseTree)
			if buildErr != nil {
				t.Fatalf("unexpected AST BUILD error: %v", buildErr)
			}
			t.Logf("[DEBUG] Turn 3: AST built.")

			if err := h.Interpreter.Load(&interfaces.Tree{Root: programAST}); err != nil {
				t.Fatalf("failed to load program into interpreter: %v", err)
			}
			t.Logf("[DEBUG] Turn 4: Program loaded.")

			var gotVal lang.Value
			var execErr error

			if len(programAST.Commands) > 0 {
				gotVal, execErr = h.Interpreter.Execute(programAST)
			} else if len(programAST.Procedures) >= 1 {
				procToRun := "main"
				if _, ok := programAST.Procedures["main"]; !ok {
					for pName := range programAST.Procedures {
						procToRun = pName
						break
					}
				}
				gotVal, execErr = h.Interpreter.Run(procToRun)
			} else {
				if _, statErr := os.Stat(goldenPath); os.IsNotExist(statErr) {
					t.Log("No golden file found, skipping execution for script with no commands or single function.")
					return
				}
				t.Fatalf("test script '%s' must contain either commands or at least one procedure for execution testing", name)
			}
			t.Logf("[DEBUG] Turn 5: Execution complete.")

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
				t.Fatalf("failed to unmarshal golden file %s: %v", goldenPath, err)
			}
			t.Logf("[DEBUG] Turn 6: Golden file loaded.")

			nativeGotVal := lang.Unwrap(gotVal)
			gotMap := map[string]any{"return": nativeGotVal}

			if diff := cmp.Diff(wantMap, gotMap); diff != "" {
				gotJSONBytes, _ := json.MarshalIndent(gotMap, "", "  ")
				t.Fatalf("result mismatch (-want +got):\n%s\n\nGot payload:\n%s", diff, gotJSONBytes)
			}
			t.Logf("[DEBUG] Turn 7: Assertion passed.")
		})
	}
}
