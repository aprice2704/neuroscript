// NeuroScript Version: 0.3.5
// File version: 9.6.0
// Purpose: Removed the local test interpreter helper, which has been moved to testing_bits.go to be exported for cross-package use.
// filename: pkg/interpreter/interpreter_suite_test.go
package interpreter

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// --- Local Test Case Struct ---

type localExecuteStepsTestCase struct {
	name            string
	inputSteps      []ast.Step
	initialVars     map[string]lang.Value
	lastResult      lang.Value
	expectError     bool
	expectedErrorIs error
	expectedResult  lang.Value
	expectedVars    map[string]lang.Value
}

// --- Local AST Creation Helpers ---

var localTestPos = &types.Position{Line: 1, Column: 1, File: "test"}

func createTestStep(stepType, lvalueName string, value ast.Expression, call *ast.CallableExprNode) ast.Step {
	step := ast.Step{
		Type:     stepType,
		BaseNode: ast.BaseNode{StartPos: localTestPos},
	}
	if lvalueName != "" {
		step.LValues = []*ast.LValueNode{{
			BaseNode:   ast.BaseNode{StartPos: localTestPos},
			Identifier: lvalueName,
		}}
	}
	if value != nil {
		step.Values = []ast.Expression{value}
	}
	if call != nil {
		step.Call = call
	}
	return step
}

func createIfStep(pos *types.Position, cond ast.Expression, body []ast.Step, elseBody []ast.Step) ast.Step {
	return ast.Step{
		Type:     "if",
		BaseNode: ast.BaseNode{StartPos: pos},
		Cond:     cond,
		Body:     body,
		ElseBody: elseBody,
	}
}

func createWhileStep(pos *types.Position, cond ast.Expression, body []ast.Step) ast.Step {
	return ast.Step{
		Type:     "while",
		BaseNode: ast.BaseNode{StartPos: pos},
		Cond:     cond,
		Body:     body,
	}
}

func createForStep(pos *types.Position, loopVar string, collection ast.Expression, body []ast.Step) ast.Step {
	return ast.Step{
		Type:        "for",
		BaseNode:    ast.BaseNode{StartPos: pos},
		LoopVarName: loopVar,
		Collection:  collection,
		Body:        body,
	}
}

// --- Local Test Execution Helper ---

func runLocalExecuteStepsTest(t *testing.T, tc localExecuteStepsTestCase) {
	t.Helper()
	// Most step tests do not require privileges.
	interp, err := NewTestInterpreter(t, tc.initialVars, tc.lastResult, false)
	if err != nil {
		t.Fatalf("Test %q: Failed to create test interpreter: %v", tc.name, err)
	}
	finalResultFromExec, wasReturn, _, err := interp.executeSteps(tc.inputSteps, false, nil)
	if (err != nil) != tc.expectError {
		t.Fatalf("Test %q: Unexpected error state. Got err: %v, wantErr: %t", tc.name, err, tc.expectError)
	}
	if tc.expectError {
		if tc.expectedErrorIs != nil && !errors.Is(err, tc.expectedErrorIs) {
			t.Fatalf("Test %q: Expected error wrapping: [%v], got: [%v]", tc.name, tc.expectedErrorIs, err)
		}
		return
	}
	var actualResult lang.Value
	if wasReturn {
		actualResult = finalResultFromExec
	} else {
		actualResult = interp.lastCallResult
	}
	if !reflect.DeepEqual(actualResult, tc.expectedResult) {
		t.Errorf("Test %q: Final execution result mismatch:\n Expected: %#v (%T)\n      Got: %#v (%T)", tc.name, tc.expectedResult, tc.expectedResult, actualResult, actualResult)
	}
	if tc.expectedVars != nil {
		for expectedKey, expectedValue := range tc.expectedVars {
			gotValue, ok := interp.GetVariable(expectedKey)
			if !ok {
				t.Errorf("Test %q: Expected variable '%s' not found in final state", tc.name, expectedKey)
				continue
			}
			if !reflect.DeepEqual(gotValue, expectedValue) {
				t.Errorf("Test %q: Variable state mismatch for '%s':\n Expected: %#v (%T)\n      Got: %#v (%T)", tc.name, expectedKey, expectedValue, expectedValue, gotValue, gotValue)
			}
		}
	}
}

// --- Main Test Function ---

func TestExecuteStepsBlocksAndLoops(t *testing.T) {
	initialList := lang.NewListValue([]lang.Value{
		lang.StringValue{Value: "item1"},
		lang.NumberValue{Value: 2},
		lang.BoolValue{Value: true},
	})

	mustBeVars := map[string]lang.Value{
		"s":      lang.StringValue{Value: "a string"},
		"n":      lang.NumberValue{Value: 10},
		"f":      lang.NumberValue{Value: 3.14},
		"b":      lang.BoolValue{Value: true},
		"l":      lang.NewListValue([]lang.Value{lang.NumberValue{Value: 1}, lang.NumberValue{Value: 2}}),
		"m":      lang.NewMapValue(map[string]lang.Value{"a": lang.NumberValue{Value: 1}}),
		"emptyS": lang.StringValue{Value: ""},
		"emptyL": lang.NewListValue(nil),
		"emptyM": lang.NewMapValue(nil),
		"zeroN":  lang.NumberValue{Value: 0},
		"nilV":   &lang.NilValue{},
	}

	testCases := []localExecuteStepsTestCase{
		{
			name: "IF true literal",
			inputSteps: []ast.Step{
				createIfStep(localTestPos, &ast.BooleanLiteralNode{Value: true},
					[]ast.Step{createTestStep("set", "x", &ast.StringLiteralNode{Value: "Inside"}, nil)},
					nil,
				),
			},
			expectedVars:   map[string]lang.Value{"x": lang.StringValue{Value: "Inside"}},
			expectedResult: lang.StringValue{Value: "Inside"},
			expectError:    false,
		},
		{
			name: "IF block with RETURN",
			inputSteps: []ast.Step{
				createTestStep("set", "status", &ast.StringLiteralNode{Value: "Started"}, nil),
				createIfStep(localTestPos, &ast.BooleanLiteralNode{Value: true},
					[]ast.Step{
						createTestStep("set", "x", &ast.StringLiteralNode{Value: "Inside"}, nil),
						createTestStep("return", "", &ast.StringLiteralNode{Value: "ReturnedFromIf"}, nil),
					},
					nil),
				createTestStep("set", "status", &ast.StringLiteralNode{Value: "Finished"}, nil),
			},
			initialVars:    map[string]lang.Value{},
			expectedVars:   map[string]lang.Value{"status": lang.StringValue{Value: "Started"}, "x": lang.StringValue{Value: "Inside"}},
			expectedResult: lang.StringValue{Value: "ReturnedFromIf"},
			expectError:    false,
		},
		{
			name: "WHILE loop basic",
			inputSteps: []ast.Step{
				createTestStep("set", "i", &ast.NumberLiteralNode{Value: 0}, nil),
				createWhileStep(localTestPos,
					&ast.BinaryOpNode{Left: &ast.VariableNode{Name: "i"}, Operator: "<", Right: &ast.NumberLiteralNode{Value: 3}},
					[]ast.Step{
						createTestStep("set", "i", &ast.BinaryOpNode{
							Left:     &ast.VariableNode{Name: "i"},
							Operator: "+",
							Right:    &ast.NumberLiteralNode{Value: 1},
						}, nil),
					},
				),
			},
			expectedVars:   map[string]lang.Value{"i": lang.NumberValue{Value: 3}},
			expectedResult: lang.NumberValue{Value: 3},
			expectError:    false,
		},
		{
			name: "FOR EACH loop",
			inputSteps: []ast.Step{
				createTestStep("set", "l", &ast.ListLiteralNode{Elements: []ast.Expression{&ast.NumberLiteralNode{Value: 10}, &ast.NumberLiteralNode{Value: 20}, &ast.NumberLiteralNode{Value: 30}}}, nil),
				createTestStep("set", "sum", &ast.NumberLiteralNode{Value: 0}, nil),
				createForStep(localTestPos, "item", &ast.VariableNode{Name: "l"},
					[]ast.Step{
						createTestStep("set", "sum", &ast.BinaryOpNode{Left: &ast.VariableNode{Name: "sum"}, Operator: "+", Right: &ast.VariableNode{Name: "item"}}, nil),
					},
				),
			},
			expectedVars: map[string]lang.Value{
				"l":    lang.NewListValue([]lang.Value{lang.NumberValue{Value: 10}, lang.NumberValue{Value: 20}, lang.NumberValue{Value: 30}}),
				"sum":  lang.NumberValue{Value: 60},
				"item": lang.NumberValue{Value: 30},
			},
			expectedResult: lang.NumberValue{Value: 60},
			expectError:    false,
		},
		{
			name: "MUST evaluation error",
			inputSteps: []ast.Step{
				createTestStep("must", "", &ast.BinaryOpNode{Left: &ast.NumberLiteralNode{Value: 1}, Operator: ">", Right: &ast.NumberLiteralNode{Value: 5}}, nil),
			},
			expectError:     true,
			expectedErrorIs: lang.ErrMustConditionFailed,
		},
		{
			name:            "MUST non-empty string ('other')",
			inputSteps:      []ast.Step{createTestStep("must", "", &ast.StringLiteralNode{Value: "other"}, nil)},
			initialVars:     mustBeVars,
			expectError:     true,
			expectedErrorIs: lang.ErrMustConditionFailed,
		},
		{
			name: "Tool Call List.Append",
			inputSteps: []ast.Step{
				createTestStep("set", "lvar", nil, &ast.CallableExprNode{
					// FIX: Use types.MakeFullName to construct the robust, fully-qualified tool name.
					Target:    ast.CallTarget{IsTool: true, Name: string(types.MakeFullName("list", "Append"))},
					Arguments: []ast.Expression{&ast.VariableNode{Name: "initialListVar"}, &ast.StringLiteralNode{Value: "newItem"}},
				}),
			},
			initialVars:    map[string]lang.Value{"initialListVar": initialList},
			expectedResult: lang.NewListValue([]lang.Value{lang.StringValue{Value: "item1"}, lang.NumberValue{Value: 2}, lang.BoolValue{Value: true}, lang.StringValue{Value: "newItem"}}),
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runLocalExecuteStepsTest(t, tc)
		})
	}
}
