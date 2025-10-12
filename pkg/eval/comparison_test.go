// NeuroScript Version: 0.8.0
// File version: 3.0.0
// Purpose: Refactored to use the local mock runtime for isolated testing of comparison expressions.
// filename: pkg/eval/comparison_test.go
// nlines: 45
// risk_rating: LOW

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestEvaluateComparison(t *testing.T) {
	testCases := []localEvalTestCase{
		{
			Name: "Equal Numbers",
			InputNode: &ast.BinaryOpNode{
				Left:     &ast.NumberLiteralNode{Value: 5.0},
				Operator: "==",
				Right:    &ast.NumberLiteralNode{Value: 5.0},
			},
			Expected: lang.BoolValue{Value: true},
		},
		{
			Name: "Unequal Strings",
			InputNode: &ast.BinaryOpNode{
				Left:     &ast.StringLiteralNode{Value: "hello"},
				Operator: "!=",
				Right:    &ast.StringLiteralNode{Value: "world"},
			},
			Expected: lang.BoolValue{Value: true},
		},
		{
			Name: "Greater Than",
			InputNode: &ast.BinaryOpNode{
				Left:     &ast.NumberLiteralNode{Value: 10.0},
				Operator: ">",
				Right:    &ast.NumberLiteralNode{Value: 5.0},
			},
			Expected: lang.BoolValue{Value: true},
		},
	}

	for _, tc := range testCases {
		runLocalExpressionTest(t, tc)
	}
}
