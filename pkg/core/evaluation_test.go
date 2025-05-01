// NeuroScript Version: 0.3.5
// Last Modified: 2025-05-01 15:20:15 PDT
// filename: pkg/core/evaluation_test.go
package core

import (
	"testing"
	// Assuming EvalTestCase and runEvalExpressionTest are defined in testing_helpers_test.go
	// Assuming AST node types (StringLiteralNode, BinaryOpNode, etc.) and Position are defined in ast.go
	// Assuming error variables (ErrInvalidOperandType, ErrVariableNotFound) are defined in errors.go
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
		{Name: "String Literal (Raw)", InputNode: &StringLiteralNode{Pos: dummyPos, Value: "Hello {{name}}"}, InitialVars: vars, LastResult: lastResult, Expected: `Hello {{name}}`, WantErr: false, ExpectedErrorIs: nil},
		{Name: "Variable String (Raw)", InputNode: &VariableNode{Pos: dummyPos, Name: "greeting"}, InitialVars: vars, LastResult: lastResult, Expected: `Hello {{name}}`, WantErr: false, ExpectedErrorIs: nil},
		{Name: "Last Call Result (Raw)", InputNode: &LastNode{Pos: dummyPos}, InitialVars: vars, LastResult: lastResult, Expected: `LastCallResult {{name}}`, WantErr: false, ExpectedErrorIs: nil},
		{Name: "Placeholder to String (Raw Ref Value)", InputNode: &PlaceholderNode{Pos: dummyPos, Name: "greeting"}, InitialVars: vars, LastResult: lastResult, Expected: `Hello {{name}}`, WantErr: false, ExpectedErrorIs: nil},
		{Name: "Placeholder LAST (Raw Ref Value)", InputNode: &PlaceholderNode{Pos: dummyPos, Name: "LAST"}, InitialVars: vars, LastResult: lastResult, Expected: `LastCallResult {{name}}`, WantErr: false, ExpectedErrorIs: nil},

		// --- Concatenation Tests using BinaryOpNode ---
		{Name: "Concat Lit(raw) + Var(raw)", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "A={{name}} "}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "greeting"}}, InitialVars: vars, LastResult: lastResult, Expected: "A={{name}} Hello {{name}}", WantErr: false, ExpectedErrorIs: nil},
		{Name: "Concat Var(raw) + Lit(raw)", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "greeting"}, Operator: "+", Right: &StringLiteralNode{Pos: dummyPos, Value: " B={{name}}"}}, InitialVars: vars, LastResult: lastResult, Expected: "Hello {{name}} B={{name}}", WantErr: false, ExpectedErrorIs: nil},
		{Name: "Concat Var(raw) + Var(raw)", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "greeting"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "name"}}, InitialVars: vars, LastResult: lastResult, Expected: "Hello {{name}}World", WantErr: false, ExpectedErrorIs: nil},
		// --- MODIFIED: Expect successful string concatenation ---
		{Name: "Concat with Number", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "Count: "}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "numVar"}}, InitialVars: vars, LastResult: lastResult, Expected: "Count: 123", WantErr: false, ExpectedErrorIs: nil},
		// --- END MODIFICATION ---
		{Name: "Concat Eval + StringLit(Raw)", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &EvalNode{Pos: dummyPos, Argument: &VariableNode{Pos: dummyPos, Name: "greeting"}}, Operator: "+", Right: &StringLiteralNode{Pos: dummyPos, Value: " end {{name}}"}}, InitialVars: vars, LastResult: lastResult, Expected: "Hello World end {{name}}", WantErr: false, ExpectedErrorIs: nil},
		{Name: "Concat Error Operand", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "Val: "}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "missing"}}, InitialVars: vars, LastResult: lastResult, Expected: nil, WantErr: true, ExpectedErrorIs: ErrVariableNotFound},
		{Name: "Concat Nil Operand", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "Start:"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, Operator: "+", Right: &StringLiteralNode{Pos: dummyPos, Value: ":End {{name}}"}}, InitialVars: vars, LastResult: lastResult, Expected: "Start::End {{name}}", WantErr: false, ExpectedErrorIs: nil}, // Concatenating nil results in empty string representation

		// --- Arithmetic Tests using BinaryOpNode ---
		{Name: "Add Numbers", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(5)}, Operator: "+", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, LastResult: lastResult, Expected: int64(8), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Add Num + NumStr", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "numVar"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "numStr"}}, InitialVars: vars, LastResult: lastResult, Expected: int64(579), WantErr: false, ExpectedErrorIs: nil}, // Arithmetic preferred if both look like numbers
		{Name: "Subtract Numbers", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(5)}, Operator: "-", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, LastResult: lastResult, Expected: int64(2), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Multiply Numbers", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(5)}, Operator: "*", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, LastResult: lastResult, Expected: int64(15), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Divide Numbers (Int)", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(6)}, Operator: "/", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, LastResult: lastResult, Expected: int64(2), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Divide Numbers (Float)", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(7)}, Operator: "/", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}}, InitialVars: vars, LastResult: lastResult, Expected: float64(3.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Modulo Numbers", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(7)}, Operator: "%", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}}, InitialVars: vars, LastResult: lastResult, Expected: int64(1), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Power Numbers", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}, Operator: "**", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, LastResult: lastResult, Expected: float64(8), WantErr: false, ExpectedErrorIs: nil},

		// --- Lists, Maps -> Expect RAW values inside ---
		{Name: "Simple List (Raw)", InputNode: &ListLiteralNode{Pos: dummyPos, Elements: []Expression{&NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, &StringLiteralNode{Pos: dummyPos, Value: "{{name}}"}, &VariableNode{Pos: dummyPos, Name: "boolProp"}}}, InitialVars: vars, LastResult: lastResult, Expected: []interface{}{int64(1), "{{name}}", true}, WantErr: false, ExpectedErrorIs: nil},
		// *** CORRECTED: MapEntryNode Key needs VALUE, Value needs POINTER (Expression) ***
		{Name: "Simple Map (Raw)", InputNode: &MapLiteralNode{Pos: dummyPos, Entries: []MapEntryNode{{Pos: dummyPos, Key: StringLiteralNode{Pos: dummyPos, Value: "k1"}, Value: &StringLiteralNode{Pos: dummyPos, Value: "{{name}}"}}}}, InitialVars: vars, LastResult: lastResult, Expected: map[string]interface{}{"k1": "{{name}}"}, WantErr: false, ExpectedErrorIs: nil},
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
