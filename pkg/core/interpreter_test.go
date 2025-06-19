// NeuroScript Version: 0.3.5
// File version: 1.1.0
// Purpose: Added test cases for 'while' and 'for each' loops to improve block statement coverage.
// filename: pkg/core/interpreter_test.go
package core

import (
	"testing"
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
				createIfStep(dummyPos, &BooleanLiteralNode{Pos: dummyPos, Value: true},
					[]Step{createTestStep("set", "x", &StringLiteralNode{Pos: dummyPos, Value: "Inside"}, nil)},
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
				createTestStep("set", "status", &StringLiteralNode{Pos: dummyPos, Value: "Started"}, nil),
				createIfStep(dummyPos, &BooleanLiteralNode{Pos: dummyPos, Value: true},
					[]Step{
						createTestStep("set", "x", &StringLiteralNode{Pos: dummyPos, Value: "Inside"}, nil),
						createTestStep("return", "", &StringLiteralNode{Pos: dummyPos, Value: "ReturnedFromIf"}, nil),
					},
					nil),
				createTestStep("set", "status", &StringLiteralNode{Pos: dummyPos, Value: "Finished"}, nil),
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
				createTestStep("set", "i", &NumberLiteralNode{Value: 0}, nil),
				createWhileStep(dummyPos,
					&BinaryOpNode{ // Condition: i < 3
						Left:     &VariableNode{Name: "i"},
						Operator: "<",
						Right:    &NumberLiteralNode{Value: 3},
					},
					[]Step{ // Body: set i = i + 1
						createTestStep("set", "i", &BinaryOpNode{
							Left:     &VariableNode{Name: "i"},
							Operator: "+",
							Right:    &NumberLiteralNode{Value: 1},
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
				createTestStep("set", "l", &ListLiteralNode{
					Elements: []Expression{
						&NumberLiteralNode{Value: 10},
						&NumberLiteralNode{Value: 20},
						&NumberLiteralNode{Value: 30},
					},
				}, nil),
				createTestStep("set", "sum", &NumberLiteralNode{Value: 0}, nil),
				createForStep(dummyPos, "item", &VariableNode{Name: "l"},
					[]Step{ // Body: set sum = sum + item
						createTestStep("set", "sum", &BinaryOpNode{
							Left:     &VariableNode{Name: "sum"},
							Operator: "+",
							Right:    &VariableNode{Name: "item"},
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
					Type: "return",
					Pos:  dummyPos,
					Values: []Expression{
						&StringLiteralNode{Value: "hello"},
						&NumberLiteralNode{Value: int64(10)},
						&BooleanLiteralNode{Value: true},
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
				createTestStep("set", "myVar", &StringLiteralNode{Value: "data"}, nil),
				createTestStep("return", "", &VariableNode{Name: "myVar"}, nil),
			},
			expectedVars:   map[string]Value{"myVar": StringValue{Value: "data"}},
			expectedResult: StringValue{Value: "data"},
			expectError:    false,
		},
		{
			name: "RETURN multiple values including variable",
			inputSteps: []Step{
				createTestStep("set", "myVar", &BooleanLiteralNode{Value: false}, nil),
				{
					Type: "return",
					Pos:  dummyPos,
					Values: []Expression{
						&NumberLiteralNode{Value: int64(1)},
						&VariableNode{Name: "myVar"},
						&NumberLiteralNode{Value: 3.14},
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
			inputSteps:     []Step{createTestStep("must", "", &BooleanLiteralNode{Value: true}, nil)},
			expectedResult: BoolValue{Value: true},
			expectError:    false,
		},
		{
			name:           "MUST non-empty string ('true')",
			inputSteps:     []Step{createTestStep("must", "", &StringLiteralNode{Value: "true"}, nil)},
			expectedResult: StringValue{Value: "true"},
			expectError:    false,
		},
		{
			name:           "MUST non-empty string ('1')",
			inputSteps:     []Step{createTestStep("must", "", &StringLiteralNode{Value: "1"}, nil)},
			expectedResult: StringValue{Value: "1"},
			expectError:    false,
		},
		{
			name:            "MUST non-empty string ('other')",
			inputSteps:      []Step{createTestStep("must", "", &StringLiteralNode{Value: "other"}, nil)},
			initialVars:     mustBeVars,
			expectError:     true,
			ExpectedErrorIs: ErrMustConditionFailed,
		},
		{
			name:            "MUST last result (false)",
			inputSteps:      []Step{createTestStep("must", "", &LastNode{}, nil)},
			lastResult:      BoolValue{Value: false},
			expectError:     true,
			ExpectedErrorIs: ErrMustConditionFailed,
		},
		{
			name: "MUST evaluation error",
			inputSteps: []Step{createTestStep("must", "", &BinaryOpNode{
				Left:     &NumberLiteralNode{Value: 1},
				Operator: "-",
				Right:    &StringLiteralNode{Value: "a"},
			}, nil)},
			expectError:     true,
			ExpectedErrorIs: ErrInvalidOperandTypeNumeric,
		},
		{
			name: "Tool Call List.Append",
			inputSteps: []Step{
				createTestStep("set", "appendedList", &CallableExprNode{
					Pos:       dummyPos,
					Target:    CallTarget{Pos: dummyPos, IsTool: true, Name: "List.Append"},
					Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "lvar"}, &StringLiteralNode{Value: "newItem"}},
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
				createTestStep("set", "gotItemOrDefault", &CallableExprNode{
					Pos:    dummyPos,
					Target: CallTarget{Pos: dummyPos, IsTool: true, Name: "List.Get"},
					Arguments: []Expression{
						&VariableNode{Pos: dummyPos, Name: "lvar"},
						&NumberLiteralNode{Value: int64(99)},
						&StringLiteralNode{Value: "default"},
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
