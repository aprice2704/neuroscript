// NeuroScript Version: 0.8.0
// File version: 17
// Purpose: Refactored to use a local mock runtime, removing the circular dependency on the interpreter package.
// filename: pkg/eval/access_test.go
// nlines: 95
// risk_rating: HIGH

package eval

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// --- Local Mock Runtime and Test Helpers ---

type mockRuntime struct {
	vars map[string]lang.Value
}

func (m *mockRuntime) GetVariable(name string) (lang.Value, bool) {
	v, ok := m.vars[name]
	return v, ok
}
func (m *mockRuntime) ExecuteTool(toolName types.FullName, args map[string]lang.Value) (lang.Value, error) {
	return nil, errors.New("ExecuteTool not implemented in mock")
}
func (m *mockRuntime) RunProcedure(procName string, args ...lang.Value) (lang.Value, error) {
	return nil, errors.New("RunProcedure not implemented in mock")
}
func (m *mockRuntime) GetToolSpec(toolName types.FullName) (tool.ToolSpec, bool) {
	return tool.ToolSpec{}, false
}

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
		mock := &mockRuntime{vars: tc.InitialVars}
		result, err := Expression(mock, tc.InputNode)

		if (err != nil) != tc.WantErr {
			t.Fatalf("Expression() error = %v, wantErr %v", err, tc.WantErr)
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

func TestElementAccess(t *testing.T) {
	initialVars := map[string]lang.Value{
		"myList": lang.ListValue{Value: []lang.Value{
			lang.StringValue{Value: "apple"},
			lang.NumberValue{Value: 42},
		}},
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
