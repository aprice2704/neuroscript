// filename: neuroscript/pkg/core/interpreter_test.go
package core

import (
	"errors"
	"reflect"
	"sort"
	"testing"
)

// --- (executeStepsTestCase struct and runExecuteStepsTest helper unchanged) ---
type executeStepsTestCase struct {
	name            string
	inputSteps      []Step
	initialVars     map[string]interface{}
	expectedVars    map[string]interface{}
	expectedResult  interface{}
	expectError     bool
	ExpectedErrorIs error
}

func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	interp, _ := NewTestInterpreter(t, tc.initialVars, nil)
	finalResult, wasReturn, _, err := interp.executeSteps(tc.inputSteps, false, nil)
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
	} else {
		if err != nil {
			t.Errorf("Test %q: Unexpected error: %v", tc.name, err)
		}
		expectedExecResult := tc.expectedResult
		actualExecResult := finalResult
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
			actualExecResult = nil
		}
		if !reflect.DeepEqual(actualExecResult, expectedExecResult) {
			t.Errorf("Test %q: Final execution result mismatch:\nExpected: %v (%T)\nGot:      %v (%T) (WasReturn: %t)", tc.name, expectedExecResult, expectedExecResult, actualExecResult, actualExecResult, wasReturn)
		}
	}
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
	// Add some common variables for mustbe tests
	mustBeVars := map[string]interface{}{
		"s":      "a string",
		"n":      int64(10),
		"f":      3.14,
		"b":      true,
		"l":      []interface{}{1, 2},
		"m":      map[string]interface{}{"a": 1},
		"emptyS": "",
		"emptyL": []interface{}{},
		"emptyM": map[string]interface{}{},
		"zeroN":  int64(0),
		"nilV":   nil,
	}

	testCases := []executeStepsTestCase{
		// --- Existing Tests ---
		// ... (Keep all previous test cases) ...
		{name: "IF true literal", inputSteps: []Step{createIfStep(BooleanLiteralNode{Value: true}, []Step{createTestStep("set", "x", StringLiteralNode{Value: "Inside"}, nil)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"x": "Inside"}, expectedResult: nil, expectError: false},
		{name: "IF block with RETURN", inputSteps: []Step{createTestStep("set", "status", StringLiteralNode{Value: "Started"}, nil), createIfStep(BooleanLiteralNode{Value: true}, []Step{createTestStep("set", "x", StringLiteralNode{Value: "Inside"}, nil), createTestStep("return", "", []interface{}{StringLiteralNode{Value: "ReturnedFromIf"}}, nil), createTestStep("set", "y", StringLiteralNode{Value: "NotReached"}, nil)}, nil), createTestStep("set", "status", StringLiteralNode{Value: "Finished"}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"status": "Started", "x": "Inside"}, expectedResult: "ReturnedFromIf", expectError: false},
		{name: "CALL TOOL JoinStrings with ListLiteral", inputSteps: []Step{createTestStep("call", "tool.JoinStrings", nil, []interface{}{ListLiteralNode{Elements: []interface{}{StringLiteralNode{Value: "A"}, NumberLiteralNode{Value: int64(1)}, BooleanLiteralNode{Value: true}}}, StringLiteralNode{Value: "-"}}), createTestStep("set", "joined", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"joined": "A-1-true"}, expectedResult: nil, expectError: false},
		{name: "CALL TOOL Substring Wrong Arg Type AST", inputSteps: []Step{createTestStep("call", "tool.Substring", nil, []interface{}{StringLiteralNode{Value: "hello"}, StringLiteralNode{Value: "one"}, NumberLiteralNode{Value: int64(3)}})}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrValidationTypeMismatch},
		{name: "RETURN single value (backward compat)", inputSteps: []Step{createTestStep("return", "", []interface{}{NumberLiteralNode{Value: int64(42)}}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: int64(42), expectError: false},
		{name: "RETURN multiple values", inputSteps: []Step{createTestStep("return", "", []interface{}{StringLiteralNode{Value: "hello"}, NumberLiteralNode{Value: int64(10)}, BooleanLiteralNode{Value: true}}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: []interface{}{"hello", int64(10), true}, expectError: false},
		{name: "RETURN no value", inputSteps: []Step{createTestStep("return", "", nil, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "RETURN value from variable", inputSteps: []Step{createTestStep("set", "myVar", StringLiteralNode{Value: "data"}, nil), createTestStep("return", "", []interface{}{VariableNode{Name: "myVar"}}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myVar": "data"}, expectedResult: "data", expectError: false},
		{name: "RETURN multiple values including variable", inputSteps: []Step{createTestStep("set", "myVar", BooleanLiteralNode{Value: false}, nil), createTestStep("return", "", []interface{}{NumberLiteralNode{Value: int64(1)}, VariableNode{Name: "myVar"}, NumberLiteralNode{Value: 3.14}}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myVar": false}, expectedResult: []interface{}{int64(1), false, 3.14}, expectError: false},

		// +++ ADDED: MUST / MUSTBE Tests +++
		// --- MUST Tests ---
		{name: "MUST true literal", inputSteps: []Step{createTestStep("must", "", BooleanLiteralNode{Value: true}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST false literal", inputSteps: []Step{createTestStep("must", "", BooleanLiteralNode{Value: false}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST non-zero number", inputSteps: []Step{createTestStep("must", "", NumberLiteralNode{Value: int64(1)}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST zero number", inputSteps: []Step{createTestStep("must", "", NumberLiteralNode{Value: int64(0)}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST non-empty string ('true')", inputSteps: []Step{createTestStep("must", "", StringLiteralNode{Value: "true"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},                                           // Truthy string
		{name: "MUST non-empty string ('1')", inputSteps: []Step{createTestStep("must", "", StringLiteralNode{Value: "1"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},                                                 // Truthy string
		{name: "MUST empty string", inputSteps: []Step{createTestStep("must", "", StringLiteralNode{Value: ""}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},                    // Falsy string
		{name: "MUST non-empty string ('other')", inputSteps: []Step{createTestStep("must", "", StringLiteralNode{Value: "other"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed}, // Falsy string
		{name: "MUST nil", inputSteps: []Step{createTestStep("must", "", VariableNode{Name: "nilVar"}, nil)}, initialVars: map[string]interface{}{"nilVar": nil}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},                // Nil is falsy
		{name: "MUST true variable", inputSteps: []Step{createTestStep("must", "", VariableNode{Name: "t"}, nil)}, initialVars: map[string]interface{}{"t": true}, expectedResult: nil, expectError: false},
		{name: "MUST last result (true)", inputSteps: []Step{createTestStep("set", "_ignored", BooleanLiteralNode{Value: true}, nil), createTestStep("must", "", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST last result (false)", inputSteps: []Step{createTestStep("set", "_ignored", BooleanLiteralNode{Value: false}, nil), createTestStep("must", "", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST expression (1 > 0)", inputSteps: []Step{createTestStep("must", "", BinaryOpNode{Left: NumberLiteralNode{Value: int64(1)}, Operator: ">", Right: NumberLiteralNode{Value: int64(0)}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST expression (1 < 0)", inputSteps: []Step{createTestStep("must", "", BinaryOpNode{Left: NumberLiteralNode{Value: int64(1)}, Operator: "<", Right: NumberLiteralNode{Value: int64(0)}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST evaluation error", inputSteps: []Step{createTestStep("must", "", BinaryOpNode{Left: NumberLiteralNode{Value: int64(1)}, Operator: "+", Right: StringLiteralNode{Value: "a"}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed}, // Wrapped ErrInvalidOperandType

		// --- MUSTBE Tests ---
		{name: "MUSTBE is_string pass", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_string", Arguments: []interface{}{VariableNode{Name: "s"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_string fail", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_string", Arguments: []interface{}{VariableNode{Name: "n"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_number pass int", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_number", Arguments: []interface{}{VariableNode{Name: "n"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_number pass float", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_number", Arguments: []interface{}{VariableNode{Name: "f"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_number fail", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_number", Arguments: []interface{}{VariableNode{Name: "s"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_int pass", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_int", Arguments: []interface{}{VariableNode{Name: "n"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_int fail (float)", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_int", Arguments: []interface{}{VariableNode{Name: "f"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_float pass", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_float", Arguments: []interface{}{VariableNode{Name: "f"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_float fail (int)", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_float", Arguments: []interface{}{VariableNode{Name: "n"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_bool pass", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_bool", Arguments: []interface{}{VariableNode{Name: "b"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_bool fail", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_bool", Arguments: []interface{}{VariableNode{Name: "s"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_list pass", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_list", Arguments: []interface{}{VariableNode{Name: "l"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_list fail", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_list", Arguments: []interface{}{VariableNode{Name: "m"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_map pass", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_map", Arguments: []interface{}{VariableNode{Name: "m"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_map fail", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_map", Arguments: []interface{}{VariableNode{Name: "l"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE not_empty pass (string)", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "not_empty", Arguments: []interface{}{VariableNode{Name: "s"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE not_empty pass (list)", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "not_empty", Arguments: []interface{}{VariableNode{Name: "l"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE not_empty pass (map)", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "not_empty", Arguments: []interface{}{VariableNode{Name: "m"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE not_empty fail (empty string)", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "not_empty", Arguments: []interface{}{VariableNode{Name: "emptyS"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE not_empty fail (empty list)", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "not_empty", Arguments: []interface{}{VariableNode{Name: "emptyL"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE not_empty fail (empty map)", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "not_empty", Arguments: []interface{}{VariableNode{Name: "emptyM"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE not_empty fail (nil)", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "not_empty", Arguments: []interface{}{VariableNode{Name: "nilV"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE unknown function", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_banana", Arguments: []interface{}{BooleanLiteralNode{Value: true}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE wrong arg count", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_string", Arguments: []interface{}{StringLiteralNode{Value: "a"}, StringLiteralNode{Value: "b"}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE argument evaluation error", inputSteps: []Step{createTestStep("mustbe", "", FunctionCallNode{FunctionName: "is_string", Arguments: []interface{}{VariableNode{Name: "missing"}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed}, // Wrapped ErrVariableNotFound

		// +++ END ADDED Tests +++

	} // End testCases slice

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { runExecuteStepsTest(t, tc) })
	}
} // End TestExecuteStepsBlocksAndLoops
