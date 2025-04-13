// filename: pkg/core/testing_helpers_test.go
package core

import (
	// Keep fmt
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
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
