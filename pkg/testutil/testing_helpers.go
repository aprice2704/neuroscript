// NeuroScript Version: 0.4.1
// File version: 13
// Purpose: Updated createTestStep helper to use the new LValues field.
// filename: pkg/testutil/testing_helpers.go

package testutil

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

var dummyPos = &lang.Position{Line: 1, Column: 1, File: "test"}

// ... (struct definitions and other functions are unchanged) ...

type EvalTestCase struct {
	Name		string
	InputNode	ast.Expression
	InitialVars	map[string]lang.Value
	LastResult	lang.Value
	Expected	lang.Value
	WantErr		bool
	ExpectedErrorIs	error
}

type executeStepsTestCase struct {
	name		string
	inputSteps	[]ast.Step
	initialVars	map[string]lang.Value
	lastResult	lang.Value
	expectError	bool
	ExpectedErrorIs	error
	errContains	string
	expectedResult	lang.Value
	expectedVars	map[string]lang.Value
}

type ValidationTestCase struct {
	Name		string
	InputArgs	[]interface{}
	ExpectedError	error
}

func runValidationTestCases(t *testing.T, toolName string, testCases []ValidationTestCase) {
	t.Helper()
	interp, _ := llm.NewDefaultTestInterpreter(t)
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

func ExpressionTest(t *testing.T, tc EvalTestCase) {
	t.Helper()
	t.Run(tc.Name, func(t *testing.T) {
		i, err := llm.NewTestInterpreter(t, tc.InitialVars, tc.LastResult)
		if err != nil {
			t.Fatalf("NewTestInterpreter failed: %v", err)
		}

		got, err := i.evaluate.Expression(tc.InputNode)

		if (err != nil) != tc.WantErr {
			t.Fatalf("Test %q: Error expectation mismatch.\n got err = %v, wantErr %t", tc.Name, err, tc.WantErr)
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
		   Expected: 	%#v (%T)
		   Got: 	 	%#v (%T)`,
				tc.Name, tc.Expected, tc.Expected, got, got)
		}
	})
}

func runExecuteStepsTest(t *testing.T, tc executeStepsTestCase) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		i, err := llm.NewTestInterpreter(t, tc.initialVars, tc.lastResult)
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

		var actualResult lang.Value
		if wasReturn {
			actualResult = finalResultFromExec
		} else {
			actualResult = i.lastCallResult
		}

		if !reflect.DeepEqual(actualResult, tc.expectedResult) {
			t.Errorf("Test %q: Final execution result mismatch:\n Expected: %#v (%T)\n Got: 	  %#v (%T)", tc.name, tc.expectedResult, tc.expectedResult, actualResult, actualResult)
		}

		if tc.expectedVars != nil {
			for expectedKey, expectedValue := range tc.expectedVars {
				gotValue, ok := i.GetVariable(expectedKey)
				if !ok {
					t.Errorf("Test %q: Expected variable '%s' not found in final interpreter state", tc.name, expectedKey)
					continue
				}
				if !reflect.DeepEqual(gotValue, expectedValue) {
					t.Errorf("Test %q: Variable state mismatch for key '%s':\n Expected: %#v (%T)\n Got: 	  %#v (%T)", tc.name, expectedKey, expectedValue, expectedValue, gotValue, gotValue)
				}
			}
		}
	})
}

// createTestStep is a robust helper for creating ast.Step structs for tests.
func createTestStep(stepType, target string, value ast.Expression, callArgs []ast.Expression) ast.Step {
	s := ast.Step{Position: dummyPos, Type: stepType}
	switch strings.ToLower(stepType) {
	case "set":
		// MODIFIED: Use the new LValues field.
		lval := &ast.LValueNode{Identifier: target, Position: s.Pos}
		s.LValues = []ast.Expression{lval}
		s.Value = value
	case "emit", "return":
		s.Values = []ast.Expression{value}
	case "must":
		s.Cond = value
	case "call":
		s.Call = &ast.CallableExprNode{
			Position:	dummyPos,
			Target:		ast.CallTarget{Position: dummyPos, Name: target, IsTool: true},
			Arguments:	callArgs,
		}
	default:
		// This default case might be problematic if 'value' is an LValue, but for now, we leave it.
		// It seems designed for simple expression assignments to 'Value' which is not always correct.
	}
	return s
}

func createIfStep(pos *lang.Position, condNode ast.Expression, thenSteps, elseSteps []ast.Step) ast.Step {
	return ast.Step{Position: pos, Type: "if", Cond: condNode, Body: thenSteps, Else: elseSteps}
}

func createWhileStep(pos *lang.Position, condNode ast.Expression, bodySteps []ast.Step) ast.Step {
	return ast.Step{
		Position:	pos,
		Type:		"while",
		Cond:		condNode,
		Body:		bodySteps,
	}
}

func createForStep(pos *lang.Position, loopVarName string, collectionExpr ast.Expression, bodySteps []ast.Step) ast.Step {
	return ast.Step{
		Position:	pos,
		Type:		"for",
		LoopVarName:	loopVarName,
		Collection:	collectionExpr,
		Body:		bodySteps,
	}
}

func NewTestStringLiteral(val string) *ast.StringLiteralNode {
	return &ast.StringLiteralNode{Position: dummyPos, Value: val}
}

func NewTestNumberLiteral(val float64) *ast.NumberLiteralNode {
	return &ast.NumberLiteralNode{Position: dummyPos, Value: val}
}

func NewTestBooleanLiteral(val bool) *ast.BooleanLiteralNode {
	return &ast.BooleanLiteralNode{Position: dummyPos, Value: val}
}

func NewVariableNode(name string) *ast.VariableNode {
	return &ast.VariableNode{Position: dummyPos, Name: name}
}

func DebugDumpVariables(i *neurogo.Interpreter, t *testing.T) {
	i.variablesMu.RLock()
	defer i.variablesMu.RUnlock()
	t.Log("--- INTERPRETER VARIABLE DUMP ---")
	t.Log("  (variable dumping needs update to get all keys)")
	t.Log("--- END VARIABLE DUMP ---")
}