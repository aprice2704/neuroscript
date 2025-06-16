// NeuroScript Version: 0.4.0
// File version: 9
// Purpose: Reverted runValidationTestCases to pass raw primitives, correctly testing the ToolFunc signature without performing the outer adapter's wrapping step.
// filename: pkg/core/testing_helpers.go
// nlines: 215
// risk_rating: HIGH

package core

import (
	"errors"
	"fmt"
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

// ValidationTestCase is for testing input validation of tool functions.
type ValidationTestCase struct {
	Name          string
	InputArgs     []interface{}
	ExpectedError error
}

// runValidationTestCases runs a set of validation test cases for a given tool.
// It tests the ToolFunc directly by passing it raw primitive types, which
// aligns with the ToolFunc signature and tests the tool's internal validation logic.
func runValidationTestCases(t *testing.T, toolName string, testCases []ValidationTestCase) {
	t.Helper()
	interp, _ := NewDefaultTestInterpreter(t)
	tool, ok := interp.ToolRegistry().GetTool(toolName)
	if !ok {
		t.Fatalf("Tool %s not found in registry", toolName)
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Call the tool's Func with raw primitives, spreading the slice
			// into the variadic arguments, to match the ToolFunc signature.
			_, err := tool.Func(interp, tc.InputArgs)
			if !errors.Is(err, tc.ExpectedError) {
				t.Errorf("Expected error [%v], but got [%v]", tc.ExpectedError, err)
			}
		})
	}
}

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		message := fmt.Sprintf("Expected no error, but got: %v", err)
		if len(msgAndArgs) > 0 {
			format, ok := msgAndArgs[0].(string)
			if !ok {
				message += fmt.Sprintf("\nContext: %+v", msgAndArgs)
			} else {
				message += "\nContext: " + fmt.Sprintf(format, msgAndArgs[1:]...)
			}
		}
		t.Fatal(message)
	}
}

// runEvalExpressionTest executes a single EvalTestCase.
func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	t.Run(tc.Name, func(t *testing.T) {
		i, _ := NewTestInterpreter(t, nil, nil)

		if tc.LastResult != nil {
			i.lastCallResult = tc.LastResult
		}

		if tc.InitialVars != nil {
			for k, v := range tc.InitialVars {
				if err := i.SetVariable(k, v); err != nil {
					t.Fatalf("test setup: failed to set initial variable %q: %v", k, err)
				}
			}
		}

		got, err := i.evaluateExpression(tc.InputNode)

		if (err != nil) != tc.WantErr {
			t.Fatalf("Test %q: Error expectation mismatch.\n  got err = %v, wantErr %t",
				tc.Name, err, tc.WantErr)
		}
		if tc.WantErr {
			if tc.ExpectedErrorIs != nil && !errors.Is(err, tc.ExpectedErrorIs) {
				t.Fatalf("Test %q: Expected error wrapping [%v], but got [%v]",
					tc.Name, tc.ExpectedErrorIs, err)
			}
			return
		}
		if err != nil {
			t.Fatalf("Test %q: unexpected error: %v", tc.Name, err)
		}

		if !reflect.DeepEqual(got, tc.Expected) {
			t.Fatalf(`Test %q: Result mismatch.
	  Input:       %#v
	  Vars:        %#v
	  Last:        %#v
	  Expected:    %#v (%T)
	  Got:         %#v (%T)`,
				tc.Name, tc.InputNode, tc.InitialVars,
				tc.LastResult, tc.Expected, tc.Expected,
				got, got)
		}
	})
}

func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		i, _ := NewTestInterpreter(t, nil, nil)
		if tc.lastResult != nil {
			i.lastCallResult = tc.lastResult
		}
		if tc.initialVars != nil {
			for k, v := range tc.initialVars {
				if err := i.SetVariable(k, v); err != nil {
					t.Fatalf("test setup: failed to set initial variable %q: %v", k, err)
				}
			}
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

func createTestStep(stepType, targetOrLoopVarOrInto string, valueOrCollectionOrCall, _ interface{}) Step {
	s := Step{Pos: dummyPos, Type: stepType}
	switch strings.ToLower(stepType) {
	case "set":
		s.LValue = &LValueNode{Identifier: targetOrLoopVarOrInto, Pos: s.Pos}
		s.Value = valueOrCollectionOrCall.(Expression)
	default:
		if expr, ok := valueOrCollectionOrCall.(Expression); ok {
			s.Value = expr
		}
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
	// As GetVariable is now the safe way, we can't inspect the map directly.
	// This function would need a way to get all variable keys to be effective.
	// For now, it's left as a template.
	t.Log("  (variable dumping needs update to get all keys)")
	t.Log("--- END VARIABLE DUMP ---")
}
