// filename: pkg/core/interpreter_test.go
package core

import (
	"testing"
	// Assuming Position is defined in this package (ast.go)
	// Assuming Step struct is defined in ast.go
	// Assuming error variables like ErrMustConditionFailed are defined in errors.go
	// Assuming helper functions like createTestStep, createIfStep are defined in testing_helpers_test.go
	// Assuming EvalTestCase and runExecuteStepsTest are defined in testing_helpers_test.go
)

// --- (executeStepsTestCase struct and runExecuteStepsTest helper unchanged) ---

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

	// Define dummyPos using local Position type pointer
	dummyPos := &Position{Line: 1, Column: 1}

	testCases := []executeStepsTestCase{
		// --- Existing Tests ---
		{name: "IF true literal", inputSteps: []Step{createIfStep(&BooleanLiteralNode{Pos: dummyPos, Value: true}, []Step{createTestStep("set", "x", &StringLiteralNode{Pos: dummyPos, Value: "Inside"}, nil)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"x": "Inside"}, expectedResult: nil, expectError: false},
		// --- MODIFIED RETURN STEPS: Use []Expression{} type literal ---
		{name: "IF block with RETURN", inputSteps: []Step{
			createTestStep("set", "status", &StringLiteralNode{Pos: dummyPos, Value: "Started"}, nil),
			createIfStep(&BooleanLiteralNode{Pos: dummyPos, Value: true}, []Step{
				createTestStep("set", "x", &StringLiteralNode{Pos: dummyPos, Value: "Inside"}, nil),
				// Changed []interface{} to []Expression
				createTestStep("return", "", []Expression{&StringLiteralNode{Pos: dummyPos, Value: "ReturnedFromIf"}}, nil),
				createTestStep("set", "y", &StringLiteralNode{Pos: dummyPos, Value: "NotReached"}, nil),
			}, nil),
			createTestStep("set", "status", &StringLiteralNode{Pos: dummyPos, Value: "Finished"}, nil),
		}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"status": "Started", "x": "Inside"}, expectedResult: "ReturnedFromIf", expectError: false},

		// --- RETURN Tests (MODIFIED) ---
		{name: "RETURN single value", inputSteps: []Step{createTestStep("return", "", []Expression{&NumberLiteralNode{Pos: dummyPos, Value: int64(42)}}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: int64(42), expectError: false},
		{name: "RETURN multiple values", inputSteps: []Step{createTestStep("return", "", []Expression{&StringLiteralNode{Pos: dummyPos, Value: "hello"}, &NumberLiteralNode{Pos: dummyPos, Value: int64(10)}, &BooleanLiteralNode{Pos: dummyPos, Value: true}}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: []interface{}{"hello", int64(10), true}, expectError: false},
		// RETURN no value still uses nil, which is correct
		{name: "RETURN no value", inputSteps: []Step{createTestStep("return", "", nil, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "RETURN value from variable", inputSteps: []Step{
			createTestStep("set", "myVar", &StringLiteralNode{Pos: dummyPos, Value: "data"}, nil),
			createTestStep("return", "", []Expression{&VariableNode{Pos: dummyPos, Name: "myVar"}}, nil),
		}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myVar": "data"}, expectedResult: "data", expectError: false},
		{name: "RETURN multiple values including variable", inputSteps: []Step{
			createTestStep("set", "myVar", &BooleanLiteralNode{Pos: dummyPos, Value: false}, nil),
			createTestStep("return", "", []Expression{&NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, &VariableNode{Pos: dummyPos, Name: "myVar"}, &NumberLiteralNode{Pos: dummyPos, Value: 3.14}}, nil),
		}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myVar": false}, expectedResult: []interface{}{int64(1), false, 3.14}, expectError: false},
		// --- END RETURN MODIFICATIONS ---

		// --- MUST Tests (Corrected based on previous pointer fix) ---
		{name: "MUST true literal", inputSteps: []Step{createTestStep("must", "", &BooleanLiteralNode{Pos: dummyPos, Value: true}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST false literal", inputSteps: []Step{createTestStep("must", "", &BooleanLiteralNode{Pos: dummyPos, Value: false}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST non-zero number", inputSteps: []Step{createTestStep("must", "", &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST zero number", inputSteps: []Step{createTestStep("must", "", &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST non-empty string ('true')", inputSteps: []Step{createTestStep("must", "", &StringLiteralNode{Pos: dummyPos, Value: "true"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST non-empty string ('1')", inputSteps: []Step{createTestStep("must", "", &StringLiteralNode{Pos: dummyPos, Value: "1"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST empty string", inputSteps: []Step{createTestStep("must", "", &StringLiteralNode{Pos: dummyPos, Value: ""}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST non-empty string ('other')", inputSteps: []Step{createTestStep("must", "", &StringLiteralNode{Pos: dummyPos, Value: "other"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST nil", inputSteps: []Step{createTestStep("must", "", &VariableNode{Pos: dummyPos, Name: "nilVar"}, nil)}, initialVars: map[string]interface{}{"nilVar": nil}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST true variable", inputSteps: []Step{createTestStep("must", "", &VariableNode{Pos: dummyPos, Name: "t"}, nil)}, initialVars: map[string]interface{}{"t": true}, expectedResult: nil, expectError: false},
		{name: "MUST last result (true)", inputSteps: []Step{createTestStep("set", "_ignored", &BooleanLiteralNode{Pos: dummyPos, Value: true}, nil), createTestStep("must", "", &LastNode{Pos: dummyPos}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST last result (false)", inputSteps: []Step{createTestStep("set", "_ignored", &BooleanLiteralNode{Pos: dummyPos, Value: false}, nil), createTestStep("must", "", &LastNode{Pos: dummyPos}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST expression (1 > 0)", inputSteps: []Step{createTestStep("must", "", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, Operator: ">", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "MUST expression (1 < 0)", inputSteps: []Step{createTestStep("must", "", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, Operator: "<", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST evaluation error", inputSteps: []Step{createTestStep("must", "", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, Operator: "+", Right: &StringLiteralNode{Pos: dummyPos, Value: "a"}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed}, // Error wrapped by MUST

		// --- MUSTBE Tests (Corrected based on previous pointer fix) ---
		{name: "MUSTBE is_string pass", inputSteps: []Step{createTestStep("mustbe", "is_string", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "is_string"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "s"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: false},
		{name: "MUSTBE is_string fail", inputSteps: []Step{createTestStep("mustbe", "is_string", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "is_string"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "n"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		// ... other MUSTBE tests assumed correct ...
		{name: "MUSTBE not_empty fail (nil)", inputSteps: []Step{createTestStep("mustbe", "not_empty", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "not_empty"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "nilV"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE unknown function", inputSteps: []Step{createTestStep("mustbe", "is_banana", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "is_banana"}, Arguments: []Expression{&BooleanLiteralNode{Pos: dummyPos, Value: true}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE wrong arg count", inputSteps: []Step{createTestStep("mustbe", "is_string", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "is_string"}, Arguments: []Expression{&StringLiteralNode{Pos: dummyPos, Value: "a"}, &StringLiteralNode{Pos: dummyPos, Value: "b"}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE argument evaluation error", inputSteps: []Step{createTestStep("mustbe", "is_string", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "is_string"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "missing"}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
	} // End testCases slice

	// Run tests
	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() // Disable parallel for now
			runExecuteStepsTest(t, tc) // Assumes runExecuteStepsTest is defined elsewhere
		})
	}
} // End TestExecuteStepsBlocksAndLoops
