// NeuroScript Version: 0.3.5
// File version: 12
// Purpose: Aligned tests with compliant helpers by expecting core.Value types instead of raw primitives.
// filename: pkg/core/evaluation_access_test.go

package core

import (
	"testing"
)

func TestEvaluateElementAccess(t *testing.T) {
	initialVars := map[string]Value{
		"myList": NewListValue([]Value{
			StringValue{Value: "apple"},
			NumberValue{Value: 42},
		}),
		"myMap": NewMapValue(map[string]Value{
			"key1": StringValue{Value: "value1"},
		}),
		"idx": NumberValue{Value: 1},
	}

	testCases := []EvalTestCase{
		{
			Name:        "List Access Valid Index 0",
			InputNode:   &ElementAccessNode{Collection: &VariableNode{Name: "myList"}, Accessor: &NumberLiteralNode{Value: int64(0)}},
			InitialVars: initialVars,
			Expected:    StringValue{Value: "apple"},
		},
		{
			Name:        "List Access Valid Index Var",
			InputNode:   &ElementAccessNode{Collection: &VariableNode{Name: "myList"}, Accessor: &VariableNode{Name: "idx"}},
			InitialVars: initialVars,
			Expected:    NumberValue{Value: 42},
		},
		{
			Name:            "List Access Index Out of Bounds (High)",
			InputNode:       &ElementAccessNode{Collection: &VariableNode{Name: "myList"}, Accessor: &NumberLiteralNode{Value: int64(99)}},
			InitialVars:     initialVars,
			WantErr:         true,
			ExpectedErrorIs: ErrListIndexOutOfBounds,
		},
		{
			Name:        "Map Access Valid Key",
			InputNode:   &ElementAccessNode{Collection: &VariableNode{Name: "myMap"}, Accessor: &StringLiteralNode{Value: "key1"}},
			InitialVars: initialVars,
			Expected:    StringValue{Value: "value1"},
		},
	}

	for _, tc := range testCases {
		runEvalExpressionTest(t, tc)
	}
}
