// filename: pkg/core/evaluation_test.go
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

	// Dummy position for nodes
	dummyPos := &Position{Line: 1, Column: 1}

	// Use EvalTestCase struct (updated with ExpectedErrorIs)
	tests := []EvalTestCase{
		// Literals, Variables, Placeholders, LastNode -> Expect RAW
		// *** CORRECTED: Add & to node literals where Expression is expected ***
		{"String Literal (Raw)", &StringLiteralNode{Pos: dummyPos, Value: "Hello {{name}}"}, vars, lastResult, `Hello {{name}}`, false, nil},
		{"Variable String (Raw)", &VariableNode{Pos: dummyPos, Name: "greeting"}, vars, lastResult, `Hello {{name}}`, false, nil},
		{"Last Call Result (Raw)", &LastNode{Pos: dummyPos}, vars, lastResult, `LastCallResult {{name}}`, false, nil},
		{"Placeholder to String (Raw Ref Value)", &PlaceholderNode{Pos: dummyPos, Name: "greeting"}, vars, lastResult, `Hello {{name}}`, false, nil},
		{"Placeholder LAST (Raw Ref Value)", &PlaceholderNode{Pos: dummyPos, Name: "LAST"}, vars, lastResult, `LastCallResult {{name}}`, false, nil},

		// --- Concatenation Tests using BinaryOpNode ---
		{"Concat Lit(raw) + Var(raw)", &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "A={{name}} "}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "greeting"}}, vars, lastResult, "A={{name}} Hello {{name}}", false, nil},
		{"Concat Var(raw) + Lit(raw)", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "greeting"}, Operator: "+", Right: &StringLiteralNode{Pos: dummyPos, Value: " B={{name}}"}}, vars, lastResult, "Hello {{name}} B={{name}}", false, nil},
		{"Concat Var(raw) + Var(raw)", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "greeting"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "name"}}, vars, lastResult, "Hello {{name}}World", false, nil},
		{"Concat with Number", &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "Count: "}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "numVar"}}, vars, lastResult, nil, true, ErrInvalidOperandType},
		{"Concat Eval + StringLit(Raw)", &BinaryOpNode{Pos: dummyPos, Left: &EvalNode{Pos: dummyPos, Argument: &VariableNode{Pos: dummyPos, Name: "greeting"}}, Operator: "+", Right: &StringLiteralNode{Pos: dummyPos, Value: " end {{name}}"}}, vars, lastResult, "Hello World end {{name}}", false, nil},
		{"Concat Error Operand", &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "Val: "}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "missing"}}, vars, lastResult, nil, true, ErrVariableNotFound},
		{"Concat Nil Operand", &BinaryOpNode{Pos: dummyPos, Left: &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "Start:"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, Operator: "+", Right: &StringLiteralNode{Pos: dummyPos, Value: ":End {{name}}"}}, vars, lastResult, "Start::End {{name}}", false, nil}, // Concatenating nil results in empty string representation

		// --- Arithmetic Tests using BinaryOpNode ---
		{"Add Numbers", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(5)}, Operator: "+", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, vars, lastResult, int64(8), false, nil},
		{"Add Num + NumStr", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "numVar"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "numStr"}}, vars, lastResult, int64(579), false, nil},
		{"Subtract Numbers", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(5)}, Operator: "-", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, vars, lastResult, int64(2), false, nil},
		{"Multiply Numbers", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(5)}, Operator: "*", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, vars, lastResult, int64(15), false, nil},
		{"Divide Numbers (Int)", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(6)}, Operator: "/", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, vars, lastResult, int64(2), false, nil},
		{"Divide Numbers (Float)", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(7)}, Operator: "/", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}}, vars, lastResult, float64(3.5), false, nil},
		{"Modulo Numbers", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(7)}, Operator: "%", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}}, vars, lastResult, int64(1), false, nil},
		{"Power Numbers", &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}, Operator: "**", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, vars, lastResult, float64(8), false, nil},

		// --- Lists, Maps -> Expect RAW values inside ---
		{"Simple List (Raw)", &ListLiteralNode{Pos: dummyPos, Elements: []Expression{&NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, &StringLiteralNode{Pos: dummyPos, Value: "{{name}}"}, &VariableNode{Pos: dummyPos, Name: "boolProp"}}}, vars, lastResult, []interface{}{int64(1), "{{name}}", true}, false, nil},
		// *** CORRECTED: MapEntryNode Key needs VALUE, Value needs POINTER (Expression) ***
		{"Simple Map (Raw)", &MapLiteralNode{Pos: dummyPos, Entries: []MapEntryNode{{Pos: dummyPos, Key: StringLiteralNode{Pos: dummyPos, Value: "k1"}, Value: &StringLiteralNode{Pos: dummyPos, Value: "{{name}}"}}}}, vars, lastResult, map[string]interface{}{"k1": "{{name}}"}, false, nil},
		// *** END CORRECTIONS ***
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Assuming runEvalExpressionTest is defined in testing_helpers_test.go
			// and handles the updated EvalTestCase struct correctly
			// Also ensure runEvalExpressionTest correctly handles InputNode being Expression
			runEvalExpressionTest(t, tt) // Use the helper
		})
	}
}
