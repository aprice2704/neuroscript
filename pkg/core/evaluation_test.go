// filename: pkg/core/evaluation_test.go
package core

import (
	"testing"
	// Import errors package if sentinel errors are used directly here
	// "errors"
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

	// Use EvalTestCase struct (updated with ExpectedErrorIs)
	tests := []EvalTestCase{
		// Literals, Variables, Placeholders, LastNode -> Expect RAW
		// *** Use nil for ExpectedErrorIs when WantErr is false ***
		{"String Literal (Raw)", StringLiteralNode{Value: "Hello {{name}}"}, vars, lastResult, `Hello {{name}}`, false, nil},
		{"Variable String (Raw)", VariableNode{Name: "greeting"}, vars, lastResult, `Hello {{name}}`, false, nil},
		{"Last Call Result (Raw)", LastNode{}, vars, lastResult, `LastCallResult {{name}}`, false, nil},
		{"Placeholder to String (Raw Ref Value)", PlaceholderNode{Name: "greeting"}, vars, lastResult, `Hello {{name}}`, false, nil},
		{"Placeholder LAST (Raw Ref Value)", PlaceholderNode{Name: "LAST"}, vars, lastResult, `LastCallResult {{name}}`, false, nil},

		// --- Concatenation Tests using BinaryOpNode ---
		{"Concat Lit(raw) + Var(raw)", BinaryOpNode{Left: StringLiteralNode{Value: "A={{name}} "}, Operator: "+", Right: VariableNode{Name: "greeting"}}, vars, lastResult, "A={{name}} Hello {{name}}", false, nil},
		{"Concat Var(raw) + Lit(raw)", BinaryOpNode{Left: VariableNode{Name: "greeting"}, Operator: "+", Right: StringLiteralNode{Value: " B={{name}}"}}, vars, lastResult, "Hello {{name}} B={{name}}", false, nil},
		{"Concat Var(raw) + Var(raw)", BinaryOpNode{Left: VariableNode{Name: "greeting"}, Operator: "+", Right: VariableNode{Name: "name"}}, vars, lastResult, "Hello {{name}}World", false, nil},
		// *** FIX: Concat with Number should now ERROR ***
		{"Concat with Number", BinaryOpNode{Left: StringLiteralNode{Value: "Count: "}, Operator: "+", Right: VariableNode{Name: "numVar"}}, vars, lastResult, nil, true, ErrInvalidOperandType}, // Expect specific error
		// *** END FIX ***
		{"Concat Eval + StringLit(Raw)", BinaryOpNode{Left: EvalNode{Argument: VariableNode{Name: "greeting"}}, Operator: "+", Right: StringLiteralNode{Value: " end {{name}}"}}, vars, lastResult, "Hello World end {{name}}", false, nil},
		// Concat Error Operand: Use ExpectedErrorIs
		{"Concat Error Operand", BinaryOpNode{Left: StringLiteralNode{Value: "Val: "}, Operator: "+", Right: VariableNode{Name: "missing"}}, vars, lastResult, nil, true, ErrVariableNotFound},                                                                                       // Use sentinel error
		{"Concat Nil Operand", BinaryOpNode{Left: BinaryOpNode{Left: StringLiteralNode{Value: "Start:"}, Operator: "+", Right: VariableNode{Name: "nilVar"}}, Operator: "+", Right: StringLiteralNode{Value: ":End {{name}}"}}, vars, lastResult, "Start::End {{name}}", false, nil}, // Concatenating nil results in empty string representation

		// --- Arithmetic Tests using BinaryOpNode ---
		{"Add Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(5)}, Operator: "+", Right: NumberLiteralNode{Value: int64(3)}}, vars, lastResult, int64(8), false, nil},
		{"Add Num + NumStr", BinaryOpNode{Left: VariableNode{Name: "numVar"}, Operator: "+", Right: VariableNode{Name: "numStr"}}, vars, lastResult, int64(579), false, nil},
		{"Subtract Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(5)}, Operator: "-", Right: NumberLiteralNode{Value: int64(3)}}, vars, lastResult, int64(2), false, nil},
		{"Multiply Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(5)}, Operator: "*", Right: NumberLiteralNode{Value: int64(3)}}, vars, lastResult, int64(15), false, nil},
		{"Divide Numbers (Int)", BinaryOpNode{Left: NumberLiteralNode{Value: int64(6)}, Operator: "/", Right: NumberLiteralNode{Value: int64(3)}}, vars, lastResult, int64(2), false, nil},
		{"Divide Numbers (Float)", BinaryOpNode{Left: NumberLiteralNode{Value: int64(7)}, Operator: "/", Right: NumberLiteralNode{Value: int64(2)}}, vars, lastResult, float64(3.5), false, nil},
		{"Modulo Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(7)}, Operator: "%", Right: NumberLiteralNode{Value: int64(2)}}, vars, lastResult, int64(1), false, nil},
		{"Power Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "**", Right: NumberLiteralNode{Value: int64(3)}}, vars, lastResult, float64(8), false, nil}, // Power results in float

		// --- Lists, Maps -> Expect RAW values inside ---
		{"Simple List (Raw)", ListLiteralNode{Elements: []interface{}{NumberLiteralNode{Value: int64(1)}, StringLiteralNode{Value: "{{name}}"}, VariableNode{Name: "boolProp"}}}, vars, lastResult, []interface{}{int64(1), "{{name}}", true}, false, nil},
		{"Simple Map (Raw)", MapLiteralNode{Entries: []MapEntryNode{{Key: StringLiteralNode{Value: "k1"}, Value: StringLiteralNode{Value: "{{name}}"}}}}, vars, lastResult, map[string]interface{}{"k1": "{{name}}"}, false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Assuming runEvalExpressionTest is defined in testing_helpers_test.go
			// and handles the updated EvalTestCase struct correctly
			runEvalExpressionTest(t, tt) // Use the helper
		})
	}
}
