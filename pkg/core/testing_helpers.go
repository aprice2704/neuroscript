// filename: pkg/core/testing_helpers.go
package core

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"testing"
	// Assuming Position is defined in ast.go
)

// --- Shared Test Struct Definitions ---

// EvalTestCase defines the structure for testing evaluateExpression
type EvalTestCase struct {
	Name            string
	InputNode       interface{} // AST node or raw value (asserted to Expression by helper if needed)
	InitialVars     map[string]interface{}
	LastResult      interface{} // Mocked result of previous step if needed
	Expected        interface{} // Expected result of evaluation
	WantErr         bool
	ExpectedErrorIs error // Use sentinel error or nil for errors.Is checks
}

// executeStepsTestCase defines the structure for testing interp.executeSteps
type executeStepsTestCase struct {
	name            string
	inputSteps      []Step                 // Sequence of steps to execute
	initialVars     map[string]interface{} // Initial variable state
	expectError     bool                   // Whether an error is expected during execution
	ExpectedErrorIs error                  // Specific sentinel error expected (if expectError is true)
	expectedResult  interface{}            // Expected final result (if a RETURN step occurs)
	expectedVars    map[string]interface{} // Expected final variable state (excluding built-ins)
}

// --- Shared Helper Functions ---

// AssertNoError fails the test if err is not nil.
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
		t.Fatal(message) // Use Fatal to stop the test on unexpected error
	}
}

// deepEqualWithTolerance compares two values, allowing for float tolerance.
const defaultTolerance = 1e-9

func deepEqualWithTolerance(a, b interface{}) bool {
	// Handle nil cases first
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false // One is nil, the other isn't
	}

	// Use reflect.DeepEqual for non-numeric types or if types differ fundamentally
	// However, try float conversion first if types *might* be compatible numbers
	aFloat, aIsNum := toFloat64(a)
	bFloat, bIsNum := toFloat64(b)

	if aIsNum && bIsNum {
		// If both are numbers (even different types like int/float), compare as floats
		return math.Abs(aFloat-bFloat) < defaultTolerance
	}

	// If types are identical and not floats, or if one/both are not numbers, use DeepEqual
	if reflect.TypeOf(a) == reflect.TypeOf(b) {
		// Special check for identical floats (handles NaN, Inf correctly)
		if reflect.TypeOf(a).Kind() == reflect.Float64 {
			aF64 := a.(float64)
			bF64 := b.(float64)
			if math.IsNaN(aF64) && math.IsNaN(bF64) {
				return true
			}
			return math.Abs(aF64-bF64) < defaultTolerance
		}
		// Fallback for non-float identical types
		return reflect.DeepEqual(a, b)
	}

	// Types are different and at least one is not numerically convertible to float
	return false
}

// runEvalExpressionTest executes a single expression evaluation test case.
func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	interp, _ := NewTestInterpreter(t, tc.InitialVars, tc.LastResult) // Use helper

	// Assert InputNode to Expression before passing to evaluateExpression
	var inputExpr Expression
	if tc.InputNode != nil {
		var ok bool
		inputExpr, ok = tc.InputNode.(Expression)
		if !ok {
			// Check if it's a raw value that evaluateExpression might handle directly
			// For now, assume if it's not Expression, it's an error in test setup
			t.Fatalf("Test setup error in %q: InputNode (%T) does not implement Expression and is not nil", tc.Name, tc.InputNode)
		}
	} // else inputExpr remains nil, evaluateExpression should handle nil input if necessary

	got, err := interp.evaluateExpression(inputExpr) // Pass asserted Expression

	// Error checking
	if (err != nil) != tc.WantErr {
		t.Errorf("Test %q: Error expectation mismatch. got err = %v, wantErr %v", tc.Name, err, tc.WantErr)
		if err != nil {
			t.Logf("Input: %#v, Vars: %#v, Last: %#v", tc.InputNode, tc.InitialVars, tc.LastResult)
		}
		return
	}
	if tc.WantErr {
		if tc.ExpectedErrorIs != nil && !errors.Is(err, tc.ExpectedErrorIs) {
			t.Errorf("Test %q: Expected error wrapping [%v], but got [%v]", tc.Name, tc.ExpectedErrorIs, err)
		} else if tc.ExpectedErrorIs == nil {
			t.Logf("Test %q: Got expected error: %v", tc.Name, err)
		}
		return
	}

	// Result comparison
	if !deepEqualWithTolerance(got, tc.Expected) {
		t.Errorf("Test %q: Result mismatch.\nInput:    %#v\nVars:     %#v\nLast:     %#v\nExpected: %v (%T)\nGot:      %v (%T)",
			tc.Name, tc.InputNode, tc.InitialVars, tc.LastResult,
			tc.Expected, tc.Expected, got, got)
	}
}

// runExecuteStepsTest executes a sequence of steps and checks results/errors/variables.
func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	interp, _ := NewTestInterpreter(t, tc.initialVars, nil) // Use helper, lastResult often irrelevant for steps

	finalResult, wasReturn, _, err := interp.executeSteps(tc.inputSteps, false, nil) // Execute the steps

	// Error Checking
	if tc.expectError {
		if err == nil {
			t.Errorf("Test %q: Expected an error, but got nil", tc.name)
			return
		}
		if tc.ExpectedErrorIs != nil {
			if !errors.Is(err, tc.ExpectedErrorIs) {
				t.Errorf("Test %q: Error mismatch.\nExpected error wrapping: [%v]\nGot error:               [%v]", tc.name, tc.ExpectedErrorIs, err)
			} else {
				t.Logf("Test %q: Got expected error wrapping [%v]: %v", tc.name, tc.ExpectedErrorIs, err)
			}
		} else {
			t.Logf("Test %q: Got expected error (no specific sentinel check provided): %v", tc.name, err)
		}
	} else { // Not expecting error
		if err != nil {
			t.Errorf("Test %q: Unexpected error: %+v", tc.name, err)
		}
		// Result Checking (only if no error expected)
		expectedExecResult := tc.expectedResult
		actualExecResult := finalResult

		if wasReturn {
			// If expected result is not a slice, but return gave a single-element slice, unwrap it.
			if resultSlice, ok := actualExecResult.([]interface{}); ok && len(resultSlice) == 1 {
				if _, expectedSlice := expectedExecResult.([]interface{}); !expectedSlice {
					actualExecResult = resultSlice[0]
				}
			}
			// Add case: If expected result *is* a slice, compare directly with actual slice.
		} else {
			// If no return occurred, result should be nil.
			if tc.expectedResult != nil {
				t.Errorf("Test %q: Expected non-nil result (%v) but no RETURN occurred.", tc.name, tc.expectedResult)
			}
			actualExecResult = nil // Set to nil for comparison
		}

		if !deepEqualWithTolerance(actualExecResult, expectedExecResult) {
			t.Errorf("Test %q: Final execution result mismatch:\nExpected: %v (%T)\nGot:      %v (%T) (WasReturn: %t)",
				tc.name, expectedExecResult, expectedExecResult,
				actualExecResult, actualExecResult, wasReturn)
		}
	}

	// Variable state check (only if no error was expected)
	if !tc.expectError && tc.expectedVars != nil {
		// Create clean interp just to get base vars without test setup side effects
		cleanInterp, _ := NewDefaultTestInterpreter(t)
		baseVars := cleanInterp.variables

		// Check expected variables exist and match
		for key, expectedValue := range tc.expectedVars {
			actualValue, exists := interp.variables[key]
			if !exists {
				if _, isBuiltIn := baseVars[key]; !isBuiltIn {
					t.Errorf("Test %q: Expected variable '%s' not found in final state", tc.name, key)
				}
				continue // Skip checking built-ins if not explicitly expected
			}
			if !deepEqualWithTolerance(actualValue, expectedValue) {
				t.Errorf("Test %q: Variable '%s' mismatch:\nExpected: %v (%T)\nGot:      %v (%T)",
					tc.name, key, expectedValue, expectedValue, actualValue, actualValue)
			}
		}

		// Check for any extra non-builtin variables
		extraVars := []string{}
		for k := range interp.variables {
			// Ignore built-ins unless explicitly listed in expectedVars
			if _, isBuiltIn := baseVars[k]; isBuiltIn {
				if _, expected := tc.expectedVars[k]; !expected {
					continue
				}
			}
			// If not built-in (or expected built-in) and not in expectedVars map, it's extra.
			if _, expected := tc.expectedVars[k]; !expected {
				extraVars = append(extraVars, k)
			}
		}
		if len(extraVars) > 0 {
			sort.Strings(extraVars)
			t.Errorf("Test %q: Unexpected variables found in final state: %v", tc.name, extraVars)
		}
	}
}

// createTestStep creates a generic step, often used for simple steps like set/must/return.
// Relies on newStep (assumed defined in ast.go based on user files).
func createTestStep(typ string, target string, valueNode interface{}, elseValueNode interface{}) Step {
	// Pass nil for cond; specific helpers below will handle conditions.
	// The valueNode might be an Expression pointer, raw value, or []Expression pointer slice depending on 'typ'.
	// The caller must ensure valueNode has the correct type structure.
	return newStep(typ, target, nil, valueNode, elseValueNode) // Calls newStep from ast.go
}

// createIfStep creates an 'if' step struct.
// Requires condNode to be assertable to Expression.
func createIfStep(condNode interface{}, thenSteps []Step, elseSteps []Step) Step {
	condExpr, ok := condNode.(Expression)
	if !ok {
		panic(fmt.Sprintf("createIfStep: test provided a condNode argument (%T) that does not implement Expression", condNode))
	}

	var elseVal interface{} = nil
	if elseSteps != nil {
		elseVal = elseSteps // Assign the slice directly
	}
	// Directly create Step struct, ensuring Cond gets the asserted Expression.
	return Step{
		Type:      "if",
		Cond:      condExpr, // Use asserted Expression
		Value:     thenSteps,
		ElseValue: elseVal,
		Metadata:  make(map[string]string),
		// Pos: condExpr.GetPos(), // Set position from condition
	}
}

// createWhileStep creates a 'while' step struct.
// Requires condNode to be assertable to Expression.
func createWhileStep(condNode interface{}, bodySteps []Step) Step {
	condExpr, ok := condNode.(Expression)
	if !ok {
		panic(fmt.Sprintf("createWhileStep: test provided a condNode argument (%T) that does not implement Expression", condNode))
	}
	return Step{
		Type:     "while",
		Cond:     condExpr, // Use asserted Expression
		Value:    bodySteps,
		Metadata: make(map[string]string),
		// Pos: condExpr.GetPos(), // Set position from condition
	}
}

// createForStep creates a 'for' step struct.
// Requires collectionNode to be assertable to Expression.
func createForStep(loopVar string, collectionNode interface{}, bodySteps []Step) Step {
	collectionExpr, ok := collectionNode.(Expression)
	if !ok {
		panic(fmt.Sprintf("createForStep: test provided a collectionNode argument (%T) that does not implement Expression", collectionNode))
	}
	return Step{
		Type:     "for",
		Target:   loopVar,
		Cond:     collectionExpr, // Use asserted Expression for the collection
		Value:    bodySteps,
		Metadata: make(map[string]string),
		// Pos: collectionExpr.GetPos(), // Set position from collection
	}
}
