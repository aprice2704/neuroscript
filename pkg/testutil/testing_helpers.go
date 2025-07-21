// NeuroScript Version: 0.5.2
// File version: 23
// Purpose: Removed legacy 'Pos' field assignments from all test helpers to align with the updated AST.
// filename: pkg/testutil/testing_helpers.go

package testutil

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

var dummyPos = &types.Position{Line: 1, Column: 1, File: "test"}

// --- Test Case Structs (Exported) ---

type ExecuteStepsTestCase struct {
	Name            string
	InputSteps      []ast.Step
	InitialVars     map[string]lang.Value
	LastResult      lang.Value
	ExpectError     bool
	ExpectedErrorIs error
	ErrContains     string
	ExpectedResult  lang.Value
	ExpectedVars    map[string]lang.Value
}

type EvalTestCase struct {
	Name            string
	InputNode       ast.Expression
	InitialVars     map[string]lang.Value
	LastResult      lang.Value
	Expected        lang.Value
	WantErr         bool
	ExpectedErrorIs error
}

// --- Generic AST Creation Helpers (Exported) ---

func NewTestStringLiteral(val string) *ast.StringLiteralNode {
	return &ast.StringLiteralNode{Value: val}
}

func NewTestNumberLiteral(val float64) *ast.NumberLiteralNode {
	return &ast.NumberLiteralNode{Value: val}
}

func NewTestBooleanLiteral(val bool) *ast.BooleanLiteralNode {
	return &ast.BooleanLiteralNode{Value: val}
}

func NewVariableNode(name string) *ast.VariableNode {
	return &ast.VariableNode{Name: name}
}

// --- Test Execution Helpers (Exported) ---

func NewTestInterpreter(t *testing.T, initialVars map[string]lang.Value, lastResult lang.Value) (*interpreter.Interpreter, error) {
	t.Helper()
	testLogger := logging.NewTestLogger(t)
	sandboxDir := t.TempDir()

	interp := interpreter.NewInterpreter(
		interpreter.WithLogger(testLogger),
		interpreter.WithSandboxDir(sandboxDir),
	)

	if initialVars != nil {
		for k, v := range initialVars {
			if err := interp.SetInitialVariable(k, v); err != nil {
				return nil, fmt.Errorf("failed to set initial variable %q: %w", k, err)
			}
		}
	}

	if lastResult != nil {
		interp.SetLastResult(lastResult)
	}

	if err := tool.RegisterCoreTools(interp.ToolRegistry()); err != nil {
		return nil, fmt.Errorf("failed to register core tools for test interpreter: %w", err)
	}
	return interp, nil
}

// ExpressionTest runs a single expression evaluation test case.
func ExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	t.Run(tc.Name, func(t *testing.T) {
		i, err := NewTestInterpreter(t, tc.InitialVars, tc.LastResult)
		if err != nil {
			t.Fatalf("NewTestInterpreter failed: %v", err)
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

func RunExecuteStepsTest(t *testing.T, tc ExecuteStepsTestCase) {
	t.Helper()
	t.Run(tc.Name, func(t *testing.T) {
		i, err := NewTestInterpreter(t, tc.InitialVars, tc.LastResult)
		if err != nil {
			t.Fatalf("NewTestInterpreter failed: %v", err)
		}

		finalResultFromExec, wasReturn, _, err := i.RunSteps(tc.InputSteps)

		if (err != nil) != tc.ExpectError {
			t.Fatalf("Test %q: Unexpected error state. Got err: %v, wantErr: %t", tc.Name, err, tc.ExpectError)
		}
		if tc.ExpectError {
			if tc.ExpectedErrorIs != nil && !errors.Is(err, tc.ExpectedErrorIs) {
				t.Fatalf("Test %q: Expected error wrapping: [%v], got: [%v]", tc.Name, tc.ExpectedErrorIs, err)
			}
			return
		}

		var actualResult lang.Value
		if wasReturn {
			actualResult = finalResultFromExec
		} else {
			actualResult = i.GetLastResult()
		}

		if !reflect.DeepEqual(actualResult, tc.ExpectedResult) {
			t.Errorf("Test %q: Final execution result mismatch:\n Expected: %#v (%T)\n Got:      %#v (%T)", tc.Name, tc.ExpectedResult, tc.ExpectedResult, actualResult, actualResult)
		}

		if tc.ExpectedVars != nil {
			for expectedKey, expectedValue := range tc.ExpectedVars {
				gotValue, ok := i.GetVariable(expectedKey)
				if !ok {
					t.Errorf("Test %q: Expected variable '%s' not found", tc.Name, expectedKey)
					continue
				}
				if !reflect.DeepEqual(gotValue, expectedValue) {
					t.Errorf("Test %q: Variable state mismatch for key '%s':\n Expected: %#v (%T)\n Got:      %#v (%T)", tc.Name, expectedKey, expectedValue, expectedValue, gotValue, gotValue)
				}
			}
		}
	})
}
