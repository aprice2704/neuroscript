// NeuroScript Version: 0.3.1
// File version: 0.1.3 // Corrected expectedVars for list/map number types and LastNode usage.
// Purpose: Tests for the 'typeof' operator evaluation.
// filename: pkg/core/eval_typeof_test.go
// nlines: 201 // Approximate, please adjust after pasting
// risk_rating: MEDIUM

package core

import (
	"testing"
	// core/testing_helpers.go is used for executeStepsTestCase, runExecuteStepsTest, createTestStep, NewTestXXXLiteral/Node
	// core/type_names.go is used for TypeString, TypeNumber etc.
	// ast.go definitions (Position, TypeOfNode, various LiteralNodes, VariableNode, LastNode, Expression interface) are assumed
)

func TestTypeOfOperator(t *testing.T) {
	pos := &Position{Line: 1, Column: 1, File: "typeof_test.go"} // Common position for test AST nodes

	// Assumed AST Node structures (must implement Expression interface from ast.go):
	// type TypeOfNode struct { Pos *Position; Argument Expression }
	// func (n *TypeOfNode) expressionNode() {}
	// func (n *TypeOfNode) GetPos() *Position { return n.Pos }

	// type NilLiteralNode struct { Pos *Position }
	// func (n *NilLiteralNode) expressionNode() {}
	// func (n *NilLiteralNode) GetPos() *Position { return n.Pos }

	// type ListLiteralNode struct { Pos *Position; Elements []Expression }
	// func (n *ListLiteralNode) expressionNode() {}
	// func (n *ListLiteralNode) GetPos() *Position { return n.Pos }

	// type MapLiteralNode struct { Pos *Position; Entries []*MapEntryNode }
	// func (n *MapLiteralNode) expressionNode() {}
	// func (n *MapLiteralNode) GetPos() *Position { return n.Pos }

	// type MapEntryNode struct { Pos *Position; Key *StringLiteralNode; Value Expression }
	// func (n *MapEntryNode) expressionNode() {}
	// func (n *MapEntryNode) GetPos() *Position { return n.Pos }

	// type LastNode struct { Pos *Position }
	// func (n *LastNode) expressionNode() {}
	// func (n *LastNode) GetPos() *Position { return n.Pos }
	// StringLiteralNode, NumberLiteralNode, BooleanLiteralNode, VariableNode are provided by NewTest... helpers in testing_helpers.go

	tests := []executeStepsTestCase{
		{
			name: "typeof string literal",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: NewTestStringLiteral("hello")}, nil),
			},
			expectedResult: string(TypeString),
		},
		{
			name: "typeof number literal (int)",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: NewTestNumberLiteral(123)}, nil),
			},
			expectedResult: string(TypeNumber),
		},
		{
			name: "typeof number literal (float)",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: NewTestNumberLiteral(123.45)}, nil),
			},
			expectedResult: string(TypeNumber),
		},
		{
			name: "typeof boolean literal (true)",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: NewTestBooleanLiteral(true)}, nil),
			},
			expectedResult: string(TypeBoolean),
		},
		{
			name: "typeof boolean literal (false)",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: NewTestBooleanLiteral(false)}, nil),
			},
			expectedResult: string(TypeBoolean),
		},
		{
			name: "typeof nil literal",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: &NilLiteralNode{Pos: pos}}, nil),
			},
			expectedResult: string(TypeNil),
		},
		{
			name: "typeof list literal",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: &ListLiteralNode{Pos: pos, Elements: []Expression{NewTestNumberLiteral(1), NewTestStringLiteral("a")}}}, nil),
			},
			expectedResult: string(TypeList),
		},
		{
			name: "typeof map literal",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: &MapLiteralNode{Pos: pos, Entries: []*MapEntryNode{
					{Pos: pos, Key: NewTestStringLiteral("key"), Value: NewTestStringLiteral("value")},
					{Pos: pos, Key: NewTestStringLiteral("num"), Value: NewTestNumberLiteral(1)},
				}}}, nil),
			},
			expectedResult: string(TypeMap),
		},
		{
			name: "typeof arithmetic expression",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: &BinaryOpNode{
					Pos:      pos,
					Left:     NewTestNumberLiteral(1),
					Operator: "+",
					Right:    NewTestNumberLiteral(2),
				}}, nil),
			},
			expectedResult: string(TypeNumber),
		},
		{
			name: "typeof string concatenation",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: &BinaryOpNode{
					Pos:      pos,
					Left:     NewTestStringLiteral("hello"),
					Operator: "+",
					Right:    NewTestStringLiteral("world"),
				}}, nil),
			},
			expectedResult: string(TypeString),
		},
		{
			name: "typeof logical expression",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: &BinaryOpNode{
					Pos:      pos,
					Left:     NewTestBooleanLiteral(true),
					Operator: "and",
					Right:    NewTestBooleanLiteral(false),
				}}, nil),
			},
			expectedResult: string(TypeBoolean),
		},
		{
			name: "typeof variable (string)",
			inputSteps: []Step{
				createTestStep("set", "myVar", NewTestStringLiteral("test"), nil),
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: NewTestVariableNode("myVar")}, nil),
			},
			expectedResult: string(TypeString),
			expectedVars:   map[string]interface{}{"myVar": "test"},
		},
		{
			name: "typeof variable (number)",
			inputSteps: []Step{
				createTestStep("set", "myNum", NewTestNumberLiteral(42), nil),
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: NewTestVariableNode("myNum")}, nil),
			},
			expectedResult: string(TypeNumber),
			expectedVars:   map[string]interface{}{"myNum": 42}, // Expect int for comparison
		},
		{
			name: "typeof variable (list)",
			inputSteps: []Step{
				createTestStep("set", "myList", &ListLiteralNode{Pos: pos, Elements: []Expression{NewTestNumberLiteral(1)}}, nil),
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: NewTestVariableNode("myList")}, nil),
			},
			expectedResult: string(TypeList),
			expectedVars:   map[string]interface{}{"myList": []interface{}{1}}, // Expect int(1) in the list
		},
		{
			name: "typeof variable (map)",
			inputSteps: []Step{
				createTestStep("set", "myMap", &MapLiteralNode{Pos: pos, Entries: []*MapEntryNode{
					{Pos: pos, Key: NewTestStringLiteral("a"), Value: NewTestNumberLiteral(10)},
				}}, nil),
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: NewTestVariableNode("myMap")}, nil),
			},
			expectedResult: string(TypeMap),
			expectedVars:   map[string]interface{}{"myMap": map[string]interface{}{"a": 10}}, // Expect int(10) in the map
		},
		{
			name: "typeof variable (nil)",
			inputSteps: []Step{
				createTestStep("set", "myNil", &NilLiteralNode{Pos: pos}, nil),
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: NewTestVariableNode("myNil")}, nil),
			},
			expectedResult: string(TypeNil),
			expectedVars:   map[string]interface{}{"myNil": nil},
		},
		{
			name: "typeof last expression (number)",
			inputSteps: []Step{
				createTestStep("emit", "", NewTestNumberLiteral(100), nil),                            // This sets 'last' to int 100
				createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: &LastNode{Pos: pos}}, nil), // Use LastNode here
			},
			expectedResult: string(TypeNumber),
		},
		// --- Tests requiring specific interpreter setup (functions/tools) ---
		// These remain placeholders as their setup is environment-specific.
		/*
			{
				name: "typeof user function",
				// Setup: Requires 'myFunc' to be defined in the interpreter.
				inputSteps: []Step{
					// Hypothetical: createTestStep for function definition (complex)
					// Then:
					// createTestStep("emit", "", &TypeOfNode{Pos: pos, Argument: NewTestVariableNode("myFunc")}, nil),
				},
				expectedResult: string(TypeFunction),
			},
		*/
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// The runExecuteStepsTest function from testing_helpers.go is used.
			// It internally creates a NewTestInterpreter.
			runExecuteStepsTest(t, tc)
		})
	}
}
