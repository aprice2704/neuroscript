// filename: neuroscript/pkg/core/evaluation_comparison_test.go
package core

import (
	"strings"
	"testing"
)

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
		"strOne":    "1",
		"x":         "A",
		"y":         "A",
		"z":         "B",
		"n1":        int64(10),
		"n2":        int64(5),
		"n3":        int64(10),
		"nilVar":    nil,
		"phVar":     "{{inner}}", // Raw value
		"inner":     "resolved",  // Added for placeholder test case resolution within EVAL if needed elsewhere
	}
	lastValue := "true {{inner}}" // Raw value for LAST tests

	tests := []struct {
		name        string
		node        interface{} // Input is AST node
		expected    bool        // Expected bool result if no error
		wantErr     bool        // Expect an error during evaluation
		errContains string      // Substring expected in the error message
	}{
		// Single Expression Conditions (Truthiness)
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
		{"Var String Other", VariableNode{Name: "strOther"}, false, false, ""}, // "hello" is not "true" or "1"
		{"Var String One", VariableNode{Name: "strOne"}, true, false, ""},
		{"String Literal True", StringLiteralNode{Value: "true"}, true, false, ""},
		{"String Literal False", StringLiteralNode{Value: "false"}, false, false, ""},
		{"String Literal One", StringLiteralNode{Value: "1"}, true, false, ""},
		{"String Literal Other", StringLiteralNode{Value: "yes"}, false, false, ""}, // "yes" is not "true" or "1"
		{"Number Literal NonZero", NumberLiteralNode{Value: int64(1)}, true, false, ""},
		{"Number Literal Zero", NumberLiteralNode{Value: int64(0)}, false, false, ""},
		{"Variable Not Found Condition", VariableNode{Name: "not_found"}, false, false, ""},      // Not found evaluates to false in condition
		{"List Literal Condition", ListLiteralNode{Elements: []interface{}{}}, false, false, ""}, // Collections are falsey
		{"Map Literal Condition", MapLiteralNode{Entries: []MapEntryNode{}}, false, false, ""},   // Collections are falsey
		{"Nil Variable Condition", VariableNode{Name: "nilVar"}, false, false, ""},               // nil is falsey
		{"LAST Condition (Truthy String)", LastNode{}, false, false, ""},                         // LAST contains "true {{inner}}", which is not "true" or "1", so false

		// --- Comparison Conditions using BinaryOpNode ---
		{"Comp EQ String True", BinaryOpNode{Left: VariableNode{Name: "x"}, Operator: "==", Right: VariableNode{Name: "y"}}, true, false, ""},
		{"Comp EQ Var(raw placeholder) vs String", BinaryOpNode{Left: VariableNode{Name: "phVar"}, Operator: "==", Right: StringLiteralNode{Value: "{{inner}}"}}, true, false, ""},
		{"Comp EQ LAST(raw) vs String", BinaryOpNode{Left: LastNode{}, Operator: "==", Right: StringLiteralNode{Value: "true {{inner}}"}}, true, false, ""},
		{"Comp NEQ String True", BinaryOpNode{Left: VariableNode{Name: "x"}, Operator: "!=", Right: VariableNode{Name: "z"}}, true, false, ""},
		{"Comp NEQ Num True", BinaryOpNode{Left: VariableNode{Name: "n1"}, Operator: "!=", Right: VariableNode{Name: "n2"}}, true, false, ""},
		{"Comp GT True", BinaryOpNode{Left: VariableNode{Name: "n1"}, Operator: ">", Right: VariableNode{Name: "n2"}}, true, false, ""},
		{"Comp LT True", BinaryOpNode{Left: VariableNode{Name: "n2"}, Operator: "<", Right: VariableNode{Name: "n1"}}, true, false, ""},
		{"Comp GTE Equal", BinaryOpNode{Left: VariableNode{Name: "n1"}, Operator: ">=", Right: VariableNode{Name: "n3"}}, true, false, ""},
		{"Comp LTE Equal", BinaryOpNode{Left: VariableNode{Name: "n1"}, Operator: "<=", Right: VariableNode{Name: "n3"}}, true, false, ""},
		// Error cases for comparisons
		{"Comp Numeric Error Types", BinaryOpNode{Left: VariableNode{Name: "x"}, Operator: ">", Right: VariableNode{Name: "n1"}}, false, true, "requires numeric operands"},
		{"Comp Numeric Error String Lit", BinaryOpNode{Left: StringLiteralNode{Value: "a"}, Operator: "<", Right: StringLiteralNode{Value: "b"}}, false, true, "requires numeric operands"},
		// Variable not found cases for comparisons (evaluate as nil)
		{"Comp Error Evaluating LHS Var Not Found EQ", BinaryOpNode{Left: VariableNode{Name: "missing"}, Operator: "==", Right: VariableNode{Name: "x"}}, false, false, ""}, // nil == "A" -> false
		{"Comp Error Evaluating RHS Var Not Found EQ", BinaryOpNode{Left: VariableNode{Name: "x"}, Operator: "==", Right: VariableNode{Name: "missing"}}, false, false, ""}, // "A" == nil -> false
		// Mixed type comparisons
		{"Comp String Num vs Num EQ", BinaryOpNode{Left: VariableNode{Name: "strNum10"}, Operator: "==", Right: VariableNode{Name: "n1"}}, true, false, ""}, // "10" == 10 -> true
		{"Comp String Num vs Num GT", BinaryOpNode{Left: VariableNode{Name: "strNum10"}, Operator: ">", Right: VariableNode{Name: "n2"}}, true, false, ""},  // 10 > 5 -> true

		// Nil Comparisons using BinaryOpNode
		{"Comp EQ Nil vs Nil", BinaryOpNode{Left: VariableNode{Name: "nilVar"}, Operator: "==", Right: VariableNode{Name: "nilVar"}}, true, false, ""},
		{"Comp EQ Nil vs String", BinaryOpNode{Left: VariableNode{Name: "nilVar"}, Operator: "==", Right: StringLiteralNode{Value: "A"}}, false, false, ""},
		{"Comp NEQ Nil vs Nil", BinaryOpNode{Left: VariableNode{Name: "nilVar"}, Operator: "!=", Right: VariableNode{Name: "nilVar"}}, false, false, ""},
		{"Comp NEQ Nil vs String", BinaryOpNode{Left: VariableNode{Name: "nilVar"}, Operator: "!=", Right: StringLiteralNode{Value: "A"}}, true, false, ""},
		// Comparison operators > < >= <= should fail with nil
		{"Comp GT Nil vs Num", BinaryOpNode{Left: VariableNode{Name: "nilVar"}, Operator: ">", Right: NumberLiteralNode{Value: int64(5)}}, false, true, "cannot be applied to nil operand"},
		// Comparing not found variables (evaluated as nil)
		{"Comp Var Not Found vs Nil EQ", BinaryOpNode{Left: VariableNode{Name: "not_found"}, Operator: "==", Right: VariableNode{Name: "nilVar"}}, true, false, ""},                // nil == nil -> true
		{"Comp Var Not Found vs Var Not Found EQ", BinaryOpNode{Left: VariableNode{Name: "not_found1"}, Operator: "==", Right: VariableNode{Name: "not_found2"}}, true, false, ""}, // nil == nil -> true
		{"Comp Var Not Found vs String NEQ", BinaryOpNode{Left: VariableNode{Name: "not_found"}, Operator: "!=", Right: StringLiteralNode{Value: "A"}}, true, false, ""},           // nil != "A" -> true
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// *** FIXED: Use newTestInterpreter from test scope ***
			interp, _ := newTestInterpreter(t, vars, lastValue) // Get interpreter, ignore sandbox path
			got, err := interp.evaluateCondition(tt.node)

			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateCondition(%s): Error expectation mismatch. got err = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Errorf("evaluateCondition(%s): Expected error containing %q, got: %v", tt.name, tt.errContains, err)
				}
			} else {
				if got != tt.expected {
					t.Errorf("evaluateCondition(%s)\nNode:       %+v\nGot bool:   %v\nWant bool:  %v", tt.name, tt.node, got, tt.expected)
				}
			}
		})
	}
}
