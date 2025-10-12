// NeuroScript Version: 0.8.0
// File version: 13
// Purpose: Refactored to use a local mock runtime for isolated testing.
// filename: pkg/eval/logical_bitwise_test.go
// nlines: 45
// risk_rating: LOW

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestLogicalBitwiseOps(t *testing.T) {
	vars := map[string]lang.Value{
		"trueVar":  lang.BoolValue{Value: true},
		"falseVar": lang.BoolValue{Value: false},
		"num3":     lang.NumberValue{Value: 3}, // 011
		"num5":     lang.NumberValue{Value: 5}, // 101
	}

	testCases := []localEvalTestCase{
		{
			Name:        "NOT True",
			InputNode:   &ast.UnaryOpNode{Operator: "NOT", Operand: &ast.VariableNode{Name: "trueVar"}},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: false},
		},
		{
			Name:        "Bitwise AND",
			InputNode:   &ast.BinaryOpNode{Left: &ast.VariableNode{Name: "num5"}, Operator: "&", Right: &ast.VariableNode{Name: "num3"}},
			InitialVars: vars,
			Expected:    lang.NumberValue{Value: 1}, // 101 & 011 = 001
		},
		{
			Name:            "Bitwise AND Error Float",
			InputNode:       &ast.BinaryOpNode{Left: &ast.NumberLiteralNode{Value: 3.14}, Operator: "&", Right: &ast.VariableNode{Name: "num5"}},
			InitialVars:     vars,
			WantErr:         true,
			ExpectedErrorIs: lang.ErrInvalidOperandTypeInteger,
		},
	}

	for _, tc := range testCases {
		runLocalExpressionTest(t, tc)
	}
}
