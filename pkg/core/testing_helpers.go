// NeuroScript Version: 0.3.5
// File version: 0.0.4 // Corrected result checking in runExecuteStepsTest for non-return cases.
// filename: pkg/core/testing_helpers.go
package core

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings" // Required for tc.errContains
	"testing"
	// Position is defined in ast.go
	// Expression is defined in ast.go
	// Step is defined in ast.go
)

// --- Shared Test Struct Definitions ---

// EvalTestCase defines the structure for testing evaluateExpression
type EvalTestCase struct {
	Name            string
	InputNode       interface{} // AST node or raw value (asserted to Expression by helper if needed)
	InitialVars     map[string]interface{}
	LastResult      interface{} // Mocked result of previous step if needed
	Expected        interface{} // Expected result of evaluation
	WantErr         bool
	ExpectedErrorIs error // Use sentinel error or nil for errors.Is checks
}

// executeStepsTestCase defines the structure for testing interp.executeSteps
type executeStepsTestCase struct {
	name            string
	inputSteps      []Step
	initialVars     map[string]interface{}
	lastResult      interface{} // For LAST keyword testing in initial state for NewTestInterpreter
	expectError     bool
	ExpectedErrorIs error       // Specific sentinel error expected
	errContains     string      // Substring to check in error message
	expectedResult  interface{} // Expected final result (from RETURN or LAST if no RETURN)
	expectedVars    map[string]interface{}
}

// --- Shared Helper Functions ---

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

const defaultTolerance = 1e-9

func deepEqualWithTolerance(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if reflect.TypeOf(a).Kind() == reflect.Float64 && reflect.TypeOf(b).Kind() == reflect.Float64 {
		aF64 := a.(float64)
		bF64 := b.(float64)
		if math.IsNaN(aF64) && math.IsNaN(bF64) {
			return true
		}
		return math.Abs(aF64-bF64) < defaultTolerance
	}
	aFloat, aIsNum := toFloat64(a)
	bFloat, bIsNum := toFloat64(b)
	if aIsNum && bIsNum {
		return math.Abs(aFloat-bFloat) < defaultTolerance
	}
	return reflect.DeepEqual(a, b)
}

func runEvalExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	interp, _ := NewTestInterpreter(t, tc.InitialVars, tc.LastResult)
	var inputExpr Expression
	if tc.InputNode != nil {
		var ok bool
		inputExpr, ok = tc.InputNode.(Expression)
		if !ok {
			t.Fatalf("Test setup error in %q: InputNode (%T) does not implement Expression and is not nil", tc.Name, tc.InputNode)
		}
	}
	got, err := interp.evaluateExpression(inputExpr)
	if (err != nil) != tc.WantErr {
		t.Errorf("Test %q: Error expectation mismatch. got err = %v, wantErr %v", tc.Name, err, tc.WantErr)
		if err != nil {
			t.Logf("Input: %#v, Vars: %#v, Last: %#v", tc.InputNode, tc.InitialVars, tc.LastResult)
		}
		return
	}
	if tc.WantErr {
		if err == nil { // Ensure an error was actually returned when one was expected
			t.Errorf("Test %q: Expected error, but got nil", tc.Name)
			return
		}
		if tc.ExpectedErrorIs != nil && !errors.Is(err, tc.ExpectedErrorIs) {
			t.Errorf("Test %q: Expected error wrapping [%v], but got [%v]", tc.Name, tc.ExpectedErrorIs, err)
		} else if tc.ExpectedErrorIs == nil {
			t.Logf("Test %q: Got expected error: %v", tc.Name, err)
		}
		return
	}
	if !deepEqualWithTolerance(got, tc.Expected) {
		t.Errorf("Test %q: Result mismatch.\nInput:    %#v\nVars:     %#v\nLast:     %#v\nExpected: %v (%T)\nGot:      %v (%T)",
			tc.Name, tc.InputNode, tc.InitialVars, tc.LastResult,
			tc.Expected, tc.Expected, got, got)
	}
}

func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	// Use tc.lastResult for initializing the interpreter's LAST state
	interp, _ := NewTestInterpreter(t, tc.initialVars, tc.lastResult)

	finalResultFromExec, wasReturn, _, err := interp.executeSteps(tc.inputSteps, false, nil)

	if tc.expectError {
		if err == nil {
			t.Errorf("Test %q: Expected an error, but got nil", tc.name)
			return
		}
		if tc.ExpectedErrorIs != nil {
			if !errors.Is(err, tc.ExpectedErrorIs) {
				t.Errorf("Test %q: Error mismatch.\nExpected error wrapping: [%v]\nGot error:               [%v]", tc.name, tc.ExpectedErrorIs, err)
			} else {
				t.Logf("Test %q: Got expected error wrapping [%v]: %v", tc.name, tc.ExpectedErrorIs, err)
			}
		}
		if tc.errContains != "" {
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("Test %q: Error message mismatch.\nExpected to contain: %q\nGot:                 %v", tc.name, tc.errContains, err.Error())
			}
		} else if tc.ExpectedErrorIs == nil { // Only log if no specific error was expected, but an error occurred.
			t.Logf("Test %q: Got expected error (no specific sentinel/contains check): %v", tc.name, err)
		}
		return
	}

	if err != nil {
		t.Errorf("Test %q: Unexpected error: %+v", tc.name, err)
		return
	}

	var actualExecResult interface{}
	logMessageDetail := ""

	if wasReturn {
		actualExecResult = finalResultFromExec
		logMessageDetail = "(from RETURN)"
		// Handle single-element slice unwrapping if expected is not a slice
		if resultSlice, ok := actualExecResult.([]interface{}); ok && len(resultSlice) == 1 {
			if _, expectedSlice := tc.expectedResult.([]interface{}); !expectedSlice {
				actualExecResult = resultSlice[0]
			}
		}
	} else {
		// If no RETURN, the "result" of the script execution for testing purposes
		// is considered to be the interpreter's lastCallResult.
		// finalResultFromExec will be nil in this case.
		actualExecResult = interp.lastCallResult
		logMessageDetail = "(from LAST)"
	}

	if !deepEqualWithTolerance(actualExecResult, tc.expectedResult) {
		t.Errorf("Test %q: Final execution result %s mismatch:\nExpected: %v (%T)\nGot:      %v (%T)",
			tc.name, logMessageDetail, tc.expectedResult, tc.expectedResult,
			actualExecResult, actualExecResult)
	}

	if tc.expectedVars != nil {
		cleanInterp, _ := NewDefaultTestInterpreter(t)
		baseVars := cleanInterp.variables
		for key, expectedValue := range tc.expectedVars {
			actualValue, exists := interp.variables[key]
			if !exists {
				if _, isBuiltIn := baseVars[key]; !isBuiltIn {
					t.Errorf("Test %q: Expected variable '%s' not found in final state", tc.name, key)
				}
				continue
			}
			if !deepEqualWithTolerance(actualValue, expectedValue) {
				t.Errorf("Test %q: Variable '%s' mismatch:\nExpected: %v (%T)\nGot:      %v (%T)",
					tc.name, key, expectedValue, expectedValue, actualValue, actualValue)
			}
		}
		extraVars := []string{}
		for k := range interp.variables {
			if _, isBuiltIn := baseVars[k]; isBuiltIn {
				if _, expected := tc.expectedVars[k]; !expected {
					continue
				}
			}
			if _, expected := tc.expectedVars[k]; !expected {
				extraVars = append(extraVars, k)
			}
		}
		if len(extraVars) > 0 {
			sort.Strings(extraVars)
			t.Errorf("Test %q: Unexpected variables found in final state: %v", tc.name, extraVars)
		}
	}
}

func createTestStep(stepType string, target string, valueOrValuesOrCall interface{}, _ignoredCallArg interface{}) Step {
	s := Step{Pos: &Position{Line: 1, Column: 1, File: "test"}, Type: stepType, Target: target}
	switch val := valueOrValuesOrCall.(type) {
	case *CallableExprNode:
		if stepType == "call" {
			s.Call = val
		} else {
			s.Value = val
		}
	case []Expression:
		s.Values = val
	case Expression:
		s.Value = val
	case nil:
		// Valid
	default:
		// Potential unhandled type, might cause issues if not Expression or nil
	}
	return s
}

func createIfStep(pos *Position, condNode Expression, thenSteps []Step, elseSteps []Step) Step {
	if condNode == nil {
		panic("createIfStep: test provided a nil condNode argument")
	}
	return Step{
		Pos:  pos,
		Type: "if",
		Cond: condNode,
		Body: thenSteps,
		Else: elseSteps,
	}
}

func createWhileStep(pos *Position, condNode Expression, bodySteps []Step) Step {
	if condNode == nil {
		panic("createWhileStep: test provided a nil condNode argument")
	}
	return Step{
		Pos:  pos,
		Type: "while",
		Cond: condNode,
		Body: bodySteps,
	}
}

func createForStep(pos *Position, loopVar string, collectionNode Expression, bodySteps []Step) Step {
	if collectionNode == nil {
		panic("createForStep: test provided a nil collectionNode argument")
	}
	return Step{
		Pos:    pos,
		Type:   "for",
		Target: loopVar,
		Cond:   collectionNode,
		Body:   bodySteps,
	}
}

func NewTestStringLiteral(val string) *StringLiteralNode {
	return &StringLiteralNode{Pos: &Position{Line: 1, Column: 1, File: "test"}, Value: val}
}

func NewTestNumberLiteral(val interface{}) *NumberLiteralNode {
	return &NumberLiteralNode{Pos: &Position{Line: 1, Column: 1, File: "test"}, Value: val}
}

func NewTestBooleanLiteral(val bool) *BooleanLiteralNode {
	return &BooleanLiteralNode{Pos: &Position{Line: 1, Column: 1, File: "test"}, Value: val}
}

func NewTestVariableNode(name string) *VariableNode {
	return &VariableNode{Pos: &Position{Line: 1, Column: 1, File: "test"}, Name: name}
}

// nlines: 268
// risk_rating: MEDIUM
