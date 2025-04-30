// filename: pkg/core/evaluation_logical_bitwise_test.go
package core

import (
	"testing"
	// Assuming Position is defined in this package (e.g., ast.go)
	// No sub-package import needed.
)

// --- Test Suite for Logical and Bitwise Operations ---
func TestLogicalBitwiseOps(t *testing.T) {
	vars := map[string]interface{}{
		"trueVar":      true,
		"falseVar":     false,
		"numOne":       int64(1),
		"numZero":      int64(0),
		"floatOne":     float64(1.0),
		"floatZero":    float64(0.0),
		"strTrue":      "true",   // Truthy string
		"strFalse":     "false",  // Falsy string
		"strEmpty":     "",       // Falsy string (Zero Value)
		"strOther":     "hello",  // Falsy string (but non-zero value)
		"strOne":       "1",      // Truthy string
		"nilVar":       nil,      // Falsy (Zero Value)
		"num5":         int64(5), // 0101
		"num3":         int64(3), // 0011
		"floatNonInt":  float64(3.14),
		"emptyListVar": []interface{}{},                      // Falsy (Zero Value)
		"listVar":      []interface{}{int64(1), "a"},         // Truthy (non-zero value)
		"emptyMapVar":  map[string]interface{}{},             // Falsy (Zero Value)
		"mapVar":       map[string]interface{}{"key": "val"}, // Truthy (non-zero value)
	}
	lastResult := "LastResult" // Placeholder for LAST tests, value doesn't matter here

	// Define dummyPos using local Position type pointer
	dummyPos := &Position{Line: 1, Column: 1}

	tests := []EvalTestCase{
		// --- NOT Operator ---
		// *** CORRECTED: Add & to node literals where Expression is expected ***
		{"NOT True Literal", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &BooleanLiteralNode{Pos: dummyPos, Value: true}}, vars, lastResult, false, false, nil},
		{"NOT False Literal", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &BooleanLiteralNode{Pos: dummyPos, Value: false}}, vars, lastResult, true, false, nil},
		{"NOT True Var", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "trueVar"}}, vars, lastResult, false, false, nil},
		{"NOT False Var", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "falseVar"}}, vars, lastResult, true, false, nil},
		{"NOT Num NonZero", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "numOne"}}, vars, lastResult, false, false, nil},
		{"NOT Num Zero", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "numZero"}}, vars, lastResult, true, false, nil},
		{"NOT Str Empty", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "strEmpty"}}, vars, lastResult, true, false, nil},
		{"NOT Str Other", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "strOther"}}, vars, lastResult, true, false, nil},
		{"NOT Str True", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "strTrue"}}, vars, lastResult, false, false, nil},
		{"NOT Nil", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, vars, lastResult, true, false, nil},
		{"NOT Empty List", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "emptyListVar"}}, vars, lastResult, true, false, nil},
		{"NOT List", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "listVar"}}, vars, lastResult, false, false, nil},
		{"NOT Empty Map", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "emptyMapVar"}}, vars, lastResult, true, false, nil},
		{"NOT Map", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "mapVar"}}, vars, lastResult, false, false, nil},
		{"NOT Not Found", &UnaryOpNode{Pos: dummyPos, Operator: "NOT", Operand: &VariableNode{Pos: dummyPos, Name: "missing"}}, vars, lastResult, nil, true, ErrVariableNotFound},

		// --- AND Operator ---
		{"AND True True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "trueVar"}, Operator: "AND", Right: &BooleanLiteralNode{Pos: dummyPos, Value: true}}, vars, lastResult, true, false, nil},
		{"AND True False", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "trueVar"}, Operator: "AND", Right: &VariableNode{Pos: dummyPos, Name: "falseVar"}}, vars, lastResult, false, false, nil},
		{"AND False True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "falseVar"}, Operator: "AND", Right: &VariableNode{Pos: dummyPos, Name: "trueVar"}}, vars, lastResult, false, false, nil},
		{"AND False False", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "falseVar"}, Operator: "AND", Right: &VariableNode{Pos: dummyPos, Name: "falseVar"}}, vars, lastResult, false, false, nil},
		{"AND Num1 StrTrue", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "numOne"}, Operator: "AND", Right: &VariableNode{Pos: dummyPos, Name: "strTrue"}}, vars, lastResult, true, false, nil},
		{"AND Num0 StrTrue", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "numZero"}, Operator: "AND", Right: &VariableNode{Pos: dummyPos, Name: "strTrue"}}, vars, lastResult, false, false, nil},
		{"AND Nil True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: "AND", Right: &VariableNode{Pos: dummyPos, Name: "trueVar"}}, vars, lastResult, false, false, nil},
		{"AND True Nil", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "trueVar"}, Operator: "AND", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, vars, lastResult, false, false, nil},
		{"AND Not Found Left", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "missing"}, Operator: "AND", Right: &VariableNode{Pos: dummyPos, Name: "trueVar"}}, vars, lastResult, nil, true, ErrVariableNotFound},

		// --- OR Operator ---
		{"OR True True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "trueVar"}, Operator: "OR", Right: &BooleanLiteralNode{Pos: dummyPos, Value: true}}, vars, lastResult, true, false, nil},
		{"OR True False", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "trueVar"}, Operator: "OR", Right: &VariableNode{Pos: dummyPos, Name: "falseVar"}}, vars, lastResult, true, false, nil},
		{"OR False True", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "falseVar"}, Operator: "OR", Right: &VariableNode{Pos: dummyPos, Name: "trueVar"}}, vars, lastResult, true, false, nil},
		{"OR False False", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "falseVar"}, Operator: "OR", Right: &VariableNode{Pos: dummyPos, Name: "falseVar"}}, vars, lastResult, false, false, nil},
		{"OR Num0 StrFalse", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "numZero"}, Operator: "OR", Right: &VariableNode{Pos: dummyPos, Name: "strFalse"}}, vars, lastResult, false, false, nil},
		{"OR Num1 StrFalse", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "numOne"}, Operator: "OR", Right: &VariableNode{Pos: dummyPos, Name: "strFalse"}}, vars, lastResult, true, false, nil},
		{"OR StrOther Nil", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "strOther"}, Operator: "OR", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, vars, lastResult, false, false, nil},
		{"OR Nil False", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "nilVar"}, Operator: "OR", Right: &VariableNode{Pos: dummyPos, Name: "falseVar"}}, vars, lastResult, false, false, nil},
		{"OR Not Found Right", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "falseVar"}, Operator: "OR", Right: &VariableNode{Pos: dummyPos, Name: "missing"}}, vars, lastResult, nil, true, ErrVariableNotFound},

		// --- Bitwise AND (&) ---
		{"Bitwise AND 5&3", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "&", Right: &VariableNode{Pos: dummyPos, Name: "num3"}}, vars, lastResult, int64(1), false, nil},
		{"Bitwise AND 5&0", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "&", Right: &VariableNode{Pos: dummyPos, Name: "numZero"}}, vars, lastResult, int64(0), false, nil},
		{"Bitwise AND Error Float", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "&", Right: &VariableNode{Pos: dummyPos, Name: "floatOne"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise AND Error String", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "strOne"}, Operator: "&", Right: &VariableNode{Pos: dummyPos, Name: "num3"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise AND Error Nil", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "&", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},

		// --- Bitwise OR (|) ---
		{"Bitwise OR 5|3", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "|", Right: &VariableNode{Pos: dummyPos, Name: "num3"}}, vars, lastResult, int64(7), false, nil},
		{"Bitwise OR 5|0", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "|", Right: &VariableNode{Pos: dummyPos, Name: "numZero"}}, vars, lastResult, int64(5), false, nil},
		{"Bitwise OR Error Float", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "|", Right: &VariableNode{Pos: dummyPos, Name: "floatOne"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise OR Error String", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "strOne"}, Operator: "|", Right: &VariableNode{Pos: dummyPos, Name: "num3"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise OR Error Nil", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "|", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},

		// --- Bitwise XOR (^) ---
		{"Bitwise XOR 5^3", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "^", Right: &VariableNode{Pos: dummyPos, Name: "num3"}}, vars, lastResult, int64(6), false, nil},
		{"Bitwise XOR 5^5", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "^", Right: &VariableNode{Pos: dummyPos, Name: "num5"}}, vars, lastResult, int64(0), false, nil},
		{"Bitwise XOR Error Float", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "^", Right: &VariableNode{Pos: dummyPos, Name: "floatOne"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise XOR Error String", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "strOne"}, Operator: "^", Right: &VariableNode{Pos: dummyPos, Name: "num3"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise XOR Error Nil", &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "num5"}, Operator: "^", Right: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},

		// --- Bitwise NOT (~) ---
		{"Bitwise NOT ~5", &UnaryOpNode{Pos: dummyPos, Operator: "~", Operand: &VariableNode{Pos: dummyPos, Name: "num5"}}, vars, lastResult, int64(-6), false, nil},
		{"Bitwise NOT ~0", &UnaryOpNode{Pos: dummyPos, Operator: "~", Operand: &VariableNode{Pos: dummyPos, Name: "numZero"}}, vars, lastResult, int64(-1), false, nil},
		{"Bitwise NOT Error Float", &UnaryOpNode{Pos: dummyPos, Operator: "~", Operand: &VariableNode{Pos: dummyPos, Name: "floatNonInt"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise NOT Error String", &UnaryOpNode{Pos: dummyPos, Operator: "~", Operand: &VariableNode{Pos: dummyPos, Name: "strOther"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise NOT Error Nil", &UnaryOpNode{Pos: dummyPos, Operator: "~", Operand: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},

		// --- no Operator (Zero Value Check) ---
		{"no nilVar", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, vars, lastResult, true, false, nil},
		{"no strEmpty", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "strEmpty"}}, vars, lastResult, true, false, nil},
		{"no numZero", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "numZero"}}, vars, lastResult, true, false, nil},
		{"no floatZero", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "floatZero"}}, vars, lastResult, true, false, nil},
		{"no falseVar", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "falseVar"}}, vars, lastResult, true, false, nil},
		{"no emptyListVar", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "emptyListVar"}}, vars, lastResult, true, false, nil},
		{"no emptyMapVar", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "emptyMapVar"}}, vars, lastResult, true, false, nil},
		{"no notFoundVar", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "notFoundVar"}}, vars, lastResult, nil, true, ErrVariableNotFound},
		{"no strOther", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "strOther"}}, vars, lastResult, false, false, nil},
		{"no numOne", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "numOne"}}, vars, lastResult, false, false, nil},
		{"no floatOne", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "floatOne"}}, vars, lastResult, false, false, nil},
		{"no trueVar", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "trueVar"}}, vars, lastResult, false, false, nil},
		{"no listVar", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "listVar"}}, vars, lastResult, false, false, nil},
		{"no mapVar", &UnaryOpNode{Pos: dummyPos, Operator: "no", Operand: &VariableNode{Pos: dummyPos, Name: "mapVar"}}, vars, lastResult, false, false, nil},

		// --- some Operator (Non-Zero Value Check) ---
		{"some nilVar", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "nilVar"}}, vars, lastResult, false, false, nil},
		{"some strEmpty", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "strEmpty"}}, vars, lastResult, false, false, nil},
		{"some numZero", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "numZero"}}, vars, lastResult, false, false, nil},
		{"some floatZero", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "floatZero"}}, vars, lastResult, false, false, nil},
		{"some falseVar", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "falseVar"}}, vars, lastResult, false, false, nil},
		{"some emptyListVar", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "emptyListVar"}}, vars, lastResult, false, false, nil},
		{"some emptyMapVar", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "emptyMapVar"}}, vars, lastResult, false, false, nil},
		{"some notFoundVar", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "notFoundVar"}}, vars, lastResult, nil, true, ErrVariableNotFound},
		{"some strOther", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "strOther"}}, vars, lastResult, true, false, nil},
		{"some numOne", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "numOne"}}, vars, lastResult, true, false, nil},
		{"some floatOne", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "floatOne"}}, vars, lastResult, true, false, nil},
		{"some trueVar", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "trueVar"}}, vars, lastResult, true, false, nil},
		{"some listVar", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "listVar"}}, vars, lastResult, true, false, nil},
		{"some mapVar", &UnaryOpNode{Pos: dummyPos, Operator: "some", Operand: &VariableNode{Pos: dummyPos, Name: "mapVar"}}, vars, lastResult, true, false, nil},
		// *** END CORRECTIONS ***
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Assuming runEvalExpressionTest is defined in testing_helpers_test.go
			// It takes EvalTestCase which has InputNode interface{}
			// The helper itself should handle asserting InputNode to Expression if needed
			runEvalExpressionTest(t, tt)
		})
	}
}

// NOTE: Commented out helper stubs below as they are defined elsewhere or not needed
/*
// Helper Function (already provided in testing_helpers_test.go)
// Assume runEvalExpressionTest is correctly defined elsewhere and handles the EvalTestCase struct.

// EvalTestCase struct (already provided in testing_helpers_test.go)
type EvalTestCase struct {
    Name            string
    InputNode       interface{} // The AST node (or sometimes a literal value for simplicity)
    InitialVars     map[string]interface{}
    LastResult      interface{} // Mocked result of previous step if needed
    Expected        interface{} // Expected result of evaluation
    WantErr         bool
    ExpectedErrorIs error // Use sentinel error or nil
}

// runEvalExpressionTest (already provided in testing_helpers_test.go)
func runEvalExpressionTest(t *testing.T, tt EvalTestCase) {
    t.Helper()
    interp, _ := NewDefaultTestInterpreter(t) // Use default interpreter
    if tt.InitialVars != nil {
        for k, v := range tt.InitialVars {
            interp.variables[k] = v
        }
    }
    interp.lastCallResult = tt.LastResult

    // Assert InputNode to Expression before passing to evaluateExpression
    var inputExpr Expression
    if tt.InputNode != nil {
 		var ok bool
 		inputExpr, ok = tt.InputNode.(Expression)
 		if !ok {
 			t.Fatalf("Test setup error in %s: InputNode (%T) does not implement Expression", tt.Name, tt.InputNode)
 		}
 	} // else inputExpr remains nil, evaluateExpression should handle nil input if necessary

    got, err := interp.evaluateExpression(inputExpr) // Pass asserted Expression

    // ... rest of the validation logic using reflect.DeepEqual or errors.Is ...
	// (Keep the existing validation logic from the pasted file)
    if tt.WantErr {
        if err == nil {
            t.Errorf("%s: Expected error, but got nil", tt.Name)
            return
        }
        if tt.ExpectedErrorIs != nil && !errors.Is(err, tt.ExpectedErrorIs) {
            t.Errorf("%s: Error mismatch.\nExpected error wrapping [%v]\nGot:               [%v]", tt.Name, tt.ExpectedErrorIs, err)
        } else if tt.ExpectedErrorIs == nil {
            t.Logf("%s: Got expected error: %v", tt.Name, err)
        }
    } else { // No error wanted
        if err != nil {
            t.Errorf("%s: Unexpected error: %v", tt.Name, err)
        } else if !reflect.DeepEqual(got, tt.Expected) { // Use reflect.DeepEqual here
            t.Errorf("%s: Result mismatch.\nInput Node: %+v\nExpected:   %v (%T)\nGot:        %v (%T)",
                tt.Name, tt.InputNode, tt.Expected, tt.Expected, got, got)
        }
    }
}

// Minimal Position struct if not imported
type Position struct { Line int; Column int }
func (p *Position) GetPos() *Position { return p }
// Minimal Node structs if not imported
type UnaryOpNode struct { Pos *Position; Operator string; Operand Expression }
func (n *UnaryOpNode) GetPos() *Position { return n.Pos }
func (n *UnaryOpNode) expressionNode() {} // Marker method if Expression requires it

type BinaryOpNode struct { Pos *Position; Left Expression; Operator string; Right Expression }
func (n *BinaryOpNode) GetPos() *Position { return n.Pos }
func (n *BinaryOpNode) expressionNode() {} // Marker method

type BooleanLiteralNode struct { Pos *Position; Value bool }
func (n *BooleanLiteralNode) GetPos() *Position { return n.Pos }
func (n *BooleanLiteralNode) expressionNode() {} // Marker method

type VariableNode struct { Pos *Position; Name string }
func (n *VariableNode) GetPos() *Position { return n.Pos }
func (n *VariableNode) expressionNode() {} // Marker method
*/
