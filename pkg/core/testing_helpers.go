// pkg/tools/testing_helpers_test.go
package core

import (
	"fmt"
	"reflect"
	"strings"
	"testing" // Needed if you add test helper functions like AssertNoError
)

// EvalTestCase defines the structure for testing expression evaluation.
type EvalTestCase struct {
	Name        string                 // Name of the test case
	InputNode   interface{}            // The AST expression node to evaluate
	InitialVars map[string]interface{} // Initial variable state for the interpreter
	LastResult  interface{}            // Optional: Initial 'LAST' value
	Expected    interface{}            // Expected result if evaluation succeeds
	WantErr     bool                   // Whether an error is expected during evaluation
	ErrContains string                 // Substring expected in the error message if WantErr is true
}

// You'll likely also need the runEvalExpressionTest helper function
// that uses this struct. Here's a plausible reconstruction based on
// how the tests seem to use it:

// runEvalExpressionTest executes a single expression evaluation test case.
func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper() // Marks this function as a test helper

	// Use the appropriate test interpreter setup function
	// interp := newTestInterpreterEval(tc.InitialVars, tc.LastResult) // Or use newTestInterpreter
	interp := newTestInterpreter(tc.InitialVars, tc.LastResult) // Use the consolidated helper

	// Evaluate the expression node
	got, err := interp.evaluateExpression(tc.InputNode)

	// Check error expectation
	if tc.WantErr {
		if err == nil {
			t.Errorf("%s: Expected an error, but got nil", tc.Name)
			return
		}
		if tc.ErrContains != "" && !strings.Contains(err.Error(), tc.ErrContains) {
			t.Errorf("%s: Expected error containing %q, got: %v", tc.Name, tc.ErrContains, err)
		}
		// Don't check result if error was expected
		return
	}

	// If no error was expected, but one occurred
	if err != nil {
		t.Errorf("%s: Unexpected error: %v", tc.Name, err)
		return
	}

	// Check the result if no error occurred
	if !reflect.DeepEqual(got, tc.Expected) {
		// Provide detailed output on mismatch
		t.Errorf("%s: Result mismatch.\nInput Node: %+v\nExpected:   %v (%T)\nGot:        %v (%T)",
			tc.Name, tc.InputNode, tc.Expected, tc.Expected, got, got)
	}
}

// func (d *dummyContext) Logger() *log.Logger {
// 	if d.logger != nil {
// 		return d.logger
// 	}
// 	// Return a discarding logger by default for tests
// 	return log.New(io.Discard, "", 0)
// }

// func (d *dummyContext) GetVectorIndex() map[string][]float32 {
// 	if d.vectorIdx == nil {
// 		d.vectorIdx = make(map[string][]float32)
// 	}
// 	return d.vectorIdx
// }

// func (d *dummyContext) SetVectorIndex(vi map[string][]float32) {
// 	d.vectorIdx = vi
// }

// // GenerateEmbedding provides a simple mock embedding for testing.
// func (d *dummyContext) GenerateEmbedding(text string) ([]float32, error) {
// 	if d.embedDim <= 0 {
// 		d.embedDim = 4
// 	} // Use a small default dim for tests

// 	embedding := make([]float32, d.embedDim)
// 	// Simple deterministic embedding based on text length/content for testing
// 	// Use a fixed seed based on text for reproducibility in tests
// 	var seed int64
// 	for _, r := range text {
// 		seed = (seed*31 + int64(r)) & 0xFFFFFFFF
// 	}
// 	if seed < 0 {
// 		seed = -seed
// 	}
// 	rng := rand.New(rand.NewSource(seed))

// 	norm := float32(0.0)
// 	for i := range embedding {
// 		val := rng.Float32()*2.0 - 1.0 // Values in [-1, 1]
// 		embedding[i] = val
// 		norm += val * val
// 	}

// 	// Normalize
// 	norm = float32(math.Sqrt(float64(norm)))
// 	if norm > 1e-6 {
// 		for i := range embedding {
// 			embedding[i] /= norm
// 		}
// 	} else if d.embedDim > 0 {
// 		embedding[0] = 1.0 // Avoid zero vector
// 	}

// 	return embedding, nil
// }

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
// ... (helpers remain the same) ...
func newTestInterpreter(vars map[string]interface{}, lastResult interface{}) *Interpreter {
	interp := NewInterpreter(nil) // Assumes nil logger for tests
	if vars != nil {
		interp.variables = make(map[string]interface{}, len(vars))
		for k, v := range vars {
			interp.variables[k] = v
		}
	} else {
		interp.variables = make(map[string]interface{})
	}
	interp.lastCallResult = lastResult // Use the specific field name
	return interp
}

// newDefaultTestInterpreter creates a new interpreter with default settings
// and a discarding logger, suitable for tests not needing specific setup.
func newDefaultTestInterpreter() *Interpreter {
	// NewInterpreter already sets up the logger to discard if nil is passed,
	// initializes the variable map with built-ins, and registers tools.
	interp := NewInterpreter(nil)
	return interp
}

// You might also want the makeArgs helper if moving this:
// func makeArgs(vals ...interface{}) []interface{} {
//  if vals == nil {
//      return []interface{}{}
//  }
//  return vals
// }
