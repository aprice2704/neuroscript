// NeuroScript Version: 0.3.5
// Last Modified: 2025-05-01 15:05:01 PDT
// filename: pkg/core/evaluation_arithmetic_test.go
package core

import (
	"testing"
	// Need Position definition
)

// Assumes NewDefaultTestInterpreter and runEvalExpressionTest are defined in testing_helpers_test.go
// Assumes EvalTestCase struct is defined in testing_helpers_test.go

func TestArithmeticOps(t *testing.T) {
	vars := map[string]interface{}{
		"int5":     int64(5),
		"int3":     int64(3),
		"float2_5": float64(2.5),
		"float1_5": float64(1.5),
		"str10":    "10",  // String representing a number
		"strABC":   "ABC", // String not representing a number
		"int0":     int64(0),
		"float0":   float64(0.0),
	}

	// Define dummyPos using local Position type pointer
	dummyPos := &Position{Line: 1, Column: 1}

	// Use the named EvalTestCase struct (updated with ExpectedErrorIs)
	testCases := []EvalTestCase{
		// Addition
		// *** CORRECTED: Add & to node literals where Expression is expected ***
		{Name: "Add Int+Int", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "int3"}}, InitialVars: vars, Expected: int64(8), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Add Float+Float", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "float2_5"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "float1_5"}}, InitialVars: vars, Expected: float64(4.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Add Int+Float", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "float1_5"}}, InitialVars: vars, Expected: float64(6.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Add Float+Int", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "float2_5"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "int3"}}, InitialVars: vars, Expected: float64(5.5), WantErr: false, ExpectedErrorIs: nil},
		// --- MODIFIED: Expect successful string concatenation ---
		{Name: "Add Int+StrABC", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "+", Right: &VariableNode{Pos: dummyPos, Name: "strABC"}}, InitialVars: vars, Expected: "5ABC", WantErr: false, ExpectedErrorIs: nil},
		// --- END MODIFICATION ---

		// Subtraction
		{Name: "Sub Int-Int", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "-", Right: &VariableNode{Pos: dummyPos, Name: "int3"}}, InitialVars: vars, Expected: int64(2), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Sub Float-Float", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "float2_5"}, Operator: "-", Right: &VariableNode{Pos: dummyPos, Name: "float1_5"}}, InitialVars: vars, Expected: float64(1.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Sub Int-Float", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "-", Right: &VariableNode{Pos: dummyPos, Name: "float1_5"}}, InitialVars: vars, Expected: float64(3.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Sub Float-Int", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "float2_5"}, Operator: "-", Right: &VariableNode{Pos: dummyPos, Name: "int3"}}, InitialVars: vars, Expected: float64(-0.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Sub Error Str", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "-", Right: &VariableNode{Pos: dummyPos, Name: "strABC"}}, InitialVars: vars, Expected: nil, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric}, // Subtraction should still require numeric

		// Multiplication
		{Name: "Mul Int*Int", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "*", Right: &VariableNode{Pos: dummyPos, Name: "int3"}}, InitialVars: vars, Expected: int64(15), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Mul Float*Float", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "float2_5"}, Operator: "*", Right: &VariableNode{Pos: dummyPos, Name: "float1_5"}}, InitialVars: vars, Expected: float64(3.75), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Mul Int*Float", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int3"}, Operator: "*", Right: &VariableNode{Pos: dummyPos, Name: "float2_5"}}, InitialVars: vars, Expected: float64(7.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Mul Error Str", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "*", Right: &VariableNode{Pos: dummyPos, Name: "strABC"}}, InitialVars: vars, Expected: nil, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric}, // Multiplication should still require numeric

		// Division
		{Name: "Div Int/Int Exact", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(6)}, Operator: "/", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, Expected: int64(2), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Div Int/Int Inexact", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(7)}, Operator: "/", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, Expected: float64(7.0 / 3.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Div Float/Int", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: float64(7.5)}, Operator: "/", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, Expected: float64(2.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Div Int/Float", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(10)}, Operator: "/", Right: &NumberLiteralNode{Pos: dummyPos, Value: float64(2.5)}}, InitialVars: vars, Expected: float64(4.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Div By Int Zero", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "/", Right: &VariableNode{Pos: dummyPos, Name: "int0"}}, InitialVars: vars, Expected: nil, WantErr: true, ExpectedErrorIs: ErrDivisionByZero},
		{Name: "Div By Float Zero", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "/", Right: &VariableNode{Pos: dummyPos, Name: "float0"}}, InitialVars: vars, Expected: nil, WantErr: true, ExpectedErrorIs: ErrDivisionByZero},
		{Name: "Div Error Str Denom", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "/", Right: &VariableNode{Pos: dummyPos, Name: "strABC"}}, InitialVars: vars, Expected: nil, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric},
		{Name: "Div Error Str Num", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "strABC"}, Operator: "/", Right: &VariableNode{Pos: dummyPos, Name: "int3"}}, InitialVars: vars, Expected: nil, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric},

		// Modulo
		{Name: "Mod Int%Int", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(7)}, Operator: "%", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, Expected: int64(1), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Mod Negative", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(-7)}, Operator: "%", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, Expected: int64(-1), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Mod Error Float", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: float64(7.5)}, Operator: "%", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, Expected: nil, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeInteger},
		{Name: "Mod By Int Zero", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "%", Right: &VariableNode{Pos: dummyPos, Name: "int0"}}, InitialVars: vars, Expected: nil, WantErr: true, ExpectedErrorIs: ErrDivisionByZero},

		// Power
		{Name: "Pow Int**Int", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}, Operator: "**", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, Expected: float64(8.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Pow Int**Float", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(4)}, Operator: "**", Right: &NumberLiteralNode{Pos: dummyPos, Value: float64(0.5)}}, InitialVars: vars, Expected: float64(2.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Pow Float**Int", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: float64(2.5)}, Operator: "**", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}}, InitialVars: vars, Expected: float64(6.25), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Pow Error Str", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &VariableNode{Pos: dummyPos, Name: "int5"}, Operator: "**", Right: &VariableNode{Pos: dummyPos, Name: "strABC"}}, InitialVars: vars, Expected: nil, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric},

		// Unary Minus
		{Name: "Unary Minus Int", InputNode: &UnaryOpNode{Pos: dummyPos, Operator: "-", Operand: &NumberLiteralNode{Pos: dummyPos, Value: int64(5)}}, InitialVars: vars, Expected: int64(-5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Unary Minus Float", InputNode: &UnaryOpNode{Pos: dummyPos, Operator: "-", Operand: &NumberLiteralNode{Pos: dummyPos, Value: float64(2.5)}}, InitialVars: vars, Expected: float64(-2.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Unary Minus Error Str", InputNode: &UnaryOpNode{Pos: dummyPos, Operator: "-", Operand: &VariableNode{Pos: dummyPos, Name: "strABC"}}, InitialVars: vars, Expected: nil, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric},

		// Precedence
		{Name: "Prec Add Mul: 2 + 3 * 4", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}, Operator: "+", Right: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}, Operator: "*", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(4)}}}, InitialVars: vars, Expected: int64(14), WantErr: false, ExpectedErrorIs: nil},
		// Corrected Precedence case: Multiply first
		{Name: "Prec Mul Add: 2 * 3 + 4", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}, Operator: "*", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, Operator: "+", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(4)}}, InitialVars: vars, Expected: int64(10), WantErr: false, ExpectedErrorIs: nil},
		// Corrected Precedence case: Parens simulated by AST structure
		{Name: "Prec Parens: (2 + 3) * 4", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}, Operator: "+", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, Operator: "*", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(4)}}, InitialVars: vars, Expected: int64(20), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Prec Pow Right Assoc: 2 ** 3 ** 2", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}, Operator: "**", Right: &BinaryOpNode{Pos: dummyPos, Left: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}, Operator: "**", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}}}, InitialVars: vars, Expected: float64(512.0), WantErr: false, ExpectedErrorIs: nil}, // 2**(3**2) = 2**9 = 512
		{Name: "Prec Unary Minus High: -2 + 3", InputNode: &BinaryOpNode{Pos: dummyPos, Left: &UnaryOpNode{Pos: dummyPos, Operator: "-", Operand: &NumberLiteralNode{Pos: dummyPos, Value: int64(2)}}, Operator: "+", Right: &NumberLiteralNode{Pos: dummyPos, Value: int64(3)}}, InitialVars: vars, Expected: int64(1), WantErr: false, ExpectedErrorIs: nil},
		// *** END CORRECTIONS ***
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel() // Optional
			// Assuming runEvalExpressionTest asserts InputNode to Expression if needed
			runEvalExpressionTest(t, tc)
		})
	}
}
