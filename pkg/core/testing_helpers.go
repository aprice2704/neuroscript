// NeuroScript Version: 0.4.0
// File version: 11
// Purpose: Implemented a robust createTestStep helper to correctly build Step structs for various statement types, fixing numerous test failures.
// filename: pkg/core/testing_helpers.go

package core

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

var dummyPos = &Position{Line: 1, Column: 1, File: "test"}

type EvalTestCase struct {
	Name            string
	InputNode       Expression
	InitialVars     map[string]Value
	LastResult      Value
	Expected        Value
	WantErr         bool
	ExpectedErrorIs error
}

type executeStepsTestCase struct {
	name            string
	inputSteps      []Step
	initialVars     map[string]Value
	lastResult      Value
	expectError     bool
	ExpectedErrorIs error
	errContains     string
	expectedResult  Value
	expectedVars    map[string]Value
}

type ValidationTestCase struct {
	Name          string
	InputArgs     []interface{}
	ExpectedError error
}

func runValidationTestCases(t *testing.T, toolName string, testCases []ValidationTestCase) {
	t.Helper()
	interp, _ := NewDefaultTestInterpreter(t)
	tool, ok := interp.ToolRegistry().GetTool(toolName)
	if !ok {
		t.Fatalf("Tool %s not found in registry", toolName)
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := tool.Func(interp, tc.InputArgs)
			if !errors.Is(err, tc.ExpectedError) {
				t.Errorf("Expected error [%v], but got [%v]", tc.ExpectedError, err)
			}
		})
	}
}

func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
}

func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	t.Run(tc.Name, func(t *testing.T) {
		i, err := NewTestInterpreter(t, tc.InitialVars, tc.LastResult)
		if err != nil {
			t.Fatalf("NewTestInterpreter failed: %v", err)
		}

		got, err := i.evaluateExpression(tc.InputNode)

		if (err != nil) != tc.WantErr {
			t.Fatalf("Test %q: Error expectation mismatch.\n  got err = %v, wantErr %t", tc.Name, err, tc.WantErr)
		}
		if tc.WantErr {
			if tc.ExpectedErrorIs != nil && !errors.Is(err, tc.ExpectedErrorIs) {
				t.Fatalf("Test %q: Expected error wrapping [%v], but got [%v]", tc.Name, tc.ExpectedErrorIs, err)
			}
			return
		}
		if err != nil {
			t.Fatalf("Test %q: unexpected error: %v", tc.Name, err)
		}

		if !reflect.DeepEqual(got, tc.Expected) {
			t.Fatalf(`Test %q: Result mismatch.
			Expected:    %#v (%T)
			Got:         %#v (%T)`,
				tc.Name, tc.Expected, tc.Expected, got, got)
		}
	})
}

func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		i, err := NewTestInterpreter(t, tc.initialVars, tc.lastResult)
		if err != nil {
			t.Fatalf("NewTestInterpreter failed: %v", err)
		}

		finalResultFromExec, wasReturn, _, err := i.executeSteps(tc.inputSteps, false, nil)

		if (err != nil) != tc.expectError {
			t.Fatalf("Test %q: Unexpected error state. Got err: %v, wantErr: %t", tc.name, err, tc.expectError)
		}
		if tc.expectError {
			if tc.ExpectedErrorIs != nil && !errors.Is(err, tc.ExpectedErrorIs) {
				t.Fatalf("Test %q: Expected error wrapping: [%v], got: [%v]", tc.name, tc.ExpectedErrorIs, err)
			}
			return
		}

		var actualResult Value
		if wasReturn {
			actualResult = finalResultFromExec
		} else {
			actualResult = i.lastCallResult
		}

		if !reflect.DeepEqual(actualResult, tc.expectedResult) {
			t.Errorf("Test %q: Final execution result mismatch:\n  Expected: %#v (%T)\n  Got:      %#v (%T)", tc.name, tc.expectedResult, tc.expectedResult, actualResult, actualResult)
		}

		if tc.expectedVars != nil {
			for expectedKey, expectedValue := range tc.expectedVars {
				gotValue, ok := i.GetVariable(expectedKey)
				if !ok {
					t.Errorf("Test %q: Expected variable '%s' not found in final interpreter state", tc.name, expectedKey)
					continue
				}
				if !reflect.DeepEqual(gotValue, expectedValue) {
					t.Errorf("Test %q: Variable state mismatch for key '%s':\n  Expected: %#v (%T)\n  Got:      %#v (%T)", tc.name, expectedKey, expectedValue, expectedValue, gotValue, gotValue)
				}
			}
		}
	})
}

// createTestStep is a robust helper for creating Step structs for tests.
func createTestStep(stepType, target string, value Expression, callArgs []Expression) Step {
	s := Step{Pos: dummyPos, Type: stepType}
	switch strings.ToLower(stepType) {
	case "set":
		s.LValue = &LValueNode{Identifier: target, Pos: s.Pos}
		s.Value = value
	case "emit", "return":
		s.Values = []Expression{value}
	case "must":
		// 'must' can operate on a condition or a 'last' node.
		s.Cond = value
	case "call":
		s.Call = &CallableExprNode{
			Pos:       dummyPos,
			Target:    CallTarget{Pos: dummyPos, Name: target, IsTool: true},
			Arguments: callArgs,
		}
	default:
		// Fallback for simple cases, though explicit cases are better.
		s.Value = value
	}
	return s
}

func createIfStep(pos *Position, condNode Expression, thenSteps, elseSteps []Step) Step {
	return Step{Pos: pos, Type: "if", Cond: condNode, Body: thenSteps, Else: elseSteps}
}

func NewTestStringLiteral(val string) *StringLiteralNode {
	return &StringLiteralNode{Pos: dummyPos, Value: val}
}

func NewTestNumberLiteral(val float64) *NumberLiteralNode {
	return &NumberLiteralNode{Pos: dummyPos, Value: val}
}

func NewTestBooleanLiteral(val bool) *BooleanLiteralNode {
	return &BooleanLiteralNode{Pos: dummyPos, Value: val}
}

func NewTestVariableNode(name string) *VariableNode {
	return &VariableNode{Pos: dummyPos, Name: name}
}

func DebugDumpVariables(i *Interpreter, t *testing.T) {
	i.variablesMu.RLock()
	defer i.variablesMu.RUnlock()
	t.Log("--- INTERPRETER VARIABLE DUMP ---")
	// This would need updating to get all keys from the interpreter's variable map.
	t.Log("  (variable dumping needs update to get all keys)")
	t.Log("--- END VARIABLE DUMP ---")
}
