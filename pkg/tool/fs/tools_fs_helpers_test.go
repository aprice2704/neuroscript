// filename: pkg/tool/fs/tools_fs_helpers_test.go
package fs

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fsTestCase defines the structure for a single filesystem tool test case.
type fsTestCase struct {
	name		string
	toolName	string
	args		[]interface{}
	setupFunc	func(sandboxRoot string) error
	checkFunc	func(t *testing.T, interp *Interpreter, result interface{}, err error, setupCtx interface{})
	wantResult	interface{}
	wantContent	string	// New field to check file content
	wantToolErrIs	error
}

// testFsToolHelper runs a single filesystem tool test case.
func testFsToolHelper(t *testing.T, interp *Interpreter, tc fsTestCase) {
	t.Helper()

	sandboxRoot := interp.SandboxDir()
	if sandboxRoot == "" {
		t.Fatal("Interpreter provided to testFsToolHelper has no sandbox directory set.")
	}

	if tc.setupFunc != nil {
		if err := tc.setupFunc(sandboxRoot); err != nil {
			t.Fatalf("Setup function failed for test '%s': %v", tc.name, err)
		}
	}

	toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
	if !found {
		t.Fatalf("Tool %q not found in registry", tc.toolName)
	}

	result, err := toolImpl.Func(interp, tc.args)

	// Custom check function takes precedence
	if tc.checkFunc != nil {
		tc.checkFunc(t, interp, result, err, nil)
		return
	}

	// Standard error and result checking
	if tc.wantToolErrIs != nil {
		if !errors.Is(err, tc.wantToolErrIs) {
			t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantToolErrIs, err)
		}
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if err == nil {
		if tc.wantResult != nil {
			// A more detailed comparison might be needed here, especially for maps/slices
			if !reflect.DeepEqual(result, tc.wantResult) {
				t.Errorf("Result mismatch.\nGot:    %#v\nWanted: %#v", result, tc.wantResult)
			}
		}

		// New check for file content
		if tc.wantContent != "" {
			// Assume the first argument is the file path
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

// testFsToolHelperWithCompare is a variant that uses a custom comparison function.
func testFsToolHelperWithCompare(t *testing.T, interp *Interpreter, tc fsTestCase, compareFunc func(t *testing.T, tc fsTestCase, expected, actual interface{})) {
	t.Helper()

	sandboxRoot := interp.SandboxDir()
	if sandboxRoot == "" {
		t.Fatal("Interpreter provided to testFsToolHelperWithCompare has no sandbox directory set.")
	}

	if tc.setupFunc != nil {
		if err := tc.setupFunc(sandboxRoot); err != nil {
			t.Fatalf("Setup function failed for test '%s': %v", tc.name, err)
		}
	}

	toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
	if !found {
		t.Fatalf("Tool %q not found in registry", tc.toolName)
	}

	result, err := toolImpl.Func(interp, tc.args)

	if tc.wantToolErrIs != nil {
		if !errors.Is(err, tc.wantToolErrIs) {
			t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantToolErrIs, err)
		}
		if err != nil && tc.wantResult != nil {
			if errMsg, ok := tc.wantResult.(string); ok {
				if !strings.Contains(err.Error(), errMsg) {
					t.Errorf("Expected error message to contain %q, but got %q", errMsg, err.Error())
				}
			}
		}
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	} else if compareFunc != nil {
		compareFunc(t, tc, tc.wantResult, result)
	} else if tc.wantResult != nil {
		if !reflect.DeepEqual(result, tc.wantResult) {
			t.Errorf("Result mismatch.\nGot:    %#v\nWanted: %#v", result, tc.wantResult)
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

// NewTestInterpreterWithSandbox creates a test interpreter with a dedicated sandbox directory.
func NewTestInterpreterWithSandbox(t *testing.T, sandboxDir string) *Interpreter {
	t.Helper()
	interp, _ := NewDefaultTestInterpreter(t)
	err := interp.SetSandboxDir(sandboxDir)
	if err != nil {
		t.Fatalf("Failed to set sandbox dir in test helper: %v", err)
	}
	return interp
}