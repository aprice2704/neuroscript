// NeuroScript Version: 0.3.5
// File version: 0.0.5 // Corrected Step struct initialization in createTestStep and createForStep.
// filename: pkg/core/testing_helpers.go
package core

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"testing"
	// Position, Expression, Step, LValueNode, AccessorNode, CallableExprNode, CallTarget are defined in ast.go
)

// --- Shared Test Struct Definitions ---

// EvalTestCase defines the structure for testing evaluateExpression
type EvalTestCase struct {
	Name            string
	InputNode       interface{}
	InitialVars     map[string]interface{}
	LastResult      interface{}
	Expected        interface{}
	WantErr         bool
	ExpectedErrorIs error
}

// executeStepsTestCase defines the structure for testing interp.executeSteps
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
	aFloat, aIsNum := toFloat64(a) // toFloat64 is assumed to be defined elsewhere (e.g., evaluation_helpers.go)
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
		if err == nil {
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
		} else if tc.ExpectedErrorIs == nil {
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
		if resultSlice, ok := actualExecResult.([]interface{}); ok && len(resultSlice) == 1 {
			if _, expectedSlice := tc.expectedResult.([]interface{}); !expectedSlice {
				actualExecResult = resultSlice[0]
			}
		}
	} else {
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

// MODIFIED createTestStep
func createTestStep(stepType string, targetOrLoopVarOrInto string, valueOrCollectionOrCall interface{}, _ignoredCallArg interface{}) Step {
	s := Step{Pos: &Position{Line: 1, Column: 1, File: "test"}, Type: stepType}

	switch strings.ToLower(stepType) {
	case "set":
		s.LValue = &LValueNode{Identifier: targetOrLoopVarOrInto, Pos: s.Pos} // Assuming simple LValue for tests
		if expr, ok := valueOrCollectionOrCall.(Expression); ok {
			s.Value = expr
		} else if valueOrCollectionOrCall != nil {
			// Handle cases where a raw value might be passed for testing simple literals
			// This might need a helper to convert raw Go values to Expression nodes for tests
			// For now, assume valueOrCollectionOrCall is already an Expression or nil
			panic(fmt.Sprintf("createTestStep 'set': valueOrCollectionOrCall (%T) is not Expression", valueOrCollectionOrCall))
		}
	case "call":
		if callNode, ok := valueOrCollectionOrCall.(*CallableExprNode); ok {
			s.Call = callNode
		} else {
			panic(fmt.Sprintf("createTestStep 'call': valueOrCollectionOrCall (%T) is not *CallableExprNode", valueOrCollectionOrCall))
		}
	case "ask":
		s.AskIntoVar = targetOrLoopVarOrInto
		if expr, ok := valueOrCollectionOrCall.(Expression); ok {
			s.Value = expr // Prompt expression
		} else {
			panic(fmt.Sprintf("createTestStep 'ask': valueOrCollectionOrCall (%T) for prompt is not Expression", valueOrCollectionOrCall))
		}
	case "for", "for_each": // Assuming "for" is alias for "for_each" in tests
		s.Type = "for_each" // Normalize
		s.LoopVarName = targetOrLoopVarOrInto
		if expr, ok := valueOrCollectionOrCall.(Expression); ok {
			s.Collection = expr
		} else {
			panic(fmt.Sprintf("createTestStep 'for_each': valueOrCollectionOrCall (%T) for collection is not Expression", valueOrCollectionOrCall))
		}
	case "mustbe":
		// For 'mustbe', the 'targetOrLoopVarOrInto' is the callable's name (often not directly used in Step if Call is set)
		// 'valueOrCollectionOrCall' should be the CallableExprNode
		if callNode, ok := valueOrCollectionOrCall.(*CallableExprNode); ok {
			s.Call = callNode
			s.Value = callNode // As per current AST builder logic for mustbe
		} else {
			panic(fmt.Sprintf("createTestStep 'mustbe': valueOrCollectionOrCall (%T) is not *CallableExprNode", valueOrCollectionOrCall))
		}
	// For 'return', 'emit', 'must', 'fail', they usually only use 'Value' or 'Values'
	default:
		if expr, ok := valueOrCollectionOrCall.(Expression); ok {
			s.Value = expr
		} else if exprs, ok := valueOrCollectionOrCall.([]Expression); ok {
			s.Values = exprs
		} else if valueOrCollectionOrCall != nil {
			// This case might be problematic if the type doesn't fit Value or Values
			// For simplicity in a test helper, this might be acceptable if tests pass appropriate types.
			// Consider panicking for unhandled types if strictness is required.
			// For now, if it's not an Expression or []Expression, it might be left nil or cause issues.
		}
		// If 'target' was used for other step types, that logic needs to be added here.
		// e.g. if stepType == "someOtherTypeThatUsedTarget", s.SomeOtherField = targetOrLoopVarOrInto
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

// MODIFIED createForStep
func createForStep(pos *Position, loopVar string, collectionNode Expression, bodySteps []Step) Step {
	if collectionNode == nil {
		panic("createForStep: test provided a nil collectionNode argument")
	}
	return Step{
		Pos:         pos,
		Type:        "for_each", // Standardize to "for_each" as per AST builder
		LoopVarName: loopVar,
		Collection:  collectionNode,
		Body:        bodySteps,
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

// --- Debugging Helpers ---

// DebugDumpVariables prints the current state of all global variables to the test log.
// It safely locks the interpreter's variable map for reading.
func DebugDumpVariables(i *Interpreter, t *testing.T) {
	i.variablesMu.RLock()
	defer i.variablesMu.RUnlock()

	t.Log("--- INTERPRETER VARIABLE DUMP ---")
	if len(i.variables) == 0 {
		t.Log("  (no variables set)")
		t.Log("--- END VARIABLE DUMP ---")
		return
	}
	for key, value := range i.variables {
		t.Logf("  - %s (%T) = %v", key, value, value)
	}
	t.Log("--- END VARIABLE DUMP ---")
}
