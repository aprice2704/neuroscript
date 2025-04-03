package core

import (
	"reflect" // Import sort package for stable map key iteration
	"strings"
	"testing"
	// Import sort
)

// --- Interpreter Test Specific Helper ---
func newTestInterpreter(vars map[string]interface{}, lastResult interface{}) *Interpreter {
	interp := NewInterpreter(nil)
	if vars != nil {
		interp.variables = make(map[string]interface{}, len(vars))
		for k, v := range vars {
			interp.variables[k] = v
		}
	} else {
		interp.variables = make(map[string]interface{})
	}
	interp.lastCallResult = lastResult
	return interp
}
func createTestStep(typ string, target string, condNode interface{}, valueNode interface{}, argNodes []interface{}) Step {
	// Use the correct signature for newStep
	return newStep(typ, target, condNode, valueNode, nil, argNodes) // Pass nil for ElseValue
}
func createIfStep(condNode interface{}, thenSteps []Step, elseSteps []Step) Step {
	// Helper specifically for IF steps, including ElseValue
	return Step{Type: "IF", Cond: condNode, Value: thenSteps, ElseValue: elseSteps}
}

// --- Test Suite for executeSteps (Blocks, Loops, Tools) ---
type executeStepsTestCase struct {
	name           string
	inputSteps     []Step
	initialVars    map[string]interface{}
	expectedVars   map[string]interface{}
	expectedResult interface{}
	expectError    bool
	errorContains  string
}

func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	interp := newTestInterpreter(nil, nil)
	if tc.initialVars != nil {
		interp.variables = make(map[string]interface{}, len(tc.initialVars))
		for k, v := range tc.initialVars {
			interp.variables[k] = v
		}
	}
	finalResult, _, err := interp.executeSteps(tc.inputSteps) // Ignore wasReturn for now
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
	}
	if !tc.expectError && !reflect.DeepEqual(finalResult, tc.expectedResult) {
		t.Errorf("Test %q: Final result mismatch:\nExpected: %v (%T)\nGot:      %v (%T)", tc.name, tc.expectedResult, tc.expectedResult, finalResult, finalResult)
	}
	if !tc.expectError && err == nil && tc.expectedVars != nil {
		// Sort map keys before comparison if maps are involved in expectedVars
		for key, expectedValue := range tc.expectedVars {
			actualValue, exists := interp.variables[key]
			if !exists {
				t.Errorf("Test %q: Expected variable '%s' not found", tc.name, key)
				continue
			}

			// Deep comparison handles slices well, need special care only if comparing maps directly
			if !reflect.DeepEqual(actualValue, expectedValue) {
				t.Errorf("Test %q: Variable '%s' mismatch:\nExpected: %v (%T)\nGot:      %v (%T)", tc.name, key, expectedValue, expectedValue, actualValue, actualValue)
			}
		}
		if len(interp.variables) != len(tc.expectedVars) {
			extraVars := []string{}
			for k := range interp.variables {
				if _, expected := tc.expectedVars[k]; !expected {
					extraVars = append(extraVars, k)
				}
			}
			if len(extraVars) > 0 {
				t.Errorf("Test %q: Unexpected variables: %v", tc.name, extraVars)
			}
		}
	}
}

// TestExecuteStepsBlocksAndLoops - Includes List/Map iteration
func TestExecuteStepsBlocksAndLoops(t *testing.T) {
	testCases := []executeStepsTestCase{
		// --- Basic IF/WHILE Tests ---
		{name: "IF true literal", inputSteps: []Step{createTestStep("IF", "", BooleanLiteralNode{Value: true}, []Step{createTestStep("SET", "x", nil, StringLiteralNode{Value: "Inside"}, nil)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"x": "Inside"}, expectedResult: nil, expectError: false},
		{name: "IF false literal", inputSteps: []Step{createTestStep("IF", "", BooleanLiteralNode{Value: false}, []Step{createTestStep("SET", "x", nil, StringLiteralNode{Value: "Inside"}, nil)}, nil), createTestStep("SET", "y", nil, StringLiteralNode{Value: "Outside"}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"y": "Outside"}, expectedResult: nil, expectError: false},
		{name: "IF condition var true", inputSteps: []Step{createTestStep("IF", "", VariableNode{Name: "cond_var"}, []Step{createTestStep("SET", "x", nil, StringLiteralNode{Value: "Inside"}, nil)}, nil)}, initialVars: map[string]interface{}{"cond_var": true}, expectedVars: map[string]interface{}{"cond_var": true, "x": "Inside"}, expectedResult: nil, expectError: false},
		{name: "IF block with RETURN", inputSteps: []Step{createTestStep("SET", "status", nil, StringLiteralNode{Value: "Started"}, nil), createTestStep("IF", "", BooleanLiteralNode{Value: true}, []Step{createTestStep("SET", "x", nil, StringLiteralNode{Value: "Inside"}, nil), createTestStep("RETURN", "", nil, StringLiteralNode{Value: "ReturnedFromIf"}, nil), createTestStep("SET", "y", nil, StringLiteralNode{Value: "NotReached"}, nil)}, nil), createTestStep("SET", "status", nil, StringLiteralNode{Value: "Finished"}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"status": "Started", "x": "Inside"}, expectedResult: "ReturnedFromIf", expectError: false},
		{name: "WHILE runs once", inputSteps: []Step{createTestStep("SET", "run", nil, BooleanLiteralNode{Value: true}, nil), createTestStep("SET", "counter", nil, NumberLiteralNode{Value: int64(0)}, nil), createTestStep("WHILE", "", VariableNode{Name: "run"}, []Step{createTestStep("SET", "run", nil, BooleanLiteralNode{Value: false}, nil), createTestStep("SET", "counter", nil, NumberLiteralNode{Value: int64(1)}, nil)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"run": false, "counter": int64(1)}, expectedResult: nil, expectError: false},
		{name: "WHILE false initially", inputSteps: []Step{createTestStep("WHILE", "", BooleanLiteralNode{Value: false}, []Step{createTestStep("SET", "x", nil, StringLiteralNode{Value: "Never"}, nil)}, nil), createTestStep("SET", "y", nil, StringLiteralNode{Value: "After"}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"y": "After"}, expectedResult: nil, expectError: false},

		// --- FOR EACH String Iteration ---
		{name: "FOR EACH string char iteration", inputSteps: []Step{createTestStep("SET", "input", nil, StringLiteralNode{Value: "Hi!"}, nil), createTestStep("SET", "output", nil, StringLiteralNode{Value: ""}, nil), createTestStep("FOR", "char", VariableNode{Name: "input"}, []Step{createTestStep("SET", "output", nil, ConcatenationNode{Operands: []interface{}{VariableNode{Name: "output"}, VariableNode{Name: "char"}, StringLiteralNode{Value: "-"}}}, nil)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"input": "Hi!", "output": "H-i-!-"}, expectedResult: nil, expectError: false},
		{name: "FOR EACH comma split", inputSteps: []Step{createTestStep("SET", "input", nil, StringLiteralNode{Value: "a, b ,c"}, nil), createTestStep("SET", "output", nil, StringLiteralNode{Value: ""}, nil), createTestStep("FOR", "item", VariableNode{Name: "input"}, []Step{createTestStep("SET", "output", nil, ConcatenationNode{Operands: []interface{}{VariableNode{Name: "output"}, StringLiteralNode{Value: "-"}, VariableNode{Name: "item"}}}, nil)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"input": "a, b ,c", "output": "-a-b-c"}, expectedResult: nil, expectError: false}, // Fixed expected output

		// --- FOR EACH List Iteration (NEW) ---
		{
			name: "FOR EACH list literal",
			inputSteps: []Step{
				createTestStep("SET", "output", nil, StringLiteralNode{Value: ""}, nil),
				createTestStep("FOR", "item", ListLiteralNode{Elements: []interface{}{
					NumberLiteralNode{Value: int64(1)},
					StringLiteralNode{Value: "X"},
					BooleanLiteralNode{Value: true},
					ListLiteralNode{Elements: []interface{}{StringLiteralNode{Value: "nest"}}}, // Nested list
				}}, []Step{
					createTestStep("SET", "output", nil, ConcatenationNode{Operands: []interface{}{
						VariableNode{Name: "output"},
						VariableNode{Name: "item"}, // Item will hold element value
						StringLiteralNode{Value: "|"},
					}}, nil),
				}, nil),
			},
			initialVars: map[string]interface{}{},
			// Expected output reflects how Go prints slices/bools within fmt.Sprintf("%v") used by concat
			expectedVars:   map[string]interface{}{"output": "1|X|true|[nest]|"},
			expectedResult: nil,
			expectError:    false,
		},
		{
			name: "FOR EACH list variable",
			inputSteps: []Step{
				createTestStep("SET", "output", nil, StringLiteralNode{Value: ""}, nil),
				createTestStep("FOR", "val", VariableNode{Name: "myListVar"}, []Step{
					createTestStep("SET", "output", nil, ConcatenationNode{Operands: []interface{}{
						VariableNode{Name: "output"}, VariableNode{Name: "val"}}}, nil),
				}, nil),
			},
			initialVars:    map[string]interface{}{"myListVar": []interface{}{"A", "B", int64(3)}},
			expectedVars:   map[string]interface{}{"myListVar": []interface{}{"A", "B", int64(3)}, "output": "AB3"},
			expectedResult: nil,
			expectError:    false,
		},

		// --- FOR EACH Map Iteration (Keys) ---
		{name: "FOR EACH map literal keys", inputSteps: []Step{createTestStep("SET", "output", nil, StringLiteralNode{Value: ""}, nil), createTestStep("FOR", "key", MapLiteralNode{Entries: []MapEntryNode{{Key: StringLiteralNode{Value: "b"}, Value: NumberLiteralNode{Value: int64(2)}}, {Key: StringLiteralNode{Value: "a"}, Value: NumberLiteralNode{Value: int64(1)}}}}, []Step{createTestStep("SET", "output", nil, ConcatenationNode{Operands: []interface{}{VariableNode{Name: "output"}, VariableNode{Name: "key"}, StringLiteralNode{Value: ","}}}, nil)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"output": "a,b,"}, expectedResult: nil, expectError: false},
		{name: "FOR EACH map variable keys", inputSteps: []Step{createTestStep("SET", "output", nil, StringLiteralNode{Value: ""}, nil), createTestStep("FOR", "k", VariableNode{Name: "myMap"}, []Step{createTestStep("SET", "output", nil, ConcatenationNode{Operands: []interface{}{VariableNode{Name: "output"}, VariableNode{Name: "k"}}}, nil)}, nil)}, initialVars: map[string]interface{}{"myMap": map[string]interface{}{"z": true, "x": "hello", "a": 1}}, expectedVars: map[string]interface{}{"myMap": map[string]interface{}{"z": true, "x": "hello", "a": 1}, "output": "axz"}, expectedResult: nil, expectError: false},

		// --- Tool Call Tests --- (Remain the same)
		{name: "CALL TOOL StringLength AST", inputSteps: []Step{createTestStep("SET", "myStr", nil, StringLiteralNode{Value: "Test"}, nil), createTestStep("CALL", "TOOL.StringLength", nil, nil, []interface{}{VariableNode{Name: "myStr"}}), createTestStep("SET", "lenResult", nil, LastCallResultNode{}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myStr": "Test", "lenResult": int64(4)}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL Substring AST", inputSteps: []Step{createTestStep("CALL", "TOOL.Substring", nil, nil, []interface{}{StringLiteralNode{Value: "ABCDE"}, NumberLiteralNode{Value: int64(1)}, NumberLiteralNode{Value: int64(4)}}), createTestStep("SET", "sub", nil, LastCallResultNode{}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"sub": "BCD"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL Substring Wrong Arg Type AST", inputSteps: []Step{createTestStep("CALL", "TOOL.Substring", nil, nil, []interface{}{StringLiteralNode{Value: "hello"}, StringLiteralNode{Value: "one"}, NumberLiteralNode{Value: int64(3)}})}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, errorContains: "cannot be converted to int"},
		{name: "CALL TOOL JoinStrings with ListLiteral", inputSteps: []Step{createTestStep("CALL", "TOOL.JoinStrings", nil, nil, []interface{}{ListLiteralNode{Elements: []interface{}{StringLiteralNode{Value: "A"}, NumberLiteralNode{Value: int64(1)}, BooleanLiteralNode{Value: true}}}, StringLiteralNode{Value: "-"}}), createTestStep("SET", "joined", nil, LastCallResultNode{}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"joined": "A-1-true"}, expectedResult: nil, expectError: false},
	} // End testCases slice

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { runExecuteStepsTest(t, tc) })
	}
} // End TestExecuteStepsBlocksAndLoops
