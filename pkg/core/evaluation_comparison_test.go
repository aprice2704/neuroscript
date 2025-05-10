// NeuroScript Version: 0.3.5
// File version: 0.0.3 // Used ExpectedErrorIs for specific sentinel error checks.
// filename: pkg/core/evaluation_comparison_test.go
package core

import (
	"errors"
	"strings"
	"testing"
	// Need Position definition, assuming it's in this package (ast.go)
	// Assuming EvalTestCase and runEvalExpressionTest are defined in testing_helpers_test.go
	// Assuming AST node types (BooleanLiteralNode, etc.) and Position are defined in ast.go
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
		"phVar":     "{{inner}}",
		"inner":     "resolved",
	}
	lastValue := "true"

	dummyPos := &Position{Line: 1, Column: 1}

	tests := []struct {
		name            string
		node            Expression
		expected        bool
		wantErr         bool
		ExpectedErrorIs error  // Used for errors.Is checks
		errContains     string // Kept for cases where specific string content is still relevant AFTER sentinel check
	}{
		{"Bool Literal True", &BooleanLiteralNode{Pos: dummyPos, Value: true}, true, false, nil, ""},
		{"Bool Literal False", &BooleanLiteralNode{Pos: dummyPos, Value: false}, false, false, nil, ""},
		{"Var Bool True", &VariableNode{Pos: dummyPos, Name: "trueVar"}, true, false, nil, ""},
		{"Var Bool False", &VariableNode{Pos: dummyPos, Name: "falseVar"}, false, false, nil, ""},
		{"Var Num NonZero", &VariableNode{Pos: dummyPos, Name: "numOne"}, true, false, nil, ""},
		{"Var Num Zero", &VariableNode{Pos: dummyPos, Name: "numZero"}, false, false, nil, ""},
		{"Var Float NonZero", &VariableNode{Pos: dummyPos, Name: "floatOne"}, true, false, nil, ""},
		{"Var Float Zero", &VariableNode{Pos: dummyPos, Name: "floatZero"}, false, false, nil, ""},
		{"Var String True", &VariableNode{Pos: dummyPos, Name: "strTrue"}, true, false, nil, ""},
		{"Var String False", &VariableNode{Pos: dummyPos, Name: "strFalse"}, false, false, nil, ""},
		{"Var String Other", &VariableNode{Pos: dummyPos, Name: "strOther"}, false, false, nil, ""},
		{"Var String One", &VariableNode{Pos: dummyPos, Name: "strOne"}, true, false, nil, ""},
		{"String Literal True", &StringLiteralNode{Pos: dummyPos, Value: "true"}, true, false, nil, ""},
		{"String Literal False", &StringLiteralNode{Pos: dummyPos, Value: "false"}, false, false, nil, ""},
		{"String Literal One", &StringLiteralNode{Pos: dummyPos, Value: "1"}, true, false, nil, ""},
		{"String Literal Other", &StringLiteralNode{Pos: dummyPos, Value: "yes"}, false, false, nil, ""},
		{"Number Literal NonZero", &NumberLiteralNode{Pos: dummyPos, Value: int64(1)}, true, false, nil, ""},
		{"Number Literal Zero", &NumberLiteralNode{Pos: dummyPos, Value: int64(0)}, false, false, nil, ""},
		{"Variable Not Found Condition", &VariableNode{Pos: dummyPos, Name: "not_found"}, false, false, nil, ""}, // evaluates to nil, which is falsy
		{"List Literal Condition", &ListLiteralNode{Pos: dummyPos, Elements: []Expression{}}, false, false, nil, ""},
		{"Map Literal Condition", &MapLiteralNode{Pos: dummyPos, Entries: []*MapEntryNode{}}, false, false, nil, ""},
		{"Nil Variable Condition", &VariableNode{Pos: dummyPos, Name: "nilVar"}, false, false, nil, ""},
		{"LAST Condition (Truthy String)", &LastNode{Pos: dummyPos}, true, false, nil, ""},

		{"Comp EQ String True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "x"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "y"}}, true, false, nil, ""},
		{"Comp EQ Var(raw placeholder) vs String", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "phVar"}, Operator: "==", Right: &StringLiteralNode{Pos: dummyPos, Value: "{{inner}}"}}, true, false, nil, ""},
		{"Comp EQ LAST(raw) vs String", &BinaryOpNode{Pos: dummyPos, Left: &LastNode{Pos: dummyPos}, Operator: "==", Right: &StringLiteralNode{Pos: dummyPos, Value: "true"}}, true, false, nil, ""},
		{"Comp NEQ String True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "x"}, Operator: "!=", Right: &VariableNode{Pos: dummyPos, Name: "z"}}, true, false, nil, ""},
		{"Comp NEQ Num True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "n1"}, Operator: "!=", Right: &VariableNode{Pos: dummyPos, Name: "n2"}}, true, false, nil, ""},
		{"Comp GT True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "n1"}, Operator: ">", Right: &VariableNode{Pos: dummyPos, Name: "n2"}}, true, false, nil, ""},
		{"Comp LT True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "n2"}, Operator: "<", Right: &VariableNode{Pos: dummyPos, Name: "n1"}}, true, false, nil, ""},
		{"Comp GTE Equal", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "n1"}, Operator: ">=", Right: &VariableNode{Pos: dummyPos, Name: "n3"}}, true, false, nil, ""},
		{"Comp LTE Equal", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "n1"}, Operator: "<=", Right: &VariableNode{Pos: dummyPos, Name: "n3"}}, true, false, nil, ""},

		// --- CORRECTED to use ExpectedErrorIs ---
		{"Comp Numeric Error Types", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "x"}, Operator: ">", Right: &VariableNode{Pos: dummyPos, Name: "n1"}}, false, true, ErrInvalidOperandTypeNumeric, ""},
		{"Comp Numeric Error String Lit", &BinaryOpNode{Pos: dummyPos, Left: &StringLiteralNode{Pos: dummyPos, Value: "a"}, Operator: "<", Right: &StringLiteralNode{Pos: dummyPos, Value: "b"}}, false, true, ErrInvalidOperandTypeNumeric, ""},
		// --- END CORRECTION ---

		{"Comp Error Evaluating LHS Var Not Found EQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "missing"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "x"}}, false, false, nil, ""}, // nil == "A" -> false
		{"Comp Error Evaluating RHS Var Not Found EQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "x"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "missing"}}, false, false, nil, ""}, // "A" == nil -> false
		{"Comp String Num vs Num EQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "strNum10"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "n1"}}, true, false, nil, ""},
		{"Comp String Num vs Num GT", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "strNum10"}, Operator: ">", Right: &VariableNode{Pos: dummyPos, Name: "n2"}}, true, false, nil, ""},

		{"Comp EQ Nil vs Nil", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, true, false, nil, ""},
		{"Comp EQ Nil vs String", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: "==", Right: &StringLiteralNode{Pos: dummyPos, Value: "A"}}, false, false, nil, ""},
		{"Comp NEQ Nil vs Nil", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: "!=", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, false, false, nil, ""},
		{"Comp NEQ Nil vs String", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: "!=", Right: &StringLiteralNode{Pos: dummyPos, Value: "A"}}, true, false, nil, ""},
		// --- CORRECTED to use ExpectedErrorIs ---
		{"Comp GT Nil vs Num", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: ">", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(5)}}, false, true, ErrNilOperand, ""},
		// --- END CORRECTION ---
		{"Comp Var Not Found vs Nil EQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "not_found"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, true, false, nil, ""},                // nil == nil -> true
		{"Comp Var Not Found vs Var Not Found EQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "not_found1"}, Operator: "==", Right: &VariableNode{Pos: dummyPos, Name: "not_found2"}}, true, false, nil, ""}, // nil == nil -> true
		{"Comp Var Not Found vs String NEQ", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "not_found"}, Operator: "!=", Right: &StringLiteralNode{Pos: dummyPos, Value: "A"}}, true, false, nil, ""},           // nil != "A" -> true
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp, _ := NewTestInterpreter(t, vars, lastValue) // From testing_helpers.go

			inputExpr, ok := tt.node.(Expression)
			if !ok {
				t.Fatalf("Test setup error: InputNode (%T) in test '%s' does not implement Expression", tt.node, tt.name)
			}

			// evaluateCondition is part of the interpreter, which calls evaluateExpression
			got, err := interp.evaluateCondition(inputExpr)

			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateCondition(%s): Error expectation mismatch. got err = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("evaluateCondition(%s): Expected error, but got nil", tt.name)
					return // Important to return after this to prevent further checks on nil error
				}
				// Check for specific sentinel error if ExpectedErrorIs is set
				if tt.ExpectedErrorIs != nil {
					if !errors.Is(err, tt.ExpectedErrorIs) {
						t.Errorf("evaluateCondition(%s): Expected error to wrap [%v], but got [%v]", tt.name, tt.ExpectedErrorIs, err)
					}
				}
				// Check for errContains if provided (can be used alongside ExpectedErrorIs for more context if needed, or if ExpectedErrorIs is nil)
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("evaluateCondition(%s): Expected error message to contain %q, got: %v", tt.name, tt.errContains, err)
				}
				// If we expected an error and got one (and it passed the checks above), log it and finish.
				t.Logf("evaluateCondition(%s): Got expected error: %v", tt.name, err)

			} else { // No error wanted
				if err != nil {
					t.Errorf("evaluateCondition(%s): Unexpected error: %v", tt.name, err)
				} else if got != tt.expected {
					t.Errorf("evaluateCondition(%s)\nNode:       %+v\nGot bool:   %t\nWant bool:  %t", tt.name, tt.node, got, tt.expected)
				}
			}
		})
	}
}

// nlines: 144
// risk_rating: LOW
