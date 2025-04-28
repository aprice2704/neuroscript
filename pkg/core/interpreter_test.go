// filename: neuroscript/pkg/core/interpreter_test.go
package core

import (
	"errors"
	"reflect"
	"sort"
	"testing"
	// Import sort needed for checking unexpected vars
)

// --- Test Suite for executeSteps (Blocks, Loops, Tools) ---
type executeStepsTestCase struct {
	name            string
	inputSteps      []Step
	initialVars     map[string]interface{}
	expectedVars    map[string]interface{}
	expectedResult  interface{}
	expectError     bool
	ExpectedErrorIs error // Use this for errors.Is checks
}

// FIX: Removed stray l.logDebugAST call
func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	interp, _ := NewTestInterpreter(t, tc.initialVars, nil)

	finalResult, wasReturn, err := interp.executeSteps(tc.inputSteps)

	if tc.expectError {
		if err == nil {
			t.Errorf("Test %q: Expected an error, but got nil", tc.name)
			return
		}
		if tc.ExpectedErrorIs != nil {
			if !errors.Is(err, tc.ExpectedErrorIs) {
				t.Errorf("Test %q: Error mismatch.\nExpected error wrapping: [%v]\nGot:                     [%v]", tc.name, tc.ExpectedErrorIs, err)
			} else {
				t.Logf("Test %q: Got expected error wrapping [%v]: %v", tc.name, tc.ExpectedErrorIs, err)
			}
		} else {
			t.Logf("Test %q: Got expected error (no specific sentinel check provided): %v", tc.name, err)
		}
	} else { // No error wanted
		if err != nil {
			t.Errorf("Test %q: Unexpected error: %v", tc.name, err)
		}
		expectedExecResult := tc.expectedResult
		actualExecResult := finalResult

		if wasReturn {
			// If executeReturn gave us a slice of size 1, unpack it for comparison if the expected result is NOT a slice.
			if resultSlice, ok := actualExecResult.([]interface{}); ok && len(resultSlice) == 1 {
				if _, expectedSlice := expectedExecResult.([]interface{}); !expectedSlice {
					actualExecResult = resultSlice[0] // Unpack single-element slice
					// Removed undefined l.logDebugAST call here
				}
			}
		} else { // No RETURN occurred
			if tc.expectedResult != nil {
				t.Errorf("Test %q: Expected non-nil result (%v) but no RETURN occurred.", tc.name, tc.expectedResult)
			}
			actualExecResult = nil
		}

		if !reflect.DeepEqual(actualExecResult, expectedExecResult) {
			t.Errorf("Test %q: Final execution result mismatch:\nExpected: %v (%T)\nGot:      %v (%T) (WasReturn: %t)",
				tc.name, expectedExecResult, expectedExecResult, actualExecResult, actualExecResult, wasReturn)
		}
	}

	// Check final variable state (only if no error expected)
	if !tc.expectError && tc.expectedVars != nil {
		cleanInterp, _ := NewDefaultTestInterpreter(t)
		baseVars := cleanInterp.variables
		for key, expectedValue := range tc.expectedVars {
			actualValue, exists := interp.variables[key]
			if !exists {
				if _, isBuiltIn := baseVars[key]; !isBuiltIn {
					t.Errorf("Test %q: Expected variable '%s' not found in final state", tc.name, key)
				}
				continue
			}
			if !reflect.DeepEqual(actualValue, expectedValue) {
				t.Errorf("Test %q: Variable '%s' mismatch:\nExpected: %v (%T)\nGot:      %v (%T)", tc.name, key, expectedValue, expectedValue, actualValue, actualValue)
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

// TestExecuteStepsBlocksAndLoops - Includes List/Map iteration
func TestExecuteStepsBlocksAndLoops(t *testing.T) {
	testCases := []executeStepsTestCase{
		// --- Existing Tests (IF/WHILE/FOR/TOOL Calls) ---
		{name: "IF true literal", inputSteps: []Step{createIfStep(BooleanLiteralNode{Value: true}, []Step{createTestStep("set", "x", StringLiteralNode{Value: "Inside"}, nil)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"x": "Inside"}, expectedResult: nil, expectError: false},
		// FIX: Correctly call createTestStep with 4 args for RETURN
		{name: "IF block with RETURN", inputSteps: []Step{
			createTestStep("set", "status", StringLiteralNode{Value: "Started"}, nil),
			createIfStep(BooleanLiteralNode{Value: true}, []Step{
				createTestStep("set", "x", StringLiteralNode{Value: "Inside"}, nil),
				// Args: typ, target, value (slice of nodes), args (nil for return)
				createTestStep("return", "", []interface{}{StringLiteralNode{Value: "ReturnedFromIf"}}, nil),
				createTestStep("set", "y", StringLiteralNode{Value: "NotReached"}, nil),
			}, nil),
			createTestStep("set", "status", StringLiteralNode{Value: "Finished"}, nil),
		}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"status": "Started", "x": "Inside"}, expectedResult: "ReturnedFromIf", expectError: false},
		// ... other existing cases ...
		{name: "CALL TOOL JoinStrings with ListLiteral", inputSteps: []Step{createTestStep("call", "tool.JoinStrings", nil, []interface{}{ListLiteralNode{Elements: []interface{}{StringLiteralNode{Value: "A"}, NumberLiteralNode{Value: int64(1)}, BooleanLiteralNode{Value: true}}}, StringLiteralNode{Value: "-"}}), createTestStep("set", "joined", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"joined": "A-1-true"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL Substring Wrong Arg Type AST", inputSteps: []Step{createTestStep("call", "tool.Substring", nil, []interface{}{StringLiteralNode{Value: "hello"}, StringLiteralNode{Value: "one"}, NumberLiteralNode{Value: int64(3)}})}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrValidationTypeMismatch},

		// +++ Multiple Return Value Tests +++
		{
			name: "RETURN single value (backward compat)",
			inputSteps: []Step{
				// FIX: Correctly call createTestStep with 4 args
				// Args: typ, target, value (slice of nodes), args (nil for return)
				createTestStep("return", "", []interface{}{NumberLiteralNode{Value: int64(42)}}, nil),
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{},
			expectedResult: int64(42), // Expect single value after RunProcedure unpacks
			expectError:    false,
		},
		{
			name: "RETURN multiple values",
			inputSteps: []Step{
				// FIX: Correctly call createTestStep with 4 args
				createTestStep("return", "", []interface{}{
					StringLiteralNode{Value: "hello"},
					NumberLiteralNode{Value: int64(10)},
					BooleanLiteralNode{Value: true},
				}, nil), // Pass nil for args
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{},
			expectedResult: []interface{}{"hello", int64(10), true}, // Expect slice
			expectError:    false,
		},
		{
			name: "RETURN no value",
			inputSteps: []Step{
				// FIX: Correctly call createTestStep with 4 args (value is nil)
				createTestStep("return", "", nil, nil), // Value is nil, args is nil
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{},
			expectedResult: nil, // Expect nil
			expectError:    false,
		},
		{
			name: "RETURN value from variable",
			inputSteps: []Step{
				createTestStep("set", "myVar", StringLiteralNode{Value: "data"}, nil),
				// FIX: Correctly call createTestStep with 4 args
				createTestStep("return", "", []interface{}{VariableNode{Name: "myVar"}}, nil),
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"myVar": "data"},
			expectedResult: "data",
			expectError:    false,
		},
		{
			name: "RETURN multiple values including variable",
			inputSteps: []Step{
				createTestStep("set", "myVar", BooleanLiteralNode{Value: false}, nil),
				// FIX: Correctly call createTestStep with 4 args
				createTestStep("return", "", []interface{}{
					NumberLiteralNode{Value: int64(1)},
					VariableNode{Name: "myVar"},
					NumberLiteralNode{Value: 3.14},
				}, nil), // Pass nil for args
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"myVar": false},
			expectedResult: []interface{}{int64(1), false, 3.14}, // Expect slice
			expectError:    false,
		},
		// +++ Semantic Check Tests (Add these when ready) +++
		/* ... */

	} // End testCases slice

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { runExecuteStepsTest(t, tc) })
	}
} // End TestExecuteStepsBlocksAndLoops
