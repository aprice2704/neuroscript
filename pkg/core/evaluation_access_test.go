// NeuroScript Version: 0.3.5
// File version: 11
// Purpose: Reverted tests to expect native Go types, as comparison is handled by the test runner.
// filename: pkg/core/evaluation_access_test.go

package core

import (
	"testing"
)

func TestEvaluateElementAccess(t *testing.T) {
	initialVars := map[string]interface{}{
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
			Expected:    "apple",
		},
		{
			Name:        "List Access Valid Index Var",
			InputNode:   &ElementAccessNode{Collection: &VariableNode{Name: "myList"}, Accessor: &VariableNode{Name: "idx"}},
			InitialVars: initialVars,
			Expected:    float64(42),
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
			Expected:    "value1",
		},
	}

	for _, tc := range testCases {
		runEvalExpressionTest(t, tc)
	}
}
