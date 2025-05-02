// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 20:53:03 PDT // Fix test logic for expected errors
// filename: pkg/core/tools_go_test.go

package core

import (
	"errors" // Import errors for checking
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	// If NewDefaultTestInterpreter or MakeArgs are in a testing sub-package:
	// "github.com/aprice2704/neuroscript/pkg/core/testing"
)

// Assume NewDefaultTestInterpreter, MakeArgs are defined elsewhere (e.g. testing_helpers.go)

// --- Test Go Mod Tidy ---
// func TestToolGoModTidy(t *testing.T) { ... } // Keep as is

// --- Test GoFmt ---
// func TestToolGoFmt(t *testing.T) { ... } // Keep as is

// --- Test GoListPackages ---
// func TestToolGoListPackages(t *testing.T) { ... } // Keep as is

// --- Helper struct field for tests requiring setup ---
type goGetModuleInfoTestCase struct {
	name             string
	dirArg           interface{}
	wantResult       map[string]interface{}                                                          // Expected map or nil if error expected
	wantErrResultNil bool                                                                            // Expect nil result map? (Used for "not found" case OR error cases)
	wantErrIs        error                                                                           // Expect a specific Go error type (using errors.Is)? (e.g., ErrPathViolation)
	valShouldFail    bool                                                                            // Expect validation (ValidateAndConvertArgs) to fail?
	setupFunc        func(t *testing.T, interp *Interpreter, sandboxRoot, goModPath, subDir *string) // Func to setup test state
	cleanupFunc      func(t *testing.T)                                                              // Optional cleanup
}

// +++ ADDED Test for GoGetModuleInfo +++
func TestToolGoGetModuleInfo(t *testing.T) {
	// --- Initial Setup (run once for most tests) ---
	baseInterp, baseSandboxRoot := NewDefaultTestInterpreter(t)
	if baseInterp == nil || baseSandboxRoot == "" {
		t.Fatal("Failed to create initial test interpreter or sandbox root")
	}

	moduleName := "example.com/modinfo_test"
	goVersion := "1.21"
	reqPath := "example.com/othermod"
	reqVersion := "v1.2.3"
	repOldPath := "example.com/old"
	repNewPath := "../local/new" // Use relative path for replace

	// --- Create Default go.mod Content ---
	goModContent := fmt.Sprintf("module %s\n\ngo %s\n\nrequire %s %s\n\nreplace %s => %s\n",
		moduleName, goVersion, reqPath, reqVersion, repOldPath, repNewPath)

	// --- Function to set up the sandbox filesystem (go.mod, subdir) ---
	setupSandboxFS := func(t *testing.T, sandboxDir string) (goModPath, subDirPath string) {
		goModP := filepath.Join(sandboxDir, "go.mod")
		err := os.WriteFile(goModP, []byte(goModContent), 0644)
		if err != nil {
			t.Fatalf("Setup: Failed to write go.mod: %v", err)
		}
		subDirP := filepath.Join(sandboxDir, "subdir")
		err = os.Mkdir(subDirP, 0755)
		if err != nil {
			t.Fatalf("Setup: Failed to create subdir: %v", err)
		}
		return goModP, subDirP
	}

	// --- Initial Sandbox FS Setup ---
	baseGoModPath, baseSubDirPath := setupSandboxFS(t, baseSandboxRoot)

	// --- Test Cases Definition ---
	tests := []goGetModuleInfoTestCase{ // Use the helper struct
		{
			name:   "From root dir (.)",
			dirArg: ".",
			wantResult: map[string]interface{}{
				"modulePath": moduleName,
				"goVersion":  goVersion,
				"rootDir":    baseSandboxRoot, // Expect absolute path to dir containing go.mod
				"requires": []map[string]interface{}{
					{"path": reqPath, "version": reqVersion, "indirect": false},
				},
				"replaces": []map[string]interface{}{
					{"oldPath": repOldPath, "oldVersion": "", "newPath": repNewPath, "newVersion": ""},
				},
			},
			wantErrResultNil: false,
			wantErrIs:        nil, // Expect successful execution
		},
		{
			name:   "From subdir",
			dirArg: "subdir", // Relative to sandbox root
			wantResult: map[string]interface{}{ // Expect same result as finding from root
				"modulePath": moduleName,
				"goVersion":  goVersion,
				"rootDir":    baseSandboxRoot,
				"requires": []map[string]interface{}{
					{"path": reqPath, "version": reqVersion, "indirect": false},
				},
				"replaces": []map[string]interface{}{
					{"oldPath": repOldPath, "oldVersion": "", "newPath": repNewPath, "newVersion": ""},
				},
			},
			wantErrResultNil: false,
			wantErrIs:        nil, // Expect successful execution
		},
		{
			name:   "Nil directory arg (defaults to root)",
			dirArg: nil,
			wantResult: map[string]interface{}{ // Expect same result as finding from root
				"modulePath": moduleName,
				"goVersion":  goVersion,
				"rootDir":    baseSandboxRoot,
				"requires": []map[string]interface{}{
					{"path": reqPath, "version": reqVersion, "indirect": false},
				},
				"replaces": []map[string]interface{}{
					{"oldPath": repOldPath, "oldVersion": "", "newPath": repNewPath, "newVersion": ""},
				},
			},
			wantErrResultNil: false,
			wantErrIs:        nil, // Expect successful execution
		},
		{
			name:             "Directory outside sandbox",
			dirArg:           "../outside",
			wantResult:       nil,              // Result map should be nil on error
			wantErrResultNil: true,             // Expect nil result due to path error
			wantErrIs:        ErrPathViolation, // <<-- EXPECT THIS ERROR TYPE
			valShouldFail:    false,            // Validation passes, execution fails path check
		},
		{
			name:             "Directory with no go.mod above it (root)",
			dirArg:           "/",              // Search from filesystem root (will fail security check first)
			wantResult:       nil,              // Result map should be nil on error
			wantErrResultNil: true,             // Expect nil result because path invalid
			wantErrIs:        ErrPathViolation, // <<-- EXPECT THIS ERROR TYPE
			valShouldFail:    false,            // Path check happens during execution, not validation
		},
		{
			// Test for the case where go.mod is not found *within* the sandbox
			name:   "Go.mod not found within sandbox",
			dirArg: ".", // Start search from root
			setupFunc: func(t *testing.T, interp *Interpreter, sandboxRoot, goModPath, subDir *string) {
				// Modify state for this test: remove go.mod
				err := os.Remove(*goModPath)
				if err != nil && !errors.Is(err, os.ErrNotExist) { // Ignore if already gone
					t.Fatalf("Setup: Failed to remove go.mod: %v", err)
				}
			},
			wantResult:       nil,  // Result should be nil
			wantErrResultNil: true, // Expect nil result map (tool returns nil, nil for not found)
			wantErrIs:        nil,  // <<-- EXPECT NO GO ERROR (tool returns nil, nil)
			valShouldFail:    false,
		},
		{
			name:             "Validation: Wrong arg type",
			dirArg:           123,
			wantResult:       nil,
			wantErrResultNil: true,                      // Expect nil result due to validation error
			wantErrIs:        ErrValidationTypeMismatch, // <<-- EXPECT THIS ERROR TYPE
			valShouldFail:    true,                      // Expect validation itself to fail
		},
	}

	// --- Get Tool Spec ---
	toolImpl, found := baseInterp.ToolRegistry().GetTool("GoGetModuleInfo")
	if !found {
		t.Fatalf("Tool %q not found in registry", "GoGetModuleInfo")
	}
	spec := toolImpl.Spec

	// --- Run Tests ---
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the base interpreter and sandbox by default
			currentInterp := baseInterp
			currentSandboxRoot := baseSandboxRoot
			currentGoModPath := baseGoModPath
			currentSubDirPath := baseSubDirPath

			// --- Per-test Setup (if needed) ---
			if tt.setupFunc != nil {
				// Create a NEW sandbox for tests that modify state to avoid interference
				//	var err error
				currentInterp, currentSandboxRoot = NewDefaultTestInterpreter(t) // Get new sandbox path
				if currentInterp == nil || currentSandboxRoot == "" {
					t.Fatal("Setup: Failed to create new test interpreter or sandbox root")
				}
				currentGoModPath, currentSubDirPath = setupSandboxFS(t, currentSandboxRoot) // Setup FS in new sandbox

				// Run the specific setup function for this test
				tt.setupFunc(t, currentInterp, &currentSandboxRoot, &currentGoModPath, &currentSubDirPath)
			}

			rawArgs := MakeArgs(tt.dirArg) // Only one argument

			// --- Validation ---
			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs)

			if tt.valShouldFail {
				// If validation was expected to fail...
				if valErr == nil {
					t.Errorf("ValidateAndConvertArgs() expected validation error, but got nil")
				} else if tt.wantErrIs != nil && !errors.Is(valErr, tt.wantErrIs) {
					// If we expect a specific validation error type, check it
					t.Errorf("ValidateAndConvertArgs() expected error containing [%v], but got error: %v", tt.wantErrIs, valErr)
				} else if tt.wantErrIs == nil {
					// If we expected *some* validation error, but not a specific one
					t.Logf("Got expected validation error: %v", valErr)
				}
				// Check result is nil if validation failed
				// Note: tool function isn't called if validation fails, so result will be nil anyway.
				return // Stop test here if validation failed (expectedly or not)
			} else if valErr != nil {
				// If validation failed unexpectedly...
				t.Fatalf("ValidateAndConvertArgs() returned unexpected validation error: %v", valErr)
			}

			// --- Execution (only if validation passed) ---
			resultInterface, toolErr := toolGoGetModuleInfo(currentInterp, convertedArgs)

			// --- Check Expected Go Error ---
			if tt.wantErrIs != nil { // Check if we expect a specific error type
				if toolErr == nil {
					t.Errorf("toolGoGetModuleInfo() expected Go error containing [%v], but got nil error", tt.wantErrIs)
				} else if !errors.Is(toolErr, tt.wantErrIs) { // Use errors.Is for wrapped errors
					t.Errorf("toolGoGetModuleInfo() expected Go error containing [%v], but got error: %v", tt.wantErrIs, toolErr)
				} else {
					t.Logf("Got expected Go error: %v", toolErr) // Log success
				}
				// Check result interface based on wantErrResultNil when an error is expected
				if resultInterface != nil && tt.wantErrResultNil {
					t.Errorf("Expected nil result map due to expected error, but got: %#v", resultInterface)
				} else if resultInterface == nil && !tt.wantErrResultNil {
					t.Errorf("Expected non-nil result map despite expected error, but got nil")
				}
				return // Stop processing this test case after checking the expected error
			}

			// --- Check Unexpected Go Error ---
			// This block is only reached if tt.wantErrIs == nil
			if toolErr != nil {
				t.Fatalf("toolGoGetModuleInfo() returned unexpected Go error: %v", toolErr)
			}

			// --- Check Result Map (only if NO Go error occurred or was expected) ---

			// Check if result map should be nil (e.g., go.mod not found case where wantErrIs is nil)
			if tt.wantErrResultNil {
				if resultInterface != nil {
					t.Errorf("Expected nil result map (and no Go error), but got non-nil: %#v", resultInterface)
				}
				return // Stop if nil result was correctly found
			}

			// Check if result map is non-nil when expected
			if !tt.wantErrResultNil && resultInterface == nil {
				t.Fatalf("Expected non-nil map result, but got nil (and no Go error).")
			}

			// --- Type Assertion and Map Comparison ---
			gotMap, ok := resultInterface.(map[string]interface{})
			if !ok {
				t.Fatalf("toolGoGetModuleInfo() did not return map[string]interface{}, got %T", resultInterface)
			}

			// Ensure rootDir in wantResult matches the dynamic sandboxRoot for this specific test run
			// Must update the specific map instance for this test case if sandbox changed
			expectedResultMap := tt.wantResult
			if expectedResultMap != nil {
				expectedResultMap["rootDir"] = currentSandboxRoot // Use the correct sandbox for this test run
			}

			// Compare using reflect.DeepEqual for overall structure
			if !reflect.DeepEqual(gotMap, expectedResultMap) {
				// Provide more detailed diff if mismatch occurs
				// diff := pretty.Compare(expectedResultMap, gotMap) // Using go-pretty library if available
				// t.Errorf("Result map mismatch (-want +got):\n%s", diff)

				// Fallback to basic print
				t.Errorf("Result map mismatch:\nGot:  %#v\nWant: %#v", gotMap, expectedResultMap)
			}

			// --- Per-test Cleanup ---
			if tt.cleanupFunc != nil {
				tt.cleanupFunc(t)
			}
		})
	}
}

// +++ END Test +++
