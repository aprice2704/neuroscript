// filename: pkg/core/tools_fs_helpers_test.go
package core

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// fsTestCase defines the unified struct for all filesystem tool tests.
// *** MODIFIED: setupFunc and cleanupFunc now accept sandboxRoot string ***
type fsTestCase struct {
	name          string
	toolName      string
	args          []interface{}
	wantResult    interface{}
	wantContent   string                         // Optional: For write tests
	wantToolErrIs error                          // Optional: Expect specific tool error
	valWantErrIs  error                          // Optional: Expect specific validation error
	setupFunc     func(sandboxRoot string) error // Optional setup for specific cases
	cleanupFunc   func(sandboxRoot string) error // Optional cleanup
}

// testFsToolHelper is the consolidated helper for testing FS tools.
// *** MODIFIED: Removed os.Chdir logic, passes sandboxRoot to setup/cleanup funcs ***
func testFsToolHelper(t *testing.T, interp *Interpreter, tc fsTestCase) {
	t.Helper()

	// --- Setup ---
	sandboxRoot := interp.sandboxDir                       // Get sandbox from interpreter
	if err := os.MkdirAll(sandboxRoot, 0755); err != nil { // Ensure sandbox exists
		t.Fatalf("Failed to ensure sandbox directory '%s' exists: %v", sandboxRoot, err)
	}

	if tc.setupFunc != nil {
		// *** Execute setupFunc, passing the explicit sandboxRoot ***
		// *** REMOVED os.Chdir logic ***
		if setupErr := tc.setupFunc(sandboxRoot); setupErr != nil {
			t.Fatalf("Setup failed for test '%s': %v", tc.name, setupErr)
		}
		t.Logf("[Setup Debug] Executed setup function for %s", tc.name)
	}

	// --- Cleanup (deferred) ---
	if tc.cleanupFunc != nil {
		t.Cleanup(func() {
			// *** Execute cleanupFunc, passing the explicit sandboxRoot ***
			// *** REMOVED os.Chdir logic ***
			if err := tc.cleanupFunc(sandboxRoot); err != nil {
				t.Logf("Warning: Custom cleanup failed for test '%s': %v", tc.name, err)
			}
		})
	}

	// --- Tool Lookup & Validation ---
	// (Lookup and Validation logic remains the same)
	t.Logf("[Helper Debug] Test %q: Attempting to look up tool with tc.toolName = %q", tc.name, tc.toolName)
	toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
	if !found {
		t.Fatalf("Tool %q not found in registry", tc.toolName)
	}
	spec := toolImpl.Spec
	convertedArgs, valErr := ValidateAndConvertArgs(spec, tc.args)
	if tc.valWantErrIs != nil {
		if valErr == nil {
			t.Errorf("ValidateAndConvertArgs() expected error [%v], but got nil", tc.valWantErrIs)
		} else if !errors.Is(valErr, tc.valWantErrIs) {
			t.Errorf("ValidateAndConvertArgs() expected error type [%v], but got type [%T] with value: %v", tc.valWantErrIs, valErr, valErr)
		}
		return
	}
	if valErr != nil && tc.valWantErrIs == nil {
		t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
	}

	// --- Execution ---
	// Tool execution uses the interpreter which has the correct sandboxDir set
	gotResult, toolErr := toolImpl.Func(interp, convertedArgs)

	// --- Error Handling ---
	// (Error checking logic remains the same)
	if tc.wantToolErrIs != nil {
		if toolErr == nil {
			t.Errorf("Tool function expected error [%v], but got nil.", tc.wantToolErrIs)
		} else if !errors.Is(toolErr, tc.wantToolErrIs) {
			if !strings.Contains(toolErr.Error(), tc.wantToolErrIs.Error()) {
				t.Errorf("Tool function expected error type [%v] or containing string %q, but got type [%T] with value: %v", tc.wantToolErrIs, tc.wantToolErrIs.Error(), toolErr, toolErr)
			} else {
				t.Logf("Tool error message contained expected string %q: Got %v", tc.wantToolErrIs.Error(), toolErr)
			}
		}
		return
	}
	if toolErr != nil {
		t.Errorf("Tool function unexpected error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
		return
	}

	// --- Result Comparison ---
	// (Result comparison logic remains the same)
	if tc.toolName == "ListDirectory" {
		gotSlice, gotOk := gotResult.([]interface{})
		wantSlice, wantOk := tc.wantResult.([]interface{})
		if !gotOk || !wantOk {
			if !(gotResult == nil && tc.wantResult == nil) {
				t.Errorf("ListDirectory type mismatch, got %#v (%T), want %#v (%T)", gotResult, gotResult, tc.wantResult, tc.wantResult)
			}
		} else {
			sort.SliceStable(gotSlice, func(i, j int) bool {
				iMap, iOk := gotSlice[i].(map[string]interface{})
				jMap, jOk := gotSlice[j].(map[string]interface{})
				if !iOk || !jOk {
					return false
				}
				iName, iNameOk := iMap["name"].(string)
				jName, jNameOk := jMap["name"].(string)
				if !iNameOk || !jNameOk {
					return false
				}
				return iName < jName
			})
			sort.SliceStable(wantSlice, func(i, j int) bool {
				iMap, iOk := wantSlice[i].(map[string]interface{})
				jMap, jOk := wantSlice[j].(map[string]interface{})
				if !iOk || !jOk {
					return false
				}
				iName, iNameOk := iMap["name"].(string)
				jName, jNameOk := jMap["name"].(string)
				if !iNameOk || !jNameOk {
					return false
				}
				return iName < jName
			})
			if !reflect.DeepEqual(gotSlice, wantSlice) {
				t.Errorf("Tool function result mismatch (after sorting):\n  Got:  %#v\n  Want: %#v", gotSlice, wantSlice)
			}
		}
	} else if !reflect.DeepEqual(gotResult, tc.wantResult) {
		t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)", gotResult, gotResult, tc.wantResult, tc.wantResult)
	}

	// --- File Content Verification (for WriteFile) ---
	// *** Uses interp.sandboxDir - this part was already correct ***
	if tc.wantContent != "" && tc.toolName == "WriteFile" {
		filePathRel := tc.args[0].(string)
		// Construct path using the sandbox root stored in the interpreter
		verifyPath := filepath.Join(interp.sandboxDir, filePathRel)
		contentBytes, err := os.ReadFile(verifyPath)
		if err != nil {
			t.Errorf("Test '%s': Failed to read back file '%s' after write: %v", tc.name, verifyPath, err)
		} else if string(contentBytes) != tc.wantContent {
			t.Errorf("Test '%s': File content mismatch for %s.\nGot:\n%q\nWant:\n%q", tc.name, verifyPath, string(contentBytes), tc.wantContent)
		}
	}
} // End testFsToolHelper
