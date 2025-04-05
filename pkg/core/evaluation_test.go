// pkg/core/evaluation_test.go
package core

import (
	"reflect"
	"strings"
	"testing"
)

// func newTestInterpreter( defined in test_helpers_test.go

// --- Tests for EVAL Node Resolution (Iterative) ---
func TestEvalNodeResolution(t *testing.T) {
	// ... (content remains the same as previous version) ...
}

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

	tests := []struct {
		name        string
		inputNode   interface{}
		expected    interface{}
		wantErr     bool
		errContains string
	}{
		// Literals, Variables, Placeholders, LastNode -> Expect RAW
		{"String Literal (Raw)", StringLiteralNode{Value: "Hello {{name}}"}, `Hello {{name}}`, false, ""},
		{"Variable String (Raw)", VariableNode{Name: "greeting"}, `Hello {{name}}`, false, ""},
		{"Last Call Result (Raw)", LastNode{}, `LastCallResult {{name}}`, false, ""},
		{"Placeholder to String (Raw)", PlaceholderNode{Name: "greeting"}, `Hello {{name}}`, false, ""},

		// --- Concatenation Tests using BinaryOpNode ---
		{"Concat Lit(raw) + Var(raw)", BinaryOpNode{Left: StringLiteralNode{Value: "A={{name}} "}, Operator: "+", Right: VariableNode{Name: "greeting"}}, "A={{name}} Hello {{name}}", false, ""},
		{"Concat Var(raw) + Lit(raw)", BinaryOpNode{Left: VariableNode{Name: "greeting"}, Operator: "+", Right: StringLiteralNode{Value: " B={{name}}"}}, "Hello {{name}} B={{name}}", false, ""},
		{"Concat Var(raw) + Var(raw)", BinaryOpNode{Left: VariableNode{Name: "greeting"}, Operator: "+", Right: VariableNode{Name: "name"}}, "Hello {{name}}World", false, ""},
		{"Concat with Number", BinaryOpNode{Left: StringLiteralNode{Value: "Count: "}, Operator: "+", Right: VariableNode{Name: "numVar"}}, "Count: 123", false, ""},
		{"Concat Eval + StringLit(Raw)", BinaryOpNode{Left: EvalNode{Argument: VariableNode{Name: "greeting"}}, Operator: "+", Right: StringLiteralNode{Value: " end {{name}}"}}, "Hello World end {{name}}", false, ""},
		{"Concat Error Operand", BinaryOpNode{Left: StringLiteralNode{Value: "Val: "}, Operator: "+", Right: VariableNode{Name: "missing"}}, nil, true, "variable 'missing' not found"},
		// Nested BinaryOpNode for "Start:" + nilVar + ":End {{name}}" - Expect nil to become ""
		{"Concat Nil Operand", BinaryOpNode{Left: BinaryOpNode{Left: StringLiteralNode{Value: "Start:"}, Operator: "+", Right: VariableNode{Name: "nilVar"}}, Operator: "+", Right: StringLiteralNode{Value: ":End {{name}}"}}, "Start::End {{name}}", false, ""}, // Corrected Expectation

		// --- Arithmetic Tests using BinaryOpNode ---
		{"Add Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(5)}, Operator: "+", Right: NumberLiteralNode{Value: int64(3)}}, int64(8), false, ""}, // Expect int64
		// *** CORRECTED EXPECTATION: Expect int64 as both inputs are int-like ***
		{"Add Num + NumStr", BinaryOpNode{Left: VariableNode{Name: "numVar"}, Operator: "+", Right: VariableNode{Name: "numStr"}}, int64(579), false, ""},                     // 123 + "456" -> int64(579)
		{"Subtract Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(5)}, Operator: "-", Right: NumberLiteralNode{Value: int64(3)}}, int64(2), false, ""},           // Expect int64
		{"Multiply Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(5)}, Operator: "*", Right: NumberLiteralNode{Value: int64(3)}}, int64(15), false, ""},          // Expect int64
		{"Divide Numbers (Int)", BinaryOpNode{Left: NumberLiteralNode{Value: int64(6)}, Operator: "/", Right: NumberLiteralNode{Value: int64(3)}}, int64(2), false, ""},       // Expect int64 (exact)
		{"Divide Numbers (Float)", BinaryOpNode{Left: NumberLiteralNode{Value: int64(7)}, Operator: "/", Right: NumberLiteralNode{Value: int64(2)}}, float64(3.5), false, ""}, // Expect float (inexact)
		{"Modulo Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(7)}, Operator: "%", Right: NumberLiteralNode{Value: int64(2)}}, int64(1), false, ""},             // Expect int64
		{"Power Numbers", BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "**", Right: NumberLiteralNode{Value: int64(3)}}, float64(8), false, ""},           // Power results in float

		// --- Lists, Maps -> Expect RAW values inside ---
		{"Simple List (Raw)", ListLiteralNode{Elements: []interface{}{NumberLiteralNode{Value: int64(1)}, StringLiteralNode{Value: "{{name}}"}, VariableNode{Name: "boolProp"}}}, []interface{}{int64(1), "{{name}}", true}, false, ""},
		{"Simple Map (Raw)", MapLiteralNode{Entries: []MapEntryNode{{Key: StringLiteralNode{Value: "k1"}, Value: StringLiteralNode{Value: "{{name}}"}}}}, map[string]interface{}{"k1": "{{name}}"}, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interpTest := newTestInterpreter(vars, lastResult) // Use shared helper
			got, err := interpTest.evaluateExpression(tt.inputNode)
			// Assertions
			if (err != nil) != tt.wantErr {
				t.Errorf("TestEvalExpr(%s): error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Errorf("TestEvalExpr(%s): Err substr mismatch. Got: %v, Want: %q", tt.name, err, tt.errContains)
				}
			} else {
				if !reflect.DeepEqual(got, tt.expected) {
					// Add more detail for debugging mismatches
					t.Errorf("TestEvalExpr(%s)\nInput Node: %+v\nExp Value:  %v\nExp Type:   %T\nGot Value:  %v\nGot Type:   %T",
						tt.name, tt.inputNode, tt.expected, tt.expected, got, got)
				}
			}
		})
	}
}
