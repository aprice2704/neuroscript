// NeuroScript Version: 0.5.2
// File version: 10
// Purpose: Updated tests to expect lang.Value types, aligning with new compliant testing helpers.
// filename: pkg/interpreter/evaluation_arithmetic_test.go
// nlines: 45
// risk_rating: LOW

package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/testutil"
)

func TestArithmeticOps(t *testing.T) {
	vars := map[string]lang.Value{
		"int5":		lang.NumberValue{Value: 5},
		"int3":		lang.NumberValue{Value: 3},
		"float2_5":	lang.NumberValue{Value: 2.5},
		"strABC":	lang.StringValue{Value: "ABC"},
		"int0":		lang.NumberValue{Value: 0},
	}

	testCases := []testutil.EvalTestCase{
		{
			Name:		"Add Int+Int",
			InputNode:	&ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "+", Right: &ast.VariableNode{Name: "int3"}},
			InitialVars:	vars,
			Expected:	lang.NumberValue{Value: 8},
		},
		{
			Name:		"Add Int+Float",
			InputNode:	&ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "+", Right: &ast.VariableNode{Name: "float2_5"}},
			InitialVars:	vars,
			Expected:	lang.NumberValue{Value: 7.5},
		},
		{
			Name:		"Add Int+StrABC",
			InputNode:	&ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "+", Right: &ast.VariableNode{Name: "strABC"}},
			InitialVars:	vars,
			Expected:	lang.StringValue{Value: "5ABC"},
		},
		{
			Name:			"Div By Int Zero",
			InputNode:		&ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "/", Right: &ast.VariableNode{Name: "int0"}},
			InitialVars:		vars,
			WantErr:		true,
			ExpectedErrorIs:	lang.ErrDivisionByZero,
		},
	}

	for _, tc := range testCases {
		testutil.ExpressionTest(t, tc)
	}
}