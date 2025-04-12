// filename: neuroscript/pkg/core/evaluation_test.go
package core

import (
	"testing"
)

// runEvalExpressionTest is now defined in testing_helpers_test.go

// --- Tests for General Expression Evaluation (Using BinaryOpNode for +) ---
func TestEvaluateExpressionASTGeneral(t *testing.T) {
	vars := map[string]interface{}{
		"name":     "World",
		"greeting": "Hello {{name}}", // Contains placeholder
		"numVar":   int64(123),
		"boolProp": true,
		"listVar":  []interface{}{"x", int64(99), "{{name}}"},
		"mapVar":   map[string]interface{}{"mKey": "mVal {{name}}", "mNum": int64(1)},
		"nilVar":   nil,
		"numStr":   "456",
	}
	lastResult := "LastCallResult {{name}}" // Raw value

	tests := []EvalTestCase{ // Use defined struct
		// Literals, Variables, Placeholders, LastNode -> Expect RAW
		{"String Literal (Raw)", StringLiteralNode{Value: "Hello {{name}}"}, vars, lastResult, `Hello {{name}}`, false, ""},
		{"Variable String (Raw)", VariableNode{Name: "greeting"}, vars, lastResult, `Hello {{name}}`, false, ""},
		{"Last Call Result (Raw)", LastNode{}, vars, lastResult, `LastCallResult {{name}}`, false, ""},
		// Evaluating a placeholder directly returns the raw value it references
		{"Placeholder to String (Raw Ref Value)", PlaceholderNode{Name: "greeting"}, vars, lastResult, `Hello {{name}}`, false, ""},
		{"Placeholder LAST (Raw Ref Value)", PlaceholderNode{Name: "LAST"}, vars, lastResult, `LastCallResult {{name}}`, false, ""},

		// --- Concatenation Tests using BinaryOpNode ---
		{"Concat Lit(raw) + Var(raw)", BinaryOpNode{Left: StringLiteralNode{Value: "A={{name}} "}, Operator: "+", Right: VariableNode{Name: "greeting"}}, vars, lastResult, "A={{name}} Hello {{name}}", false, ""},
		{"Concat Var(raw) + Lit(raw)", BinaryOpNode{Left: VariableNode{Name: "greeting"}, Operator: "+", Right: StringLiteralNode{Value: " B={{name}}"}}, vars, lastResult, "Hello {{name}} B={{name}}", false, ""},
		{"Concat Var(raw) + Var(raw)", BinaryOpNode{Left: VariableNode{Name: "greeting"}, Operator: "+", Right: VariableNode{Name: "name"}}, vars, lastResult, "Hello {{name}}World", false, ""},
		{"Concat with Number", BinaryOpNode{Left: StringLiteralNode{Value: "Count: "}, Operator: "+", Right: VariableNode{Name: "numVar"}}, vars, lastResult, "Count: 123", false, ""},
		{"Concat Eval + StringLit(Raw)", BinaryOpNode{Left: EvalNode{Argument: VariableNode{Name: "greeting"}}, Operator: "+", Right: StringLiteralNode{Value: " end {{name}}"}}, vars, lastResult, "Hello World end {{name}}", false, ""},
		{"Concat Error Operand", BinaryOpNode{Left: StringLiteralNode{Value: "Val: "}, Operator: "+", Right: VariableNode{Name: "missing"}}, vars, lastResult, nil, true, "variable 'missing' not found"},
		{"Concat Nil Operand", BinaryOpNode{Left: BinaryOpNode{Left: StringLiteralNode{Value: "Start:"}, Operator: "+", Right: VariableNode{Name: "nilVar"}}, Operator: "+", Right: StringLiteralNode{Value: ":End {{name}}"}}, vars, lastResult, "Start::End {{name}}", false, ""},

		// --- Arithmetic Tests using BinaryOpNode ---
		{"Add Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(5)}, Operator: "+", Right: NumberLiteralNode{Value: int64(3)}}, vars, lastResult, int64(8), false, ""},
		{"Add Num + NumStr", BinaryOpNode{Left: VariableNode{Name: "numVar"}, Operator: "+", Right: VariableNode{Name: "numStr"}}, vars, lastResult, int64(579), false, ""},
		{"Subtract Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(5)}, Operator: "-", Right: NumberLiteralNode{Value: int64(3)}}, vars, lastResult, int64(2), false, ""},
		{"Multiply Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(5)}, Operator: "*", Right: NumberLiteralNode{Value: int64(3)}}, vars, lastResult, int64(15), false, ""},
		{"Divide Numbers (Int)", BinaryOpNode{Left: NumberLiteralNode{Value: int64(6)}, Operator: "/", Right: NumberLiteralNode{Value: int64(3)}}, vars, lastResult, int64(2), false, ""},
		{"Divide Numbers (Float)", BinaryOpNode{Left: NumberLiteralNode{Value: int64(7)}, Operator: "/", Right: NumberLiteralNode{Value: int64(2)}}, vars, lastResult, float64(3.5), false, ""},
		{"Modulo Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(7)}, Operator: "%", Right: NumberLiteralNode{Value: int64(2)}}, vars, lastResult, int64(1), false, ""},
		{"Power Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "**", Right: NumberLiteralNode{Value: int64(3)}}, vars, lastResult, float64(8), false, ""},

		// --- Lists, Maps -> Expect RAW values inside ---
		{"Simple List (Raw)", ListLiteralNode{Elements: []interface{}{NumberLiteralNode{Value: int64(1)}, StringLiteralNode{Value: "{{name}}"}, VariableNode{Name: "boolProp"}}}, vars, lastResult, []interface{}{int64(1), "{{name}}", true}, false, ""},
		{"Simple Map (Raw)", MapLiteralNode{Entries: []MapEntryNode{{Key: StringLiteralNode{Value: "k1"}, Value: StringLiteralNode{Value: "{{name}}"}}}}, vars, lastResult, map[string]interface{}{"k1": "{{name}}"}, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runEvalExpressionTest(t, tt) // Use the helper
		})
	}
}
