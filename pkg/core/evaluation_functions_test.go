// filename: pkg/core/evaluation_functions_test.go
package core

import (
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
		"str_num": "3.14",
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
		{Name: "LN(e)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "e"}}}, InitialVars: vars, Expected: NumberValue{Value: 1.0}, WantErr: false},
		{Name: "LN(1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}}}, InitialVars: vars, Expected: NumberValue{Value: 0.0}, WantErr: false},
		{Name: "LN(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN(-1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "neg_one"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN Type Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN Arg Count Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}, &VariableNode{Pos: dummyPos, Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrIncorrectArgCount},

		// LOG (Base 10)
		{Name: "LOG(10)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "log"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "ten"}}}, InitialVars: vars, Expected: NumberValue{Value: 1.0}, WantErr: false},
		{Name: "LOG(1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "log"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}}}, InitialVars: vars, Expected: NumberValue{Value: 0.0}, WantErr: false},
		{Name: "LOG(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "log"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LOG Type Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "log"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// SIN / COS / TAN (using Pi/2 for known values)
		{Name: "SIN(Pi/2)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "sin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "pi_o_2"}}}, InitialVars: vars, Expected: NumberValue{Value: 1.0}, WantErr: false},
		{Name: "SIN(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "sin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: NumberValue{Value: 0.0}, WantErr: false},
		{Name: "COS(Pi/2)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "cos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "pi_o_2"}}}, InitialVars: vars, Expected: NumberValue{Value: math.Cos(math.Pi / 2.0)}, WantErr: false},
		{Name: "COS(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "cos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: NumberValue{Value: 1.0}, WantErr: false},
		{Name: "TAN(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "tan"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: NumberValue{Value: 0.0}, WantErr: false},
		{Name: "SIN Type Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "sin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// ASIN / ACOS
		{Name: "ASIN(1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "asin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}}}, InitialVars: vars, Expected: NumberValue{Value: math.Asin(1.0)}, WantErr: false},
		{Name: "ASIN(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "asin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: NumberValue{Value: 0.0}, WantErr: false},
		{Name: "ASIN(2)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "asin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "ACOS(1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "acos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}}}, InitialVars: vars, Expected: NumberValue{Value: 0.0}, WantErr: false},
		{Name: "ACOS(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "acos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: NumberValue{Value: math.Acos(0.0)}, WantErr: false},
		{Name: "ACOS(-1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "acos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "neg_one"}}}, InitialVars: vars, Expected: NumberValue{Value: math.Acos(-1.0)}, WantErr: false},
		{Name: "ACOS(2)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "acos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "ASIN Type Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "asin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// ATAN
		{Name: "ATAN(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "atan"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: NumberValue{Value: 0.0}, WantErr: false},
		{Name: "ATAN(1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "atan"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}}}, InitialVars: vars, Expected: NumberValue{Value: math.Atan(1.0)}, WantErr: false},
		{Name: "ATAN Type Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "atan"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// Unknown function
		{Name: "Unknown Func", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "SQRT"}, Arguments: []Expression{&NumberLiteralNode{Pos: dummyPos, Value: int64(4)}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrProcedureNotFound},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			runEvalExpressionTest(t, tc)
		})
	}
}
