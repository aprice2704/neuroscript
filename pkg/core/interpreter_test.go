// NeuroScript Version: 0.3.5
// File version: 0.0.5 // Corrected ExpectedErrorIs for tool arg type error.
// filename: pkg/core/interpreter_test.go
package core

import (
	"testing"
	// Assuming Position is defined in this package (ast.go)
	// Assuming Step struct is defined in ast.go
	// Assuming error variables like ErrMustConditionFailed are defined in errors.go
	// Helper functions like createTestStep, createIfStep are defined in testing_helpers.go
	// executeStepsTestCase and runExecuteStepsTest are defined in testing_helpers.go
)

// dummyPos is a shared position for test AST nodes.
var dummyPos = &Position{Line: 1, Column: 1, File: "test"}

// TestExecuteStepsBlocksAndLoops - Includes List/Map iteration and dotted tool calls
func TestExecuteStepsBlocksAndLoops(t *testing.T) {
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

	initialList := []interface{}{"item1", int64(2), true}

	testCases := []executeStepsTestCase{
		// --- Existing IF Tests ---
		{name: "IF true literal", inputSteps: []Step{createIfStep(dummyPos, &BooleanLiteralNode{Pos: dummyPos, Value: true}, []Step{createTestStep("set", "x", &StringLiteralNode{Pos: dummyPos, Value: "Inside"}, nil)}, []Step{})}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"x": "Inside"}, expectedResult: "Inside", expectError: false},
		{name: "IF block with RETURN", inputSteps: []Step{
			createTestStep("set", "status", &StringLiteralNode{Pos: dummyPos, Value: "Started"}, nil),
			createIfStep(dummyPos, &BooleanLiteralNode{Pos: dummyPos, Value: true}, []Step{
				createTestStep("set", "x", &StringLiteralNode{Pos: dummyPos, Value: "Inside"}, nil),
				createTestStep("return", "", []Expression{&StringLiteralNode{Pos: dummyPos, Value: "ReturnedFromIf"}}, nil),
				createTestStep("set", "y", &StringLiteralNode{Pos: dummyPos, Value: "NotReached"}, nil),
			}, []Step{}),
			createTestStep("set", "status", &StringLiteralNode{Pos: dummyPos, Value: "Finished"}, nil),
		}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"status": "Started", "x": "Inside"}, expectedResult: "ReturnedFromIf", expectError: false},

		// --- Existing RETURN Tests ---
		{name: "RETURN single value", inputSteps: []Step{createTestStep("return", "", &NumberLiteralNode{Pos: dummyPos, Value: int64(42)}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: int64(42), expectError: false},
		{name: "RETURN multiple values", inputSteps: []Step{createTestStep("return", "", []Expression{&StringLiteralNode{Pos: dummyPos, Value: "hello"}, &NumberLiteralNode{Pos: dummyPos, Value: int64(10)}, &BooleanLiteralNode{Pos: dummyPos, Value: true}}, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: []interface{}{"hello", int64(10), true}, expectError: false},
		{name: "RETURN no value", inputSteps: []Step{createTestStep("return", "", nil, nil)}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{}, expectedResult: nil, expectError: false},
		{name: "RETURN value from variable", inputSteps: []Step{
			createTestStep("set", "myVar", &StringLiteralNode{Pos: dummyPos, Value: "data"}, nil),
			createTestStep("return", "", &VariableNode{Pos: dummyPos, Name: "myVar"}, nil),
		}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myVar": "data"}, expectedResult: "data", expectError: false},
		{name: "RETURN multiple values including variable", inputSteps: []Step{
			createTestStep("set", "myVar", &BooleanLiteralNode{Pos: dummyPos, Value: false}, nil),
			createTestStep("return", "", []Expression{&NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, &VariableNode{Pos: dummyPos, Name: "myVar"}, &NumberLiteralNode{Pos: dummyPos, Value: 3.14}}, nil),
		}, initialVars: map[string]interface{}{}, expectedVars: map[string]interface{}{"myVar": false}, expectedResult: []interface{}{int64(1), false, 3.14}, expectError: false},

		// --- Corrected MUST Tests: expectedResult is the condition's value ---
		{name: "MUST true literal", inputSteps: []Step{createTestStep("must", "", &BooleanLiteralNode{Pos: dummyPos, Value: true}, nil)}, initialVars: map[string]interface{}{}, expectedResult: true, expectError: false},
		{name: "MUST false literal", inputSteps: []Step{createTestStep("must", "", &BooleanLiteralNode{Pos: dummyPos, Value: false}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST non-zero number", inputSteps: []Step{createTestStep("must", "", &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, nil)}, initialVars: map[string]interface{}{}, expectedResult: int64(1), expectError: false},
		{name: "MUST zero number", inputSteps: []Step{createTestStep("must", "", &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST non-empty string ('true')", inputSteps: []Step{createTestStep("must", "", &StringLiteralNode{Pos: dummyPos, Value: "true"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: "true", expectError: false},
		{name: "MUST non-empty string ('1')", inputSteps: []Step{createTestStep("must", "", &StringLiteralNode{Pos: dummyPos, Value: "1"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: "1", expectError: false},
		{name: "MUST empty string", inputSteps: []Step{createTestStep("must", "", &StringLiteralNode{Pos: dummyPos, Value: ""}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST non-empty string ('other')", inputSteps: []Step{createTestStep("must", "", &StringLiteralNode{Pos: dummyPos, Value: "other"}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST nil", inputSteps: []Step{createTestStep("must", "", &VariableNode{Pos: dummyPos, Name: "nilVar"}, nil)}, initialVars: map[string]interface{}{"nilVar": nil}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST true variable", inputSteps: []Step{createTestStep("must", "", &VariableNode{Pos: dummyPos, Name: "t"}, nil)}, initialVars: map[string]interface{}{"t": true}, expectedResult: true, expectError: false},
		{name: "MUST last result (true)", inputSteps: []Step{createTestStep("set", "_ignored", &BooleanLiteralNode{Pos: dummyPos, Value: true}, nil), createTestStep("must", "", &LastNode{Pos: dummyPos}, nil)}, initialVars: map[string]interface{}{}, lastResult: true, expectedResult: true, expectError: false},
		{name: "MUST last result (false)", inputSteps: []Step{createTestStep("set", "_ignored", &BooleanLiteralNode{Pos: dummyPos, Value: false}, nil), createTestStep("must", "", &LastNode{Pos: dummyPos}, nil)}, initialVars: map[string]interface{}{}, lastResult: false, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST expression (1 > 0)", inputSteps: []Step{createTestStep("must", "", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, Operator: ">", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: true, expectError: false},
		{name: "MUST expression (1 < 0)", inputSteps: []Step{createTestStep("must", "", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, Operator: "<", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUST evaluation error", inputSteps: []Step{createTestStep("must", "", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, Operator: "+", Right: &StringLiteralNode{Pos: dummyPos, Value: "a"}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},

		// --- Corrected MUSTBE Tests: expectedResult is the check function's result (true for pass) ---
		{name: "MUSTBE is_string pass", inputSteps: []Step{createTestStep("mustbe", "is_string", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "is_string", IsTool: false}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "s"}}}, nil)}, initialVars: mustBeVars, expectedResult: true, expectError: false},
		{name: "MUSTBE is_string fail", inputSteps: []Step{createTestStep("mustbe", "is_string", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "is_string", IsTool: false}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "n"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE not_empty fail (nil)", inputSteps: []Step{createTestStep("mustbe", "not_empty", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "not_empty", IsTool: false}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "nilV"}}}, nil)}, initialVars: mustBeVars, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE unknown function", inputSteps: []Step{createTestStep("mustbe", "is_banana", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "is_banana", IsTool: false}, Arguments: []Expression{&BooleanLiteralNode{Pos: dummyPos, Value: true}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE wrong arg count", inputSteps: []Step{createTestStep("mustbe", "is_string", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "is_string", IsTool: false}, Arguments: []Expression{&StringLiteralNode{Pos: dummyPos, Value: "a"}, &StringLiteralNode{Pos: dummyPos, Value: "b"}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},
		{name: "MUSTBE argument evaluation error", inputSteps: []Step{createTestStep("mustbe", "is_string", &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "is_string", IsTool: false}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "missing"}}}, nil)}, initialVars: map[string]interface{}{}, expectedResult: nil, expectError: true, ExpectedErrorIs: ErrMustConditionFailed},

		// --- Dotted Tool Name Test Cases ---
		{
			name: "Tool Call List.Length",
			inputSteps: []Step{
				createTestStep("set", "listLen", &CallableExprNode{
					Pos:       dummyPos,
					Target:    CallTarget{Pos: dummyPos, IsTool: true, Name: "List.Length"},
					Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "myInitialList"}},
				}, nil),
			},
			initialVars:    map[string]interface{}{"myInitialList": initialList},
			expectedVars:   map[string]interface{}{"myInitialList": initialList, "listLen": int64(3)},
			expectedResult: int64(3),
			expectError:    false,
		},
		{
			name: "Tool Call List.Append",
			inputSteps: []Step{
				createTestStep("set", "appendedList", &CallableExprNode{
					Pos:    dummyPos,
					Target: CallTarget{Pos: dummyPos, IsTool: true, Name: "List.Append"},
					Arguments: []Expression{
						&VariableNode{Pos: dummyPos, Name: "myInitialList"},
						&StringLiteralNode{Pos: dummyPos, Value: "newItem"},
					},
				}, nil),
			},
			initialVars:    map[string]interface{}{"myInitialList": initialList},
			expectedVars:   map[string]interface{}{"myInitialList": initialList, "appendedList": []interface{}{"item1", int64(2), true, "newItem"}},
			expectedResult: []interface{}{"item1", int64(2), true, "newItem"},
			expectError:    false,
		},
		{
			name: "Tool Call List.Get valid index",
			inputSteps: []Step{
				createTestStep("set", "gotItem", &CallableExprNode{
					Pos:    dummyPos,
					Target: CallTarget{Pos: dummyPos, IsTool: true, Name: "List.Get"},
					Arguments: []Expression{
						&VariableNode{Pos: dummyPos, Name: "myInitialList"},
						&NumberLiteralNode{Pos: dummyPos, Value: int64(1)},
					},
				}, nil),
			},
			initialVars:    map[string]interface{}{"myInitialList": initialList},
			expectedVars:   map[string]interface{}{"myInitialList": initialList, "gotItem": int64(2)},
			expectedResult: int64(2),
			expectError:    false,
		},
		{
			name: "Tool Call List.Get out-of-bounds with default",
			inputSteps: []Step{
				createTestStep("set", "gotItemOrDefault", &CallableExprNode{
					Pos:    dummyPos,
					Target: CallTarget{Pos: dummyPos, IsTool: true, Name: "List.Get"},
					Arguments: []Expression{
						&VariableNode{Pos: dummyPos, Name: "myInitialList"},
						&NumberLiteralNode{Pos: dummyPos, Value: int64(10)},
						&StringLiteralNode{Pos: dummyPos, Value: "default"},
					},
				}, nil),
			},
			initialVars:    map[string]interface{}{"myInitialList": initialList},
			expectedVars:   map[string]interface{}{"myInitialList": initialList, "gotItemOrDefault": "default"},
			expectedResult: "default",
			expectError:    false,
		},
		{
			name: "Tool Call List.IsEmpty (false)",
			inputSteps: []Step{
				createTestStep("set", "isEmptyResult", &CallableExprNode{
					Pos:       dummyPos,
					Target:    CallTarget{Pos: dummyPos, IsTool: true, Name: "List.IsEmpty"},
					Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "myInitialList"}},
				}, nil),
			},
			initialVars:    map[string]interface{}{"myInitialList": initialList},
			expectedVars:   map[string]interface{}{"myInitialList": initialList, "isEmptyResult": false},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "Tool Call List.IsEmpty (true)",
			inputSteps: []Step{
				createTestStep("set", "isEmptyResult", &CallableExprNode{
					Pos:       dummyPos,
					Target:    CallTarget{Pos: dummyPos, IsTool: true, Name: "List.IsEmpty"},
					Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "emptyLVar"}},
				}, nil),
			},
			initialVars:    map[string]interface{}{"emptyLVar": []interface{}{}},
			expectedVars:   map[string]interface{}{"emptyLVar": []interface{}{}, "isEmptyResult": true},
			expectedResult: true,
			expectError:    false,
		},
		{
			name: "Tool Call Error - List.Length on non-list",
			inputSteps: []Step{
				createTestStep("set", "_ignored", &CallableExprNode{
					Pos:       dummyPos,
					Target:    CallTarget{Pos: dummyPos, IsTool: true, Name: "List.Length"},
					Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "notAList"}},
				}, nil),
			},
			initialVars:     map[string]interface{}{"notAList": "this is a string"},
			expectedVars:    map[string]interface{}{"notAList": "this is a string"},
			expectedResult:  nil,
			expectError:     true,
			ExpectedErrorIs: ErrValidationTypeMismatch, // Corrected: Expecting specific type mismatch from validation
			errContains:     "expected a slice (list), got string",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runExecuteStepsTest(t, tc)
		})
	}
}

// nlines: 202
// risk_rating: MEDIUM
