// NeuroScript Version: 0.3.1
// File version: 3
// Purpose: Aligns tests with compliant helpers, fixes compiler error by using float64 in NewTestNumberLiteral.
// filename: pkg/runtime/eval_typeof_test.go
// nlines: 201
// risk_rating: LOW

package runtime

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// common lang.Position for test AST nodes
var testPos = &Position{Line: 1, Column: 1, File: "typeof_test.go"}

// testDummyProcedure for testing typeof function
var testDummyProcedure = Procedure{
	Name: "myTestFuncForTypeOf", // Unique name
	Steps: []Step{
		createTestStep("emit", "", NewTestStringLiteral("from myTestFuncForTypeOf"), nil),
	},
	Position: testPos,
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
				createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: NewTestStringLiteral("hello")}, nil),
			},
			expectedResult: StringValue{Value: string(TypeString)},
		},
		{
			name: "typeof number literal (int)",
			inputSteps: []Step{
				createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: NewTestNumberLiteral(123.0)}, nil),
			},
			expectedResult: StringValue{Value: string(TypeNumber)},
		},
		{
			name: "typeof number literal (float)",
			inputSteps: []Step{
				createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: NewTestNumberLiteral(123.45)}, nil),
			},
			expectedResult: StringValue{Value: string(TypeNumber)},
		},
		{
			name: "typeof boolean literal (true)",
			inputSteps: []Step{
				createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: NewTestBooleanLiteral(true)}, nil),
			},
			expectedResult: StringValue{Value: string(TypeBoolean)},
		},
		{
			name: "typeof nil literal",
			inputSteps: []Step{
				createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: &ast.NilLiteralNode{Position: testPos}}, nil),
			},
			expectedResult: StringValue{Value: string(TypeNil)},
		},
		{
			name: "typeof list literal",
			inputSteps: []Step{
				createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: &ast.ListLiteralNode{Position: testPos, Elements: []ast.Expression{NewTestNumberLiteral(1.0), NewTestStringLiteral("a")}}}, nil),
			},
			expectedResult: StringValue{Value: string(TypeList)},
		},
		{
			name: "typeof map literal",
			inputSteps: []Step{
				createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: &ast.MapLiteralNode{Position: testPos, Entries: []*ast.MapEntryNode{
					{Position: testPos, Key: NewTestStringLiteral("key"), Value: NewTestStringLiteral("value")},
					{Position: testPos, Key: NewTestStringLiteral("num"), Value: NewTestNumberLiteral(1.0)},
				}}}, nil),
			},
			expectedResult: StringValue{Value: string(TypeMap)},
		},
		{
			name: "typeof arithmetic expression",
			inputSteps: []Step{
				createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: &ast.BinaryOpNode{
					Position: testPos,
					Left:     NewTestNumberLiteral(1.0),
					Operator: "+",
					Right:    NewTestNumberLiteral(2.0),
				}}, nil),
			},
			expectedResult: StringValue{Value: string(TypeNumber)},
		},
		{
			name: "typeof variable (string)",
			inputSteps: []Step{
				createTestStep("set", "myVar", NewTestStringLiteral("test"), nil),
				createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: NewTestast.VariableNode("myVar")}, nil),
			},
			expectedResult: StringValue{Value: string(TypeString)},
			expectedVars:   map[string]Value{"myVar": StringValue{Value: "test"}},
		},
		{
			name: "typeof variable (list)",
			inputSteps: []Step{
				createTestStep("set", "myList", &ast.ListLiteralNode{Position: testPos, Elements: []ast.Expression{NewTestNumberLiteral(1.0)}}, nil),
				createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: NewTestast.VariableNode("myList")}, nil),
			},
			expectedResult: StringValue{Value: string(TypeList)},
			expectedVars:   map[string]Value{"myList": NewListValue([]Value{NumberValue{Value: 1}})},
		},
		{
			name: "typeof last expression (number)",
			inputSteps: []Step{
				createTestStep("emit", "", NewTestNumberLiteral(100.0), nil),
				createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: &ast.EvalNode{Position: testPos}}, nil),
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
	i, _ := NewInterpreter(NewTestLogger(t), nil, ".", nil, nil)
	err := i.AddProcedure(testDummyProcedure)
	if err != nil {
		t.Fatalf("Failed to add dummy procedure: %v", err)
	}

	err = i.SetVariable(testDummyProcedure.Name, FunctionValue{Value: testDummyProcedure})
	if err != nil {
		t.Fatalf("Failed to set variable '%s' to procedure object: %v", testDummyProcedure.Name, err)
	}

	argVarNode := NewTestast.VariableNode(testDummyProcedure.Name)
	typeOfExpr := &ast.TypeOfNode{Position: testPos, Argument: argVarNode}

	result, evalErr := i.evaluate.Expression(typeOfExpr)
	if evalErr != nil {
		t.Fatalf("evaluate.Expression failed: %v", evalErr)
	}

	expected := StringValue{Value: string(TypeFunction)}
	if result != expected {
		t.Errorf("Expected typeof(function) to be '%s', got '%s'", expected, result)
	}
}

func TestTypeOfOperator_Tool(t *testing.T) {
	i, _ := NewInterpreter(NewTestLogger(t), nil, ".", nil, nil)
	err := i.RegisterTool(testDummyTool)
	if err != nil {
		t.Fatalf("Failed to register dummy tool: %v", err)
	}

	toolVal, found := i.GetTool("MyTestToolForTypeOf")
	if !found {
		t.Fatalf("Failed to retrieve registered tool MyTestToolForTypeOf")
	}

	err = i.SetVariable("myActualTestToolVar", ToolValue{Value: toolVal})
	if err != nil {
		t.Fatalf("Failed to set variable for tool value: %v", err)
	}

	argVarNode := NewTestast.VariableNode("myActualTestToolVar")
	typeOfExpr := &ast.TypeOfNode{Position: testPos, Argument: argVarNode}

	result, evalErr := i.evaluate.Expression(typeOfExpr)
	if evalErr != nil {
		t.Fatalf("evaluate.Expression failed: %v", evalErr)
	}

	expected := StringValue{Value: string(TypeTool)}
	if result != expected {
		t.Errorf("Expected typeof(tool) to be '%s', got '%s'", expected, result)
	}
}
