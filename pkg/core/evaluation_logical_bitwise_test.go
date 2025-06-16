// NeuroScript Version: 0.3.5
// File version: 11
// Purpose: Updated tests to expect core.Value types, aligning with new compliant testing helpers.
// filename: pkg/core/evaluation_logical_bitwise_test.go

package core

import (
	"testing"
)

func TestLogicalBitwiseOps(t *testing.T) {
	vars := map[string]Value{
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
			Expected:    BoolValue{Value: false},
		},
		{
			Name:        "Bitwise AND",
			InputNode:   &BinaryOpNode{Left: &VariableNode{Name: "num5"}, Operator: "&", Right: &VariableNode{Name: "num3"}},
			InitialVars: vars,
			Expected:    NumberValue{Value: 1}, // 101 & 011 = 001
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
