// NeuroScript Version: 0.3.1
// File version: 0.1.7 // Updated tests to expect wrapped Value types.
// Purpose: Tests for the 'typeof' operator evaluation.
// filename: pkg/core/eval_typeof_test.go

package core

import (
	"testing"
	// core/testing_helpers.go for NewTestInterpreter, createTestStep, NewTestXXXLiteral/Node, etc.
	// core/type_names.go for TypeString, TypeNumber etc.
	// ast.go definitions (Position, TypeOfNode, various LiteralNodes, VariableNode, LastNode, Expression interface, Procedure, ToolImplementation, ArgSpec) are assumed
)

// common position for test AST nodes
var testPos = &Position{Line: 1, Column: 1, File: "typeof_test.go"}

// testDummyProcedure for testing typeof function
var testDummyProcedure = Procedure{
	Name: "myTestFuncForTypeOf", // Unique name
	Steps: []Step{
		createTestStep("emit", "", NewTestStringLiteral("from myTestFuncForTypeOf"), nil),
	},
	Pos: testPos,
}

// testDummyTool for testing typeof tool
var testDummyTool = ToolImplementation{
	Spec: ToolSpec{
		Name:        "MyTestToolForTypeOf", // Unique name
		Description: "A dummy tool for testing typeof.",
		Category:    "Test",
		Args:        []ArgSpec{},
		ReturnType:  ArgTypeString,
	},
	Func: func(interpreter *Interpreter, args []interface{}) (interface{}, error) {
		return "dummy tool executed", nil
	},
}

func TestTypeOfOperator_LiteralsAndVariables(t *testing.T) {
	tests := []executeStepsTestCase{
		{
			name: "typeof string literal",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: testPos, Argument: NewTestStringLiteral("hello")}, nil),
			},
			expectedResult: StringValue{Value: string(TypeString)},
		},
		{
			name: "typeof number literal (int)",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: testPos, Argument: NewTestNumberLiteral(123)}, nil),
			},
			expectedResult: StringValue{Value: string(TypeNumber)},
		},
		{
			name: "typeof number literal (float)",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: testPos, Argument: NewTestNumberLiteral(123.45)}, nil),
			},
			expectedResult: StringValue{Value: string(TypeNumber)},
		},
		{
			name: "typeof boolean literal (true)",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: testPos, Argument: NewTestBooleanLiteral(true)}, nil),
			},
			expectedResult: StringValue{Value: string(TypeBoolean)},
		},
		{
			name: "typeof nil literal",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: testPos, Argument: &NilLiteralNode{Pos: testPos}}, nil),
			},
			expectedResult: StringValue{Value: string(TypeNil)},
		},
		{
			name: "typeof list literal",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: testPos, Argument: &ListLiteralNode{Pos: testPos, Elements: []Expression{NewTestNumberLiteral(1), NewTestStringLiteral("a")}}}, nil),
			},
			expectedResult: StringValue{Value: string(TypeList)},
		},
		{
			name: "typeof map literal",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: testPos, Argument: &MapLiteralNode{Pos: testPos, Entries: []*MapEntryNode{
					{Pos: testPos, Key: NewTestStringLiteral("key"), Value: NewTestStringLiteral("value")},
					{Pos: testPos, Key: NewTestStringLiteral("num"), Value: NewTestNumberLiteral(1)},
				}}}, nil),
			},
			expectedResult: StringValue{Value: string(TypeMap)},
		},
		{
			name: "typeof arithmetic expression",
			inputSteps: []Step{
				createTestStep("emit", "", &TypeOfNode{Pos: testPos, Argument: &BinaryOpNode{
					Pos:      testPos,
					Left:     NewTestNumberLiteral(1),
					Operator: "+",
					Right:    NewTestNumberLiteral(2),
				}}, nil),
			},
			expectedResult: StringValue{Value: string(TypeNumber)},
		},
		{
			name: "typeof variable (string)",
			inputSteps: []Step{
				createTestStep("set", "myVar", NewTestStringLiteral("test"), nil),
				createTestStep("emit", "", &TypeOfNode{Pos: testPos, Argument: NewTestVariableNode("myVar")}, nil),
			},
			expectedResult: StringValue{Value: string(TypeString)},
			expectedVars:   map[string]interface{}{"myVar": StringValue{Value: "test"}},
		},
		{
			name: "typeof variable (list)",
			inputSteps: []Step{
				createTestStep("set", "myList", &ListLiteralNode{Pos: testPos, Elements: []Expression{NewTestNumberLiteral(int64(1))}}, nil),
				createTestStep("emit", "", &TypeOfNode{Pos: testPos, Argument: NewTestVariableNode("myList")}, nil),
			},
			expectedResult: StringValue{Value: string(TypeList)},
			expectedVars:   map[string]interface{}{"myList": NewListValue([]Value{NumberValue{Value: 1}})},
		},
		{
			name: "typeof last expression (number)",
			inputSteps: []Step{
				createTestStep("emit", "", NewTestNumberLiteral(100), nil),
				createTestStep("emit", "", &TypeOfNode{Pos: testPos, Argument: &LastNode{Pos: testPos}}, nil),
			},
			expectedResult: StringValue{Value: string(TypeNumber)},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runExecuteStepsTest(t, tc)
		})
	}
}

func TestTypeOfOperator_Function(t *testing.T) {
	i, _ := NewTestInterpreter(t, map[string]interface{}{}, nil)
	err := i.AddProcedure(testDummyProcedure) // Procedure is known to the interpreter
	if err != nil {
		t.Fatalf("Failed to add dummy procedure: %v", err)
	}

	// Store the Procedure object itself in a variable to test typeof(Procedure object)
	err = i.SetVariable(testDummyProcedure.Name, testDummyProcedure)
	if err != nil {
		t.Fatalf("Failed to set variable '%s' to procedure object: %v", testDummyProcedure.Name, err)
	}

	// Create AST node for: typeof(myTestFuncForTypeOf)
	argVarNode := NewTestVariableNode(testDummyProcedure.Name) // Variable now holds the Procedure struct
	typeOfExpr := &TypeOfNode{Pos: testPos, Argument: argVarNode}

	result, evalErr := i.evaluateExpression(typeOfExpr)
	if evalErr != nil {
		t.Fatalf("evaluateExpression failed: %v", evalErr)
	}

	expected := StringValue{Value: string(TypeFunction)}
	if result != expected {
		t.Errorf("Expected typeof(function) to be '%s', got '%s'", expected, result)
	}
}

func TestTypeOfOperator_Tool(t *testing.T) {
	i, _ := NewTestInterpreter(t, map[string]interface{}{}, nil)
	err := i.RegisterTool(testDummyTool)
	if err != nil {
		t.Fatalf("Failed to register dummy tool: %v", err)
	}

	toolVal, found := i.GetTool("MyTestToolForTypeOf")
	if !found {
		t.Fatalf("Failed to retrieve registered tool MyTestToolForTypeOf")
	}
	err = i.SetVariable("myActualTestToolVar", toolVal) // Variable holds the ToolImplementation struct
	if err != nil {
		t.Fatalf("Failed to set variable for tool value: %v", err)
	}

	// Create AST node for: typeof(myActualTestToolVar)
	argVarNode := NewTestVariableNode("myActualTestToolVar")
	typeOfExpr := &TypeOfNode{Pos: testPos, Argument: argVarNode}

	result, evalErr := i.evaluateExpression(typeOfExpr)
	if evalErr != nil {
		t.Fatalf("evaluateExpression failed: %v", evalErr)
	}

	expected := StringValue{Value: string(TypeTool)}
	if result != expected {
		t.Errorf("Expected typeof(tool) to be '%s', got '%s'", expected, result)
	}
}
