// NeuroScript Version: 0.5.2
// File version: 16
// Purpose: Broke an import cycle by moving test helpers into the file, removing the need to import the testutil package.
// filename: pkg/interpreter/evaluation_access_test.go
// nlines: 95
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// --- Local Test Helpers to Avoid Import Cycle ---

type localEvalTestCase struct {
	Name            string
	InputNode       ast.Expression
	InitialVars     map[string]lang.Value
	Expected        lang.Value
	WantErr         bool
	ExpectedErrorIs error
}

func runLocalExpressionTest(t *testing.T, tc localEvalTestCase) {
	t.Helper()
	t.Run(tc.Name, func(t *testing.T) {
		i := NewInterpreter(WithLogger(logging.NewTestLogger(t)))
		if tc.InitialVars != nil {
			for k, v := range tc.InitialVars {
				i.SetInitialVariable(k, v)
			}
		}

		result, err := i.EvaluateExpression(tc.InputNode)

		if (err != nil) != tc.WantErr {
			t.Fatalf("EvaluateExpression() error = %v, wantErr %v", err, tc.WantErr)
		}
		if tc.WantErr {
			if tc.ExpectedErrorIs != nil && !errors.Is(err, tc.ExpectedErrorIs) {
				t.Fatalf("Expected error wrapping: [%v], got: [%v]", tc.ExpectedErrorIs, err)
			}
			return
		}

		if !reflect.DeepEqual(result, tc.Expected) {
			t.Errorf("Expression evaluation result mismatch:\n Expected: %#v (%T)\n      Got: %#v (%T)", tc.Expected, tc.Expected, result, result)
		}
	})
}

// --- Test ---

func TestEvaluateElementAccess(t *testing.T) {
	initialVars := map[string]lang.Value{
		"myList": lang.NewListValue([]lang.Value{
			lang.StringValue{Value: "apple"},
			lang.NumberValue{Value: 42},
		}),
		"myMap": lang.NewMapValue(map[string]lang.Value{
			"key1": lang.StringValue{Value: "value1"},
		}),
		"idx": lang.NumberValue{Value: 1},
	}

	testCases := []localEvalTestCase{
		{
			Name:        "List Access Valid Index 0",
			InputNode:   &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myList"}, Accessor: &ast.NumberLiteralNode{Value: int64(0)}},
			InitialVars: initialVars,
			Expected:    lang.StringValue{Value: "apple"},
		},
		{
			Name:        "List Access Valid Index Var",
			InputNode:   &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myList"}, Accessor: &ast.VariableNode{Name: "idx"}},
			InitialVars: initialVars,
			Expected:    lang.NumberValue{Value: 42},
		},
		{
			Name:            "List Access Index Out of Bounds (High)",
			InputNode:       &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myList"}, Accessor: &ast.NumberLiteralNode{Value: int64(99)}},
			InitialVars:     initialVars,
			WantErr:         true,
			ExpectedErrorIs: lang.ErrListIndexOutOfBounds,
		},
		{
			Name:        "Map Access Valid Key",
			InputNode:   &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myMap"}, Accessor: &ast.StringLiteralNode{Value: "key1"}},
			InitialVars: initialVars,
			Expected:    lang.StringValue{Value: "value1"},
		},
	}

	for _, tc := range testCases {
		runLocalExpressionTest(t, tc)
	}
}
