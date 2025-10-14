// NeuroScript Version: 0.8.0
// File version: 3.0.2
// Purpose: Added test cases for string equality using variables to ensure the evaluator handles them correctly.
// filename: pkg/eval/comparison_test.go
// nlines: 83
// risk_rating: LOW

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestEvaluateComparison(t *testing.T) {
	vars := map[string]lang.Value{
		"s1": lang.StringValue{Value: "hello"},
		"s2": lang.StringValue{Value: "hello"},
		"s3": lang.StringValue{Value: "world"},
	}

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
			Name: "Equal Strings Literals",
			InputNode: &ast.BinaryOpNode{
				Left:     &ast.StringLiteralNode{Value: "hello"},
				Operator: "==",
				Right:    &ast.StringLiteralNode{Value: "hello"},
			},
			Expected: lang.BoolValue{Value: true},
		},
		{
			Name: "Equal Strings Variables",
			InputNode: &ast.BinaryOpNode{
				Left:     &ast.VariableNode{Name: "s1"},
				Operator: "==",
				Right:    &ast.VariableNode{Name: "s2"},
			},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: true},
		},
		{
			Name: "Unequal Strings Variables",
			InputNode: &ast.BinaryOpNode{
				Left:     &ast.VariableNode{Name: "s1"},
				Operator: "!=",
				Right:    &ast.VariableNode{Name: "s3"},
			},
			InitialVars: vars,
			Expected:    lang.BoolValue{Value: true},
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
