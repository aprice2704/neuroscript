// NeuroScript Version: 0.3.5
// File version: 1.1.0
// Purpose: Added test cases for 'while' and 'for each' loops to improve block statement coverage.
// filename: pkg/runtime/interpreter_test.go
package runtime

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestExecuteStepsBlocksAndLoops tests the execution of various control flow blocks and tool calls.
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
		"l":      lang.NewListValue([]lang.Value{lang.lang.NumberValue{Value: 1}, lang.lang.NumberValue{Value: 2}}),
		"m":      lang.NewMapValue(map[string]lang.Value{"a": lang.NumberValue{Value: 1}}),
		"emptyS": lang.StringValue{Value: ""},
		"emptyL": lang.NewListValue(nil),
		"emptyM": lang.NewMapValue(nil),
		"zeroN":  lang.NumberValue{Value: 0},
		"nilV":   lang.NilValue{},
	}

	testCases := []testutil.executeStepsTestCase{
		{
			name: "IF true literal",
			inputSteps: []ast.Step{
				testutil.createIfStep(testutil.testutil.dummyPos, &ast.BooleanLiteralNode{Position: testutil.testutil.dummyPos, Value: true},
					[]ast.Step{testutil.createTestStep("set", "x", &ast.StringLiteralNode{Position: testutil.dummyPos, Value: "Inside"}, nil)},
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
				testutil.createTestStep("set", "status", &ast.StringLiteralNode{Position: testutil.dummyPos, Value: "Started"}, nil),
				testutil.createIfStep(testutil.testutil.dummyPos, &ast.BooleanLiteralNode{Position: testutil.testutil.dummyPos, Value: true},
					[]ast.Step{
						testutil.createTestStep("set", "x", &ast.StringLiteralNode{Position: testutil.dummyPos, Value: "Inside"}, nil),
						testutil.createTestStep("return", "", &ast.StringLiteralNode{Position: testutil.dummyPos, Value: "ReturnedFromIf"}, nil),
					},
					nil),
				testutil.createTestStep("set", "status", &ast.StringLiteralNode{Position: testutil.dummyPos, Value: "Finished"}, nil),
			},
			initialVars:    map[string]lang.Value{},
			expectedVars:   map[string]lang.Value{"status": lang.lang.StringValue{Value: "Started"}, "x": lang.lang.StringValue{Value: "Inside"}},
			expectedResult: lang.StringValue{Value: "ReturnedFromIf"},
			expectError:    false,
		},
		// ADDED: Test case for a basic 'while' loop.
		{
			name: "WHILE loop basic",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "i", &ast.NumberLiteralNode{Value: 0}, nil),
				testutil.createWhileStep(testutil.dummyPos,
					&ast.BinaryOpNode{ // Condition: i < 3
						Left:     &ast.VariableNode{Name: "i"},
						Operator: "<",
						Right:    &ast.NumberLiteralNode{Value: 3},
					},
					[]ast.Step{ // Body: set i = i + 1
						testutil.createTestStep("set", "i", &ast.BinaryOpNode{
							Left:     &ast.VariableNode{Name: "i"},
							Operator: "+",
							Right:    &ast.NumberLiteralNode{Value: 1},
						}, nil),
					},
				),
			},
			expectedVars:   map[string]lang.Value{"i": lang.NumberValue{Value: 3}},
			expectedResult: lang.NumberValue{Value: 3}, // The result of the last set statement in the loop
			expectError:    false,
		},
		// ADDED: Test case for a 'for each' loop.
		{
			name: "FOR EACH loop",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "l", &ast.ListLiteralNode{
					Elements: []ast.Expression{
						&ast.NumberLiteralNode{Value: 10},
						&ast.NumberLiteralNode{Value: 20},
						&ast.NumberLiteralNode{Value: 30},
					},
				}, nil),
				testutil.createTestStep("set", "sum", &ast.NumberLiteralNode{Value: 0}, nil),
				testutil.createForStep(testutil.dummyPos, "item", &ast.VariableNode{Name: "l"},
					[]ast.Step{ // Body: set sum = sum + item
						testutil.createTestStep("set", "sum", &ast.BinaryOpNode{
							Left:     &ast.VariableNode{Name: "sum"},
							Operator: "+",
							Right:    &ast.VariableNode{Name: "item"},
						}, nil),
					},
				),
			},
			expectedVars: map[string]lang.Value{
				"l":    lang.NewListValue([]lang.Value{lang.lang.lang.NumberValue{Value: 10}, lang.lang.lang.NumberValue{Value: 20}, lang.lang.lang.NumberValue{Value: 30}}),
				"sum":  lang.NumberValue{Value: 60},
				"item": lang.NumberValue{Value: 30}, // Loop variable holds the last value after the loop
			},
			expectedResult: lang.NumberValue{Value: 60}, // Result of the last set statement
			expectError:    false,
		},
		{
			name: "RETURN multiple values",
			inputSteps: []ast.Step{
				{
					Type:     "return",
					Position: testutil.dummyPos,
					Values: []ast.Expression{
						&ast.StringLiteralNode{Value: "hello"},
						&ast.NumberLiteralNode{Value: int64(10)},
						&ast.BooleanLiteralNode{Value: true},
					},
				},
			},
			expectedResult: lang.NewListValue([]lang.Value{
				lang.StringValue{Value: "hello"},
				lang.NumberValue{Value: 10},
				lang.BoolValue{Value: true},
			}),
			expectError: false,
		},
		{
			name: "RETURN value from variable",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "myVar", &ast.StringLiteralNode{Value: "data"}, nil),
				testutil.createTestStep("return", "", &ast.VariableNode{Name: "myVar"}, nil),
			},
			expectedVars:   map[string]lang.Value{"myVar": lang.StringValue{Value: "data"}},
			expectedResult: lang.StringValue{Value: "data"},
			expectError:    false,
		},
		{
			name: "RETURN multiple values including variable",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "myVar", &ast.BooleanLiteralNode{Value: false}, nil),
				{
					Type:     "return",
					Position: testutil.dummyPos,
					Values: []ast.Expression{
						&ast.NumberLiteralNode{Value: int64(1)},
						&ast.VariableNode{Name: "myVar"},
						&ast.NumberLiteralNode{Value: 3.14},
					},
				},
			},
			expectedVars: map[string]lang.Value{"myVar": lang.BoolValue{Value: false}},
			expectedResult: lang.NewListValue([]lang.Value{
				lang.NumberValue{Value: 1},
				lang.BoolValue{Value: false},
				lang.NumberValue{Value: 3.14},
			}),
			expectError: false,
		},
		{
			name:           "MUST true literal",
			inputSteps:     []ast.Step{testutil.createTestStep("must", "", &ast.BooleanLiteralNode{Value: true}, nil)},
			expectedResult: lang.BoolValue{Value: true},
			expectError:    false,
		},
		{
			name:           "MUST non-empty string ('true')",
			inputSteps:     []ast.Step{testutil.createTestStep("must", "", &ast.StringLiteralNode{Value: "true"}, nil)},
			expectedResult: lang.StringValue{Value: "true"},
			expectError:    false,
		},
		{
			name:           "MUST non-empty string ('1')",
			inputSteps:     []ast.Step{testutil.createTestStep("must", "", &ast.StringLiteralNode{Value: "1"}, nil)},
			expectedResult: lang.StringValue{Value: "1"},
			expectError:    false,
		},
		{
			name:            "MUST non-empty string ('other')",
			inputSteps:      []ast.Step{testutil.createTestStep("must", "", &ast.StringLiteralNode{Value: "other"}, nil)},
			initialVars:     mustBeVars,
			expectError:     true,
			ExpectedErrorIs: lang.ErrMustConditionFailed,
		},
		{
			name:            "MUST last result (false)",
			inputSteps:      []ast.Step{testutil.createTestStep("must", "", &ast.EvalNode{}, nil)},
			lastResult:      lang.BoolValue{Value: false},
			expectError:     true,
			ExpectedErrorIs: lang.ErrMustConditionFailed,
		},
		{
			name: "MUST evaluation error",
			inputSteps: []ast.Step{testutil.createTestStep("must", "", &ast.BinaryOpNode{
				Left:     &ast.NumberLiteralNode{Value: 1},
				Operator: "-",
				Right:    &ast.StringLiteralNode{Value: "a"},
			}, nil)},
			expectError:     true,
			ExpectedErrorIs: lang.ErrInvalidOperandTypeNumeric,
		},
		{
			name: "Tool Call List.Append",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "appendedList", &ast.CallableExprNode{
					Position:  testutil.dummyPos,
					Target:    ast.CallTarget{Position: testutil.dummyPos, IsTool: true, Name: "List.Append"},
					Arguments: []ast.Expression{&ast.VariableNode{Position: testutil.dummyPos, Name: "lvar"}, &ast.StringLiteralNode{Value: "newItem"}},
				}, nil),
			},
			initialVars: map[string]lang.Value{"lvar": initialList},
			expectedResult: lang.NewListValue([]lang.Value{
				lang.StringValue{Value: "item1"},
				lang.NumberValue{Value: 2},
				lang.BoolValue{Value: true},
				lang.StringValue{Value: "newItem"},
			}),
			expectError: false,
		},
		{
			name: "Tool Call List.Get out-of-bounds with default",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "gotItemOrDefault", &ast.CallableExprNode{
					Position: testutil.dummyPos,
					Target:   ast.CallTarget{Position: testutil.dummyPos, IsTool: true, Name: "List.Get"},
					Arguments: []ast.Expression{
						&ast.VariableNode{Position: testutil.dummyPos, Name: "lvar"},
						&ast.NumberLiteralNode{Value: int64(99)},
						&ast.StringLiteralNode{Value: "default"},
					},
				}, nil),
			},
			initialVars:    map[string]lang.Value{"lvar": initialList},
			expectedResult: lang.StringValue{Value: "default"},
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testutil.runExecuteStepsTest(t, tc)
		})
	}
}
