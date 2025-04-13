// filename: pkg/core/testing_helpers_test.go
package core

import (
	// *** ADDED for errors.Is ***
	"fmt"
	"io"
	"log" // *** ADDED for float comparison ***
	"path/filepath"
	"reflect" // *** ADDED for map iteration in tests ***
	"strings"
	"testing"
)

// --- Interpreter Test Specific Helpers ---

// newTestInterpreter creates an interpreter instance for testing,
// initializing variables, last result, registering core tools, and setting up sandbox.
// It now expects a *testing.T argument and sets the interpreter's sandboxDir.
// *** MODIFIED: Sets interpreter.sandboxDir, removes os.Chdir ***
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
	sandboxDirRel := t.TempDir() // This automatically handles cleanup via t.Cleanup()
	absSandboxDir, err := filepath.Abs(sandboxDirRel)
	if err != nil {
		t.Fatalf("Failed to get absolute path for sandbox %s: %v", sandboxDirRel, err)
	}

	// *** Store the absolute sandbox path in the interpreter using the field name you added ***
	interp.sandboxDir = absSandboxDir // Use the field name 'sandboxDir'
	testLogger.Printf("Sandbox root set in interpreter: %s", absSandboxDir)

	// *** REMOVED os.Chdir() ***

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

	return interp, absSandboxDir // Return interpreter and the sandbox path (path useful for setup funcs)
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
// (These helpers remain unchanged)
func createTestStep(typ string, target string, valueNode interface{}, argNodes []interface{}) Step {
	// Ensure Value and Args are distinct concepts in Step struct if needed
	// Assuming Step struct has fields like Type, Target, Value, Args
	return newStep(typ, target, nil, valueNode, nil, argNodes) // Pass nil for Cond, ElseValue if not applicable
}
func createIfStep(condNode interface{}, thenSteps []Step, elseSteps []Step) Step {
	var thenVal, elseVal interface{}
	if thenSteps != nil {
		thenVal = thenSteps
	}
	if elseSteps != nil {
		elseVal = elseSteps
	}
	// Assuming Step struct has fields Cond, Value (for then), ElseValue
	return Step{Type: "IF", Cond: condNode, Value: thenVal, ElseValue: elseVal}
}
func createWhileStep(condNode interface{}, bodySteps []Step) Step {
	var bodyVal interface{}
	if bodySteps != nil {
		bodyVal = bodySteps
	}
	// Assuming Step struct has fields Cond, Value (for body)
	return Step{Type: "WHILE", Cond: condNode, Value: bodyVal}
}
func createForStep(loopVar string, collectionNode interface{}, bodySteps []Step) Step {
	var bodyVal interface{}
	if bodySteps != nil {
		bodyVal = bodySteps
	}
	// Assuming Step struct has fields Target (loopVar), Cond (collection), Value (body)
	return Step{Type: "FOR", Target: loopVar, Cond: collectionNode, Value: bodyVal}
}

// runEvalExpressionTest executes a single expression evaluation test case.
// *** MODIFIED: Uses newDefaultTestInterpreter ***
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
// (makeArgs, AssertNoError remain unchanged)
func makeArgs(vals ...interface{}) []interface{} {
	if vals == nil {
		return []interface{}{}
	}
	return vals
}
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, but got: %v\nContext: %s", err, fmt.Sprint(msgAndArgs...))
	}
}

// --- Filesystem Test Helper (Consolidated) ---
// (testFsToolHelper needs updating in tools_fs_helpers_test.go - see next step)

// EvalTestCase defines the structure for testing expression evaluation.
// (Struct definition remains unchanged)
type EvalTestCase struct {
	Name        string
	InputNode   interface{}
	InitialVars map[string]interface{}
	LastResult  interface{}
	Expected    interface{}
	WantErr     bool
	ErrContains string
}
