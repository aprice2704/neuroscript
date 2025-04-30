// filename: pkg/core/evaluations_functions_test.go
package core

import (
	// Needed for errors.Is checks if done directly here (helper handles it now)
	"math"
	"testing"
	// Assuming Position is defined in this package (ast.go)
	// Assuming EvalTestCase is defined in testing_helpers_test.go
	// Assuming runEvalExpressionTest is defined in testing_helpers_test.go
	// Assuming error variables like ErrInvalidFunctionArgument are defined in errors.go
)

// Assumes NewDefaultTestInterpreter is defined in helpers.go

func TestMathFunctions(t *testing.T) {
	vars := map[string]interface{}{
		"e":       float64(math.E), // Ensure float64
		"ten":     int64(10),
		"pi_o_2":  float64(math.Pi / 2.0), // Ensure float64
		"one":     int64(1),
		"zero":    int64(0),
		"neg_one": int64(-1),
		"two":     float64(2.0), // Ensure float64
		"str_num": "3.14",
		"str_abc": "abc",
	}

	// Define dummyPos using local Position type pointer
	dummyPos := &Position{Line: 1, Column: 1}

	// Use the EvalTestCase struct (assumed updated with ExpectedErrorIs)
	testCases := []EvalTestCase{
		// LN
		// *** CORRECTED: Use & for nodes assigned to Expression fields/slices ***
		{Name: "LN(e)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "e"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "LN(1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "LN(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN(-1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "neg_one"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN Type Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN Arg Count Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "ln"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}, &VariableNode{Pos: dummyPos, Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrIncorrectArgCount},

		// LOG (Base 10)
		{Name: "LOG(10)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "log"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "ten"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "LOG(1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "log"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "LOG(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "log"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LOG Type Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "log"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// SIN / COS / TAN (using Pi/2 for known values)
		{Name: "SIN(Pi/2)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "sin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "pi_o_2"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "SIN(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "sin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "COS(Pi/2)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "cos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "pi_o_2"}}}, InitialVars: vars, Expected: math.Cos(math.Pi / 2.0), WantErr: false}, // Expect near-zero float
		{Name: "COS(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "cos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "TAN(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "tan"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "SIN Type Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "sin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// ASIN / ACOS
		{Name: "ASIN(1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "asin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}}}, InitialVars: vars, Expected: math.Asin(1.0), WantErr: false},
		{Name: "ASIN(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "asin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "ASIN(2)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "asin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument}, // Domain error
		{Name: "ACOS(1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "acos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "ACOS(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "acos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: math.Acos(0.0), WantErr: false},
		{Name: "ACOS(-1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "acos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "neg_one"}}}, InitialVars: vars, Expected: math.Acos(-1.0), WantErr: false},
		{Name: "ACOS(2)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "acos"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument}, // Domain error
		{Name: "ASIN Type Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "asin"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// ATAN
		{Name: "ATAN(0)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "atan"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "ATAN(1)", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "atan"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "one"}}}, InitialVars: vars, Expected: math.Atan(1.0), WantErr: false},
		{Name: "ATAN Type Error", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "atan"}, Arguments: []Expression{&VariableNode{Pos: dummyPos, Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// Unknown function
		{Name: "Unknown Func", InputNode: &CallableExprNode{Pos: dummyPos, Target: CallTarget{Pos: dummyPos, Name: "SQRT"}, Arguments: []Expression{&NumberLiteralNode{Pos: dummyPos, Value: int64(4)}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrProcedureNotFound}, // Assuming proc/func not found error
		// *** END CORRECTIONS ***
	}

	for _, tc := range testCases {
		// Create a new scope for t.Run
		tc := tc // Capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel() // Optional: Mark test for parallel execution
			// Assuming runEvalExpressionTest is defined in testing_helpers_test.go
			// and handles the EvalTestCase struct correctly (including ExpectedErrorIs)
			runEvalExpressionTest(t, tc)
		})
	}
}
