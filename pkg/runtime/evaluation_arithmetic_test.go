// NeuroScript Version: 0.3.2
// File version: 9
// Purpose: Updated tests to expect core.Value types, aligning with new compliant testing helpers.
// filename: pkg/runtime/evaluation_arithmetic_test.go

package runtime

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestArithmeticOps(t *testing.T) {
	vars := map[string]Value{
		"int5":     NumberValue{Value: 5},
		"int3":     NumberValue{Value: 3},
		"float2_5": NumberValue{Value: 2.5},
		"strABC":   StringValue{Value: "ABC"},
		"int0":     NumberValue{Value: 0},
	}

	testCases := []EvalTestCase{
		{
			Name:        "Add Int+Int",
			InputNode:   &ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "+", Right: &ast.VariableNode{Name: "int3"}},
			InitialVars: vars,
			Expected:    NumberValue{Value: 8},
		},
		{
			Name:        "Add Int+Float",
			InputNode:   &ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "+", Right: &ast.VariableNode{Name: "float2_5"}},
			InitialVars: vars,
			Expected:    NumberValue{Value: 7.5},
		},
		{
			Name:        "Add Int+StrABC",
			InputNode:   &ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "+", Right: &ast.VariableNode{Name: "strABC"}},
			InitialVars: vars,
			Expected:    StringValue{Value: "5ABC"},
		},
		{
			Name:            "Div By Int Zero",
			InputNode:       &ast.BinaryOpNode{Left: &ast.VariableNode{Name: "int5"}, Operator: "/", Right: &ast.VariableNode{Name: "int0"}},
			InitialVars:     vars,
			WantErr:         true,
			ExpectedErrorIs: ErrDivisionByZero,
		},
	}

	for _, tc := range testCases {
		runEval.ExpressionTest(t, tc)
	}
}
