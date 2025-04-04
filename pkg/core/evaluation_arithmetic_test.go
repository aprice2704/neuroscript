// pkg/core/evaluation_arithmetic_test.go
package core

import (
	"testing"
	// Removed reflect and strings - now handled by runEvalExpressionTest in helpers
)

// Assumes newTestInterpreter( and runEvalExpressionTest are defined in test_helpers_test.go

func TestArithmeticOps(t *testing.T) {
	vars := map[string]interface{}{
		"int5":     int64(5),
		"int3":     int64(3),
		"float2_5": float64(2.5),
		"float1_5": float64(1.5),
		"str10":    "10",
		"strABC":   "ABC",
		"int0":     int64(0),
		"float0":   float64(0.0),
	}

	// Use the named EvalTestCase struct
	testCases := []EvalTestCase{
		// ... (Addition, Subtraction, Multiplication tests remain the same) ...
		{Name: "Add Int+Int", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "+", Right: VariableNode{Name: "int3"}}, InitialVars: vars, Expected: int64(8), WantErr: false},
		{Name: "Add Float+Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "float2_5"}, Operator: "+", Right: VariableNode{Name: "float1_5"}}, InitialVars: vars, Expected: float64(4.0), WantErr: false},
		{Name: "Add Int+Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "+", Right: VariableNode{Name: "float1_5"}}, InitialVars: vars, Expected: float64(6.5), WantErr: false},
		{Name: "Add Float+Int", InputNode: BinaryOpNode{Left: VariableNode{Name: "float2_5"}, Operator: "+", Right: VariableNode{Name: "int3"}}, InitialVars: vars, Expected: float64(5.5), WantErr: false},
		{Name: "Add Int+StrNum", InputNode: BinaryOpNode{Left: VariableNode{Name: "int3"}, Operator: "+", Right: VariableNode{Name: "str10"}}, InitialVars: vars, Expected: int64(13), WantErr: false},
		{Name: "Add Float+StrNum", InputNode: BinaryOpNode{Left: VariableNode{Name: "float1_5"}, Operator: "+", Right: VariableNode{Name: "str10"}}, InitialVars: vars, Expected: float64(11.5), WantErr: false},
		{Name: "Sub Int-Int", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "-", Right: VariableNode{Name: "int3"}}, InitialVars: vars, Expected: int64(2), WantErr: false},
		{Name: "Sub Float-Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "float2_5"}, Operator: "-", Right: VariableNode{Name: "float1_5"}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "Sub Int-Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "-", Right: VariableNode{Name: "float1_5"}}, InitialVars: vars, Expected: float64(3.5), WantErr: false},
		{Name: "Sub Float-Int", InputNode: BinaryOpNode{Left: VariableNode{Name: "float2_5"}, Operator: "-", Right: VariableNode{Name: "int3"}}, InitialVars: vars, Expected: float64(-0.5), WantErr: false},
		{Name: "Sub Error Str", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "-", Right: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ErrContains: "requires numeric operands"},
		{Name: "Mul Int*Int", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "*", Right: VariableNode{Name: "int3"}}, InitialVars: vars, Expected: int64(15), WantErr: false},
		{Name: "Mul Float*Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "float2_5"}, Operator: "*", Right: VariableNode{Name: "float1_5"}}, InitialVars: vars, Expected: float64(3.75), WantErr: false},
		{Name: "Mul Int*Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "int3"}, Operator: "*", Right: VariableNode{Name: "float2_5"}}, InitialVars: vars, Expected: float64(7.5), WantErr: false},
		{Name: "Mul Error Str", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "*", Right: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ErrContains: "requires numeric operands"},

		// Division
		{Name: "Div Int/Int Exact", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(6)}, Operator: "/", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: int64(2), WantErr: false},
		{Name: "Div Int/Int Inexact", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(7)}, Operator: "/", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: float64(7.0 / 3.0), WantErr: false},
		{Name: "Div Float/Int", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: float64(7.5)}, Operator: "/", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: float64(2.5), WantErr: false},
		{Name: "Div Int/Float", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(10)}, Operator: "/", Right: NumberLiteralNode{Value: float64(2.5)}}, InitialVars: vars, Expected: float64(4.0), WantErr: false},
		{Name: "Div By Int Zero", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "/", Right: VariableNode{Name: "int0"}}, InitialVars: vars, WantErr: true, ErrContains: "division by zero"},
		{Name: "Div By Float Zero", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "/", Right: VariableNode{Name: "float0"}}, InitialVars: vars, WantErr: true, ErrContains: "division by zero"},
		// *** CORRECTED Error Expectations for Division Type Errors ***
		{Name: "Div Error Str Denom", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "/", Right: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ErrContains: "requires numeric operands"},
		{Name: "Div Error Str Num", InputNode: BinaryOpNode{Left: VariableNode{Name: "strABC"}, Operator: "/", Right: VariableNode{Name: "int3"}}, InitialVars: vars, WantErr: true, ErrContains: "requires numeric operands"},

		// ... (Modulo, Power, Unary Minus, Precedence tests remain the same) ...
		{Name: "Mod Int%Int", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(7)}, Operator: "%", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: int64(1), WantErr: false},
		{Name: "Mod Negative", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(-7)}, Operator: "%", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: int64(-1), WantErr: false},
		{Name: "Mod Error Float", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: float64(7.5)}, Operator: "%", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, WantErr: true, ErrContains: "requires integer operands"},
		{Name: "Mod By Int Zero", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "%", Right: VariableNode{Name: "int0"}}, InitialVars: vars, WantErr: true, ErrContains: "division by zero in modulo"},
		{Name: "Pow Int**Int", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "**", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: float64(8.0), WantErr: false},
		{Name: "Pow Int**Float", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(4)}, Operator: "**", Right: NumberLiteralNode{Value: float64(0.5)}}, InitialVars: vars, Expected: float64(2.0), WantErr: false},
		{Name: "Pow Float**Int", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: float64(2.5)}, Operator: "**", Right: NumberLiteralNode{Value: int64(2)}}, InitialVars: vars, Expected: float64(6.25), WantErr: false},
		{Name: "Pow Error Str", InputNode: BinaryOpNode{Left: VariableNode{Name: "int5"}, Operator: "**", Right: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ErrContains: "requires numeric operands"},
		{Name: "Unary Minus Int", InputNode: UnaryOpNode{Operator: "-", Operand: NumberLiteralNode{Value: int64(5)}}, InitialVars: vars, Expected: int64(-5), WantErr: false},
		{Name: "Unary Minus Float", InputNode: UnaryOpNode{Operator: "-", Operand: NumberLiteralNode{Value: float64(2.5)}}, InitialVars: vars, Expected: float64(-2.5), WantErr: false},
		{Name: "Unary Minus StrNum", InputNode: UnaryOpNode{Operator: "-", Operand: StringLiteralNode{Value: "10"}}, InitialVars: vars, Expected: int64(-10), WantErr: false},
		{Name: "Unary Minus Error Str", InputNode: UnaryOpNode{Operator: "-", Operand: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ErrContains: "requires a numeric operand"},
		{Name: "Prec Add Mul: 2 + 3 * 4", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "+", Right: BinaryOpNode{Left: NumberLiteralNode{Value: int64(3)}, Operator: "*", Right: NumberLiteralNode{Value: int64(4)}}}, InitialVars: vars, Expected: int64(14), WantErr: false},
		{Name: "Prec Mul Add: 2 * 3 + 4", InputNode: BinaryOpNode{Left: BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "*", Right: NumberLiteralNode{Value: int64(3)}}, Operator: "+", Right: NumberLiteralNode{Value: int64(4)}}, InitialVars: vars, Expected: int64(10), WantErr: false},
		{Name: "Prec Parens: (2 + 3) * 4", InputNode: BinaryOpNode{Left: BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "+", Right: NumberLiteralNode{Value: int64(3)}}, Operator: "*", Right: NumberLiteralNode{Value: int64(4)}}, InitialVars: vars, Expected: int64(20), WantErr: false},
		{Name: "Prec Pow Right Assoc: 2 ** 3 ** 2", InputNode: BinaryOpNode{Left: NumberLiteralNode{Value: int64(2)}, Operator: "**", Right: BinaryOpNode{Left: NumberLiteralNode{Value: int64(3)}, Operator: "**", Right: NumberLiteralNode{Value: int64(2)}}}, InitialVars: vars, Expected: float64(512), WantErr: false},
		{Name: "Prec Unary Minus High: -2 + 3", InputNode: BinaryOpNode{Left: UnaryOpNode{Operator: "-", Operand: NumberLiteralNode{Value: int64(2)}}, Operator: "+", Right: NumberLiteralNode{Value: int64(3)}}, InitialVars: vars, Expected: int64(1), WantErr: false},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel() // Optional
			runEvalExpressionTest(t, tc)
		})
	}
}
