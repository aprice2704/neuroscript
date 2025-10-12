// NeuroScript Version: 0.8.0
// File version: 11
// Purpose: Refactored to use a local mock runtime for isolated testing.
// filename: pkg/eval/arithmetic_test.go
// nlines: 45
// risk_rating: LOW

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestArithmeticOps(t *testing.T) {
	vars := map[string]lang.Value{
		"int5":     lang.NumberValue{Value: 5},
		"int3":     lang.NumberValue{Value: 3},
		"float2_5": lang.NumberValue{Value: 2.5},
		"strABC":   lang.StringValue{Value: "ABC"},
		"int0":     lang.NumberValue{Value: 0},
	}

	testCases := []localEvalTestCase{
		{
			Name:        "Add Int+Int",
			InputNode:   &ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "+", Right: &ast.VariableNode{Name: "int3"}},
			InitialVars: vars,
			Expected:    lang.NumberValue{Value: 8},
		},
		{
			Name:        "Add Int+Float",
			InputNode:   &ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "+", Right: &ast.VariableNode{Name: "float2_5"}},
			InitialVars: vars,
			Expected:    lang.NumberValue{Value: 7.5},
		},
		{
			Name:        "Add Int+StrABC",
			InputNode:   &ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "+", Right: &ast.VariableNode{Name: "strABC"}},
			InitialVars: vars,
			Expected:    lang.StringValue{Value: "5ABC"},
		},
		{
			Name:            "Div By Int Zero",
			InputNode:       &ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "/", Right: &ast.VariableNode{Name: "int0"}},
			InitialVars:     vars,
			WantErr:         true,
			ExpectedErrorIs: lang.ErrDivisionByZero,
		},
	}

	for _, tc := range testCases {
		runLocalExpressionTest(t, tc)
	}
}
