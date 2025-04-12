// filename: pkg/core/testing_helpers_test.go
package core

import (
	"errors" // Keep fmt
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings" // Keep strings for error message check
	"testing"
)

// --- Interpreter Test Specific Helpers ---

// newTestInterpreter creates an interpreter instance for testing,
// initializing variables, last result, registering core tools, and setting up sandbox.
// It now expects a *testing.T argument.
func newTestInterpreter(t *testing.T, vars map[string]interface{}, lastResult interface{}) (*Interpreter, string) {
	t.Helper()
	// Use a discarding logger for tests unless explicitly needed otherwise
	testLogger := log.New(io.Discard, "[TEST-INTERP] ", log.Lshortfile)
	// uncomment below to enable test logging
	// testLogger = log.New(os.Stderr, "[TEST-INTERP] ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
	interp := NewInterpreter(testLogger) // Pass the logger

	// *** ADDED: Log before registration attempt ***
	testLogger.Println("Attempting to register core tools...")

	// Register core tools and check for errors immediately
	if err := RegisterCoreTools(interp.ToolRegistry()); err != nil {
		// *** ADDED: Log details BEFORE failing ***
		testLogger.Printf("FATAL: Failed to register core tools during test setup: %v", err) // Log the error
		t.Fatalf("FATAL: Failed to register core tools during test setup: %v", err)          // Fail fast
	}
	// *** ADDED: Log success ***
	testLogger.Println("Successfully registered core tools.")

	// Create a temporary directory for sandboxing
	sandboxDir := t.TempDir() // This automatically handles cleanup via t.Cleanup()

	// Initialize variables if provided
	if vars != nil {
		// Start with built-ins already in interp.variables
		for k, v := range vars {
			interp.variables[k] = v
		}
	}
	// else: interp.variables already initialized with built-ins by NewInterpreter

	// Set last result if provided
	interp.lastCallResult = lastResult

	// Change working directory to the sandbox for the test's duration
	err := os.Chdir(sandboxDir) // Keep Chdir to sandbox
	if err != nil {
		t.Fatalf("Failed to change working directory to sandbox %s: %v", sandboxDir, err)
	}

	// Get absolute path for consistency if needed
	absSandboxDir, _ := filepath.Abs(".")
	return interp, absSandboxDir
}

// newDefaultTestInterpreter creates a new interpreter with default settings
// and registers core tools, setting up a sandbox directory.
// It now expects a *testing.T argument.
func newDefaultTestInterpreter(t *testing.T) (*Interpreter, string) {
	t.Helper()
	// Delegate to the more general helper with nil initial vars/last result
	return newTestInterpreter(t, nil, nil)
}

// --- Step Creation Helpers (Moved from interpreter_test.go) ---

// Helper functions to create Step structs for tests
func createTestStep(typ string, target string, valueNode interface{}, argNodes []interface{}) Step {
	return newStep(typ, target, nil, valueNode, nil, argNodes)
}
func createIfStep(condNode interface{}, thenSteps []Step, elseSteps []Step) Step {
	// Ensure Value and ElseValue are correctly typed ([]Step or nil)
	var thenVal, elseVal interface{}
	if thenSteps != nil {
		thenVal = thenSteps
	}
	if elseSteps != nil {
		elseVal = elseSteps
	}
	return Step{Type: "IF", Cond: condNode, Value: thenVal, ElseValue: elseVal}
}
func createWhileStep(condNode interface{}, bodySteps []Step) Step {
	var bodyVal interface{}
	if bodySteps != nil {
		bodyVal = bodySteps
	}
	return Step{Type: "WHILE", Cond: condNode, Value: bodyVal}
}
func createForStep(loopVar string, collectionNode interface{}, bodySteps []Step) Step {
	var bodyVal interface{}
	if bodySteps != nil {
		bodyVal = bodySteps
	}
	return Step{Type: "FOR", Target: loopVar, Cond: collectionNode, Value: bodyVal}
}

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

// testFsToolHelper is the consolidated helper for testing FS tools.
// It accepts the unified fsTestCase struct.
// *** MODIFIED: Removed unused vars and simplified result check ***
func testFsToolHelper(t *testing.T, interp *Interpreter, tc fsTestCase) {
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
		return
	}

	// Check for Unexpected Validation Error
	if valErr != nil && tc.valWantErrIs == nil {
		t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
	}

	// --- Execution ---
	gotResult, toolErr := toolImpl.Func(interp, convertedArgs)

	// Check Specific Tool Error
	if tc.wantToolErrIs != nil {
		if toolErr == nil {
			t.Errorf("Tool function expected error [%v], but got nil. Result: %v (%T)", tc.wantToolErrIs, gotResult, gotResult)
		} else if !errors.Is(toolErr, tc.wantToolErrIs) {
			if !strings.Contains(toolErr.Error(), tc.wantToolErrIs.Error()) {
				t.Errorf("Tool function expected error type [%v] or containing string %q, but got type [%T] with value: %v", tc.wantToolErrIs, tc.wantToolErrIs.Error(), toolErr, toolErr)
			}
		}
		// Compare result even if error was expected, if wantResult is non-nil
		if tc.wantResult != nil && !reflect.DeepEqual(gotResult, tc.wantResult) {
			gotStr, gotIsStr := gotResult.(string)
			wantStr, wantIsStr := tc.wantResult.(string)
			if gotIsStr && wantIsStr && strings.Contains(gotStr, wantStr) {
				t.Logf("Tool result error message contains expected string: Got %q, Want contains %q", gotStr, wantStr)
			} else {
				t.Errorf("Tool function result mismatch (even with expected error):\n  Got:  %#v (%T)\n  Want: %#v (%T)",
					gotResult, gotResult, tc.wantResult, tc.wantResult)
			}
		}
		return
	}

	// Check for Unexpected Tool Error
	if toolErr != nil && tc.wantToolErrIs == nil {
		t.Fatalf("Tool function unexpected error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
	}

	// --- Result Comparison (No Tool Error Expected) ---
	if tc.wantToolErrIs == nil {
		// Special handling for ListDirectory results (sort before compare)
		if tc.toolName == "ListDirectory" {
			gotSlice, gotOk := gotResult.([]interface{})
			wantSlice, wantOk := tc.wantResult.([]interface{})
			if !gotOk || !wantOk {
				if wantErrStr, ok := tc.wantResult.(string); ok {
					if gotStr, ok2 := gotResult.(string); ok2 && strings.Contains(gotStr, wantErrStr) {
						t.Logf("ListDirectory error message matched expected: %q", gotResult)
					} else {
						t.Errorf("ListDirectory result/wantResult type mismatch, expected []interface{} or error string %q, got %T", wantErrStr, gotResult)
					}
				} else {
					t.Errorf("ListDirectory result/wantResult type mismatch, expected []interface{}, got %T, want %T", gotResult, tc.wantResult)
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
				if !reflect.DeepEqual(gotSlice, wantSlice) {
					t.Errorf("Tool function result mismatch (after sorting):\n  Got:  %#v\n  Want: %#v", gotSlice, wantSlice)
				}
			}
		} else if !reflect.DeepEqual(gotResult, tc.wantResult) { // Standard comparison for other tools
			// *** MODIFIED: Removed unused string variables and string comparison logic ***
			t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
				gotResult, gotResult, tc.wantResult, tc.wantResult)
		}
	}

	// --- File Content Verification (for WriteFile) ---
	if tc.wantContent != "" && tc.toolName == "WriteFile" && tc.valWantErrIs == nil && tc.wantToolErrIs == nil {
		filePathRel := tc.args[0].(string)
		contentBytes, err := os.ReadFile(filePathRel)
		if err != nil {
			t.Errorf("Test '%s': Failed to read back file '%s' after write: %v", tc.name, filePathRel, err)
		} else if string(contentBytes) != tc.wantContent {
			t.Errorf("Test '%s': File content mismatch for %s. Got %q, want %q", tc.name, filePathRel, string(contentBytes), tc.wantContent)
		}
	}
} // End testFsToolHelper

// runEvalExpressionTest executes a single expression evaluation test case.
func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	// Use newDefaultTestInterpreter which now returns sandboxDir (ignored here)
	interp, _ := newDefaultTestInterpreter(t) // Ignore sandboxDir for eval tests
	// Initialize variables if provided
	if tc.InitialVars != nil {
		for k, v := range tc.InitialVars {
			interp.variables[k] = v
		}
	}
	// Set last result if provided
	interp.lastCallResult = tc.LastResult

	got, err := interp.evaluateExpression(tc.InputNode)

	if tc.WantErr {
		if err == nil {
			t.Errorf("%s: Expected an error, but got nil", tc.Name)
			return
		}
		if tc.ErrContains != "" && !strings.Contains(err.Error(), tc.ErrContains) {
			t.Errorf("%s: Expected error containing %q, got: %v", tc.Name, tc.ErrContains, err)
		}
		return
	}
	if err != nil {
		t.Errorf("%s: Unexpected error: %v", tc.Name, err)
		return
	}
	if !reflect.DeepEqual(got, tc.Expected) {
		t.Errorf("%s: Result mismatch.\nInput Node: %+v\nExpected:   %v (%T)\nGot:        %v (%T)",
			tc.Name, tc.InputNode, tc.Expected, tc.Expected, got, got)
	}
}

// --- General Test Helpers ---

// makeArgs simplifies creating []interface{} slices for tool arguments.
func makeArgs(vals ...interface{}) []interface{} {
	if vals == nil {
		return []interface{}{}
	}
	return vals
}

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, but got: %v\nContext: %s", err, fmt.Sprint(msgAndArgs...))
	}
}

// --- Interpreter Test Specific Helper ---
// *** REMOVED newTestInterpreter and newDefaultTestInterpreter - MOVED TO _test.go FILE ***

// --- Filesystem Test Helper (Consolidated) ---
// *** REMOVED testFsToolHelper - MOVED TO _test.go FILE ***

// EvalTestCase defines the structure for testing expression evaluation.
type EvalTestCase struct {
	Name        string
	InputNode   interface{}
	InitialVars map[string]interface{}
	LastResult  interface{}
	Expected    interface{}
	WantErr     bool
	ErrContains string
}
