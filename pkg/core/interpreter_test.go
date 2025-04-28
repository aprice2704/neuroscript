// filename: neuroscript/pkg/core/interpreter_test.go
package core

import (
	"errors" // Import errors package for errors.Is
	"reflect"
	"sort" // Import sort package for stable map key iteration
	"testing"
)

// Remove helper functions - MOVED to testing_helpers_test.go
/*
 func createTestStep(...) Step { ... }
 func createIfStep(...) Step { ... }
 func createWhileStep(...) Step { ... }
 func createForStep(...) Step { ... }
*/

// --- Test Suite for executeSteps (Blocks, Loops, Tools) ---
// FIX: Removed errorContains, added ExpectedErrorIs
type executeStepsTestCase struct {
	name            string
	inputSteps      []Step
	initialVars     map[string]interface{}
	expectedVars    map[string]interface{}
	expectedResult  interface{} // Expected result from RETURN or last expression if no RETURN
	expectError     bool
	ExpectedErrorIs error // Use this for errors.Is checks
	// errorContains  string // REMOVED
}

// FIX: Updated to use ExpectedErrorIs and errors.Is
func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	// Use NewTestInterpreter from test scope, passing t and handling 2 return values
	interp, _ := NewTestInterpreter(t, tc.initialVars, nil) // Use initialVars, ignore sandbox path

	finalResult, wasReturn, err := interp.executeSteps(tc.inputSteps)

	if tc.expectError {
		if err == nil {
			t.Errorf("Test %q: Expected an error, but got nil", tc.name)
			return
		}
		// *** Check for specific sentinel error using errors.Is ***
		if tc.ExpectedErrorIs != nil {
			if !errors.Is(err, tc.ExpectedErrorIs) {
				t.Errorf("Test %q: Error mismatch.\nExpected error wrapping: [%v]\nGot:                     [%v]", tc.name, tc.ExpectedErrorIs, err)
			} else {
				// Log confirmation when the correct specific error is found
				t.Logf("Test %q: Got expected error wrapping [%v]: %v", tc.name, tc.ExpectedErrorIs, err)
			}
		} else {
			// If expectError is true, but no specific error is set, log it.
			t.Logf("Test %q: Got expected error (no specific sentinel check provided): %v", tc.name, err)
		}
		// *** Removed errorContains check ***

	} else { // No error wanted
		if err != nil {
			t.Errorf("Test %q: Unexpected error: %v", tc.name, err)
		}
		// Check final result (only if no error expected)
		expectedExecResult := tc.expectedResult
		actualExecResult := finalResult
		if !wasReturn {
			// If no RETURN statement was executed, the final result of the procedure/block
			// is implicitly nil, regardless of the last step's evaluation.
			// Check if the expected result was also nil for this case.
			if tc.expectedResult != nil {
				t.Errorf("Test %q: Expected non-nil result (%v) but no RETURN occurred.", tc.name, tc.expectedResult)
			}
			actualExecResult = nil // Set actual to nil if no RETURN
		}

		// Now compare actualExecResult (which is correctly nil if no RETURN) with expectedExecResult
		if !reflect.DeepEqual(actualExecResult, expectedExecResult) {
			t.Errorf("Test %q: Final execution result mismatch:\nExpected: %v (%T)\nGot:      %v (%T) (WasReturn: %t)",
				tc.name, expectedExecResult, expectedExecResult, actualExecResult, actualExecResult, wasReturn)
		}
	}

	// Check final variable state (only if no error expected)
	if !tc.expectError && tc.expectedVars != nil {
		// Get built-ins from a clean interpreter
		cleanInterp, _ := NewDefaultTestInterpreter(t) // Pass t
		baseVars := cleanInterp.variables              // Get base built-ins

		// Check expected variables exist and match
		for key, expectedValue := range tc.expectedVars {
			actualValue, exists := interp.variables[key]
			if !exists {
				// Allow built-ins like NEUROSCRIPT_DEVELOP_PROMPT to exist implicitly
				if _, isBuiltIn := baseVars[key]; !isBuiltIn {
					t.Errorf("Test %q: Expected variable '%s' not found in final state", tc.name, key)
				}
				continue
			}
			if !reflect.DeepEqual(actualValue, expectedValue) {
				t.Errorf("Test %q: Variable '%s' mismatch:\nExpected: %v (%T)\nGot:      %v (%T)", tc.name, key, expectedValue, expectedValue, actualValue, actualValue)
			}
		}
		// Check for unexpected variables (excluding built-ins)
		extraVars := []string{}
		for k := range interp.variables {
			if _, isBuiltIn := baseVars[k]; !isBuiltIn { // Skip built-ins
				if _, expected := tc.expectedVars[k]; !expected { // If not expected
					extraVars = append(extraVars, k)
				}
			}
		}
		if len(extraVars) > 0 {
			sort.Strings(extraVars) // Sort for deterministic output
			t.Errorf("Test %q: Unexpected non-builtin variables found in final state: %v", tc.name, extraVars)
		}
	}
}

// TestExecuteStepsBlocksAndLoops - Includes List/Map iteration
func TestExecuteStepsBlocksAndLoops(t *testing.T) {
	testCases := []executeStepsTestCase{
		// IF/WHILE tests
		{name: "IF true literal", inputSteps: []Step{createIfStep(BooleanLiteralNode{Value: true}, []Step{createTestStep("SET", "x", StringLiteralNode{Value: "Inside"}, nil)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"x": "Inside"}, expectedResult: nil, expectError: false},
		{name: "IF false literal", inputSteps: []Step{createIfStep(BooleanLiteralNode{Value: false}, []Step{createTestStep("SET", "x", StringLiteralNode{Value: "Inside"}, nil)}, nil), createTestStep("SET", "y", StringLiteralNode{Value: "Outside"}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"y": "Outside"}, expectedResult: nil, expectError: false},
		{name: "IF condition var true", inputSteps: []Step{createIfStep(VariableNode{Name: "cond_var"}, []Step{createTestStep("SET", "x", StringLiteralNode{Value: "Inside"}, nil)}, nil)}, initialVars: map[string]interface{}{"cond_var": true}, expectedVars: map[string]interface{}{"cond_var": true, "x": "Inside"}, expectedResult: nil, expectError: false},
		{name: "IF block with RETURN", inputSteps: []Step{createTestStep("SET", "status", StringLiteralNode{Value: "Started"}, nil), createIfStep(BooleanLiteralNode{Value: true}, []Step{createTestStep("SET", "x", StringLiteralNode{Value: "Inside"}, nil), createTestStep("RETURN", "", StringLiteralNode{Value: "ReturnedFromIf"}, nil), createTestStep("SET", "y", StringLiteralNode{Value: "NotReached"}, nil)}, nil), createTestStep("SET", "status", StringLiteralNode{Value: "Finished"}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"status": "Started", "x": "Inside"}, expectedResult: "ReturnedFromIf", expectError: false},
		{name: "WHILE runs once", inputSteps: []Step{createTestStep("SET", "run", BooleanLiteralNode{Value: true}, nil), createTestStep("SET", "counter", NumberLiteralNode{Value: int64(0)}, nil), createWhileStep(VariableNode{Name: "run"}, []Step{createTestStep("SET", "run", BooleanLiteralNode{Value: false}, nil), createTestStep("SET", "counter", NumberLiteralNode{Value: int64(1)}, nil)})}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"run": false, "counter": int64(1)}, expectedResult: nil, expectError: false},
		{name: "WHILE false initially", inputSteps: []Step{createWhileStep(BooleanLiteralNode{Value: false}, []Step{createTestStep("SET", "x", StringLiteralNode{Value: "Never"}, nil)}), createTestStep("SET", "y", StringLiteralNode{Value: "After"}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"y": "After"}, expectedResult: nil, expectError: false},

		// FOR EACH String Iteration
		{
			name: "FOR EACH string char iteration",
			inputSteps: []Step{
				createTestStep("SET", "input", StringLiteralNode{Value: "Hi!"}, nil),
				createTestStep("SET", "output", StringLiteralNode{Value: ""}, nil),
				createForStep("char", VariableNode{Name: "input"}, []Step{
					createTestStep("SET", "output",
						BinaryOpNode{
							Left:     BinaryOpNode{Left: VariableNode{Name: "output"}, Operator: "+", Right: VariableNode{Name: "char"}},
							Operator: "+", Right: StringLiteralNode{Value: "-"},
						}, nil),
				}),
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"input": "Hi!", "output": "H-i-!-"},
			expectedResult: nil, expectError: false,
		},
		{
			name: "FOR EACH comma split",
			inputSteps: []Step{
				createTestStep("SET", "input", StringLiteralNode{Value: "a, b ,c"}, nil),
				createTestStep("SET", "output", StringLiteralNode{Value: ""}, nil),
				createForStep("item", VariableNode{Name: "input"}, []Step{
					createTestStep("SET", "output",
						BinaryOpNode{
							Left:     BinaryOpNode{Left: VariableNode{Name: "output"}, Operator: "+", Right: StringLiteralNode{Value: "-"}},
							Operator: "+", Right: VariableNode{Name: "item"},
						}, nil),
				}),
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"input": "a, b ,c", "output": "-a-b-c"},
			expectedResult: nil, expectError: false,
		},

		// FOR EACH List/Map tests (errors expected due to type mismatches with '+')
		{
			name: "FOR EACH list literal",
			inputSteps: []Step{
				createTestStep("SET", "output", StringLiteralNode{Value: ""}, nil),
				createForStep("item", ListLiteralNode{Elements: []interface{}{NumberLiteralNode{Value: int64(1)}, StringLiteralNode{Value: "X"}, BooleanLiteralNode{Value: true}, ListLiteralNode{Elements: []interface{}{StringLiteralNode{Value: "nest"}}}}}, []Step{
					createTestStep("SET", "output",
						BinaryOpNode{
							Left:     BinaryOpNode{Left: VariableNode{Name: "output"}, Operator: "+", Right: VariableNode{Name: "item"}},
							Operator: "+", Right: StringLiteralNode{Value: "|"},
						}, nil),
				}),
			},
			initialVars:  map[string]interface{}{},
			expectedVars: map[string]interface{}{"output": ""}, // Output shouldn't change before error
			expectError:  true,
			// FIX: Expect the sentinel error for invalid '+' operation
			ExpectedErrorIs: ErrInvalidOperandType, // '+' error between numeric/string
		},
		{
			name: "FOR EACH list variable",
			inputSteps: []Step{
				createTestStep("SET", "output", StringLiteralNode{Value: ""}, nil),
				createForStep("val", VariableNode{Name: "myListVar"}, []Step{
					createTestStep("SET", "output", BinaryOpNode{Left: VariableNode{Name: "output"}, Operator: "+", Right: VariableNode{Name: "val"}}, nil),
				}),
			},
			initialVars:  map[string]interface{}{"myListVar": []interface{}{"A", "B", int64(3)}},
			expectedVars: map[string]interface{}{"myListVar": []interface{}{"A", "B", int64(3)}, "output": "AB"}, // Output stops at "AB" before error
			expectError:  true,
			// FIX: Expect the sentinel error for invalid '+' operation
			ExpectedErrorIs: ErrInvalidOperandType, // '+' error between string/numeric
		},

		// FOR EACH Map Iteration (Keys)
		{
			name: "FOR EACH map literal keys",
			inputSteps: []Step{
				createTestStep("SET", "output", StringLiteralNode{Value: ""}, nil),
				createForStep("key", MapLiteralNode{Entries: []MapEntryNode{{Key: StringLiteralNode{Value: "b"}, Value: NumberLiteralNode{Value: int64(2)}}, {Key: StringLiteralNode{Value: "a"}, Value: NumberLiteralNode{Value: int64(1)}}}}, []Step{
					createTestStep("SET", "output",
						BinaryOpNode{
							Left:     BinaryOpNode{Left: VariableNode{Name: "output"}, Operator: "+", Right: VariableNode{Name: "key"}},
							Operator: "+", Right: StringLiteralNode{Value: ","},
						}, nil),
				}),
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"output": "a,b,"}, // Map iteration order is stable
			expectedResult: nil, expectError: false,
		},
		{
			name: "FOR EACH map variable keys",
			inputSteps: []Step{
				createTestStep("SET", "output", StringLiteralNode{Value: ""}, nil),
				createForStep("k", VariableNode{Name: "myMap"}, []Step{
					createTestStep("SET", "output", BinaryOpNode{Left: VariableNode{Name: "output"}, Operator: "+", Right: VariableNode{Name: "k"}}, nil),
				}),
			},
			initialVars:    map[string]interface{}{"myMap": map[string]interface{}{"z": true, "x": "hello", "a": 1}},
			expectedVars:   map[string]interface{}{"myMap": map[string]interface{}{"z": true, "x": "hello", "a": 1}, "output": "axz"}, // Map iteration order is stable
			expectedResult: nil, expectError: false,
		},

		// Tool Call Tests
		{name: "CALL TOOL StringLength AST", inputSteps: []Step{createTestStep("SET", "myStr", StringLiteralNode{Value: "Test"}, nil), createTestStep("CALL", "TOOL.StringLength", nil, []interface{}{VariableNode{Name: "myStr"}}), createTestStep("SET", "lenResult", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "Test", "lenResult": int64(4)}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL Substring AST", inputSteps: []Step{createTestStep("CALL", "TOOL.Substring", nil, []interface{}{StringLiteralNode{Value: "ABCDE"}, NumberLiteralNode{Value: int64(1)}, NumberLiteralNode{Value: int64(4)}}), createTestStep("SET", "sub", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"sub": "BCD"}, expectedResult: nil, expectError: false},
		// FIX: Use ExpectedErrorIs with ErrValidationTypeMismatch per Rule #16
		{name: "CALL TOOL Substring Wrong Arg Type AST", inputSteps: []Step{createTestStep("CALL", "TOOL.Substring", nil, []interface{}{StringLiteralNode{Value: "hello"}, StringLiteralNode{Value: "one"}, NumberLiteralNode{Value: int64(3)}})}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrValidationTypeMismatch},
		{name: "CALL TOOL JoinStrings with ListLiteral", inputSteps: []Step{createTestStep("CALL", "TOOL.JoinStrings", nil, []interface{}{ListLiteralNode{Elements: []interface{}{StringLiteralNode{Value: "A"}, NumberLiteralNode{Value: int64(1)}, BooleanLiteralNode{Value: true}}}, StringLiteralNode{Value: "-"}}), createTestStep("SET", "joined", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"joined": "A-1-true"}, expectedResult: nil, expectError: false},
	} // End testCases slice

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { runExecuteStepsTest(t, tc) })
	}
} // End TestExecuteStepsBlocksAndLoops
