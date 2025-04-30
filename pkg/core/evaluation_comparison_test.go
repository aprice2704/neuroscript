// filename: pkg/core/evaluation_comparison_test.go
package core

import (
	"strings"
	"testing"
	// Need Position definition, assuming it's in this package (ast.go)
)

func TestEvaluateCondition(t *testing.T) {
	vars := map[string]interface{}{
		"trueVar":   true,
		"falseVar":  false,
		"numOne":    int64(1),
		"numZero":   int64(0),
		"floatOne":  float64(1.0),
		"floatZero": float64(0.0),
		"strTrue":   "true",  // Truthy string
		"strFalse":  "false", // Falsy string
		"strOther":  "hello", // Falsy string (now)
		"strNum10":  "10",    // Falsy string (now)
		"strOne":    "1",     // Truthy string
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
	lastValue := "true" // Raw value for LAST tests, should evaluate as true

	// Define dummyPos using local Position type pointer
	dummyPos := &Position{Line: 1, Column: 1}

	tests := []struct {
		name        string
		node        Expression // Change node type to Expression for clarity where applicable
		expected    bool       // Expected bool result if no error
		wantErr     bool       // Expect an error during evaluation
		errContains string     // Substring expected in the error message
	}{
		// Single Expression Conditions (Truthiness)
		// *** CORRECTED: Add & to node literals where Expression is expected ***
		{"Bool Literal True", &BooleanLiteralNode{Pos: dummyPos, Value: true}, true, false, ""},
		{"Bool Literal False", &BooleanLiteralNode{Pos: dummyPos, Value: false}, false, false, ""},
		{"Var Bool True", &VariableNode{Pos: dummyPos, Name: "trueVar"}, true, false, ""},
		{"Var Bool False", &VariableNode{Pos: dummyPos, Name: "falseVar"}, false, false, ""},
		{"Var Num NonZero", &VariableNode{Pos: dummyPos, Name: "numOne"}, true, false, ""},
		{"Var Num Zero", &VariableNode{Pos: dummyPos, Name: "numZero"}, false, false, ""},
		{"Var Float NonZero", &VariableNode{Pos: dummyPos, Name: "floatOne"}, true, false, ""},
		{"Var Float Zero", &VariableNode{Pos: dummyPos, Name: "floatZero"}, false, false, ""},
		{"Var String True", &VariableNode{Pos: dummyPos, Name: "strTrue"}, true, false, ""},
		{"Var String False", &VariableNode{Pos: dummyPos, Name: "strFalse"}, false, false, ""},
		{"Var String Other", &VariableNode{Pos: dummyPos, Name: "strOther"}, false, false, ""},
		{"Var String One", &VariableNode{Pos: dummyPos, Name: "strOne"}, true, false, ""},
		{"String Literal True", &StringLiteralNode{Pos: dummyPos, Value: "true"}, true, false, ""},
		{"String Literal False", &StringLiteralNode{Pos: dummyPos, Value: "false"}, false, false, ""},
		{"String Literal One", &StringLiteralNode{Pos: dummyPos, Value: "1"}, true, false, ""},
		{"String Literal Other", &StringLiteralNode{Pos: dummyPos, Value: "yes"}, false, false, ""},
		{"Number Literal NonZero", &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, true, false, ""},
		{"Number Literal Zero", &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}, false, false, ""},
		{"Variable Not Found Condition", &VariableNode{Pos: dummyPos, Name: "not_found"}, false, false, ""},
		// Add & for List/Map literals assigned to Expression (node field)
		{"List Literal Condition", &ListLiteralNode{Pos: dummyPos, Elements: []Expression{}}, false, false, ""}, // Empty list is falsy
		{"Map Literal Condition", &MapLiteralNode{Pos: dummyPos, Entries: []MapEntryNode{}}, false, false, ""},  // Empty map is falsy
		{"Nil Variable Condition", &VariableNode{Pos: dummyPos, Name: "nilVar"}, false, false, ""},
		{"LAST Condition (Truthy String)", &LastNode{Pos: dummyPos}, true, false, ""},

		// --- Comparison Conditions using BinaryOpNode ---
		{"Comp EQ String True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "x"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "y"}}, true, false, ""},
		{"Comp EQ Var(raw placeholder) vs String", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "phVar"}, Operator: "==", Right: &StringLiteralNode{Pos: dummyPos, Value: "{{inner}}"}}, true, false, ""},
		{"Comp EQ LAST(raw) vs String", &BinaryOpNode{Pos: dummyPos, Left: &LastNode{Pos: dummyPos}, Operator: "==", Right: &StringLiteralNode{Pos: dummyPos, Value: "true"}}, true, false, ""},
		{"Comp NEQ String True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "x"}, Operator: "!=", Right: &VariableNode{Pos: dummyPos, Name: "z"}}, true, false, ""},
		{"Comp NEQ Num True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "n1"}, Operator: "!=", Right: &VariableNode{Pos: dummyPos, Name: "n2"}}, true, false, ""},
		{"Comp GT True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "n1"}, Operator: ">", Right: &VariableNode{Pos: dummyPos, Name: "n2"}}, true, false, ""},
		{"Comp LT True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "n2"}, Operator: "<", Right: &VariableNode{Pos: dummyPos, Name: "n1"}}, true, false, ""},
		{"Comp GTE Equal", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "n1"}, Operator: ">=", Right: &VariableNode{Pos: dummyPos, Name: "n3"}}, true, false, ""},
		{"Comp LTE Equal", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "n1"}, Operator: "<=", Right: &VariableNode{Pos: dummyPos, Name: "n3"}}, true, false, ""},
		{"Comp Numeric Error Types", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "x"}, Operator: ">", Right: &VariableNode{Pos: dummyPos, Name: "n1"}}, false, true, "requires numeric operands"},
		{"Comp Numeric Error String Lit", &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "a"}, Operator: "<", Right: &StringLiteralNode{Pos: dummyPos, Value: "b"}}, false, true, "requires numeric operands"},
		{"Comp Error Evaluating LHS Var Not Found EQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "missing"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "x"}}, false, false, ""},
		{"Comp Error Evaluating RHS Var Not Found EQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "x"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "missing"}}, false, false, ""},
		{"Comp String Num vs Num EQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "strNum10"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "n1"}}, true, false, ""},
		{"Comp String Num vs Num GT", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "strNum10"}, Operator: ">", Right: &VariableNode{Pos: dummyPos, Name: "n2"}}, true, false, ""},

		// Nil Comparisons using BinaryOpNode
		{"Comp EQ Nil vs Nil", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, true, false, ""},
		{"Comp EQ Nil vs String", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: "==", Right: &StringLiteralNode{Pos: dummyPos, Value: "A"}}, false, false, ""},
		{"Comp NEQ Nil vs Nil", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: "!=", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, false, false, ""},
		{"Comp NEQ Nil vs String", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: "!=", Right: &StringLiteralNode{Pos: dummyPos, Value: "A"}}, true, false, ""},
		{"Comp GT Nil vs Num", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: ">", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(5)}}, false, true, "operation received nil operand"},
		{"Comp Var Not Found vs Nil EQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "not_found"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, true, false, ""},
		{"Comp Var Not Found vs Var Not Found EQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "not_found1"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "not_found2"}}, true, false, ""},
		{"Comp Var Not Found vs String NEQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "not_found"}, Operator: "!=", Right: &StringLiteralNode{Pos: dummyPos, Value: "A"}}, true, false, ""},
		// *** END CORRECTIONS ***
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp, _ := NewTestInterpreter(t, vars, lastValue) // Use test helper

			// Assert node to Expression before passing
			inputExpr, ok := tt.node.(Expression)
			if !ok {
				t.Fatalf("Test setup error: InputNode (%T) in test '%s' does not implement Expression", tt.node, tt.name)
			}

			got, err := interp.evaluateCondition(inputExpr) // Pass asserted Expression

			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateCondition(%s): Error expectation mismatch. got err = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
					t.Errorf("evaluateCondition(%s): Expected error containing %q, got: %v", tt.name, tt.errContains, err)
				} else if tt.errContains == "" && err != nil {
					t.Logf("evaluateCondition(%s): Got expected error: %v", tt.name, err)
				}
			} else {
				if got != tt.expected {
					t.Errorf("evaluateCondition(%s)\nNode:       %+v\nGot bool:   %t\nWant bool:  %t", tt.name, tt.node, got, tt.expected)
				}
			}
		})
	}
}
