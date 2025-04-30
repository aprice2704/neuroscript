// filename: pkg/core/interpreter_test.go
package core

import (
	"testing"
)

// --- (executeStepsTestCase struct and runExecuteStepsTest helper unchanged) ---
// Assuming these are defined correctly in testing_helpers_test.go or similar
type executeStepsTestCase struct {
	name            string
	inputSteps      []Step
	initialVars     map[string]interface{}
	expectedVars    map[string]interface{}
	expectedResult  interface{}
	expectError     bool
	ExpectedErrorIs error // Sentinel error to check with errors.Is
}

// Assuming runExecuteStepsTest helper exists and works correctly
// func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) { ... }

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
		// --- Existing Tests (Keep all valid ones, remove old CALL tests) ---
		{name: "IF true literal", inputSteps: []Step{createIfStep(BooleanLiteralNode{Value: true}, []Step{createTestStep("set", "x", StringLiteralNode{Value: "Inside"}, nil)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"x": "Inside"}, expectedResult: nil, expectError: false},
		{name: "IF block with RETURN", inputSteps: []Step{createTestStep("set", "status", StringLiteralNode{Value: "Started"}, nil), createIfStep(BooleanLiteralNode{Value: true}, []Step{createTestStep("set", "x", StringLiteralNode{Value: "Inside"}, nil), createTestStep("return", "", []interface{}{StringLiteralNode{Value: "ReturnedFromIf"}}, nil), createTestStep("set", "y", StringLiteralNode{Value: "NotReached"}, nil)}, nil), createTestStep("set", "status", StringLiteralNode{Value: "Finished"}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"status": "Started", "x": "Inside"}, expectedResult: "ReturnedFromIf", expectError: false},

		// --- RETURN Tests ---
		{name: "RETURN single value", inputSteps: []Step{createTestStep("return", "", []interface{}{NumberLiteralNode{Value: int64(42)}}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: int64(42), expectError: false},
		{name: "RETURN multiple values", inputSteps: []Step{createTestStep("return", "", []interface{}{StringLiteralNode{Value: "hello"}, NumberLiteralNode{Value: int64(10)}, BooleanLiteralNode{Value: true}}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: []interface{}{"hello", int64(10), true}, expectError: false},
		{name: "RETURN no value", inputSteps: []Step{createTestStep("return", "", nil, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "RETURN value from variable", inputSteps: []Step{createTestStep("set", "myVar", StringLiteralNode{Value: "data"}, nil), createTestStep("return", "", []interface{}{VariableNode{Name: "myVar"}}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myVar": "data"}, expectedResult: "data", expectError: false},
		{name: "RETURN multiple values including variable", inputSteps: []Step{createTestStep("set", "myVar", BooleanLiteralNode{Value: false}, nil), createTestStep("return", "", []interface{}{NumberLiteralNode{Value: int64(1)}, VariableNode{Name: "myVar"}, NumberLiteralNode{Value: 3.14}}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myVar": false}, expectedResult: []interface{}{int64(1), false, 3.14}, expectError: false},

		// --- MUST Tests (Unchanged) ---
		{name: "MUST true literal", inputSteps: []Step{createTestStep("must", "", BooleanLiteralNode{Value: true}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST false literal", inputSteps: []Step{createTestStep("must", "", BooleanLiteralNode{Value: false}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST non-zero number", inputSteps: []Step{createTestStep("must", "", NumberLiteralNode{Value: int64(1)}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST zero number", inputSteps: []Step{createTestStep("must", "", NumberLiteralNode{Value: int64(0)}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST non-empty string ('true')", inputSteps: []Step{createTestStep("must", "", StringLiteralNode{Value: "true"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST non-empty string ('1')", inputSteps: []Step{createTestStep("must", "", StringLiteralNode{Value: "1"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST empty string", inputSteps: []Step{createTestStep("must", "", StringLiteralNode{Value: ""}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST non-empty string ('other')", inputSteps: []Step{createTestStep("must", "", StringLiteralNode{Value: "other"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST nil", inputSteps: []Step{createTestStep("must", "", VariableNode{Name: "nilVar"}, nil)}, initialVars: map[string]interface{}{"nilVar": nil}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST true variable", inputSteps: []Step{createTestStep("must", "", VariableNode{Name: "t"}, nil)}, initialVars: map[string]interface{}{"t": true}, expectedResult: nil, expectError: false},
		{name: "MUST last result (true)", inputSteps: []Step{createTestStep("set", "_ignored", BooleanLiteralNode{Value: true}, nil), createTestStep("must", "", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST last result (false)", inputSteps: []Step{createTestStep("set", "_ignored", BooleanLiteralNode{Value: false}, nil), createTestStep("must", "", LastNode{}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST expression (1 > 0)", inputSteps: []Step{createTestStep("must", "", BinaryOpNode{Left: NumberLiteralNode{Value: int64(1)}, Operator: ">", Right: NumberLiteralNode{Value: int64(0)}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST expression (1 < 0)", inputSteps: []Step{createTestStep("must", "", BinaryOpNode{Left: NumberLiteralNode{Value: int64(1)}, Operator: "<", Right: NumberLiteralNode{Value: int64(0)}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST evaluation error", inputSteps: []Step{createTestStep("must", "", BinaryOpNode{Left: NumberLiteralNode{Value: int64(1)}, Operator: "+", Right: StringLiteralNode{Value: "a"}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},

		// --- MUSTBE Tests (UPDATED: Using CallableExprNode in Value field, Target field updated) ---
		{name: "MUSTBE is_string pass", inputSteps: []Step{createTestStep("mustbe", "is_string", CallableExprNode{Target: CallTarget{Name: "is_string"}, Arguments: []interface{}{VariableNode{Name: "s"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_string fail", inputSteps: []Step{createTestStep("mustbe", "is_string", CallableExprNode{Target: CallTarget{Name: "is_string"}, Arguments: []interface{}{VariableNode{Name: "n"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_number pass int", inputSteps: []Step{createTestStep("mustbe", "is_number", CallableExprNode{Target: CallTarget{Name: "is_number"}, Arguments: []interface{}{VariableNode{Name: "n"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_number pass float", inputSteps: []Step{createTestStep("mustbe", "is_number", CallableExprNode{Target: CallTarget{Name: "is_number"}, Arguments: []interface{}{VariableNode{Name: "f"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_number fail", inputSteps: []Step{createTestStep("mustbe", "is_number", CallableExprNode{Target: CallTarget{Name: "is_number"}, Arguments: []interface{}{VariableNode{Name: "s"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_int pass", inputSteps: []Step{createTestStep("mustbe", "is_int", CallableExprNode{Target: CallTarget{Name: "is_int"}, Arguments: []interface{}{VariableNode{Name: "n"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_int fail (float)", inputSteps: []Step{createTestStep("mustbe", "is_int", CallableExprNode{Target: CallTarget{Name: "is_int"}, Arguments: []interface{}{VariableNode{Name: "f"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_float pass", inputSteps: []Step{createTestStep("mustbe", "is_float", CallableExprNode{Target: CallTarget{Name: "is_float"}, Arguments: []interface{}{VariableNode{Name: "f"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_float fail (int)", inputSteps: []Step{createTestStep("mustbe", "is_float", CallableExprNode{Target: CallTarget{Name: "is_float"}, Arguments: []interface{}{VariableNode{Name: "n"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_bool pass", inputSteps: []Step{createTestStep("mustbe", "is_bool", CallableExprNode{Target: CallTarget{Name: "is_bool"}, Arguments: []interface{}{VariableNode{Name: "b"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_bool fail", inputSteps: []Step{createTestStep("mustbe", "is_bool", CallableExprNode{Target: CallTarget{Name: "is_bool"}, Arguments: []interface{}{VariableNode{Name: "s"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_list pass", inputSteps: []Step{createTestStep("mustbe", "is_list", CallableExprNode{Target: CallTarget{Name: "is_list"}, Arguments: []interface{}{VariableNode{Name: "l"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_list fail", inputSteps: []Step{createTestStep("mustbe", "is_list", CallableExprNode{Target: CallTarget{Name: "is_list"}, Arguments: []interface{}{VariableNode{Name: "m"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE is_map pass", inputSteps: []Step{createTestStep("mustbe", "is_map", CallableExprNode{Target: CallTarget{Name: "is_map"}, Arguments: []interface{}{VariableNode{Name: "m"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_map fail", inputSteps: []Step{createTestStep("mustbe", "is_map", CallableExprNode{Target: CallTarget{Name: "is_map"}, Arguments: []interface{}{VariableNode{Name: "l"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE not_empty pass (string)", inputSteps: []Step{createTestStep("mustbe", "not_empty", CallableExprNode{Target: CallTarget{Name: "not_empty"}, Arguments: []interface{}{VariableNode{Name: "s"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE not_empty pass (list)", inputSteps: []Step{createTestStep("mustbe", "not_empty", CallableExprNode{Target: CallTarget{Name: "not_empty"}, Arguments: []interface{}{VariableNode{Name: "l"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE not_empty pass (map)", inputSteps: []Step{createTestStep("mustbe", "not_empty", CallableExprNode{Target: CallTarget{Name: "not_empty"}, Arguments: []interface{}{VariableNode{Name: "m"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE not_empty fail (empty string)", inputSteps: []Step{createTestStep("mustbe", "not_empty", CallableExprNode{Target: CallTarget{Name: "not_empty"}, Arguments: []interface{}{VariableNode{Name: "emptyS"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE not_empty fail (empty list)", inputSteps: []Step{createTestStep("mustbe", "not_empty", CallableExprNode{Target: CallTarget{Name: "not_empty"}, Arguments: []interface{}{VariableNode{Name: "emptyL"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE not_empty fail (empty map)", inputSteps: []Step{createTestStep("mustbe", "not_empty", CallableExprNode{Target: CallTarget{Name: "not_empty"}, Arguments: []interface{}{VariableNode{Name: "emptyM"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE not_empty fail (nil)", inputSteps: []Step{createTestStep("mustbe", "not_empty", CallableExprNode{Target: CallTarget{Name: "not_empty"}, Arguments: []interface{}{VariableNode{Name: "nilV"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE unknown function", inputSteps: []Step{createTestStep("mustbe", "is_banana", CallableExprNode{Target: CallTarget{Name: "is_banana"}, Arguments: []interface{}{BooleanLiteralNode{Value: true}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},                             // Fails evaluating the check function
		{name: "MUSTBE wrong arg count", inputSteps: []Step{createTestStep("mustbe", "is_string", CallableExprNode{Target: CallTarget{Name: "is_string"}, Arguments: []interface{}{StringLiteralNode{Value: "a"}, StringLiteralNode{Value: "b"}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed}, // Fails evaluating the check function (incorrect arg count)
		{name: "MUSTBE argument evaluation error", inputSteps: []Step{createTestStep("mustbe", "is_string", CallableExprNode{Target: CallTarget{Name: "is_string"}, Arguments: []interface{}{VariableNode{Name: "missing"}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},                      // Fails evaluating the argument

		// +++ END Tests +++

	} // End testCases slice

	// Run tests
	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() // Disable parallel for now if tests modify shared state or for easier debugging
			runExecuteStepsTest(t, tc) // Assumes runExecuteStepsTest is defined elsewhere
		})
	}
} // End TestExecuteStepsBlocksAndLoops

// --- Helpers defined in testing_helpers_test.go ---
// createTestStep
// createIfStep
// etc.
