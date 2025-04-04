// pkg/tools/testing_helpers_test.go // Note: Filename differs from path
package core

import (
	"fmt"
	"io"  // Import io for discard logger
	"log" // Import log
	"reflect"
	"strings"
	"testing"
)

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

// runEvalExpressionTest executes a single expression evaluation test case.
func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	interp := newTestInterpreter(tc.InitialVars, tc.LastResult)
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

// newTestInterpreter creates an interpreter instance for testing,
// initializing variables, last result, and crucially, REGISTERING CORE TOOLS.
func newTestInterpreter(vars map[string]interface{}, lastResult interface{}) *Interpreter {
	// Use a discarding logger for tests unless explicitly needed otherwise
	testLogger := log.New(io.Discard, "[TEST-INTERP] ", log.Lshortfile)
	interp := NewInterpreter(testLogger) // Pass the logger

	// *** ADDED: Register core tools ***
	RegisterCoreTools(interp.ToolRegistry())
	// Add registrations for other necessary packages like checklist if needed for specific core tests
	// For now, just core tools.

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

	return interp
}

// newDefaultTestInterpreter creates a new interpreter with default settings
// and registers core tools.
func newDefaultTestInterpreter() *Interpreter {
	interp := NewInterpreter(log.New(io.Discard, "", 0)) // Discard logs by default
	// *** ADDED: Register core tools ***
	RegisterCoreTools(interp.ToolRegistry())
	// Add other registrations if needed by default test scenarios
	return interp
}
