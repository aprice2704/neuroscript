// NeuroScript Version: 0.3.5
// File version: 1.0.0
// Purpose: Fully aligned with testing helpers, using core.Value for all initial, last, and expected values.
// filename: pkg/core/evaluation_test.go
package core

import (
	"testing"
)

func TestEvaluateExpressionASTGeneral(t *testing.T) {
	vars := map[string]Value{
		"name":     StringValue{Value: "World"},
		"greeting": StringValue{Value: "Hello {{name}}"},
		"numVar":   NumberValue{Value: 123},
		"boolProp": BoolValue{Value: true},
		"listVar": NewListValue([]Value{
			StringValue{Value: "x"},
			NumberValue{Value: 99},
			StringValue{Value: "{{name}}"},
		}),
		"mapVar": NewMapValue(map[string]Value{
			"mKey": StringValue{Value: "mVal {{name}}"},
			"mNum": NumberValue{Value: 1},
		}),
		"nilVar": NilValue{},
		"numStr": StringValue{Value: "456"},
	}
	lastResult := StringValue{Value: "LastCallResult {{name}}"}

	dummyPos := &Position{Line: 1, Column: 1}

	tests := []EvalTestCase{
		{Name: "String Literal (Raw)", InputNode: &StringLiteralNode{Pos: dummyPos, Value: "Hello {{name}}"}, InitialVars: vars, LastResult: lastResult, Expected: StringValue{Value: `Hello {{name}}`}},
		{Name: "Variable String (Raw)", InputNode: &VariableNode{Pos: dummyPos, Name: "greeting"}, InitialVars: vars, LastResult: lastResult, Expected: StringValue{Value: `Hello {{name}}`}},
		{Name: "Last Call Result (Raw)", InputNode: &LastNode{Pos: dummyPos}, InitialVars: vars, LastResult: lastResult, Expected: StringValue{Value: `LastCallResult {{name}}`}},
		{Name: "Placeholder to String (Raw Ref Value)", InputNode: &PlaceholderNode{Pos: dummyPos, Name: "greeting"}, InitialVars: vars, LastResult: lastResult, Expected: StringValue{Value: `Hello {{name}}`}},
		{Name: "Placeholder LAST (Raw Ref Value)", InputNode: &PlaceholderNode{Pos: dummyPos, Name: "LAST"}, InitialVars: vars, LastResult: lastResult, Expected: StringValue{Value: `LastCallResult {{name}}`}},

		// --- Concatenation Tests ---
		{Name: "Concat Lit(raw) + Var(raw)", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "A={{name}} "}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "greeting"}}, InitialVars: vars, Expected: StringValue{Value: "A={{name}} Hello {{name}}"}},
		{Name: "Concat Var(raw) + Lit(raw)", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "greeting"}, Operator: "+", Right: &StringLiteralNode{Pos: dummyPos, Value: " B={{name}}"}}, InitialVars: vars, Expected: StringValue{Value: "Hello {{name}} B={{name}}"}},
		{Name: "Concat with Number", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "Count: "}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "numVar"}}, InitialVars: vars, Expected: StringValue{Value: "Count: 123"}},
		{Name: "Concat Error Operand", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "Val: "}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "missing"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrVariableNotFound},
		{Name: "Concat Nil Operand", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "Start:"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, Operator: "+", Right: &StringLiteralNode{Pos: dummyPos, Value: ":End {{name}}"}}, InitialVars: vars, Expected: StringValue{Value: "Start::End {{name}}"}},

		// --- Arithmetic Tests ---
		{Name: "Add Numbers", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: 5.0}, Operator: "+", Right: &NumberLiteralNode{Pos: dummyPos, Value: 3.0}}, InitialVars: vars, Expected: NumberValue{Value: 8}},
		{Name: "Add Num + NumStr", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "numVar"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "numStr"}}, InitialVars: vars, Expected: NumberValue{Value: 579}},
		{Name: "Divide Numbers (Float)", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: 7.0}, Operator: "/", Right: &NumberLiteralNode{Pos: dummyPos, Value: 2.0}}, InitialVars: vars, Expected: NumberValue{Value: 3.5}},
		{Name: "Power Numbers", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: 2.0}, Operator: "**", Right: &NumberLiteralNode{Pos: dummyPos, Value: 3.0}}, InitialVars: vars, Expected: NumberValue{Value: 8}},

		// --- Lists, Maps ---
		{Name: "Simple List (Raw)", InputNode: &ListLiteralNode{Pos: dummyPos, Elements: []Expression{&NumberLiteralNode{Pos: dummyPos, Value: 1.0}, &StringLiteralNode{Pos: dummyPos, Value: "{{name}}"}, &VariableNode{Pos: dummyPos, Name: "boolProp"}}}, InitialVars: vars, Expected: NewListValue([]Value{NumberValue{Value: 1}, StringValue{Value: "{{name}}"}, BoolValue{Value: true}})},
		{Name: "Simple Map (Raw)", InputNode: &MapLiteralNode{Pos: dummyPos, Entries: []*MapEntryNode{{Pos: dummyPos, Key: &StringLiteralNode{Pos: dummyPos, Value: "k1"}, Value: &StringLiteralNode{Pos: dummyPos, Value: "{{name}}"}}}}, InitialVars: vars, Expected: NewMapValue(map[string]Value{"k1": StringValue{Value: "{{name}}"}})},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runEvalExpressionTest(t, tt)
		})
	}
}
