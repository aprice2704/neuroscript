// filename: pkg/core/testing_helpers_test.go
package core

import (
	// Keep for errors.Is if used by runValidationTestCases later
	"fmt" // For io.Discard in logger setup
	// Original file used slog
	"math"    // Needed for deepEqualWithTolerance
	"reflect" // For reflect.DeepEqual used by runEvalExpressionTest
	"strings"
	"testing"
)

// --- Interpreter Test Specific Helpers ---

// NewDefaultTestInterpreter provides a convenience wrapper.
// *** KEPT: This version remains as it's test-specific ***
func NewDefaultTestInterpreter(t *testing.T) (*Interpreter, string) {
	t.Helper()
	return NewTestInterpreter(t, nil, nil)
}

// --- Step Creation Helpers (Keep these as they were in original file) ---
func createTestStep(typ string, target string, valueNode interface{}, argNodes []interface{}) Step {
	// Ensure this uses the current Step struct definition correctly
	// It likely calls the newStep helper from ast.go
	return newStep(typ, target, nil, valueNode, nil, argNodes)
}

func createIfStep(condNode interface{}, thenSteps []Step, elseSteps []Step) Step {
	// Ensure this creates a Step matching the struct in ast.go
	return Step{Type: "if", Cond: condNode, Value: thenSteps, ElseValue: elseSteps}
}

func createWhileStep(condNode interface{}, bodySteps []Step) Step {
	// Ensure this creates a Step matching the struct in ast.go
	return Step{Type: "while", Cond: condNode, Value: bodySteps}
}

func createForStep(loopVar string, collectionNode interface{}, bodySteps []Step) Step {
	// Ensure this creates a Step matching the struct in ast.go
	return Step{Type: "for", Target: loopVar, Cond: collectionNode, Value: bodySteps}
}

// --- ADDED deepEqualWithTolerance function definition ---
// deepEqualWithTolerance compares two values, allowing for a small tolerance
// when comparing float64 or int64/float64 combinations.
const defaultTolerance = 1e-9

func deepEqualWithTolerance(a, b interface{}) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		// Attempt numeric coercion if one is int and other is float
		aFloat, aIsNum := toFloat64(a)
		bFloat, bIsNum := toFloat64(b)
		if aIsNum && bIsNum {
			return math.Abs(aFloat-bFloat) < defaultTolerance
		}
		// If types differ and numeric coercion failed, they are not equal
		return false
	}

	// If types are the same, handle floats specifically
	if reflect.TypeOf(a).Kind() == reflect.Float64 {
		aFloat := a.(float64)
		bFloat := b.(float64)
		return math.Abs(aFloat-bFloat) < defaultTolerance
	}

	// Use DeepEqual for all other types
	return reflect.DeepEqual(a, b)
}

// --- END ADDED function ---

// runEvalExpressionTest executes a single expression evaluation test case.
// (Keep this as it was in the original file, it calls NewDefaultTestInterpreter)
func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	interp, _ := NewDefaultTestInterpreter(t) // Uses the corrected helper now
	if tc.InitialVars != nil {
		for k, v := range tc.InitialVars {
			interp.variables[k] = v
		}
	}
	interp.lastCallResult = tc.LastResult
	got, err := interp.evaluateExpression(tc.InputNode)

	// Use deepEqualWithTolerance for comparison if defined (should be in universal_test_helpers.go)
	// --- Now uses the locally defined version ---
	deepCompareFunc := deepEqualWithTolerance

	if tc.WantErr {
		if err == nil {
			t.Errorf("%s: Expected an error, but got nil", tc.Name)
			return
		}
		if tc.ErrContains != "" && !strings.Contains(err.Error(), tc.ErrContains) {
			t.Errorf("%s: Expected error containing %q, got: %v", tc.Name, tc.ErrContains, err)
		}
		// Optionally use errors.Is if comparing against sentinel errors
	} else { // No error wanted
		if err != nil {
			t.Errorf("%s: Unexpected error: %v", tc.Name, err)
		}
		// Use the comparison function
		if !deepCompareFunc(got, tc.Expected) {
			t.Errorf("%s: Result mismatch.\nInput Node: %+v\nExpected:   %v (%T)\nGot:        %v (%T)", tc.Name, tc.InputNode, tc.Expected, tc.Expected, got, got)

		}
	}
}

// --- General Test Helpers ---

// AssertNoError fails the test if err is not nil. (Keep from original)
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		message := fmt.Sprintf("Expected no error, but got: %v", err)
		if len(msgAndArgs) > 0 {
			format, ok := msgAndArgs[0].(string)
			if !ok {
				message += fmt.Sprintf("\nContext: %+v", msgAndArgs)
			} else {
				message += "\nContext: " + fmt.Sprintf(format, msgAndArgs[1:]...)
			}
		}
		t.Fatal(message)
	}
}

// --- Struct Definitions (Keep from original) ---
type EvalTestCase struct {
	Name        string
	InputNode   interface{}
	InitialVars map[string]interface{}
	LastResult  interface{}
	Expected    interface{}
	WantErr     bool
	ErrContains string
}

type ValidationTestCase struct {
	Name string
	// ToolName  string // Removed - ToolName is passed to helper
	InputArgs     []interface{}
	ExpectedError error // Expected sentinel error from ValidateAndConvertArgs
	// CheckErrorIs  bool // Removed - Assume errors.Is is always used
}

// --- Placeholders for other helpers potentially defined in original ---
// Ensure runValidationTestCases is defined *once* (here or elsewhere)
/*
func runValidationTestCases(t *testing.T, toolName string, testCases []ValidationTestCase) {
	t.Helper()
	interp, _ := NewDefaultTestInterpreter(t)
	// ... rest of implementation ...
}
*/

// Ensure core errors are accessible if needed by helpers here
var (
	_ = ErrValidationArgCount // Use _ to avoid unused variable errors if not used directly
	_ = ErrValidationRequiredArgNil
	_ = ErrValidationTypeMismatch
)

// --- END FILE ---
