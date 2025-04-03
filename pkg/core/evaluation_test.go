// pkg/core/evaluation_test.go
package core

import (
	"reflect"
	"strings"
	"testing"
)

// func newTestInterpreterEval defined in test_helpers_test.go

// --- Tests for EVAL Node Resolution (Iterative) ---
func TestEvalNodeResolution(t *testing.T) {
	vars := map[string]interface{}{
		"name":  "World",
		"obj":   "Test {{sub}}", // Var holding placeholder
		"sub":   "Subject",
		"greet": "Hello {{name}}", // Var holding placeholder
		"depth": "{{d1}}", "d1": "{{d2}}", "d2": "{{d3}}", "d3": "{{d4}}", "d4": "{{d5}}",
		"d5": "{{d6}}", "d6": "{{d7}}", "d7": "{{d8}}", "d8": "{{d9}}", "d9": "{{d10}}",
		"d10":    "{{d11}}", // d11 does not exist
		"cycle1": "{{cycle2}}", "cycle2": "{{cycle1}}",
		"count": int64(5),
	}
	lastResultValue := "LAST_RESULT_VALUE {{name}}" // LAST value holding placeholder

	tests := []struct {
		name        string
		evalArgNode interface{} // The argument node inside EVAL(...)
		expected    interface{} // Expected *resolved* result (string or nil on error)
		wantErr     bool
		errContains string
	}{
		{"EVAL String Literal", StringLiteralNode{Value: "Hello {{name}}"}, "Hello World", false, ""},
		{"EVAL Var (Nested)", VariableNode{Name: "obj"}, "Test Subject", false, ""},    // Iterative resolution handles this
		{"EVAL Var (Resolved)", VariableNode{Name: "greet"}, "Hello World", false, ""}, // Iterative resolution handles this
		{"EVAL LAST", LastNode{}, "LAST_RESULT_VALUE World", false, ""},                // Iterative resolution handles this
		{"EVAL Placeholder", PlaceholderNode{Name: "obj"}, "Test Subject", false, ""},
		{"EVAL String w/ Not Found", StringLiteralNode{Value: "Hi {{user}}"}, nil, true, "placeholder variable '{{user}}' not found"},  // Error during resolve
		{"EVAL Var Not Found", VariableNode{Name: "missing"}, nil, true, "evaluating argument for EVAL: variable 'missing' not found"}, // Error during arg eval
		// *** CORRECTED: Expect iteration error ***
		{"EVAL Deep Recursion", VariableNode{Name: "depth"}, nil, true, "exceeded max iterations"},
		{"EVAL Cycle", VariableNode{Name: "cycle1"}, nil, true, "exceeded max iterations"},
		{"EVAL Non-String Var", VariableNode{Name: "count"}, "5", false, ""}, // EVAL(5) -> "5"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newTestInterpreterEval(vars, lastResultValue) // Use shared helper
			evalNode := EvalNode{Argument: tt.evalArgNode}
			got, err := interp.evaluateExpression(evalNode)

			if (err != nil) != tt.wantErr {
				t.Errorf("EVAL Test(%s): error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Errorf("EVAL Test(%s): expected error containing %q, got: %v", tt.name, tt.errContains, err)
				}
			} else {
				if !reflect.DeepEqual(got, tt.expected) {
					t.Errorf("EVAL Test(%s):\nExpected: %v (%T)\nGot:      %v (%T)", tt.name, tt.expected, tt.expected, got, got)
				}
			}
		})
	}
}

// --- Tests for General Expression Evaluation (Raw strings by default) ---
func TestEvaluateExpressionASTGeneral(t *testing.T) {
	vars := map[string]interface{}{
		"name":     "World",
		"greeting": "Hello {{name}}", // Contains placeholder
		"numVar":   int64(123),
		"boolProp": true,
		"listVar":  []interface{}{"x", int64(99), "{{name}}"},
		"mapVar":   map[string]interface{}{"mKey": "mVal {{name}}", "mNum": int64(1)},
		"nilVar":   nil,
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

		// Concatenation -> Expect RAW concatenation
		// *** CORRECTED EXPECTATIONS TO MATCH STRICTLY RAW CONCAT ***
		{"Concat Lit(raw) + Var(raw)", ConcatenationNode{Operands: []interface{}{StringLiteralNode{Value: "A={{name}} "}, VariableNode{Name: "greeting"}}}, "A={{name}} Hello {{name}}", false, ""},
		{"Concat Var(raw) + Lit(raw)", ConcatenationNode{Operands: []interface{}{VariableNode{Name: "greeting"}, StringLiteralNode{Value: " B={{name}}"}}}, "Hello {{name}} B={{name}}", false, ""},
		{"Concat Var(raw) + Var(raw)", ConcatenationNode{Operands: []interface{}{VariableNode{Name: "greeting"}, VariableNode{Name: "name"}}}, "Hello {{name}}World", false, ""},
		{"Concat with Number", ConcatenationNode{Operands: []interface{}{StringLiteralNode{Value: "Count: "}, VariableNode{Name: "numVar"}}}, "Count: 123", false, ""},
		{"Concat Eval + StringLit(Raw)", ConcatenationNode{Operands: []interface{}{EvalNode{Argument: VariableNode{Name: "greeting"}}, StringLiteralNode{Value: " end {{name}}"}}}, "Hello World end {{name}}", false, ""}, // Eval resolves, Lit raw
		{"Concat Error Operand", ConcatenationNode{Operands: []interface{}{StringLiteralNode{Value: "Val: "}, VariableNode{Name: "missing"}}}, nil, true, "variable 'missing' not found"},
		{"Concat Nil Operand", ConcatenationNode{Operands: []interface{}{StringLiteralNode{Value: "Start:"}, VariableNode{Name: "nilVar"}, StringLiteralNode{Value: ":End {{name}}"}}}, "Start::End {{name}}", false, ""}, // nil becomes empty string

		// Lists, Maps -> Expect RAW values inside
		{"Simple List (Raw)", ListLiteralNode{Elements: []interface{}{NumberLiteralNode{Value: int64(1)}, StringLiteralNode{Value: "{{name}}"}, VariableNode{Name: "boolProp"}}}, []interface{}{int64(1), "{{name}}", true}, false, ""},
		{"Simple Map (Raw)", MapLiteralNode{Entries: []MapEntryNode{{Key: StringLiteralNode{Value: "k1"}, Value: StringLiteralNode{Value: "{{name}}"}}}}, map[string]interface{}{"k1": "{{name}}"}, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interpTest := newTestInterpreterEval(vars, lastResult) // Use shared helper
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
					t.Errorf("TestEvalExpr(%s)\nNode: %+v\nExp:  %v (%T)\nGot:  %v (%T)", tt.name, tt.inputNode, tt.expected, tt.expected, got, got)
				}
			}
		})
	}
}
