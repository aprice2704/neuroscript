// pkg/core/evaluation_comparison_test.go
package core

import (
	// Required for DeepEqual
	"strings"
	"testing"
)

// --- Test Helper Function (Assume defined elsewhere or copy here) ---
// func newTestInterpreterEval(vars map[string]interface{}, lastResult interface{}) *Interpreter { ... }

// --- TestEvaluateCondition (Updated error expectations) ---
func TestEvaluateCondition(t *testing.T) {
	vars := map[string]interface{}{
		"trueVar":   true,
		"falseVar":  false,
		"numOne":    int64(1),
		"numZero":   int64(0),
		"floatOne":  float64(1.0),
		"floatZero": float64(0.0),
		"strTrue":   "true",
		"strFalse":  "false",
		"strOther":  "hello",
		"strNum10":  "10",
		"x":         "A",
		"y":         "A",
		"z":         "B",
		"n1":        int64(10),
		"n2":        int64(5),
		"n3":        int64(10),
		"nilVar":    nil,
	}
	interp := newTestInterpreterEval(vars, nil)

	tests := []struct {
		name        string
		node        interface{} // Input is AST node
		expected    bool        // Expected bool result if no error
		wantErr     bool        // Expect an error during evaluation
		errContains string      // Substring expected in the error message
	}{
		// Single Expression Conditions
		{"Bool Literal True", BooleanLiteralNode{Value: true}, true, false, ""},
		{"Bool Literal False", BooleanLiteralNode{Value: false}, false, false, ""},
		{"Var Bool True", VariableNode{Name: "trueVar"}, true, false, ""},
		{"Var Bool False", VariableNode{Name: "falseVar"}, false, false, ""},
		{"Var Num NonZero", VariableNode{Name: "numOne"}, true, false, ""},
		{"Var Num Zero", VariableNode{Name: "numZero"}, false, false, ""},
		{"Var Float NonZero", VariableNode{Name: "floatOne"}, true, false, ""},
		{"Var Float Zero", VariableNode{Name: "floatZero"}, false, false, ""},
		{"Var String True", VariableNode{Name: "strTrue"}, true, false, ""},
		{"Var String False", VariableNode{Name: "strFalse"}, false, false, ""},
		{"Var String Other", VariableNode{Name: "strOther"}, false, false, ""},
		{"String Literal True", StringLiteralNode{Value: "true"}, true, false, ""},
		{"String Literal False", StringLiteralNode{Value: "false"}, false, false, ""},
		{"String Literal Other", StringLiteralNode{Value: "yes"}, false, false, ""},
		{"Number Literal NonZero", NumberLiteralNode{Value: int64(1)}, true, false, ""},
		{"Number Literal Zero", NumberLiteralNode{Value: int64(0)}, false, false, ""},
		{"Number Literal Float NonZero", NumberLiteralNode{Value: float64(0.1)}, true, false, ""},
		{"Number Literal Float Zero", NumberLiteralNode{Value: float64(0.0)}, false, false, ""},
		{"Variable Not Found Condition", VariableNode{Name: "not_found"}, false, false, ""}, // Handled gracefully -> false
		{"List Literal Condition", ListLiteralNode{Elements: []interface{}{}}, false, false, ""},
		{"Map Literal Condition", MapLiteralNode{Entries: []MapEntryNode{}}, false, false, ""},
		{"Nil Variable Condition", VariableNode{Name: "nilVar"}, false, false, ""},
		{"Placeholder Not Found Condition", PlaceholderNode{Name: "missing"}, false, false, ""}, // *** UPDATED: Handled gracefully -> false ***

		// Comparison Node Conditions
		{"Comp EQ String True", ComparisonNode{Left: VariableNode{Name: "x"}, Operator: "==", Right: VariableNode{Name: "y"}}, true, false, ""},
		{"Comp EQ String False", ComparisonNode{Left: VariableNode{Name: "x"}, Operator: "==", Right: VariableNode{Name: "z"}}, false, false, ""},
		{"Comp EQ Num True", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: "==", Right: VariableNode{Name: "n3"}}, true, false, ""},
		{"Comp EQ Num False", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: "==", Right: VariableNode{Name: "n2"}}, false, false, ""},
		{"Comp EQ Mixed Type True (Str vs Num)", ComparisonNode{Left: NumberLiteralNode{Value: int64(1)}, Operator: "==", Right: StringLiteralNode{Value: "1"}}, true, false, ""},
		{"Comp NEQ String True", ComparisonNode{Left: VariableNode{Name: "x"}, Operator: "!=", Right: VariableNode{Name: "z"}}, true, false, ""},
		{"Comp NEQ String False", ComparisonNode{Left: VariableNode{Name: "x"}, Operator: "!=", Right: VariableNode{Name: "y"}}, false, false, ""},
		{"Comp NEQ Num True", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: "!=", Right: VariableNode{Name: "n2"}}, true, false, ""},
		{"Comp NEQ Num False", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: "!=", Right: VariableNode{Name: "n3"}}, false, false, ""},
		{"Comp GT True", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: ">", Right: VariableNode{Name: "n2"}}, true, false, ""},
		{"Comp GT False", ComparisonNode{Left: VariableNode{Name: "n2"}, Operator: ">", Right: VariableNode{Name: "n1"}}, false, false, ""},
		{"Comp GT Equal", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: ">", Right: VariableNode{Name: "n3"}}, false, false, ""},
		{"Comp LT True", ComparisonNode{Left: VariableNode{Name: "n2"}, Operator: "<", Right: VariableNode{Name: "n1"}}, true, false, ""},
		{"Comp LT False", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: "<", Right: VariableNode{Name: "n2"}}, false, false, ""},
		{"Comp LT Equal", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: "<", Right: VariableNode{Name: "n3"}}, false, false, ""},
		{"Comp GTE True", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: ">=", Right: VariableNode{Name: "n2"}}, true, false, ""},
		{"Comp GTE Equal", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: ">=", Right: VariableNode{Name: "n3"}}, true, false, ""},
		{"Comp GTE False", ComparisonNode{Left: VariableNode{Name: "n2"}, Operator: ">=", Right: VariableNode{Name: "n1"}}, false, false, ""},
		{"Comp LTE True", ComparisonNode{Left: VariableNode{Name: "n2"}, Operator: "<=", Right: VariableNode{Name: "n1"}}, true, false, ""},
		{"Comp LTE Equal", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: "<=", Right: VariableNode{Name: "n3"}}, true, false, ""},
		{"Comp LTE False", ComparisonNode{Left: VariableNode{Name: "n1"}, Operator: "<=", Right: VariableNode{Name: "n2"}}, false, false, ""},
		{"Comp Numeric Error Types", ComparisonNode{Left: VariableNode{Name: "x"}, Operator: ">", Right: VariableNode{Name: "n1"}}, false, true, "requires numeric operands"},
		{"Comp Numeric Error String Lit", ComparisonNode{Left: StringLiteralNode{Value: "a"}, Operator: "<", Right: StringLiteralNode{Value: "b"}}, false, true, "requires numeric operands"},
		{"Comp Error Evaluating LHS Placeholder", ComparisonNode{Left: PlaceholderNode{Name: "missing"}, Operator: "==", Right: VariableNode{Name: "x"}}, false, false, ""}, // *** UPDATED: Handled gracefully -> nil == "A" -> false ***
		{"Comp Error Evaluating RHS Placeholder", ComparisonNode{Left: VariableNode{Name: "x"}, Operator: "==", Right: PlaceholderNode{Name: "missing"}}, false, false, ""}, // *** UPDATED: Handled gracefully -> "A" == nil -> false ***
		{"Comp String Num vs Num", ComparisonNode{Left: VariableNode{Name: "strNum10"}, Operator: "==", Right: VariableNode{Name: "n1"}}, true, false, ""},
		{"Comp String Num vs Num GT", ComparisonNode{Left: VariableNode{Name: "strNum10"}, Operator: ">", Right: VariableNode{Name: "n2"}}, true, false, ""},
		{"Comp String Num vs String Num LT", ComparisonNode{Left: VariableNode{Name: "strNum10"}, Operator: "<", Right: StringLiteralNode{Value: "5"}}, false, false, ""},

		// Nil Comparisons
		{"Comp EQ Nil vs Nil", ComparisonNode{Left: VariableNode{Name: "nilVar"}, Operator: "==", Right: VariableNode{Name: "nilVar"}}, true, false, ""},
		{"Comp EQ Nil vs String", ComparisonNode{Left: VariableNode{Name: "nilVar"}, Operator: "==", Right: StringLiteralNode{Value: "A"}}, false, false, ""},
		{"Comp EQ String vs Nil", ComparisonNode{Left: StringLiteralNode{Value: "A"}, Operator: "==", Right: VariableNode{Name: "nilVar"}}, false, false, ""},
		{"Comp NEQ Nil vs Nil", ComparisonNode{Left: VariableNode{Name: "nilVar"}, Operator: "!=", Right: VariableNode{Name: "nilVar"}}, false, false, ""},
		{"Comp NEQ Nil vs String", ComparisonNode{Left: VariableNode{Name: "nilVar"}, Operator: "!=", Right: StringLiteralNode{Value: "A"}}, true, false, ""},
		{"Comp NEQ String vs Nil", ComparisonNode{Left: StringLiteralNode{Value: "A"}, Operator: "!=", Right: VariableNode{Name: "nilVar"}}, true, false, ""},
		{"Comp GT Nil vs Num", ComparisonNode{Left: VariableNode{Name: "nilVar"}, Operator: ">", Right: NumberLiteralNode{Value: int64(5)}}, false, true, "requires non-nil numeric operands"},
		{"Comp LT Num vs Nil", ComparisonNode{Left: NumberLiteralNode{Value: int64(5)}, Operator: "<", Right: VariableNode{Name: "nilVar"}}, false, true, "requires non-nil numeric operands"},
		{"Comp Var Not Found vs String EQ", ComparisonNode{Left: VariableNode{Name: "not_found"}, Operator: "==", Right: StringLiteralNode{Value: "A"}}, false, false, ""},
		{"Comp Var Not Found vs Nil EQ", ComparisonNode{Left: VariableNode{Name: "not_found"}, Operator: "==", Right: VariableNode{Name: "nilVar"}}, true, false, ""},
		{"Comp Var Not Found vs Var Not Found EQ", ComparisonNode{Left: VariableNode{Name: "not_found1"}, Operator: "==", Right: VariableNode{Name: "not_found2"}}, true, false, ""},
		{"Comp Var Not Found vs Var Not Found NEQ", ComparisonNode{Left: VariableNode{Name: "not_found1"}, Operator: "!=", Right: VariableNode{Name: "not_found2"}}, false, false, ""},
		{"Comp Var Not Found vs Num GT", ComparisonNode{Left: VariableNode{Name: "not_found"}, Operator: ">", Right: NumberLiteralNode{Value: int64(5)}}, false, true, "requires non-nil numeric operands"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp.variables = make(map[string]interface{}, len(vars)) // Reset vars
			for k, v := range vars {
				interp.variables[k] = v
			}

			got, err := interp.evaluateCondition(tt.node)

			if (err != nil) != tt.wantErr {
				// Provide more context on failure
				t.Errorf("evaluateCondition(%s): Error expectation mismatch. got err = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Errorf("evaluateCondition(%s): Expected error containing %q, got: %v", tt.name, tt.errContains, err)
				}
			} else {
				if got != tt.expected {
					// Use %+v for potentially complex node structures
					t.Errorf("evaluateCondition(%s)\nNode:       %+v\nGot bool:   %v\nWant bool:  %v", tt.name, tt.node, got, tt.expected)
				}
			}
		})
	}
}
