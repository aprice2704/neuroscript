// NeuroScript Version: 0.5.2
// File version: 12
// Purpose: Updated tests to expect lang.Value types, aligning with new compliant testing helpers.
// filename: pkg/interpreter/evaluation_logical_bitwise_test.go
// nlines: 45
// risk_rating: LOW

package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/testutil"
)

func TestLogicalBitwiseOps(t *testing.T) {
	vars := map[string]lang.Value{
		"trueVar":	lang.BoolValue{Value: true},
		"falseVar":	lang.BoolValue{Value: false},
		"num3":		lang.NumberValue{Value: 3},	// 011
		"num5":		lang.NumberValue{Value: 5},	// 101
	}

	testCases := []testutil.EvalTestCase{
		{
			Name:		"NOT True",
			InputNode:	&ast.UnaryOpNode{Operator: "NOT", Operand: &ast.VariableNode{Name: "trueVar"}},
			InitialVars:	vars,
			Expected:	lang.BoolValue{Value: false},
		},
		{
			Name:		"Bitwise AND",
			InputNode:	&ast.BinaryOpNode{Left: &ast.VariableNode{Name: "num5"}, Operator: "&", Right: &ast.VariableNode{Name: "num3"}},
			InitialVars:	vars,
			Expected:	lang.NumberValue{Value: 1},	// 101 & 011 = 001
		},
		{
			Name:			"Bitwise AND Error Float",
			InputNode:		&ast.BinaryOpNode{Left: &ast.NumberLiteralNode{Value: 3.14}, Operator: "&", Right: &ast.VariableNode{Name: "num5"}},
			InitialVars:		vars,
			WantErr:		true,
			ExpectedErrorIs:	lang.ErrInvalidOperandTypeInteger,
		},
	}

	for _, tc := range testCases {
		testutil.ExpressionTest(t, tc)
	}
}