// NeuroScript Version: 0.8.0
// File version: 20
// Purpose: Adds regression tests for accessing map-by-value and list-by-pointer.
// filename: pkg/eval/access_test.go
// nlines: 130
// risk_rating: HIGH

package eval

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
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
func (m *mockRuntime) GetToolSpec(toolName types.FullName) (ToolSpec, bool) {
	return ToolSpec{}, false
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
		// DEBUG
		fmt.Printf("--- RUNNING TEST: %s ---\n", tc.Name)
		fmt.Printf("DEBUG: Initial variables: %v\n", tc.InitialVars)
		fmt.Printf("DEBUG: INPUT NODE: %#v\n", tc.InputNode)

		result, err := Expression(mock, tc.InputNode)

		// DEBUG
		fmt.Printf("DEBUG: RESULT: %#v (%T)\n", result, result)
		fmt.Printf("DEBUG: ERROR: %v\n", err)
		fmt.Printf("DEBUG: EXPECTED: %#v (%T)\n", tc.Expected, tc.Expected)
		fmt.Printf("--- END TEST: %s ---\n\n", tc.Name)

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
	listPtr := &lang.ListValue{Value: []lang.Value{
		lang.StringValue{Value: "ptr_apple"},
		lang.NumberValue{Value: 142},
	}}

	initialVars := map[string]lang.Value{
		"myList": lang.ListValue{Value: []lang.Value{
			lang.StringValue{Value: "apple"},
			lang.NumberValue{Value: 42},
		}},
		"myListPtr": listPtr, // ADDED
		"myMapPtr": lang.NewMapValue(map[string]lang.Value{ // This is *MapValue
			"key1": lang.StringValue{Value: "value1_ptr"},
		}),
		"myMapVal": lang.MapValue{Value: map[string]lang.Value{ // This is MapValue
			"key1": lang.StringValue{Value: "value1_val"},
		}},
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
			Name:        "List Ptr Access Valid Index 0", // ADDED
			InputNode:   &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myListPtr"}, Accessor: &ast.NumberLiteralNode{Value: int64(0)}},
			InitialVars: initialVars,
			Expected:    lang.StringValue{Value: "ptr_apple"},
		},
		{
			Name:            "List Access Index Out of Bounds (High)",
			InputNode:       &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myList"}, Accessor: &ast.NumberLiteralNode{Value: int64(99)}},
			InitialVars:     initialVars,
			WantErr:         true,
			ExpectedErrorIs: lang.ErrListIndexOutOfBounds,
		},
		{
			Name:        "Map Ptr Access Valid Key",
			InputNode:   &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myMapPtr"}, Accessor: &ast.StringLiteralNode{Value: "key1"}},
			InitialVars: initialVars,
			Expected:    lang.StringValue{Value: "value1_ptr"},
		},
		{
			Name:        "Map Val Access Valid Key", // ADDED
			InputNode:   &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myMapVal"}, Accessor: &ast.StringLiteralNode{Value: "key1"}},
			InitialVars: initialVars,
			Expected:    lang.StringValue{Value: "value1_val"},
		},
		{
			Name:        "Map Val Access Non-existent Key", // ADDED
			InputNode:   &ast.ElementAccessNode{Collection: &ast.VariableNode{Name: "myMapVal"}, Accessor: &ast.StringLiteralNode{Value: "badkey"}},
			InitialVars: initialVars,
			Expected:    &lang.NilValue{},
		},
	}

	for _, tc := range testCases {
		runLocalExpressionTest(t, tc)
	}
}
