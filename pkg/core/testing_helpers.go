// NeuroScript Version: 0.4.0
// File version: 6
// Purpose: Patched runEvalExpressionTest to correctly wrap all primitives before injection, fixing a panic.
// filename: pkg/core/testing_helpers.go
// nlines: 216
// risk_rating: HIGH

package core

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const floatTolerance = 1e-9

var dummyPos = &Position{Line: 1, Column: 1, File: "test"}

type EvalTestCase struct {
	Name            string
	InputNode       interface{}
	InitialVars     map[string]interface{}
	InitVarsVal     map[string]Value
	LastResult      interface{}
	Expected        interface{}
	WantErr         bool
	ExpectedErrorIs error
}

type executeStepsTestCase struct {
	name            string
	inputSteps      []Step
	initialVars     map[string]interface{}
	lastResult      interface{}
	expectError     bool
	ExpectedErrorIs error
	errContains     string
	expectedResult  interface{}
	expectedVars    map[string]interface{}
}

// ValidationTestCase is for testing input validation of tool functions.
type ValidationTestCase struct {
	Name          string
	InputArgs     []interface{}
	ExpectedError error
}

// runValidationTestCases runs a set of validation test cases for a given tool.
// It is defined here to be accessible by all test files in the package.
func runValidationTestCases(t *testing.T, toolName string, testCases []ValidationTestCase) {
	t.Helper()
	interp, _ := NewDefaultTestInterpreter(t) // A dummy interpreter is fine for validation
	tool, ok := interp.ToolRegistry().GetTool(toolName)
	if !ok {
		t.Fatalf("Tool %s not found in registry", toolName)
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// The validation logic is now part of the tool function itself.
			// We call the tool's Func to test validation.
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

// deepEqualWithTolerance recursively compares two values, allowing for a small tolerance
// when comparing float64, int, or other numeric types.
func deepEqualWithTolerance(a, b interface{}) bool {
	// unwrap values first
	a = unwrapValue(a)
	b = unwrapValue(b)

	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Attempt numeric comparison first
	aFloat, aIsNum := toFloat64(a)
	bFloat, bIsNum := toFloat64(b)
	if aIsNum && bIsNum {
		if math.IsNaN(aFloat) && math.IsNaN(bFloat) {
			return true
		}
		return math.Abs(aFloat-bFloat) < floatTolerance
	}

	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	if valA.Kind() != valB.Kind() {
		return false
	}

	switch valA.Kind() {
	case reflect.Slice:
		if valA.Len() != valB.Len() {
			return false
		}
		for i := 0; i < valA.Len(); i++ {
			if !deepEqualWithTolerance(valA.Index(i).Interface(), valB.Index(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Map:
		if valA.Len() != valB.Len() {
			return false
		}
		for _, key := range valA.MapKeys() {
			valBValue := valB.MapIndex(key)
			if !valBValue.IsValid() || !deepEqualWithTolerance(valA.MapIndex(key).Interface(), valBValue.Interface()) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(a, b)
	}
}

// runEvalExpressionTest executes a single EvalTestCase.
// It ensures that no raw Go primitives leak into the interpreter,
// respecting the value-wrapping contract.
func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()

	// 1. Spin up a CLEAN interpreter. Do not pass raw primitives to the constructor.
	i, _ := NewTestInterpreter(t, nil, nil)

	// 2. Wrap and set the lastCallResult if it exists.
	if tc.LastResult != nil {
		wrappedLast, err := Wrap(tc.LastResult)
		if err != nil {
			t.Fatalf("test setup: cannot wrap LastResult (%T): %v", tc.LastResult, err)
		}
		i.lastCallResult = wrappedLast
	}

	// 3. Inject variables, respecting the contract (wrapping primitives).
	switch {
	case tc.InitVarsVal != nil: // This path is correct, uses pre-wrapped values.
		for k, v := range tc.InitVarsVal {
			i.SetVariable(k, v)
		}
	case tc.InitialVars != nil: // This legacy path now correctly wraps all primitives.
		for k, raw := range tc.InitialVars {
			w, err := Wrap(raw)
			if err != nil {
				t.Fatalf("test setup: cannot wrap %q (%T): %v", k, raw, err)
			}
			i.SetVariable(k, w)
		}
	}

	// 4. Evaluate expression node.
	got, err := i.evaluateExpression(tc.InputNode)

	// 5. Error expectation handling.
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

	// 6. Success path – unwrap result and compare.
	unwrappedGot := Unwrap(got)
	if !deepEqualWithTolerance(unwrappedGot, tc.Expected) {
		t.Fatalf(`Test %q: Result mismatch.
  Input:       %#v
  Vars(raw):   %#v
  Vars(val):   %#v
  Last:        %#v
  Expected:    %#v (%T)
  Got (raw):   %#v (%T)
  Unwrapped:   %#v (%T)`,
			tc.Name, tc.InputNode, tc.InitialVars, tc.InitVarsVal,
			tc.LastResult, tc.Expected, tc.Expected,
			got, got, unwrappedGot, unwrappedGot)
	}
}

func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		i, _ := NewTestInterpreter(t, tc.initialVars, tc.lastResult)
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

		var actualResult interface{}
		if wasReturn {
			actualResult = finalResultFromExec
		} else {
			actualResult = i.lastCallResult
		}

		if !deepEqualWithTolerance(actualResult, tc.expectedResult) {
			t.Errorf("Test %q: Final execution result mismatch:\n  Expected: %#v (%T)\n  Got:      %#v (%T)", tc.name, tc.expectedResult, tc.expectedResult, actualResult, actualResult)
		}

		if tc.expectedVars != nil {
			i.variablesMu.RLock()
			defer i.variablesMu.RUnlock()
			for expectedKey, expectedValue := range tc.expectedVars {
				gotValue, ok := i.variables[expectedKey]
				if !ok {
					t.Errorf("Test %q: Expected variable '%s' not found in final interpreter state", tc.name, expectedKey)
					continue
				}

				if !deepEqualWithTolerance(gotValue, expectedValue) {
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

func NewTestNumberLiteral(val interface{}) *NumberLiteralNode {
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
	if len(i.variables) == 0 {
		t.Log("  (no variables set)")
	} else {
		keys := make([]string, 0, len(i.variables))
		for k := range i.variables {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := i.variables[key]
			t.Logf("  - %s (%T) = %#v", key, value, value)
		}
	}
	t.Log("--- END VARIABLE DUMP ---")
}
