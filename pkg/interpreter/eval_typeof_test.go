// NeuroScript Version: 0.5.2
// File version: 12.0.0
// Purpose: Corrected dummy tool registration to include a Group, allowing it to be found by its full name and fixing the test.
// filename: pkg/interpreter/eval_typeof_test.go
package interpreter_test

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/testutil"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

var testPos = &types.Position{Line: 1, Column: 1, File: "typeof_test.go"}

var testDummyProcedure = ast.Procedure{}

var testDummyTool = tool.ToolImplementation{
	Spec: tool.ToolSpec{
		// FIX: Added a Group to the tool spec for proper registration.
		Name:        "MyTestToolForTypeOf",
		Group:       "Test",
		Description: "A dummy tool for testing typeof.",
		Category:    "Test",
		Args:        []tool.ArgSpec{},
		ReturnType:  "string",
	},
	Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
		return "dummy tool executed", nil
	},
}

func TestTypeOfOperator_LiteralsAndVariables(t *testing.T) {
	// FIX: These tests now check the ExpectedResult from the expression, rather than emitting output.
	tests := []testutil.EvalTestCase{
		{
			Name:      "typeof string literal",
			InputNode: &ast.TypeOfNode{BaseNode: ast.BaseNode{StartPos: testPos}, Argument: testutil.NewTestStringLiteral("hello")},
			Expected:  lang.StringValue{Value: string(lang.TypeString)},
		},
		{
			Name:      "typeof number literal (int)",
			InputNode: &ast.TypeOfNode{BaseNode: ast.BaseNode{StartPos: testPos}, Argument: testutil.NewTestNumberLiteral(123.0)},
			Expected:  lang.StringValue{Value: string(lang.TypeNumber)},
		},
		{
			Name:      "typeof nil literal",
			InputNode: &ast.TypeOfNode{BaseNode: ast.BaseNode{StartPos: testPos}, Argument: &ast.NilLiteralNode{BaseNode: ast.BaseNode{StartPos: testPos}}},
			Expected:  lang.StringValue{Value: string(lang.TypeNil)},
		},
		{
			Name:      "typeof list literal",
			InputNode: &ast.TypeOfNode{BaseNode: ast.BaseNode{StartPos: testPos}, Argument: &ast.ListLiteralNode{BaseNode: ast.BaseNode{StartPos: testPos}, Elements: []ast.Expression{testutil.NewTestNumberLiteral(1.0), testutil.NewTestStringLiteral("a")}}},
			Expected:  lang.StringValue{Value: string(lang.TypeList)},
		},
		{
			Name: "typeof map literal",
			InputNode: &ast.TypeOfNode{BaseNode: ast.BaseNode{StartPos: testPos}, Argument: &ast.MapLiteralNode{BaseNode: ast.BaseNode{StartPos: testPos}, Entries: []*ast.MapEntryNode{
				{BaseNode: ast.BaseNode{StartPos: testPos}, Key: testutil.NewTestStringLiteral("key"), Value: testutil.NewTestStringLiteral("value")},
			}}},
			Expected: lang.StringValue{Value: string(lang.TypeMap)},
		},
		{
			Name:        "typeof variable (string)",
			InputNode:   &ast.TypeOfNode{BaseNode: ast.BaseNode{StartPos: testPos}, Argument: testutil.NewVariableNode("myVar")},
			InitialVars: map[string]lang.Value{"myVar": lang.StringValue{Value: "test"}},
			Expected:    lang.StringValue{Value: string(lang.TypeString)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			testutil.ExpressionTest(t, tc)
		})
	}
}

func TestTypeOfOperator_Function(t *testing.T) {
	i := interpreter.NewInterpreter()
	procToTest := testDummyProcedure
	procToTest.SetName("myTestFuncForTypeOf")
	err := i.AddProcedure(procToTest)
	if err != nil {
		t.Fatalf("Failed to add dummy procedure: %v", err)
	}

	err = i.SetInitialVariable("myFuncVar", lang.FunctionValue{Value: &procToTest})
	if err != nil {
		t.Fatalf("Failed to set variable for function: %v", err)
	}

	argVarNode := testutil.NewVariableNode("myFuncVar")
	typeOfExpr := &ast.TypeOfNode{BaseNode: ast.BaseNode{StartPos: testPos}, Argument: argVarNode}

	result, evalErr := i.EvaluateExpression(typeOfExpr)
	if evalErr != nil {
		t.Fatalf("EvaluateExpression failed: %v", evalErr)
	}

	expected := lang.StringValue{Value: string(lang.TypeFunction)}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected typeof(function) to be '%s', got '%s'", expected, result)
	}
}

func TestTypeOfOperator_Tool(t *testing.T) {
	i := interpreter.NewInterpreter()
	_, err := i.ToolRegistry().RegisterTool(testDummyTool)
	if err != nil {
		t.Fatalf("Failed to register dummy tool: %v", err)
	}
	// FIX: Use the full name to retrieve the tool.
	fullName := types.MakeFullName(string(testDummyTool.Spec.Group), string(testDummyTool.Spec.Name))
	toolVal, found := i.ToolRegistry().GetTool(fullName)
	if !found {
		t.Fatalf("Failed to retrieve registered tool %s", fullName)
	}

	err = i.SetInitialVariable("myActualTestToolVar", lang.ToolValue{Value: &toolVal})
	if err != nil {
		t.Fatalf("Failed to set variable for tool value: %v", err)
	}

	argVarNode := testutil.NewVariableNode("myActualTestToolVar")
	typeOfExpr := &ast.TypeOfNode{BaseNode: ast.BaseNode{StartPos: testPos}, Argument: argVarNode}

	result, evalErr := i.EvaluateExpression(typeOfExpr)
	if evalErr != nil {
		t.Fatalf("EvaluateExpression failed: %v", evalErr)
	}

	expected := lang.StringValue{Value: string(lang.TypeTool)}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected typeof(tool) to be '%s', got '%s'", expected, result)
	}
}
