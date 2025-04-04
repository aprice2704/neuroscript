package core

import (
	"reflect" // Import reflect package
	// Import sort package for stable map key iteration
	"strings"
	"testing"
	// Import sort
)

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
func createTestStep(typ string, target string, valueNode interface{}, argNodes []interface{}) Step {
	return newStep(typ, target, nil, valueNode, nil, argNodes)
}
func createIfStep(condNode interface{}, thenSteps []Step, elseSteps []Step) Step {
	return Step{Type: "IF", Cond: condNode, Value: thenSteps, ElseValue: elseSteps}
}
func createWhileStep(condNode interface{}, bodySteps []Step) Step {
	return Step{Type: "WHILE", Cond: condNode, Value: bodySteps}
}
func createForStep(loopVar string, collectionNode interface{}, bodySteps []Step) Step {
	return Step{Type: "FOR", Target: loopVar, Cond: collectionNode, Value: bodySteps}
}

// --- Test Suite for executeSteps (Blocks, Loops, Tools) ---
type executeStepsTestCase struct {
	name           string
	inputSteps     []Step
	initialVars    map[string]interface{}
	expectedVars   map[string]interface{}
	expectedResult interface{} // Expected result from RETURN or last expression if no RETURN
	expectError    bool
	errorContains  string
	// checkOrder removed
}

func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	// ... (run test helper remains the same as previous version) ...
	t.Helper()
	interp := newTestInterpreter(tc.initialVars, nil) // Use initialVars directly

	finalResult, wasReturn, err := interp.executeSteps(tc.inputSteps)

	if tc.expectError {
		if err == nil {
			t.Errorf("Test %q: Expected an error, but got nil", tc.name)
			return
		}
		if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
			t.Errorf("Test %q: Expected error containing %q, but got: %v", tc.name, tc.errorContains, err)
		}
	} else {
		if err != nil {
			t.Errorf("Test %q: Unexpected error: %v", tc.name, err)
		}
		// Check final result (only if no error expected)
		expectedExecResult := tc.expectedResult
		actualExecResult := finalResult
		if !wasReturn {
			actualExecResult = nil // Implicit nil result if no RETURN occurred
		}

		if !reflect.DeepEqual(actualExecResult, expectedExecResult) {
			t.Errorf("Test %q: Final execution result mismatch:\nExpected: %v (%T)\nGot:      %v (%T) (Returned: %t)", tc.name, expectedExecResult, expectedExecResult, actualExecResult, actualExecResult, wasReturn)
		}
	}

	// Check final variable state (only if no error expected)
	if !tc.expectError && tc.expectedVars != nil {
		baseVars := NewInterpreter(nil).variables // Get base built-ins
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
			t.Errorf("Test %q: Unexpected non-builtin variables found in final state: %v", tc.name, extraVars)
		}
	}
}

// TestExecuteStepsBlocksAndLoops - Includes List/Map iteration
func TestExecuteStepsBlocksAndLoops(t *testing.T) {
	testCases := []executeStepsTestCase{
		// ... (IF/WHILE tests remain the same) ...
		{name: "IF true literal", inputSteps: []Step{createIfStep(BooleanLiteralNode{Value: true}, []Step{createTestStep("SET", "x", StringLiteralNode{Value: "Inside"}, nil)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"x": "Inside"}, expectedResult: nil, expectError: false},
		{name: "IF false literal", inputSteps: []Step{createIfStep(BooleanLiteralNode{Value: false}, []Step{createTestStep("SET", "x", StringLiteralNode{Value: "Inside"}, nil)}, nil), createTestStep("SET", "y", StringLiteralNode{Value: "Outside"}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"y": "Outside"}, expectedResult: nil, expectError: false},
		{name: "IF condition var true", inputSteps: []Step{createIfStep(VariableNode{Name: "cond_var"}, []Step{createTestStep("SET", "x", StringLiteralNode{Value: "Inside"}, nil)}, nil)}, initialVars: map[string]interface{}{"cond_var": true}, expectedVars: map[string]interface{}{"cond_var": true, "x": "Inside"}, expectedResult: nil, expectError: false},
		{name: "IF block with RETURN", inputSteps: []Step{createTestStep("SET", "status", StringLiteralNode{Value: "Started"}, nil), createIfStep(BooleanLiteralNode{Value: true}, []Step{createTestStep("SET", "x", StringLiteralNode{Value: "Inside"}, nil), createTestStep("RETURN", "", StringLiteralNode{Value: "ReturnedFromIf"}, nil), createTestStep("SET", "y", StringLiteralNode{Value: "NotReached"}, nil)}, nil), createTestStep("SET", "status", StringLiteralNode{Value: "Finished"}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"status": "Started", "x": "Inside"}, expectedResult: "ReturnedFromIf", expectError: false},
		{name: "WHILE runs once", inputSteps: []Step{createTestStep("SET", "run", BooleanLiteralNode{Value: true}, nil), createTestStep("SET", "counter", NumberLiteralNode{Value: int64(0)}, nil), createWhileStep(VariableNode{Name: "run"}, []Step{createTestStep("SET", "run", BooleanLiteralNode{Value: false}, nil), createTestStep("SET", "counter", NumberLiteralNode{Value: int64(1)}, nil)})}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"run": false, "counter": int64(1)}, expectedResult: nil, expectError: false},
		{name: "WHILE false initially", inputSteps: []Step{createWhileStep(BooleanLiteralNode{Value: false}, []Step{createTestStep("SET", "x", StringLiteralNode{Value: "Never"}, nil)}), createTestStep("SET", "y", StringLiteralNode{Value: "After"}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"y": "After"}, expectedResult: nil, expectError: false},

		// --- FOR EACH String Iteration ---
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
		// --- CORRECTED EXPECTATION based on comma split + trim + loop logic ---
		{
			name: "FOR EACH comma split", // Actually tests comma split + trim logic
			inputSteps: []Step{
				createTestStep("SET", "input", StringLiteralNode{Value: "a, b ,c"}, nil),
				createTestStep("SET", "output", StringLiteralNode{Value: ""}, nil),
				createForStep("item", VariableNode{Name: "input"}, []Step{
					createTestStep("SET", "output",
						BinaryOpNode{
							Left:     BinaryOpNode{Left: VariableNode{Name: "output"}, Operator: "+", Right: StringLiteralNode{Value: "-"}}, // output + "-"
							Operator: "+", Right: VariableNode{Name: "item"},                                                                // + item
						}, nil),
				}),
			},
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"input": "a, b ,c", "output": "-a-b-c"}, // Corrected expected output
			expectedResult: nil, expectError: false,
		},

		// --- FOR EACH List Iteration ---
		{
			name: "FOR EACH list literal", // Should now pass with fixed concat
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
			initialVars:    map[string]interface{}{},
			expectedVars:   map[string]interface{}{"output": "1|X|true|[nest]|"},
			expectedResult: nil, expectError: false,
		},
		{
			name: "FOR EACH list variable", // Should now pass with fixed concat
			inputSteps: []Step{
				createTestStep("SET", "output", StringLiteralNode{Value: ""}, nil),
				createForStep("val", VariableNode{Name: "myListVar"}, []Step{
					createTestStep("SET", "output", BinaryOpNode{Left: VariableNode{Name: "output"}, Operator: "+", Right: VariableNode{Name: "val"}}, nil),
				}),
			},
			initialVars:    map[string]interface{}{"myListVar": []interface{}{"A", "B", int64(3)}},
			expectedVars:   map[string]interface{}{"myListVar": []interface{}{"A", "B", int64(3)}, "output": "AB3"},
			expectedResult: nil, expectError: false,
		},

		// --- FOR EACH Map Iteration (Keys) ---
		{
			name: "FOR EACH map literal keys", // Corrected expectation based on loop logic
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
			expectedVars:   map[string]interface{}{"output": "a,b,"}, // Expect sorted keys with comma
			expectedResult: nil, expectError: false,
		},
		{
			name: "FOR EACH map variable keys", // Corrected expectation based on loop logic
			inputSteps: []Step{
				createTestStep("SET", "output", StringLiteralNode{Value: ""}, nil),
				createForStep("k", VariableNode{Name: "myMap"}, []Step{
					createTestStep("SET", "output", BinaryOpNode{Left: VariableNode{Name: "output"}, Operator: "+", Right: VariableNode{Name: "k"}}, nil),
				}),
			},
			initialVars:    map[string]interface{}{"myMap": map[string]interface{}{"z": true, "x": "hello", "a": 1}},
			expectedVars:   map[string]interface{}{"myMap": map[string]interface{}{"z": true, "x": "hello", "a": 1}, "output": "axz"}, // Expect sorted key concat
			expectedResult: nil, expectError: false,
		},

		// --- Tool Call Tests ---
		{name: "CALL TOOL StringLength AST", inputSteps: []Step{createTestStep("SET", "myStr", StringLiteralNode{Value: "Test"}, nil), createTestStep("CALL", "TOOL.StringLength", nil, []interface{}{VariableNode{Name: "myStr"}}), createTestStep("SET", "lenResult", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "Test", "lenResult": int64(4)}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL Substring AST", inputSteps: []Step{createTestStep("CALL", "TOOL.Substring", nil, []interface{}{StringLiteralNode{Value: "ABCDE"}, NumberLiteralNode{Value: int64(1)}, NumberLiteralNode{Value: int64(4)}}), createTestStep("SET", "sub", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"sub": "BCD"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL Substring Wrong Arg Type AST", inputSteps: []Step{createTestStep("CALL", "TOOL.Substring", nil, []interface{}{StringLiteralNode{Value: "hello"}, StringLiteralNode{Value: "one"}, NumberLiteralNode{Value: int64(3)}})}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, errorContains: "cannot be converted to int"},
		{name: "CALL TOOL JoinStrings with ListLiteral", inputSteps: []Step{createTestStep("CALL", "TOOL.JoinStrings", nil, []interface{}{ListLiteralNode{Elements: []interface{}{StringLiteralNode{Value: "A"}, NumberLiteralNode{Value: int64(1)}, BooleanLiteralNode{Value: true}}}, StringLiteralNode{Value: "-"}}), createTestStep("SET", "joined", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"joined": "A-1-true"}, expectedResult: nil, expectError: false},
	} // End testCases slice

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { runExecuteStepsTest(t, tc) })
	}
} // End TestExecuteStepsBlocksAndLoops
