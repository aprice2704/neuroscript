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

// --- Filesystem Test Helper (Consolidated) ---

// fsTestCase defines the unified struct for all filesystem tool tests.
type fsTestCase struct {
	name          string
	toolName      string
	args          []interface{}
	wantResult    interface{}
	wantContent   string       // Optional: For write tests
	wantToolErrIs error        // Optional: Expect specific tool error
	valWantErrIs  error        // Optional: Expect specific validation error
	setupFunc     func() error // Optional setup for specific cases
	cleanupFunc   func() error // Optional cleanup
}

// Placed in pkg/core/tools_fs_helpers_test.go

// testFsToolHelper is the consolidated helper for testing FS tools.
// It now accepts sandboxDir to correctly verify file operations.
// It only checks errors.Is when an error is expected, ignoring the result value.
func testFsToolHelper(t *testing.T, interp *Interpreter, sandboxDir string, tc fsTestCase) {
	t.Helper()

	// --- Setup ---
	if tc.setupFunc != nil {
		if err := tc.setupFunc(); err != nil {
			t.Fatalf("Setup failed for test '%s': %v", tc.name, err)
		}
	}
	// --- Cleanup (deferred) ---
	if tc.cleanupFunc != nil {
		t.Cleanup(func() {
			if err := tc.cleanupFunc(); err != nil {
				t.Logf("Warning: Custom cleanup failed for test '%s': %v", tc.name, err)
			}
		})
	}

	// --- Tool Lookup & Validation ---
	t.Logf("[Helper Debug] Test %q: Attempting to look up tool with tc.toolName = %q", tc.name, tc.toolName)
	toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
	if !found {
		t.Fatalf("Tool %q not found in registry (check registration in setup)", tc.toolName)
	}
	spec := toolImpl.Spec
	convertedArgs, valErr := ValidateAndConvertArgs(spec, tc.args)

	// Check Specific Validation Error
	if tc.valWantErrIs != nil {
		if valErr == nil {
			t.Errorf("ValidateAndConvertArgs() expected error [%v], but got nil", tc.valWantErrIs)
		} else if !errors.Is(valErr, tc.valWantErrIs) {
			t.Errorf("ValidateAndConvertArgs() expected error type [%v], but got type [%T] with value: %v", tc.valWantErrIs, valErr, valErr)
		}
		// Stop test execution if a VALIDATION error occurred or was expected
		return
	}
	// Check for Unexpected Validation Error
	if valErr != nil && tc.valWantErrIs == nil {
		t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
	}

	// --- Execution ---
	gotResult, toolErr := toolImpl.Func(interp, convertedArgs) // Keep gotResult for unexpected error logging

	// --- Error Handling ---

	// Case 1: An error WAS expected
	if tc.wantToolErrIs != nil {
		if toolErr == nil {
			t.Errorf("Tool function expected error [%v], but got nil.", tc.wantToolErrIs) // Removed Result from log as it's ignored
		} else if !errors.Is(toolErr, tc.wantToolErrIs) {
			// Check if the error message contains the expected error's message as a fallback
			if !strings.Contains(toolErr.Error(), tc.wantToolErrIs.Error()) {
				t.Errorf("Tool function expected error type [%v] or containing string %q, but got type [%T] with value: %v", tc.wantToolErrIs, tc.wantToolErrIs.Error(), toolErr, toolErr)
			} else {
				t.Logf("Tool error message contained expected string %q: Got %v", tc.wantToolErrIs.Error(), toolErr)
			}
		}
		// *** RETURN HERE: Do not proceed to result comparison if an error was expected ***
		return
	}

	// Case 2: An error was NOT expected, but one occurred
	if toolErr != nil { // Implicitly means tc.wantToolErrIs was nil here
		t.Errorf("Tool function unexpected error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
		// RETURN HERE: Do not proceed to result comparison if an unexpected error occurred
		return
	}

	// Case 3: No error was expected, and none occurred - Compare results
	// (This block is now only reached if tc.wantToolErrIs == nil AND toolErr == nil)

	// Special handling for ListDirectory results (sort before compare)
	if tc.toolName == "ListDirectory" {
		// Type assert and handle potential nil/type mismatch
		gotSlice, gotOk := gotResult.([]interface{})
		wantSlice, wantOk := tc.wantResult.([]interface{})

		if !gotOk || !wantOk {
			// Allow comparison if both are nil
			if !(gotResult == nil && tc.wantResult == nil) {
				t.Errorf("ListDirectory result/wantResult type mismatch, expected []interface{} or nil, got %#v (%T), want %#v (%T)", gotResult, gotResult, tc.wantResult, tc.wantResult)
			}
		} else {
			// Sort slices based on 'name' field for stable comparison
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
			// Compare sorted results
			if !reflect.DeepEqual(gotSlice, wantSlice) {
				t.Errorf("Tool function result mismatch (after sorting):\n  Got:  %#v\n  Want: %#v", gotSlice, wantSlice)
			}
		}
	} else if !reflect.DeepEqual(gotResult, tc.wantResult) { // Standard comparison for other tools
		t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
			gotResult, gotResult, tc.wantResult, tc.wantResult)
	}

	// --- File Content Verification (for WriteFile) ---
	// Use sandboxDir passed into the helper
	if tc.wantContent != "" && tc.toolName == "WriteFile" {
		filePathRel := tc.args[0].(string)
		contentBytes, err := os.ReadFile(filepath.Join(sandboxDir, filePathRel))
		if err != nil {
			t.Errorf("Test '%s': Failed to read back file '%s' after write: %v", tc.name, filePathRel, err)
		} else if string(contentBytes) != tc.wantContent {
			t.Errorf("Test '%s': File content mismatch for %s.\nGot:\n%q\nWant:\n%q", tc.name, filePathRel, string(contentBytes), tc.wantContent)
		}
	}
} // End testFsToolHelper
