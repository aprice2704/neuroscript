// NeuroScript Version: 0.3.5
// File version: 11
// Purpose: Updated tests to expect core.Value types, aligning with new compliant testing helpers.
// filename: pkg/runtime/evaluation_logical_bitwise_test.go

package runtime

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
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
			InputNode:   &ast.UnaryOpNode{Operator: "NOT", Operand: &ast.VariableNode{Name: "trueVar"}},
			InitialVars: vars,
			Expected:    BoolValue{Value: false},
		},
		{
			Name:        "Bitwise AND",
			InputNode:   &ast.BinaryOpNode{Left: &ast.VariableNode{Name: "num5"}, Operator: "&", Right: &ast.VariableNode{Name: "num3"}},
			InitialVars: vars,
			Expected:    NumberValue{Value: 1}, // 101 & 011 = 001
		},
		{
			Name:            "Bitwise AND Error Float",
			InputNode:       &ast.BinaryOpNode{Left: &ast.NumberLiteralNode{Value: 3.14}, Operator: "&", Right: &ast.VariableNode{Name: "num5"}},
			InitialVars:     vars,
			WantErr:         true,
			ExpectedErrorIs: ErrInvalidOperandTypeInteger,
		},
	}

	for _, tc := range testCases {
		runEval.ExpressionTest(t, tc)
	}
}
