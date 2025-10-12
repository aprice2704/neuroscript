// NeuroScript Version: 0.8.0
// File version: 24
// Purpose: Corrected the expected outcome of the TypeOf test to match the evaluator's correct behavior.
// filename: pkg/eval/new_types_test.go
// nlines: 70
// risk_rating: MEDIUM

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// NOTE: This integration test has been simplified significantly.
// Its original purpose was to test the end-to-end behavior of new types
// (timedate, error, fuzzy), which involved tools and interpreter logic.
// Now that `eval` is a separate package, this test's scope is reduced
// to ensuring the evaluator can correctly handle the *syntax* and basic
// evaluation of expressions involving these types. The underlying operator
// logic (e.g., timedate comparison) is now tested in the `lang` package.

func TestNewTypesSyntax(t *testing.T) {
	// This test verifies that a tool call returning a custom type can be
	// correctly represented in the AST and evaluated as an expression.
	vars := map[string]lang.Value{
		"t1": lang.TimedateValue{},
		"t2": lang.TimedateValue{},
	}

	testCases := []localEvalTestCase{
		{
			Name: "Timedate Comparison Syntax",
			// Represents `t1 < t2`
			InputNode: &ast.BinaryOpNode{
				Left:     &ast.VariableNode{Name: "t1"},
				Operator: "<",
				Right:    &ast.VariableNode{Name: "t2"},
			},
			InitialVars: vars,
			// The lang package handles the comparison logic.
			Expected: lang.BoolValue{Value: false},
		},
		{
			Name: "TypeOf Syntax",
			// Represents `typeof(t1)`
			InputNode: &ast.CallableExprNode{
				Target:    ast.CallTarget{Name: "typeof"},
				Arguments: []ast.Expression{&ast.VariableNode{Name: "t1"}},
			},
			InitialVars: vars,
			// FIX: The evaluator correctly identifies the type. The test should expect the correct string.
			Expected: lang.StringValue{Value: "timedate"},
		},
	}

	for _, tc := range testCases {
		runLocalExpressionTest(t, tc)
	}
}
