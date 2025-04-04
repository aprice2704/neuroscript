// pkg/core/test_helpers_test.go
package core

import (
	"math"    // Added for float comparison
	"reflect" // Added for DeepEqual
	"strings" // Added for error substring check
	"testing" // Added for testing context (t.Helper, t.Errorf)
	// "log" // Import if logger needs configuration, currently using nil
	// "io"  // Import if logger needs configuration
)

// --- Shared Test Helper Functions ---

// newTestInterpreterEval creates an interpreter instance for evaluation tests.
// It initializes variables and the last call result.
func newTestInterpreterEval(vars map[string]interface{}, lastResult interface{}) *Interpreter {
	// Note: Using NewInterpreter(nil) to disable logging during tests by default.
	interp := NewInterpreter(nil) // Initialize with no logger for tests
	if vars != nil {
		// Copy initial vars to avoid modification across tests if the map is reused
		interp.variables = make(map[string]interface{}, len(vars))
		for k, v := range vars {
			interp.variables[k] = v
		}
	} else {
		interp.variables = make(map[string]interface{})
	}
	// Add built-in prompts to the variables map for consistency
	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = PromptExecute

	interp.lastCallResult = lastResult // Set the last call result
	return interp
}

// newDummyInterpreter creates a basic interpreter, often used for tool tests
// where initial variable state or last result isn't the focus.
func newDummyInterpreter() *Interpreter {
	// Using NewInterpreter(nil) to disable logging during tests by default.
	interp := NewInterpreter(nil)
	// No initial vars or last result set automatically
	// Add built-in prompts
	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = PromptExecute
	return interp
}

// makeArgs simplifies creating []interface{} slices for tool arguments in tests.
func makeArgs(vals ...interface{}) []interface{} {
	args := make([]interface{}, len(vals))
	copy(args, vals)
	return args
}

// --- NEW: Named Struct for Evaluation Test Cases ---
type EvalTestCase struct {
	Name        string                 // Name of the test case
	InputNode   interface{}            // The AST node to evaluate
	InitialVars map[string]interface{} // Initial variable state
	Expected    interface{}            // Expected result value
	WantErr     bool                   // Whether an error is expected
	ErrContains string                 // Substring expected in the error message
}

// --- MOVED & UPDATED: Helper function to run expression evaluation tests ---
func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()          // Marks this function as a test helper
	testName := tc.Name // Use name from the struct

	interp := newTestInterpreterEval(tc.InitialVars, nil) // Fresh interpreter for each test case
	got, err := interp.evaluateExpression(tc.InputNode)

	// Check error expectation
	if (err != nil) != tc.WantErr {
		t.Errorf("%q: error expectation mismatch. got err = %v, wantErr %v", testName, err, tc.WantErr)
		// Optional: Log interpreter state or node details on error mismatch
		// logNodeDetailsOnError(t, testName, tc.InputNode, interp)
		return // Avoid further checks if error expectation is wrong
	}

	// If error was expected, check if it contains the expected substring
	if tc.WantErr {
		if tc.ErrContains != "" && (err == nil || !strings.Contains(err.Error(), tc.ErrContains)) {
			t.Errorf("%q: expected error containing %q, got: %v", testName, tc.ErrContains, err)
		}
		// If error was expected, usually don't compare the 'got' value
	} else {
		// If no error was expected, compare the result value
		if fExp, okExp := tc.Expected.(float64); okExp {
			if fGot, okGot := got.(float64); okGot {
				// Use tolerance check for floats
				delta := math.Abs(fExp - fGot)
				tolerance := 1e-9 // Define a suitable tolerance
				if delta > tolerance {
					t.Errorf("%q: float result mismatch (tolerance %g):\nExpected: %v (%T)\nGot:      %v (%T)", testName, tolerance, tc.Expected, tc.Expected, got, got)
				}
			} else {
				// Type mismatch if expected float but got something else
				t.Errorf("%q: result type mismatch (expected float):\nExpected: %v (%T)\nGot:      %v (%T)", testName, tc.Expected, tc.Expected, got, got)
			}
		} else if !reflect.DeepEqual(got, tc.Expected) {
			// Standard DeepEqual for non-float types
			t.Errorf("%q: result mismatch:\nExpected: %v (%T)\nGot:      %v (%T)", testName, tc.Expected, tc.Expected, got, got)
		}
	}
}

// Optional helper to log details on failure (Example)
// func logNodeDetailsOnError(t *testing.T, testName string, node interface{}, interp *Interpreter) {
// 	t.Logf("[%s] Failing Node: %+v", testName, node)
// 	// Potentially log relevant parts of interpreter state
// 	t.Logf("[%s] Variables at failure: %+v", testName, interp.variables)
// }

// Note: Keep createTestStep, createIfStep, createWhileStep, createForStep,
//       and runExecuteStepsTest helpers local to interpreter_test.go as they
//       are specific to testing the step execution logic, not just expression evaluation.
