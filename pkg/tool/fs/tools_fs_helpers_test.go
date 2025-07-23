// filename: pkg/tool/fs/tools_fs_helpers_test.go
package fs

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/testutil"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// fsTestCase defines the structure for a single filesystem tool test case.
type fsTestCase struct {
	name          string
	toolName      types.ToolName
	args          []interface{}
	setupFunc     func(sandboxRoot string) error
	checkFunc     func(t *testing.T, interp tool.Runtime, result interface{}, err error, setupCtx interface{})
	wantResult    interface{}
	wantContent   string
	wantToolErrIs error
}

// newFsTestInterpreter creates a self-contained interpreter with a sandbox for fs tool testing.
func newFsTestInterpreter(t *testing.T) *interpreter.Interpreter {
	t.Helper()
	// Use the centralized helper to get the sandbox option.
	sandboxOpt := testutil.NewTestSandbox(t)
	interp := interpreter.NewInterpreter(sandboxOpt)

	for _, toolImpl := range fsToolsToRegister {
		if err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
			t.Fatalf("Failed to register tool '%s': %v", toolImpl.Spec.Name, err)
		}
	}
	return interp
}

// testFsToolHelper provides a generic runner for fsTestCase tests.
// The signature is reverted to accept an interpreter, fixing the cascade.
func testFsToolHelper(t *testing.T, interp *interpreter.Interpreter, tc fsTestCase) {
	t.Helper()

	sandboxRoot := interp.SandboxDir()
	if tc.setupFunc != nil {
		if err := tc.setupFunc(sandboxRoot); err != nil {
			t.Fatalf("Setup function failed for test '%s': %v", tc.name, err)
		}
	}

	fullname := types.MakeFullName(group, string(tc.toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullname)
	if !found {
		t.Fatalf("Tool %q not found in registry", tc.toolName)
	}

	result, err := toolImpl.Func(interp, tc.args)

	if tc.checkFunc != nil {
		tc.checkFunc(t, interp, result, err, nil)
		return
	}

	if tc.wantToolErrIs != nil {
		if !errors.Is(err, tc.wantToolErrIs) {
			t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantToolErrIs, err)
		}
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err == nil {
		if tc.wantResult != nil {
			if !reflect.DeepEqual(result, tc.wantResult) {
				t.Errorf("Result mismatch.\nGot:    %#v\nWanted: %#v", result, tc.wantResult)
			}
		}

		if tc.wantContent != "" {
			if len(tc.args) > 0 {
				if path, ok := tc.args[0].(string); ok {
					absPath := filepath.Join(sandboxRoot, path)
					content, readErr := os.ReadFile(absPath)
					if readErr != nil {
						t.Fatalf("Failed to read file for content check '%s': %v", absPath, readErr)
					}
					if string(content) != tc.wantContent {
						t.Errorf("Content mismatch for file '%s'.\nGot:\n%s\n\nWanted:\n%s", path, string(content), tc.wantContent)
					}
				}
			}
		}
	}
}

// mustMkdir creates a directory and fails the test on error.
func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory '%s': %v", dir, err)
	}
}

// mustWriteFile writes a file and fails the test on error.
func mustWriteFile(t *testing.T, filename string, content string) {
	t.Helper()
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file '%s': %v", filename, err)
	}
}
