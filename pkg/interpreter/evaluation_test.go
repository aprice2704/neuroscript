// NeuroScript Version: 0.3.5
// File version: 1.1.0
// Purpose: Fully aligned with testing helpers, using lang.Value for all initial, last, and expected values, and added the required lang package import.
// filename: pkg/interpreter/evaluation_test.go
package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/testutil"
)

func TestExpressionASTGeneral(t *testing.T) {
	vars := map[string]lang.Value{
		"name":     lang.StringValue{Value: "World"},
		"greeting": lang.StringValue{Value: "Hello {{name}}"},
		"numVar":   lang.NumberValue{Value: 123},
		"boolProp": lang.BoolValue{Value: true},
		"listVar": lang.NewListValue([]lang.Value{
			lang.StringValue{Value: "x"},
			lang.NumberValue{Value: 99},
			lang.StringValue{Value: "{{name}}"},
		}),
		"mapVar": lang.NewMapValue(map[string]lang.Value{
			"mKey": lang.StringValue{Value: "mVal {{name}}"},
			"mNum": lang.NumberValue{Value: 1},
		}),
		"nilVar": &lang.NilValue{},
		"numStr": lang.StringValue{Value: "456"},
	}
	lastResult := lang.StringValue{Value: "LastCallResult {{name}}"}

	dummyPos := &lang.Position{Line: 1, Column: 1, File: "test"}

	tests := []testutil.EvalTestCase{
		{Name: "String Literal (Raw)", InputNode: &ast.StringLiteralNode{Pos: dummyPos, Value: "Hello {{name}}"}, InitialVars: vars, LastResult: lastResult, Expected: lang.StringValue{Value: `Hello {{name}}`}},
		{Name: "Variable String (Raw)", InputNode: &ast.VariableNode{Pos: dummyPos, Name: "greeting"}, InitialVars: vars, LastResult: lastResult, Expected: lang.StringValue{Value: `Hello {{name}}`}},
		{Name: "Last Call Result (Raw)", InputNode: &ast.EvalNode{Pos: dummyPos}, InitialVars: vars, LastResult: lastResult, Expected: lang.StringValue{Value: `LastCallResult {{name}}`}},
		{Name: "Placeholder to String (Raw Ref Value)", InputNode: &ast.PlaceholderNode{Pos: dummyPos, Name: "greeting"}, InitialVars: vars, LastResult: lastResult, Expected: lang.StringValue{Value: `Hello {{name}}`}},
		{Name: "Placeholder LAST (Raw Ref Value)", InputNode: &ast.PlaceholderNode{Pos: dummyPos, Name: "LAST"}, InitialVars: vars, LastResult: lastResult, Expected: lang.StringValue{Value: `LastCallResult {{name}}`}},

		// --- Concatenation Tests ---
		{Name: "Concat Lit(raw) + Var(raw)", InputNode: &ast.BinaryOpNode{Pos: dummyPos, Left: &ast.StringLiteralNode{Pos: dummyPos, Value: "A={{name}} "}, Operator: "+", Right: &ast.VariableNode{Pos: dummyPos, Name: "greeting"}}, InitialVars: vars, Expected: lang.StringValue{Value: "A={{name}} Hello {{name}}"}},
		{Name: "Concat Var(raw) + Lit(raw)", InputNode: &ast.BinaryOpNode{Pos: dummyPos, Left: &ast.VariableNode{Pos: dummyPos, Name: "greeting"}, Operator: "+", Right: &ast.StringLiteralNode{Pos: dummyPos, Value: " B={{name}}"}}, InitialVars: vars, Expected: lang.StringValue{Value: "Hello {{name}} B={{name}}"}},
		{Name: "Concat with Number", InputNode: &ast.BinaryOpNode{Pos: dummyPos, Left: &ast.StringLiteralNode{Pos: dummyPos, Value: "Count: "}, Operator: "+", Right: &ast.VariableNode{Pos: dummyPos, Name: "numVar"}}, InitialVars: vars, Expected: lang.StringValue{Value: "Count: 123"}},
		{Name: "Concat Error Operand", InputNode: &ast.BinaryOpNode{Pos: dummyPos, Left: &ast.StringLiteralNode{Pos: dummyPos, Value: "Val: "}, Operator: "+", Right: &ast.VariableNode{Pos: dummyPos, Name: "missing"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: lang.ErrVariableNotFound},
		{Name: "Concat Nil Operand", InputNode: &ast.BinaryOpNode{Pos: dummyPos, Left: &ast.BinaryOpNode{Pos: dummyPos, Left: &ast.StringLiteralNode{Pos: dummyPos, Value: "Start:"}, Operator: "+", Right: &ast.VariableNode{Pos: dummyPos, Name: "nilVar"}}, Operator: "+", Right: &ast.StringLiteralNode{Pos: dummyPos, Value: ":End {{name}}"}}, InitialVars: vars, Expected: lang.StringValue{Value: "Start::End {{name}}"}},

		// --- Arithmetic Tests ---
		{Name: "Add Numbers", InputNode: &ast.BinaryOpNode{Pos: dummyPos, Left: &ast.NumberLiteralNode{Value: 5.0}, Operator: "+", Right: &ast.NumberLiteralNode{Value: 3.0}}, InitialVars: vars, Expected: lang.NumberValue{Value: 8}},
		{Name: "Add Num + NumStr", InputNode: &ast.BinaryOpNode{Pos: dummyPos, Left: &ast.VariableNode{Pos: dummyPos, Name: "numVar"}, Operator: "+", Right: &ast.VariableNode{Pos: dummyPos, Name: "numStr"}}, InitialVars: vars, Expected: lang.NumberValue{Value: 579}},
		{Name: "Divide Numbers (Float)", InputNode: &ast.BinaryOpNode{Pos: dummyPos, Left: &ast.NumberLiteralNode{Value: 7.0}, Operator: "/", Right: &ast.NumberLiteralNode{Value: 2.0}}, InitialVars: vars, Expected: lang.NumberValue{Value: 3.5}},
		{Name: "Power Numbers", InputNode: &ast.BinaryOpNode{Pos: dummyPos, Left: &ast.NumberLiteralNode{Value: 2.0}, Operator: "**", Right: &ast.NumberLiteralNode{Value: 3.0}}, InitialVars: vars, Expected: lang.NumberValue{Value: 8}},

		// --- Lists, Maps ---
		{Name: "Simple List (Raw)", InputNode: &ast.ListLiteralNode{Pos: dummyPos, Elements: []ast.Expression{&ast.NumberLiteralNode{Value: 1.0}, &ast.StringLiteralNode{Value: "{{name}}"}, &ast.VariableNode{Name: "boolProp"}}}, InitialVars: vars, Expected: lang.NewListValue([]lang.Value{lang.NumberValue{Value: 1}, lang.StringValue{Value: "{{name}}"}, lang.BoolValue{Value: true}})},
		{Name: "Simple Map (Raw)", InputNode: &ast.MapLiteralNode{Pos: dummyPos, Entries: []*ast.MapEntryNode{{Key: &ast.StringLiteralNode{Value: "k1"}, Value: &ast.StringLiteralNode{Value: "{{name}}"}}}}, InitialVars: vars, Expected: lang.NewMapValue(map[string]lang.Value{"k1": lang.StringValue{Value: "{{name}}"}})},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testutil.ExpressionTest(t, tt)
		})
	}
}
