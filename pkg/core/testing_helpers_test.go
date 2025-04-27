// filename: pkg/core/testing_helpers_test.go
package core

import (
	"errors" // Import errors package for errors.Is
	"fmt"    // For io.Discard if logger setup remains here (may move to helpers.go)

	// If logger setup remains here
	"math" // Keep for deepEqualWithTolerance if defined here
	// "path/filepath" // Likely not needed here anymore
	"reflect" // Keep for deepEqual comparison
	"testing"
	// Assuming slogadapter is needed for logger setup if it remains here
	// It seems logger setup moved to helpers.go, so this might be removable later
)

// --- Interpreter Test Specific Helpers ---
// Removed NewTestInterpreter and NewDefaultTestInterpreter as they are in helpers.go

// --- Step Creation Helpers (Keep these as they were) ---
func createTestStep(typ string, target string, valueNode interface{}, argNodes []interface{}) Step {
	return newStep(typ, target, nil, valueNode, nil, argNodes)
}

func createIfStep(condNode interface{}, thenSteps []Step, elseSteps []Step) Step {
	return Step{Type: "if", Cond: condNode, Value: thenSteps, ElseValue: elseSteps}
}

func createWhileStep(condNode interface{}, bodySteps []Step) Step {
	return Step{Type: "while", Cond: condNode, Value: bodySteps}
}

func createForStep(loopVar string, collectionNode interface{}, bodySteps []Step) Step {
	return Step{Type: "for", Target: loopVar, Cond: collectionNode, Value: bodySteps}
}

// --- deepEqualWithTolerance function definition (Keep from previous steps) ---
const defaultTolerance = 1e-9

// deepEqualWithTolerance compares two values, allowing for a small tolerance
// when comparing float64 or int64/float64 combinations.
func deepEqualWithTolerance(a, b interface{}) bool {
	// If types differ, attempt numeric coercion if one is int and other is float
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		aFloat, aIsNum := toFloat64(a)
		bFloat, bIsNum := toFloat64(b)
		if aIsNum && bIsNum {
			// Compare floats with tolerance
			return math.Abs(aFloat-bFloat) < defaultTolerance
		}
		// If types differ and numeric coercion failed, they are not equal
		return false
	}

	// If types are the same, handle floats specifically
	// Need to check for nil type before calling Kind()
	if a != nil && reflect.TypeOf(a).Kind() == reflect.Float64 {
		// Assuming b is also float64 due to type check above
		aFloat := a.(float64)
		bFloat := b.(float64)
		return math.Abs(aFloat-bFloat) < defaultTolerance
	}

	// Use DeepEqual for all other types (including nil comparison)
	return reflect.DeepEqual(a, b)
}

// // Helper for deepEqualWithTolerance
// func toFloat64(v interface{}) (float64, bool) {
// 	switch val := v.(type) {
// 	case float64:
// 		return val, true
// 	case float32:
// 		return float64(val), true
// 	case int:
// 		return float64(val), true
// 	case int8:
// 		return float64(val), true
// 	case int16:
// 		return float64(val), true
// 	case int32:
// 		return float64(val), true
// 	case int64:
// 		return float64(val), true
// 	case uint:
// 		return float64(val), true
// 	case uint8:
// 		return float64(val), true
// 	case uint16:
// 		return float64(val), true
// 	case uint32:
// 		return float64(val), true
// 	case uint64:
// 		// Potential overflow, but okay for typical test values
// 		return float64(val), true
// 	default:
// 		return 0, false
// 	}
// }

// --- END deepEqualWithTolerance section ---

// runEvalExpressionTest executes a single expression evaluation test case.
// *** MODIFIED: Uses ExpectedErrorIs with errors.Is instead of ErrContains ***
func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	// Assuming NewDefaultTestInterpreter is exported from helpers.go now
	interp, _ := NewDefaultTestInterpreter(t)
	if tc.InitialVars != nil {
		for k, v := range tc.InitialVars {
			interp.variables[k] = v
		}
	}
	interp.lastCallResult = tc.LastResult
	got, err := interp.evaluateExpression(tc.InputNode)

	// Use deepEqualWithTolerance for comparison
	deepCompareFunc := deepEqualWithTolerance

	if tc.WantErr {
		if err == nil {
			t.Errorf("%s: Expected an error, but got nil", tc.Name)
			return
		}
		// *** Check for specific sentinel error using errors.Is ***
		if tc.ExpectedErrorIs != nil {
			if !errors.Is(err, tc.ExpectedErrorIs) {
				t.Errorf("%s: Expected error wrapping [%v], but got [%v]", tc.Name, tc.ExpectedErrorIs, err)
			}
		} else {
			// Optional: If no specific error is expected, maybe log a warning or just accept any error?
			// For now, if WantErr is true but ExpectedErrorIs is nil, we just verify *an* error occurred.
			t.Logf("%s: Got expected error: %v (No specific sentinel check provided)", tc.Name, err)
		}
		// *** Removed ErrContains check ***

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

// --- Struct Definitions ---
// *** MODIFIED: Added ExpectedErrorIs, removed ErrContains ***
type EvalTestCase struct {
	Name            string
	InputNode       interface{}
	InitialVars     map[string]interface{}
	LastResult      interface{}
	Expected        interface{}
	WantErr         bool
	ExpectedErrorIs error // Use this for errors.Is checks
	// ErrContains string // REMOVED
}

// *** MODIFIED: Removed CheckErrorIs (assume always use errors.Is if ExpectedError provided) ***
type ValidationTestCase struct {
	Name          string
	ToolName      string // Keep field
	InputArgs     []interface{}
	ExpectedArgs  []interface{} // Optional: Check converted args
	ExpectedError error         // Expected sentinel error from ValidateAndConvertArgs
	// CheckErrorIs  bool          // REMOVED
}

// --- Placeholders for other helpers potentially defined in original ---
// Ensure runValidationTestCases is defined *once* (here or elsewhere)
// *** NOTE: runValidationTestCases may also need updating if it checks ExpectedError ***
/*
func runValidationTestCases(t *testing.T, toolName string, testCases []ValidationTestCase) {
	t.Helper()
	// ... Check if tc.ExpectedError != nil ...
	// if tc.ExpectedError != nil && !errors.Is(actualError, tc.ExpectedError) {
	//     t.Errorf(...)
	// }
}
*/

// Ensure core errors are accessible if needed by helpers here
// Import 'errors' is now needed at the top of the file
var (
	_ = ErrValidationArgCount // Use _ to avoid unused variable errors if not used directly
	_ = ErrValidationRequiredArgNil
	_ = ErrValidationTypeMismatch
)

// --- END FILE ---
