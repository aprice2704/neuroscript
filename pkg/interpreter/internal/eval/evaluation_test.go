// NeuroScript Version: 0.3.5
// File version: 1.0.0
// Purpose: Fully aligned with testing helpers, using  Value for all initial, last, and expected values.
// filename: pkg/interpreter/internal/eval/evaluation_test.go
package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func ExpressionASTGeneral(t *testing.T) {
	vars := map[string]Value{
		"name":		StringValue{Value: "World"},
		"greeting":	StringValue{Value: "Hello {{name}}"},
		"numVar":	NumberValue{Value: 123},
		"boolProp":	BoolValue{Value: true},
		"listVar": NewListValue([]Value{
			StringValue{Value: "x"},
			NumberValue{Value: 99},
			StringValue{Value: "{{name}}"},
		}),
		"mapVar": NewMapValue(map[string]Value{
			"mKey":	StringValue{Value: "mVal {{name}}"},
			"mNum":	NumberValue{Value: 1},
		}),
		"nilVar":	NilValue{},
		"numStr":	StringValue{Value: "456"},
	}
	lastResult := StringValue{Value: "LastCallResult {{name}}"}

	dummyPos := &Position{Line: 1, Column: 1}

	tests := []EvalTestCase{
		{Name: "String Literal (Raw)", InputNode: &ast.StringLiteralNode{Position: dummyPos, Value: "Hello {{name}}"}, InitialVars: vars, LastResult: lastResult, Expected: StringValue{Value: `Hello {{name}}`}},
		{Name: "Variable String (Raw)", InputNode: &ast.VariableNode{Position: dummyPos, Name: "greeting"}, InitialVars: vars, LastResult: lastResult, Expected: StringValue{Value: `Hello {{name}}`}},
		{Name: "Last Call Result (Raw)", InputNode: &ast.EvalNode{Position: dummyPos}, InitialVars: vars, LastResult: lastResult, Expected: StringValue{Value: `LastCallResult {{name}}`}},
		{Name: "Placeholder to String (Raw Ref Value)", InputNode: &ast.Placeholder.Node{Position: dummyPos, Name: "greeting"}, InitialVars: vars, LastResult: lastResult, Expected: StringValue{Value: `Hello {{name}}`}},
		{Name: "Placeholder LAST (Raw Ref Value)", InputNode: &ast.Placeholder.Node{Position: dummyPos, Name: "LAST"}, InitialVars: vars, LastResult: lastResult, Expected: StringValue{Value: `LastCallResult {{name}}`}},

		// --- Concatenation Tests ---
		{Name: "Concat Lit(raw) + Var(raw)", InputNode: &ast.BinaryOpNode{Position: dummyPos, Left: &ast.StringLiteralNode{Position: dummyPos, Value: "A={{name}} "}, Operator: "+", Right: &ast.VariableNode{Position: dummyPos, Name: "greeting"}}, InitialVars: vars, Expected: StringValue{Value: "A={{name}} Hello {{name}}"}},
		{Name: "Concat Var(raw) + Lit(raw)", InputNode: &ast.BinaryOpNode{Position: dummyPos, Left: &ast.VariableNode{Position: dummyPos, Name: "greeting"}, Operator: "+", Right: &ast.StringLiteralNode{Position: dummyPos, Value: " B={{name}}"}}, InitialVars: vars, Expected: StringValue{Value: "Hello {{name}} B={{name}}"}},
		{Name: "Concat with Number", InputNode: &ast.BinaryOpNode{Position: dummyPos, Left: &ast.StringLiteralNode{Position: dummyPos, Value: "Count: "}, Operator: "+", Right: &ast.VariableNode{Position: dummyPos, Name: "numVar"}}, InitialVars: vars, Expected: StringValue{Value: "Count: 123"}},
		{Name: "Concat Error Operand", InputNode: &ast.BinaryOpNode{Position: dummyPos, Left: &ast.StringLiteralNode{Position: dummyPos, Value: "Val: "}, Operator: "+", Right: &ast.VariableNode{Position: dummyPos, Name: "missing"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrVariableNotFound},
		{Name: "Concat Nil Operand", InputNode: &ast.BinaryOpNode{Position: dummyPos, Left: &ast.BinaryOpNode{Position: dummyPos, Left: &ast.StringLiteralNode{Position: dummyPos, Value: "Start:"}, Operator: "+", Right: &ast.VariableNode{Position: dummyPos, Name: "nilVar"}}, Operator: "+", Right: &ast.StringLiteralNode{Position: dummyPos, Value: ":End {{name}}"}}, InitialVars: vars, Expected: StringValue{Value: "Start::End {{name}}"}},

		// --- Arithmetic Tests ---
		{Name: "Add Numbers", InputNode: &ast.BinaryOpNode{Position: dummyPos, Left: &ast.NumberLiteralNode{Position: dummyPos, Value: 5.0}, Operator: "+", Right: &ast.NumberLiteralNode{Position: dummyPos, Value: 3.0}}, InitialVars: vars, Expected: NumberValue{Value: 8}},
		{Name: "Add Num + NumStr", InputNode: &ast.BinaryOpNode{Position: dummyPos, Left: &ast.VariableNode{Position: dummyPos, Name: "numVar"}, Operator: "+", Right: &ast.VariableNode{Position: dummyPos, Name: "numStr"}}, InitialVars: vars, Expected: NumberValue{Value: 579}},
		{Name: "Divide Numbers (Float)", InputNode: &ast.BinaryOpNode{Position: dummyPos, Left: &ast.NumberLiteralNode{Position: dummyPos, Value: 7.0}, Operator: "/", Right: &ast.NumberLiteralNode{Position: dummyPos, Value: 2.0}}, InitialVars: vars, Expected: NumberValue{Value: 3.5}},
		{Name: "Power Numbers", InputNode: &ast.BinaryOpNode{Position: dummyPos, Left: &ast.NumberLiteralNode{Position: dummyPos, Value: 2.0}, Operator: "**", Right: &ast.NumberLiteralNode{Position: dummyPos, Value: 3.0}}, InitialVars: vars, Expected: NumberValue{Value: 8}},

		// --- Lists, Maps ---
		{Name: "Simple List (Raw)", InputNode: &ast.ListLiteralNode{Position: dummyPos, Elements: []ast.Expression{&ast.NumberLiteralNode{Position: dummyPos, Value: 1.0}, &ast.StringLiteralNode{Position: dummyPos, Value: "{{name}}"}, &ast.VariableNode{Position: dummyPos, Name: "boolProp"}}}, InitialVars: vars, Expected: NewListValue([]Value{NumberValue{Value: 1}, StringValue{Value: "{{name}}"}, BoolValue{Value: true}})},
		{Name: "Simple Map (Raw)", InputNode: &ast.MapLiteralNode{Position: dummyPos, Entries: []*ast.MapEntryNode{{Position: dummyPos, Key: &ast.StringLiteralNode{Position: dummyPos, Value: "k1"}, Value: &ast.StringLiteralNode{Position: dummyPos, Value: "{{name}}"}}}}, InitialVars: vars, Expected: NewMapValue(map[string]Value{"k1": StringValue{Value: "{{name}}"}})},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runEval.ExpressionTest(t, tt)
		})
	}
}