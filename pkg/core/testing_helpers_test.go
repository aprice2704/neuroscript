// filename: pkg/core/testing_helpers_test.go
package core

import (
	"errors" // Import errors package for errors.Is
	"fmt"
	"math"    // Keep for deepEqualWithTolerance
	"reflect" // Keep for deepEqual comparison
	"sort"    // Import sort for variable checking
	"testing"
	// Logger/Adapter imports likely not needed if setup moved to helpers.go
)

// --- Interpreter Test Specific Helpers ---
// Removed NewTestInterpreter and NewDefaultTestInterpreter as they are in helpers.go

// --- Step Creation Helpers ---
// MODIFIED: Corrected call signature for newStep (removed final args parameter, used elseValueNode)
func createTestStep(typ string, target string, valueNode interface{}, elseValueNode interface{}) Step {
	// Calling the updated newStep which takes 5 arguments
	// Pass nil for 'cond' here; specific helpers like createIfStep set it.
	return newStep(typ, target, nil, valueNode, elseValueNode)
}

// createIfStep (Unchanged from last correct version)
func createIfStep(condNode interface{}, thenSteps []Step, elseSteps []Step) Step {
	var elseVal interface{} = nil
	if elseSteps != nil {
		elseVal = elseSteps // Assign the slice directly
	}
	return Step{
		Type:      "if",
		Cond:      condNode,
		Value:     thenSteps, // Then block steps
		ElseValue: elseVal,   // Else block steps (or nil)
		Metadata:  make(map[string]string),
	}
}

// createWhileStep (Unchanged from last correct version)
func createWhileStep(condNode interface{}, bodySteps []Step) Step {
	return Step{
		Type:     "while",
		Cond:     condNode,
		Value:    bodySteps, // Body steps
		Metadata: make(map[string]string),
	}
}

// createForStep (Unchanged from last correct version)
func createForStep(loopVar string, collectionNode interface{}, bodySteps []Step) Step {
	return Step{
		Type:     "for",
		Target:   loopVar,        // Loop variable name
		Cond:     collectionNode, // Collection expression
		Value:    bodySteps,      // Body steps
		Metadata: make(map[string]string),
	}
}

// --- deepEqualWithTolerance function definition (Unchanged) ---
const defaultTolerance = 1e-9

func deepEqualWithTolerance(a, b interface{}) bool {
	// Assume toFloat64 helper exists (e.g., in helpers.go or evaluation_helpers.go)
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		aFloat, aIsNum := toFloat64(a)
		bFloat, bIsNum := toFloat64(b)
		if aIsNum && bIsNum {
			return math.Abs(aFloat-bFloat) < defaultTolerance
		}
		return false
	}
	if a != nil && reflect.TypeOf(a).Kind() == reflect.Float64 {
		aFloat := a.(float64)
		bFloat := b.(float64)
		return math.Abs(aFloat-bFloat) < defaultTolerance
	}
	return reflect.DeepEqual(a, b)
}

// --- END deepEqualWithTolerance section ---

// runEvalExpressionTest executes a single expression evaluation test case.
// (Unchanged from previous correct version)
func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	interp, _ := NewDefaultTestInterpreter(t) // Assumes this helper exists
	if tc.InitialVars != nil {
		for k, v := range tc.InitialVars {
			interp.variables[k] = v
		}
	}
	interp.lastCallResult = tc.LastResult
	got, err := interp.evaluateExpression(tc.InputNode)
	deepCompareFunc := deepEqualWithTolerance

	if tc.WantErr {
		if err == nil {
			t.Errorf("%s: Expected an error, but got nil", tc.Name)
			return
		}
		if tc.ExpectedErrorIs != nil {
			if !errors.Is(err, tc.ExpectedErrorIs) {
				t.Errorf("%s: Expected error wrapping [%v], but got [%v]", tc.Name, tc.ExpectedErrorIs, err)
			}
		} else {
			t.Logf("%s: Got expected error: %v (No specific sentinel check provided)", tc.Name, err)
		}
	} else { // No error wanted
		if err != nil {
			t.Errorf("%s: Unexpected error: %v", tc.Name, err)
		}
		if !deepCompareFunc(got, tc.Expected) {
			t.Errorf("%s: Result mismatch.\nInput Node: %+v\nExpected:   %v (%T)\nGot:        %v (%T)", tc.Name, tc.InputNode, tc.Expected, tc.Expected, got, got)
		}
	}
}

// --- ADDED runExecuteStepsTest definition ---
// runExecuteStepsTest executes a sequence of steps and checks results/errors/variables.
func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	// Use NewDefaultTestInterpreter for consistent setup
	interp, _ := NewDefaultTestInterpreter(t) // Assumes this helper exists and works
	if tc.initialVars != nil {
		for k, v := range tc.initialVars {
			interp.SetVariable(k, v)
		}
	}

	finalResult, wasReturn, _, err := interp.executeSteps(tc.inputSteps, false, nil) // Execute the steps

	if tc.expectError {
		if err == nil {
			t.Errorf("Test %q: Expected an error, but got nil", tc.name)
			return
		}
		if tc.ExpectedErrorIs != nil {
			if !errors.Is(err, tc.ExpectedErrorIs) {
				t.Errorf("Test %q: Error mismatch.\nExpected error wrapping: [%v]\nGot error:             [%v]", tc.name, tc.ExpectedErrorIs, err)
			} else {
				t.Logf("Test %q: Got expected error wrapping [%v]: %v", tc.name, tc.ExpectedErrorIs, err)
			}
		} else {
			t.Logf("Test %q: Got expected error (no specific sentinel check provided): %v", tc.name, err)
		}
	} else { // Not expecting error
		if err != nil {
			t.Errorf("Test %q: Unexpected error: %+v", tc.name, err) // Use %+v for potentially more detail
		}
		expectedExecResult := tc.expectedResult
		actualExecResult := finalResult // The result from executeSteps

		// Adjust result check for RETURN handling
		if wasReturn {
			if resultSlice, ok := actualExecResult.([]interface{}); ok && len(resultSlice) == 1 {
				if _, expectedSlice := expectedExecResult.([]interface{}); !expectedSlice {
					actualExecResult = resultSlice[0]
				}
			}
		} else {
			if tc.expectedResult != nil {
				t.Errorf("Test %q: Expected non-nil result (%v) but no RETURN occurred.", tc.name, tc.expectedResult)
			}
			actualExecResult = nil // Treat non-returned result as nil for comparison purposes
		}

		// Use deepEqualWithTolerance for comparison
		if !deepEqualWithTolerance(actualExecResult, expectedExecResult) {
			t.Errorf("Test %q: Final execution result mismatch:\nExpected: %v (%T)\nGot:      %v (%T) (WasReturn: %t)",
				tc.name, expectedExecResult, expectedExecResult,
				actualExecResult, actualExecResult, wasReturn)
		}
	}

	// Variable state check
	if !tc.expectError && tc.expectedVars != nil {
		cleanInterp, _ := NewDefaultTestInterpreter(t) // For base built-ins
		baseVars := cleanInterp.variables

		for key, expectedValue := range tc.expectedVars {
			actualValue, exists := interp.variables[key]
			if !exists {
				if _, isBuiltIn := baseVars[key]; !isBuiltIn {
					t.Errorf("Test %q: Expected variable '%s' not found in final state", tc.name, key)
				}
				continue
			}
			if !deepEqualWithTolerance(actualValue, expectedValue) { // Use tolerance check
				t.Errorf("Test %q: Variable '%s' mismatch:\nExpected: %v (%T)\nGot:      %v (%T)",
					tc.name, key, expectedValue, expectedValue, actualValue, actualValue)
			}
		}

		extraVars := []string{}
		for k := range interp.variables {
			if _, isBuiltIn := baseVars[k]; !isBuiltIn {
				if _, expected := tc.expectedVars[k]; !expected {
					extraVars = append(extraVars, k)
				}
			}
		}
		if len(extraVars) > 0 {
			sort.Strings(extraVars)
			t.Errorf("Test %q: Unexpected non-builtin variables found in final state: %v", tc.name, extraVars)
		}
	}
}

// --- END ADDED runExecuteStepsTest ---

// --- General Test Helpers ---

// AssertNoError fails the test if err is not nil. (Unchanged)
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
// EvalTestCase (Unchanged from previous version)
type EvalTestCase struct {
	Name            string
	InputNode       interface{}
	InitialVars     map[string]interface{}
	LastResult      interface{}
	Expected        interface{}
	WantErr         bool
	ExpectedErrorIs error // Use this for errors.Is checks
}

// ValidationTestCase (Unchanged from previous version)
type ValidationTestCase struct {
	Name          string
	ToolName      string
	InputArgs     []interface{}
	ExpectedArgs  []interface{} // Optional: Check converted args
	ExpectedError error         // Expected sentinel error from ValidateAndConvertArgs
}

// --- Placeholders for other helpers potentially defined in original ---
// func runValidationTestCases(...) { ... }

// Ensure core errors are accessible if needed by helpers here
var (
	_ = ErrValidationArgCount
	_ = ErrValidationRequiredArgNil
	_ = ErrValidationTypeMismatch
	// Add other error variables used here if needed
	_ = ErrMustConditionFailed
)

// Assumes toFloat64 helper exists (e.g., in helpers.go or evaluation_helpers.go)
// func toFloat64(v interface{}) (float64, bool) { ... }

// --- END FILE ---
