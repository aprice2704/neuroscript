// filename: pkg/core/evaluation_arithmetic_test.go
package core

import (
	"testing"
	// No longer need reflect, strings, errors directly here
)

// Assumes NewTestInterpreter( and runEvalExpressionTest are defined in testing_helpers_test.go

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

	// Use the named EvalTestCase struct (updated with ExpectedErrorIs)
	testCases := []EvalTestCase{
		// Addition
		{Name: "Add Int+Int", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "+", Right: VariableNode{Name: "int3"}}, InitialVars: vars, Expected: int64(8), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Add Float+Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "float2_5"}, Operator: "+", Right: VariableNode{Name: "float1_5"}}, InitialVars: vars, Expected: float64(4.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Add Int+Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "+", Right: VariableNode{Name: "float1_5"}}, InitialVars: vars, Expected: float64(6.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Add Float+Int", InputNode: BinaryOpNode{Left: VariableNode{Name: "float2_5"}, Operator: "+", Right: VariableNode{Name: "int3"}}, InitialVars: vars, Expected: float64(5.5), WantErr: false, ExpectedErrorIs: nil},
		// *** CORRECTED ExpectedErrorIs for Add Error Int+StrABC ***
		{Name: "Add Error Int+StrABC", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "+", Right: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidOperandType}, // Expect general type error now

		// Subtraction
		{Name: "Sub Int-Int", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "-", Right: VariableNode{Name: "int3"}}, InitialVars: vars, Expected: int64(2), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Sub Float-Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "float2_5"}, Operator: "-", Right: VariableNode{Name: "float1_5"}}, InitialVars: vars, Expected: float64(1.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Sub Int-Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "-", Right: VariableNode{Name: "float1_5"}}, InitialVars: vars, Expected: float64(3.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Sub Float-Int", InputNode: BinaryOpNode{Left: VariableNode{Name: "float2_5"}, Operator: "-", Right: VariableNode{Name: "int3"}}, InitialVars: vars, Expected: float64(-0.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Sub Error Str", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "-", Right: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric},

		// Multiplication
		{Name: "Mul Int*Int", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "*", Right: VariableNode{Name: "int3"}}, InitialVars: vars, Expected: int64(15), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Mul Float*Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "float2_5"}, Operator: "*", Right: VariableNode{Name: "float1_5"}}, InitialVars: vars, Expected: float64(3.75), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Mul Int*Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "int3"}, Operator: "*", Right: VariableNode{Name: "float2_5"}}, InitialVars: vars, Expected: float64(7.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Mul Error Str", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "*", Right: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric},

		// Division
		{Name: "Div Int/Int Exact", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(6)}, Operator: "/", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: int64(2), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Div Int/Int Inexact", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(7)}, Operator: "/", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: float64(7.0 / 3.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Div Float/Int", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: float64(7.5)}, Operator: "/", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: float64(2.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Div Int/Float", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(10)}, Operator: "/", Right: NumberLiteralNode{Value: float64(2.5)}}, InitialVars: vars, Expected: float64(4.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Div By Int Zero", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "/", Right: VariableNode{Name: "int0"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrDivisionByZero},
		{Name: "Div By Float Zero", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "/", Right: VariableNode{Name: "float0"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrDivisionByZero},
		{Name: "Div Error Str Denom", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "/", Right: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric},
		{Name: "Div Error Str Num", InputNode: BinaryOpNode{Left: VariableNode{Name: "strABC"}, Operator: "/", Right: VariableNode{Name: "int3"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric},

		// Modulo
		{Name: "Mod Int%Int", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(7)}, Operator: "%", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: int64(1), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Mod Negative", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(-7)}, Operator: "%", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: int64(-1), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Mod Error Float", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: float64(7.5)}, Operator: "%", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeInteger},
		{Name: "Mod By Int Zero", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "%", Right: VariableNode{Name: "int0"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrDivisionByZero},

		// Power
		{Name: "Pow Int**Int", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "**", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: float64(8.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Pow Int**Float", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(4)}, Operator: "**", Right: NumberLiteralNode{Value: float64(0.5)}}, InitialVars: vars, Expected: float64(2.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Pow Float**Int", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: float64(2.5)}, Operator: "**", Right: NumberLiteralNode{Value: int64(2)}}, InitialVars: vars, Expected: float64(6.25), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Pow Error Str", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "**", Right: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric},

		// Unary Minus
		{Name: "Unary Minus Int", InputNode: UnaryOpNode{Operator: "-", Operand: NumberLiteralNode{Value: int64(5)}}, InitialVars: vars, Expected: int64(-5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Unary Minus Float", InputNode: UnaryOpNode{Operator: "-", Operand: NumberLiteralNode{Value: float64(2.5)}}, InitialVars: vars, Expected: float64(-2.5), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Unary Minus Error Str", InputNode: UnaryOpNode{Operator: "-", Operand: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidOperandTypeNumeric},

		// Precedence
		{Name: "Prec Add Mul: 2 + 3 * 4", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "+", Right: BinaryOpNode{Left: NumberLiteralNode{Value: int64(3)}, Operator: "*", Right: NumberLiteralNode{Value: int64(4)}}}, InitialVars: vars, Expected: int64(14), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Prec Mul Add: 2 * 3 + 4", InputNode: BinaryOpNode{Left: BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "*", Right: NumberLiteralNode{Value: int64(3)}}, Operator: "+", Right: NumberLiteralNode{Value: int64(4)}}, InitialVars: vars, Expected: int64(10), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Prec Parens: (2 + 3) * 4", InputNode: BinaryOpNode{Left: BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "+", Right: NumberLiteralNode{Value: int64(3)}}, Operator: "*", Right: NumberLiteralNode{Value: int64(4)}}, InitialVars: vars, Expected: int64(20), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Prec Pow Right Assoc: 2 ** 3 ** 2", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "**", Right: BinaryOpNode{Left: NumberLiteralNode{Value: int64(3)}, Operator: "**", Right: NumberLiteralNode{Value: int64(2)}}}, InitialVars: vars, Expected: float64(512.0), WantErr: false, ExpectedErrorIs: nil},
		{Name: "Prec Unary Minus High: -2 + 3", InputNode: BinaryOpNode{Left: UnaryOpNode{Operator: "-", Operand: NumberLiteralNode{Value: int64(2)}}, Operator: "+", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: int64(1), WantErr: false, ExpectedErrorIs: nil},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel() // Optional
			runEvalExpressionTest(t, tc)
		})
	}
}
