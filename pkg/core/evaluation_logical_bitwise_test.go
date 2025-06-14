// NeuroScript Version: 0.3.5
// File version: 10
// Purpose: Reverted tests to expect native Go types.
// filename: pkg/core/evaluation_logical_bitwise_test.go

package core

import (
	"testing"
)

func TestLogicalBitwiseOps(t *testing.T) {
	vars := map[string]interface{}{
		"trueVar":  BoolValue{Value: true},
		"falseVar": BoolValue{Value: false},
		"num3":     NumberValue{Value: 3}, // 011
		"num5":     NumberValue{Value: 5}, // 101
	}

	testCases := []EvalTestCase{
		{
			Name:        "NOT True",
			InputNode:   &UnaryOpNode{Operator: "NOT", Operand: &VariableNode{Name: "trueVar"}},
			InitialVars: vars,
			Expected:    false,
		},
		{
			Name:        "Bitwise AND",
			InputNode:   &BinaryOpNode{Left: &VariableNode{Name: "num5"}, Operator: "&", Right: &VariableNode{Name: "num3"}},
			InitialVars: vars,
			Expected:    float64(1), // 101 & 011 = 001
		},
		{
			Name:            "Bitwise AND Error Float",
			InputNode:       &BinaryOpNode{Left: &NumberLiteralNode{Value: 3.14}, Operator: "&", Right: &VariableNode{Name: "num5"}},
			InitialVars:     vars,
			WantErr:         true,
			ExpectedErrorIs: ErrInvalidOperandTypeInteger,
		},
	}

	for _, tc := range testCases {
		runEvalExpressionTest(t, tc)
	}
}
