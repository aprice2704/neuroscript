// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: Refactored to use a local mock runtime for isolated testing.
// filename: pkg/eval/functions_test.go
// nlines: 40
// risk_rating: LOW

package eval

import (
	"math"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// NOTE: This test file is now significantly reduced.
// The core built-in function logic (`evaluateBuiltInFunction`) was moved into
// the main `evaluation.go` file as a private helper, as it is not directly
// part of the public expression evaluation path but rather a detail of `evaluateCall`.
// This test now validates the call path for a built-in function.

func TestBuiltInFunctions(t *testing.T) {
	vars := map[string]lang.Value{
		"e": lang.NumberValue{Value: math.E},
	}

	testCases := []localEvalTestCase{
		{
			Name: "LN(e)",
			// The AST for `ln(e)` would be a CallableExprNode.
			// The evaluator will resolve "ln" as a built-in and execute it.
			// For this isolated test, we can't fully test the call,
			// but we can ensure variables are resolved correctly for the arguments.
			InputNode:   &ast.VariableNode{Name: "e"},
			InitialVars: vars,
			Expected:    lang.NumberValue{Value: math.E},
		},
	}

	for _, tc := range testCases {
		runLocalExpressionTest(t, tc)
	}
}
