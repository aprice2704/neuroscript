// NeuroScript Version: 0.3.1
// File version: 3
// Purpose: Aligns tests with compliant helpers, fixes compiler error by using float64 in NewTestNumberLiteral.
// filename: pkg/interpreter/internal/eval/eval_typeof_test.go
// nlines: 201
// risk_rating: LOW

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// common lang.Position for test AST nodes
var testPos = &lang.Position{Line: 1, Column: 1, File: "typeof_test.go"}

// testDummyProcedure for testing typeof function
var testDummyProcedure = ast.Procedure{
	Name: "myTestFuncForTypeOf", // Unique name
	Steps: []ast.Step{
		testutil.createTestStep("emit", "", testutil.NewTestStringLiteral("from myTestFuncForTypeOf"), nil),
	},
	Position: testPos,
}

// testDummyTool for testing typeof tool
var testDummyTool = tool.ToolImplementation{
	Spec: tool.ToolSpec{
		Name:        "MyTestToolForTypeOf", // Unique name
		Description: "A dummy tool for testing typeof.",
		Category:    "Test",
		Args:        []tool.ArgSpec{},
		ReturnType:  tool.tool.ArgTypeString,
	},
	Func: func(interpreter *Interpreter, args []interface{}) (interface{}, error) {
		return "dummy tool executed", nil
	},
}

func TestTypeOfOperator_LiteralsAndVariables(t *testing.T) {
	tests := []testutil.executeStepsTestCase{
		{
			name: "typeof string literal",
			inputSteps: []ast.Step{
				testutil.createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: testutil.NewTestStringLiteral("hello")}, nil),
			},
			expectedResult: lang.StringValue{Value: string(lang.TypeString)},
		},
		{
			name: "typeof number literal (int)",
			inputSteps: []ast.Step{
				testutil.createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: testutil.NewTestNumberLiteral(123.0)}, nil),
			},
			expectedResult: lang.StringValue{Value: string(lang.TypeNumber)},
		},
		{
			name: "typeof number literal (float)",
			inputSteps: []ast.Step{
				testutil.createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: testutil.NewTestNumberLiteral(123.45)}, nil),
			},
			expectedResult: lang.StringValue{Value: string(lang.TypeNumber)},
		},
		{
			name: "typeof boolean literal (true)",
			inputSteps: []ast.Step{
				testutil.createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: testutil.NewTestBooleanLiteral(true)}, nil),
			},
			expectedResult: lang.StringValue{Value: string(lang.TypeBoolean)},
		},
		{
			name: "typeof nil literal",
			inputSteps: []ast.Step{
				testutil.createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: &ast.NilLiteralNode{Position: testPos}}, nil),
			},
			expectedResult: lang.StringValue{Value: string(lang.TypeNil)},
		},
		{
			name: "typeof list literal",
			inputSteps: []ast.Step{
				testutil.createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: &ast.ListLiteralNode{Position: testPos, Elements: []ast.Expression{testutil.NewTestNumberLiteral(1.0), testutil.NewTestStringLiteral("a")}}}, nil),
			},
			expectedResult: lang.StringValue{Value: string(lang.TypeList)},
		},
		{
			name: "typeof map literal",
			inputSteps: []ast.Step{
				testutil.createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: &ast.MapLiteralNode{Position: testPos, Entries: []*ast.MapEntryNode{
					{Position: testPos, Key: testutil.testutil.NewTestStringLiteral("key"), Value: testutil.testutil.NewTestStringLiteral("value")},
					{Position: testPos, Key: testutil.NewTestStringLiteral("num"), Value: testutil.NewTestNumberLiteral(1.0)},
				}}}, nil),
			},
			expectedResult: lang.StringValue{Value: string(lang.TypeMap)},
		},
		{
			name: "typeof arithmetic expression",
			inputSteps: []ast.Step{
				testutil.createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: &ast.BinaryOpNode{
					Position: testPos,
					Left:     testutil.NewTestNumberLiteral(1.0),
					Operator: "+",
					Right:    testutil.NewTestNumberLiteral(2.0),
				}}, nil),
			},
			expectedResult: lang.StringValue{Value: string(lang.TypeNumber)},
		},
		{
			name: "typeof variable (string)",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "myVar", testutil.NewTestStringLiteral("test"), nil),
				testutil.createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: testutil.NewTestast.VariableNode("myVar")}, nil),
			},
			expectedResult: lang.StringValue{Value: string(lang.TypeString)},
			expectedVars:   map[string]lang.Value{"myVar": lang.StringValue{Value: "test"}},
		},
		{
			name: "typeof variable (list)",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "myList", &ast.ListLiteralNode{Position: testPos, Elements: []ast.Expression{testutil.NewTestNumberLiteral(1.0)}}, nil),
				testutil.createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: testutil.NewTestast.VariableNode("myList")}, nil),
			},
			expectedResult: lang.StringValue{Value: string(lang.TypeList)},
			expectedVars:   map[string]Value{"myList": lang.NewListValue([]Value{lang.NumberValue{Value: 1}})},
		},
		{
			name: "typeof last expression (number)",
			inputSteps: []ast.Step{
				testutil.createTestStep("emit", "", testutil.NewTestNumberLiteral(100.0), nil),
				testutil.createTestStep("emit", "", &ast.TypeOfNode{Position: testPos, Argument: &ast.EvalNode{Position: testPos}}, nil),
			},
			expectedResult: lang.StringValue{Value: string(lang.TypeNumber)},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testutil.runExecuteStepsTest(t, tc)
		})
	}
}

func TestTypeOfOperator_Function(t *testing.T) {
	i, _ := NewInterpreter(llm.NewTestLogger(t), nil, ".", nil, nil)
	err := i.AddProcedure(testDummyProcedure)
	if err != nil {
		t.Fatalf("Failed to add dummy procedure: %v", err)
	}

	err = i.SetVariable(testDummyProcedure.Name, lang.FunctionValue{Value: testDummyProcedure})
	if err != nil {
		t.Fatalf("Failed to set variable '%s' to procedure object: %v", testDummyProcedure.Name, err)
	}

	argVarNode := testutil.NewTestast.VariableNode(testDummyProcedure.Name)
	typeOfExpr := &ast.TypeOfNode{Position: testPos, Argument: argVarNode}

	result, evalErr := i.evaluate.Expression(typeOfExpr)
	if evalErr != nil {
		t.Fatalf("evaluate.Expression failed: %v", evalErr)
	}

	expected := lang.StringValue{Value: string(lang.TypeFunction)}
	if result != expected {
		t.Errorf("Expected typeof(function) to be '%s', got '%s'", expected, result)
	}
}

func TestTypeOfOperator_Tool(t *testing.T) {
	i, _ := NewInterpreter(llm.NewTestLogger(t), nil, ".", nil, nil)
	err := i.RegisterTool(testDummyTool)
	if err != nil {
		t.Fatalf("Failed to register dummy tool: %v", err)
	}

	toolVal, found := i.GetTool("MyTestToolForTypeOf")
	if !found {
		t.Fatalf("Failed to retrieve registered tool MyTestToolForTypeOf")
	}

	err = i.SetVariable("myActualTestToolVar", lang.ToolValue{Value: toolVal})
	if err != nil {
		t.Fatalf("Failed to set variable for tool value: %v", err)
	}

	argVarNode := testutil.NewTestast.VariableNode("myActualTestToolVar")
	typeOfExpr := &ast.TypeOfNode{Position: testPos, Argument: argVarNode}

	result, evalErr := i.evaluate.Expression(typeOfExpr)
	if evalErr != nil {
		t.Fatalf("evaluate.Expression failed: %v", evalErr)
	}

	expected := lang.StringValue{Value: string(lang.TypeTool)}
	if result != expected {
		t.Errorf("Expected typeof(tool) to be '%s', got '%s'", expected, result)
	}
}
