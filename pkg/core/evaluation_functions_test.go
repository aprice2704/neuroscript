// filename: pkg/core/evaluation_functions_test.go
package core

import (
	"errors"
	"fmt"
	"math"
	"testing"
)

func TestMathFunctions(t *testing.T) {
	rawVars := map[string]interface{}{
		"e":       float64(math.E),
		"ten":     int64(10),
		"pi_o_2":  float64(math.Pi / 2.0),
		"one":     int64(1),
		"zero":    int64(0),
		"neg_one": int64(-1),
		"two":     float64(2.0),
		"str_abc": "abc",
	}

	vars := make(map[string]Value, len(rawVars))
	for k, v := range rawVars {
		w, err := Wrap(v)
		if err != nil {
			panic(fmt.Sprintf("test setup: cannot wrap %q: %v", k, err))
		}
		vars[k] = w
	}

	dummyPos := &Position{Line: 1, Column: 1}

	testCases := []EvalTestCase{
		// LN
		{Name: "LN(e)", InputNode: &CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []Expression{&VariableNode{Name: "e"}}}, InitialVars: vars, Expected: NumberValue{Value: 1.0}},
		{Name: "LN(1)", InputNode: &CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []Expression{&VariableNode{Name: "one"}}}, InitialVars: vars, Expected: NumberValue{Value: 0.0}},
		{Name: "LN(0)", InputNode: &CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []Expression{&VariableNode{Name: "zero"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN(-1)", InputNode: &CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []Expression{&VariableNode{Name: "neg_one"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN Type Error", InputNode: &CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []Expression{&VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN Arg Count Error", InputNode: &CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []Expression{&VariableNode{Name: "one"}, &VariableNode{Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrIncorrectArgCount},

		// LOG (Base 10)
		{Name: "LOG(10)", InputNode: &CallableExprNode{Target: CallTarget{Name: "log"}, Arguments: []Expression{&VariableNode{Name: "ten"}}}, InitialVars: vars, Expected: NumberValue{Value: 1.0}},
		{Name: "LOG(1)", InputNode: &CallableExprNode{Target: CallTarget{Name: "log"}, Arguments: []Expression{&VariableNode{Name: "one"}}}, InitialVars: vars, Expected: NumberValue{Value: 0.0}},
		{Name: "LOG(0)", InputNode: &CallableExprNode{Target: CallTarget{Name: "log"}, Arguments: []Expression{&VariableNode{Name: "zero"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LOG Type Error", InputNode: &CallableExprNode{Target: CallTarget{Name: "log"}, Arguments: []Expression{&VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// SIN / COS / TAN
		{Name: "SIN(Pi/2)", InputNode: &CallableExprNode{Target: CallTarget{Name: "sin"}, Arguments: []Expression{&VariableNode{Name: "pi_o_2"}}}, InitialVars: vars, Expected: NumberValue{Value: 1.0}},
		{Name: "COS(0)", InputNode: &CallableExprNode{Target: CallTarget{Name: "cos"}, Arguments: []Expression{&VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: NumberValue{Value: 1.0}},
		{Name: "TAN(0)", InputNode: &CallableExprNode{Target: CallTarget{Name: "tan"}, Arguments: []Expression{&VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: NumberValue{Value: 0.0}},
		{Name: "SIN Type Error", InputNode: &CallableExprNode{Target: CallTarget{Name: "sin"}, Arguments: []Expression{&VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// ASIN / ACOS
		{Name: "ASIN(1)", InputNode: &CallableExprNode{Target: CallTarget{Name: "asin"}, Arguments: []Expression{&VariableNode{Name: "one"}}}, InitialVars: vars, Expected: NumberValue{Value: math.Asin(1.0)}},
		{Name: "ACOS(-1)", InputNode: &CallableExprNode{Target: CallTarget{Name: "acos"}, Arguments: []Expression{&VariableNode{Name: "neg_one"}}}, InitialVars: vars, Expected: NumberValue{Value: math.Acos(-1.0)}},
		{Name: "ASIN(2)", InputNode: &CallableExprNode{Target: CallTarget{Name: "asin"}, Arguments: []Expression{&VariableNode{Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "ASIN Type Error", InputNode: &CallableExprNode{Target: CallTarget{Name: "asin"}, Arguments: []Expression{&VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// ATAN
		{Name: "ATAN(1)", InputNode: &CallableExprNode{Target: CallTarget{Name: "atan"}, Arguments: []Expression{&VariableNode{Name: "one"}}}, InitialVars: vars, Expected: NumberValue{Value: math.Atan(1.0)}},
		{Name: "ATAN Type Error", InputNode: &CallableExprNode{Target: CallTarget{Name: "atan"}, Arguments: []Expression{&VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if call, ok := tc.InputNode.(*CallableExprNode); ok {
				call.Pos = dummyPos
				call.Target.Pos = dummyPos
				for _, arg := range call.Arguments {
					if vn, okV := arg.(*VariableNode); okV {
						vn.Pos = dummyPos
					}
				}
			}
			runEvalExpressionTest(t, tc)
		})
	}
}

// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Unit tests for built-in evaluation functions, including len().
// filename: pkg/core/evaluation_functions_test.go
// nlines: 68
// risk_rating: LOW

func TestEvaluateBuiltInFunction_Len(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"string", "hello", 5},
		{"unicode string", "你好", 2},
		{"empty string", "", 0},
		{"list", []interface{}{1, true, "three"}, 3},
		{"empty list", []interface{}{}, 0},
		{"map", map[string]interface{}{"a": 1, "b": 2}, 2},
		{"empty map", map[string]interface{}{}, 0},
		{"number", 123.45, 1},
		{"boolean", true, 1},
		{"nil", nil, 0},
		{"error type", errors.New("an error"), 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := []interface{}{tc.input}
			result, err := evaluateBuiltInFunction("len", args)
			if err != nil {
				t.Fatalf("evaluateBuiltInFunction failed: %v", err)
			}

			numResult, ok := result.(NumberValue)
			if !ok {
				t.Fatalf("Expected NumberValue, got %T", result)
			}

			if numResult.Value != tc.expected {
				t.Errorf("Expected length %f, got %f", tc.expected, numResult.Value)
			}
		})
	}

	t.Run("incorrect argument count", func(t *testing.T) {
		_, err := evaluateBuiltInFunction("len", []interface{}{})
		if !errors.Is(err, ErrIncorrectArgCount) {
			t.Errorf("Expected ErrIncorrectArgCount, got %v", err)
		}

		_, err = evaluateBuiltInFunction("len", []interface{}{"a", "b"})
		if !errors.Is(err, ErrIncorrectArgCount) {
			t.Errorf("Expected ErrIncorrectArgCount, got %v", err)
		}
	})
}
