// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Corrected the call to the test helper function and moved internal tests to a separate file.
// filename: pkg/interpreter/evaluation_functions_test.go
// nlines: 100
// risk_rating: LOW

package interpreter_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/testutil"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestMathFunctions(t *testing.T) {
	rawVars := map[string]interface{}{
		"e":       float64(math.E),
		"ten":     int64(10),
		"pi_o_2":  float64(math.Pi / 2.0),
		"one":     int64(1),
		"zero":    int64(0),
		"neg_one": int64(-1),
		"two":     float64(2.0),
		"str_abc": "abc",
	}

	vars := make(map[string]lang.Value, len(rawVars))
	for k, v := range rawVars {
		w, err := lang.Wrap(v)
		if err != nil {
			panic(fmt.Sprintf("test setup: cannot wrap %q: %v", k, err))
		}
		vars[k] = w
	}

	dummyPos := &types.Position{Line: 1, Column: 1}

	testCases := []testutil.EvalTestCase{
		{Name: "LN(e)", InputNode: &ast.CallableExprNode{Target: ast.CallTarget{Name: "ln"}, Arguments: []ast.Expression{&ast.VariableNode{Name: "e"}}}, InitialVars: vars, Expected: lang.NumberValue{Value: 1.0}},
		{Name: "LN(1)", InputNode: &ast.CallableExprNode{Target: ast.CallTarget{Name: "ln"}, Arguments: []ast.Expression{&ast.VariableNode{Name: "one"}}}, InitialVars: vars, Expected: lang.NumberValue{Value: 0.0}},
		{Name: "LN(0)", InputNode: &ast.CallableExprNode{Target: ast.CallTarget{Name: "ln"}, Arguments: []ast.Expression{&ast.VariableNode{Name: "zero"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: lang.ErrInvalidFunctionArgument},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if call, ok := tc.InputNode.(*ast.CallableExprNode); ok {
				call.BaseNode.StartPos = dummyPos
				call.Target.BaseNode.StartPos = dummyPos
				for _, arg := range call.Arguments {
					if vn, okV := arg.(*ast.VariableNode); okV {
						vn.BaseNode.StartPos = dummyPos
					}
				}
			}
			// FIX: Corrected the function call.
			testutil.ExpressionTest(t, tc)
		})
	}
}
