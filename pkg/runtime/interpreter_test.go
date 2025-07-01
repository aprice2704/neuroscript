// NeuroScript Version: 0.3.5
// File version: 1.1.0
// Purpose: Added test cases for 'while' and 'for each' loops to improve block statement coverage.
// filename: pkg/core/interpreter_test.go
package runtime

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// TestExecuteStepsBlocksAndLoops tests the execution of various control flow blocks and tool calls.
func TestExecuteStepsBlocksAndLoops(t *testing.T) {
	initialList := NewListValue([]Value{
		StringValue{Value: "item1"},
		NumberValue{Value: 2},
		BoolValue{Value: true},
	})

	mustBeVars := map[string]Value{
		"s":      StringValue{Value: "a string"},
		"n":      NumberValue{Value: 10},
		"f":      NumberValue{Value: 3.14},
		"b":      BoolValue{Value: true},
		"l":      NewListValue([]Value{NumberValue{Value: 1}, NumberValue{Value: 2}}),
		"m":      NewMapValue(map[string]Value{"a": NumberValue{Value: 1}}),
		"emptyS": StringValue{Value: ""},
		"emptyL": NewListValue(nil),
		"emptyM": NewMapValue(nil),
		"zeroN":  NumberValue{Value: 0},
		"nilV":   NilValue{},
	}

	testCases := []executeStepsTestCase{
		{
			name: "IF true literal",
			inputSteps: []Step{
				createIfStep(dummyPos, &ast.BooleanLiteralNode{Position: dummyPos, Value: true},
					[]Step{createTestStep("set", "x", &ast.StringLiteralNode{Position: dummyPos, Value: "Inside"}, nil)},
					nil,
				),
			},
			expectedVars:   map[string]Value{"x": StringValue{Value: "Inside"}},
			expectedResult: StringValue{Value: "Inside"},
			expectError:    false,
		},
		{
			name: "IF block with RETURN",
			inputSteps: []Step{
				createTestStep("set", "status", &ast.StringLiteralNode{Position: dummyPos, Value: "Started"}, nil),
				createIfStep(dummyPos, &ast.BooleanLiteralNode{Position: dummyPos, Value: true},
					[]Step{
						createTestStep("set", "x", &ast.StringLiteralNode{Position: dummyPos, Value: "Inside"}, nil),
						createTestStep("return", "", &ast.StringLiteralNode{Position: dummyPos, Value: "ReturnedFromIf"}, nil),
					},
					nil),
				createTestStep("set", "status", &ast.StringLiteralNode{Position: dummyPos, Value: "Finished"}, nil),
			},
			initialVars:    map[string]Value{},
			expectedVars:   map[string]Value{"status": StringValue{Value: "Started"}, "x": StringValue{Value: "Inside"}},
			expectedResult: StringValue{Value: "ReturnedFromIf"},
			expectError:    false,
		},
		// ADDED: Test case for a basic 'while' loop.
		{
			name: "WHILE loop basic",
			inputSteps: []Step{
				createTestStep("set", "i", &ast.NumberLiteralNode{Value: 0}, nil),
				createWhileStep(dummyPos,
					&ast.BinaryOpNode{ // Condition: i < 3
						Left:     &ast.VariableNode{Name: "i"},
						Operator: "<",
						Right:    &ast.NumberLiteralNode{Value: 3},
					},
					[]Step{ // Body: set i = i + 1
						createTestStep("set", "i", &ast.BinaryOpNode{
							Left:     &ast.VariableNode{Name: "i"},
							Operator: "+",
							Right:    &ast.NumberLiteralNode{Value: 1},
						}, nil),
					},
				),
			},
			expectedVars:   map[string]Value{"i": NumberValue{Value: 3}},
			expectedResult: NumberValue{Value: 3}, // The result of the last set statement in the loop
			expectError:    false,
		},
		// ADDED: Test case for a 'for each' loop.
		{
			name: "FOR EACH loop",
			inputSteps: []Step{
				createTestStep("set", "l", &ast.ListLiteralNode{
					Elements: []ast.Expression{
						&ast.NumberLiteralNode{Value: 10},
						&ast.NumberLiteralNode{Value: 20},
						&ast.NumberLiteralNode{Value: 30},
					},
				}, nil),
				createTestStep("set", "sum", &ast.NumberLiteralNode{Value: 0}, nil),
				createForStep(dummyPos, "item", &ast.VariableNode{Name: "l"},
					[]Step{ // Body: set sum = sum + item
						createTestStep("set", "sum", &ast.BinaryOpNode{
							Left:     &ast.VariableNode{Name: "sum"},
							Operator: "+",
							Right:    &ast.VariableNode{Name: "item"},
						}, nil),
					},
				),
			},
			expectedVars: map[string]Value{
				"l":    NewListValue([]Value{NumberValue{Value: 10}, NumberValue{Value: 20}, NumberValue{Value: 30}}),
				"sum":  NumberValue{Value: 60},
				"item": NumberValue{Value: 30}, // Loop variable holds the last value after the loop
			},
			expectedResult: NumberValue{Value: 60}, // Result of the last set statement
			expectError:    false,
		},
		{
			name: "RETURN multiple values",
			inputSteps: []Step{
				{
					Type:     "return",
					Position: dummyPos,
					Values: []ast.Expression{
						&ast.StringLiteralNode{Value: "hello"},
						&ast.NumberLiteralNode{Value: int64(10)},
						&ast.BooleanLiteralNode{Value: true},
					},
				},
			},
			expectedResult: NewListValue([]Value{
				StringValue{Value: "hello"},
				NumberValue{Value: 10},
				BoolValue{Value: true},
			}),
			expectError: false,
		},
		{
			name: "RETURN value from variable",
			inputSteps: []Step{
				createTestStep("set", "myVar", &ast.StringLiteralNode{Value: "data"}, nil),
				createTestStep("return", "", &ast.VariableNode{Name: "myVar"}, nil),
			},
			expectedVars:   map[string]Value{"myVar": StringValue{Value: "data"}},
			expectedResult: StringValue{Value: "data"},
			expectError:    false,
		},
		{
			name: "RETURN multiple values including variable",
			inputSteps: []Step{
				createTestStep("set", "myVar", &ast.BooleanLiteralNode{Value: false}, nil),
				{
					Type:     "return",
					Position: dummyPos,
					Values: []ast.Expression{
						&ast.NumberLiteralNode{Value: int64(1)},
						&ast.VariableNode{Name: "myVar"},
						&ast.NumberLiteralNode{Value: 3.14},
					},
				},
			},
			expectedVars: map[string]Value{"myVar": BoolValue{Value: false}},
			expectedResult: NewListValue([]Value{
				NumberValue{Value: 1},
				BoolValue{Value: false},
				NumberValue{Value: 3.14},
			}),
			expectError: false,
		},
		{
			name:           "MUST true literal",
			inputSteps:     []Step{createTestStep("must", "", &ast.BooleanLiteralNode{Value: true}, nil)},
			expectedResult: BoolValue{Value: true},
			expectError:    false,
		},
		{
			name:           "MUST non-empty string ('true')",
			inputSteps:     []Step{createTestStep("must", "", &ast.StringLiteralNode{Value: "true"}, nil)},
			expectedResult: StringValue{Value: "true"},
			expectError:    false,
		},
		{
			name:           "MUST non-empty string ('1')",
			inputSteps:     []Step{createTestStep("must", "", &ast.StringLiteralNode{Value: "1"}, nil)},
			expectedResult: StringValue{Value: "1"},
			expectError:    false,
		},
		{
			name:            "MUST non-empty string ('other')",
			inputSteps:      []Step{createTestStep("must", "", &ast.StringLiteralNode{Value: "other"}, nil)},
			initialVars:     mustBeVars,
			expectError:     true,
			ExpectedErrorIs: ErrMustConditionFailed,
		},
		{
			name:            "MUST last result (false)",
			inputSteps:      []Step{createTestStep("must", "", &ast.EvalNode{}, nil)},
			lastResult:      BoolValue{Value: false},
			expectError:     true,
			ExpectedErrorIs: ErrMustConditionFailed,
		},
		{
			name: "MUST evaluation error",
			inputSteps: []Step{createTestStep("must", "", &ast.BinaryOpNode{
				Left:     &ast.NumberLiteralNode{Value: 1},
				Operator: "-",
				Right:    &ast.StringLiteralNode{Value: "a"},
			}, nil)},
			expectError:     true,
			ExpectedErrorIs: ErrInvalidOperandTypeNumeric,
		},
		{
			name: "Tool Call List.Append",
			inputSteps: []Step{
				createTestStep("set", "appendedList", &ast.CallableExprNode{
					Position:  dummyPos,
					Target:    ast.CallTarget{Position: dummyPos, IsTool: true, Name: "List.Append"},
					Arguments: []ast.Expression{&ast.VariableNode{Position: dummyPos, Name: "lvar"}, &ast.StringLiteralNode{Value: "newItem"}},
				}, nil),
			},
			initialVars: map[string]Value{"lvar": initialList},
			expectedResult: NewListValue([]Value{
				StringValue{Value: "item1"},
				NumberValue{Value: 2},
				BoolValue{Value: true},
				StringValue{Value: "newItem"},
			}),
			expectError: false,
		},
		{
			name: "Tool Call List.Get out-of-bounds with default",
			inputSteps: []Step{
				createTestStep("set", "gotItemOrDefault", &ast.CallableExprNode{
					Position: dummyPos,
					Target:   ast.CallTarget{Position: dummyPos, IsTool: true, Name: "List.Get"},
					Arguments: []ast.Expression{
						&ast.VariableNode{Position: dummyPos, Name: "lvar"},
						&ast.NumberLiteralNode{Value: int64(99)},
						&ast.StringLiteralNode{Value: "default"},
					},
				}, nil),
			},
			initialVars:    map[string]Value{"lvar": initialList},
			expectedResult: StringValue{Value: "default"},
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runExecuteStepsTest(t, tc)
		})
	}
}
