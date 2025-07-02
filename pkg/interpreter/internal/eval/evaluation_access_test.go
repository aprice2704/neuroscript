// NeuroScript Version: 0.3.5
// File version: 12
// Purpose: Aligned tests with compliant helpers by expecting  Value types instead of raw primitives.
// filename: pkg/interpreter/internal/eval/evaluation_access_test.go

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/testutil"
)

func TestEvaluateElementAccess(t *testing.T) {
	initialVars := map[string]lang.Value{
		"myList": lang.NewListValue([]lang.Value{
			lang.StringValue{Value: "apple"},
			lang.NumberValue{Value: 42},
		}),
		"myMap": lang.NewMapValue(map[string]lang.Value{
			"key1": lang.StringValue{Value: "value1"},
		}),
		"idx":	lang.NumberValue{Value: 1},
	}

	testCases := []testutil.EvalTestCase{
		{
			Name:		"List Access Valid Index 0",
			InputNode:	&ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myList"}, Accessor: &ast.NumberLiteralNode{Value: int64(0)}},
			InitialVars:	initialVars,
			Expected:	lang.StringValue{Value: "apple"},
		},
		{
			Name:		"List Access Valid Index Var",
			InputNode:	&ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myList"}, Accessor: &ast.VariableNode{Name: "idx"}},
			InitialVars:	initialVars,
			Expected:	lang.NumberValue{Value: 42},
		},
		{
			Name:			"List Access Index Out of Bounds (High)",
			InputNode:		&ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myList"}, Accessor: &ast.NumberLiteralNode{Value: int64(99)}},
			InitialVars:		initialVars,
			WantErr:		true,
			ExpectedErrorIs:	lang.ErrListIndexOutOfBounds,
		},
		{
			Name:		"Map Access Valid Key",
			InputNode:	&ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myMap"}, Accessor: &ast.StringLiteralNode{Value: "key1"}},
			InitialVars:	initialVars,
			Expected:	lang.StringValue{Value: "value1"},
		},
	}

	for _, tc := range testCases {
		testutil.runEval.ExpressionTest(t, tc)
	}
}