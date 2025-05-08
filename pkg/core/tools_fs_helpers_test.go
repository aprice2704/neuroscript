// NeuroScript Version: 0.3.1
// File version: 0.0.5 // Ensure tc is passed to compareFunc.
// nlines: 165 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_fs_helpers_test.go
package core

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fsTestCase defines the unified struct for all filesystem tool tests.
type fsTestCase struct {
	name          string
	toolName      string
	args          []interface{}
	wantResult    interface{}
	wantContent   string                         // Optional: For write tests
	wantToolErrIs error                          // Optional: Expect specific tool error
	setupFunc     func(sandboxRoot string) error // Optional setup for specific cases
	cleanupFunc   func(sandboxRoot string) error // Optional cleanup
}

// Define the type for the custom comparison function
// *** Ensure this signature includes tc fsTestCase ***
type compareFuncType func(t *testing.T, tc fsTestCase, expected, actual interface{})

// testFsToolHelper is the consolidated helper for testing FS tools.
func testFsToolHelper(t *testing.T, interp *Interpreter, tc fsTestCase) {
	t.Helper()
	testFsToolHelperInternal(t, interp, tc, nil) // Call internal helper without compare func
}

// testFsToolHelperWithCompare is a helper for testing FS tools using a custom comparison function.
func testFsToolHelperWithCompare(t *testing.T, interp *Interpreter, tc fsTestCase, compareFunc compareFuncType) {
	t.Helper()
	if compareFunc == nil {
		t.Fatal("testFsToolHelperWithCompare called with nil compareFunc")
	}
	testFsToolHelperInternal(t, interp, tc, compareFunc) // Call internal helper with compare func
}

// testFsToolHelperInternal is the core logic used by both public helpers.
func testFsToolHelperInternal(t *testing.T, interp *Interpreter, tc fsTestCase, compareFunc compareFuncType) {
	t.Helper()

	// --- Setup ---
	sandboxRoot := interp.SandboxDir()
	if sandboxRoot == "" {
		t.Fatalf("Interpreter sandbox directory is not set for test '%s'", tc.name)
	}
	if err := os.MkdirAll(sandboxRoot, 0755); err != nil {
		t.Fatalf("Failed to ensure sandbox directory '%s' exists: %v", sandboxRoot, err)
	}
	if tc.setupFunc != nil {
		if setupErr := tc.setupFunc(sandboxRoot); setupErr != nil {
			t.Fatalf("Setup failed for test '%s': %v", tc.name, setupErr)
		}
		t.Logf("[Setup Debug] Executed setup function for %s", tc.name)
	}

	// --- Cleanup (deferred) ---
	if tc.cleanupFunc != nil {
		t.Cleanup(func() {
			if err := tc.cleanupFunc(sandboxRoot); err != nil {
				t.Logf("Warning: Custom cleanup failed for test '%s': %v", tc.name, err)
			} else {
				t.Logf("[Cleanup Debug] Executed cleanup function for %s", tc.name)
			}
		})
	}

	// --- Tool Lookup ---
	t.Logf("[Helper Debug] Test %q: Attempting to look up tool with tc.toolName = %q", tc.name, tc.toolName)
	toolImpl, found := interp.GetTool(tc.toolName)
	if !found {
		t.Fatalf("Tool %q not found in registry", tc.toolName)
	}

	// --- Argument Preparation ---
	convertedArgs := tc.args

	// --- Execution ---
	t.Logf("[Helper Debug] Test %q: Executing tool function...", tc.name)
	gotResult, toolErr := toolImpl.Func(interp, convertedArgs)

	// --- Error Handling ---
	if tc.wantToolErrIs != nil {
		if toolErr == nil {
			t.Errorf("Test %q: Tool function expected error wrapping [%v], but got nil.", tc.name, tc.wantToolErrIs)
			return
		}
		if !errors.Is(toolErr, tc.wantToolErrIs) {
			expectedErrString := ""
			if tc.wantResult != nil {
				if s, ok := tc.wantResult.(string); ok {
					expectedErrString = s
				}
			}
			if expectedErrString == "" {
				expectedErrString = tc.wantToolErrIs.Error()
			}

			if !strings.Contains(toolErr.Error(), expectedErrString) {
				t.Errorf("Test %q: Tool function expected error wrapping [%v] OR message containing %q, but got type [%T] with value: %q", tc.name, tc.wantToolErrIs, expectedErrString, toolErr, toolErr.Error())
			} else {
				t.Logf("Test %q: Tool error message %q contained expected substring %q", tc.name, toolErr.Error(), expectedErrString)
			}
		} else {
			t.Logf("Test %q: Correctly found expected wrapped error: %v", tc.name, tc.wantToolErrIs)
		}
		if expectedMsgPart, ok := tc.wantResult.(string); ok && expectedMsgPart != "" {
			if !strings.Contains(toolErr.Error(), expectedMsgPart) {
				t.Errorf("Test %q: Tool function error message %q does not contain expected substring %q (errors.Is matched %v)", tc.name, toolErr.Error(), expectedMsgPart, tc.wantToolErrIs)
			}
		}
		if gotResult != nil {
			if _, ok := tc.wantResult.(string); !ok && tc.wantResult != nil {
				t.Logf("Test %q: Tool returned non-nil result (%#v) along with expected error (%v), allowed as wantResult was not nil/string.", tc.name, gotResult, toolErr)
			} else if tc.wantResult == nil || tc.wantResult == "" {
				t.Errorf("Test %q: Tool returned non-nil result (%#v) unexpectedly with error (%v)", tc.name, gotResult, toolErr)
			}
		}
		return
	}

	if toolErr != nil {
		t.Errorf("Test %q: Tool function unexpected error: %v. Result: %v (%T)", tc.name, toolErr, gotResult, gotResult)
		return
	}

	// --- Result Comparison ---
	t.Logf("[Helper Debug] Test %q: Comparing results...", tc.name)
	if compareFunc != nil {
		// *** Ensure tc is passed here ***
		compareFunc(t, tc, tc.wantResult, gotResult)
	} else {
		if !reflect.DeepEqual(gotResult, tc.wantResult) {
			gotStr, gotIsStr := gotResult.(string)
			wantStr, wantIsStr := tc.wantResult.(string)
			if gotIsStr && wantIsStr {
				t.Errorf("Test %q: Tool function result mismatch (strings):\n  Got:  %q (bytes: %x)\n  Want: %q (bytes: %x)",
					tc.name, gotStr, []byte(gotStr), wantStr, []byte(wantStr))
			} else {
				t.Errorf("Test %q: Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
					tc.name, gotResult, gotResult, tc.wantResult, tc.wantResult)
			}
		}
	}

	// --- File Content Verification (for WriteFile) ---
	if tc.wantContent != "" && tc.toolName == "FS.Write" {
		filePathRel, ok := tc.args[0].(string)
		if !ok || len(tc.args) < 1 {
			t.Errorf("Test %q: Cannot verify file content, first arg is not a string path", tc.name)
			return
		}
		verifyPath := filepath.Join(interp.SandboxDir(), filePathRel)
		contentBytes, err := os.ReadFile(verifyPath)
		if err != nil {
			t.Errorf("Test '%s': Failed to read back file '%s' after write: %v", tc.name, verifyPath, err)
		} else if string(contentBytes) != tc.wantContent {
			t.Errorf("Test '%s': File content mismatch for %s.\nGot:\n%q\nWant:\n%q", tc.name, verifyPath, string(contentBytes), tc.wantContent)
		}
	}
	t.Logf("[Helper Debug] Test %q: Finished.", tc.name)
}
