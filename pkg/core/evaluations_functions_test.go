// filename: pkg/core/evaluations_functions_test.go
package core

import (
	// Needed for errors.Is checks if done directly here (helper handles it now)
	"math"
	"testing"
	// Assuming EvalTestCase is defined in testing_helpers_test.go
	// Assuming runEvalExpressionTest is defined in testing_helpers_test.go
)

// Assumes NewTestInterpreter is defined in helpers.go

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

	// Use the EvalTestCase struct (assumed updated with ExpectedErrorIs)
	testCases := []EvalTestCase{
		// LN
		// *** Using ExpectedErrorIs where WantErr is true ***
		{Name: "LN(e)", InputNode: FunctionCallNode{FunctionName: "LN", Arguments: []interface{}{VariableNode{Name: "e"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "LN(1)", InputNode: FunctionCallNode{FunctionName: "LN", Arguments: []interface{}{VariableNode{Name: "one"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "LN(10 str)", InputNode: FunctionCallNode{FunctionName: "LN", Arguments: []interface{}{StringLiteralNode{Value: "10"}}}, InitialVars: vars, Expected: math.Log(10), WantErr: false},
		{Name: "LN(0)", InputNode: FunctionCallNode{FunctionName: "LN", Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},                                  // Use errors.Is
		{Name: "LN(-1)", InputNode: FunctionCallNode{FunctionName: "LN", Arguments: []interface{}{VariableNode{Name: "neg_one"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},                              // Use errors.Is
		{Name: "LN Type Error", InputNode: FunctionCallNode{FunctionName: "LN", Arguments: []interface{}{VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},                       // Use errors.Is
		{Name: "LN Arg Count Error", InputNode: FunctionCallNode{FunctionName: "LN", Arguments: []interface{}{VariableNode{Name: "one"}, VariableNode{Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrIncorrectArgCount}, // Use errors.Is

		// LOG (Base 10)
		{Name: "LOG(10)", InputNode: FunctionCallNode{FunctionName: "LOG", Arguments: []interface{}{VariableNode{Name: "ten"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "LOG(1)", InputNode: FunctionCallNode{FunctionName: "LOG", Arguments: []interface{}{VariableNode{Name: "one"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "LOG(100 str)", InputNode: FunctionCallNode{FunctionName: "LOG", Arguments: []interface{}{StringLiteralNode{Value: "100"}}}, InitialVars: vars, Expected: float64(2.0), WantErr: false},
		{Name: "LOG(0)", InputNode: FunctionCallNode{FunctionName: "LOG", Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},            // Use errors.Is
		{Name: "LOG Type Error", InputNode: FunctionCallNode{FunctionName: "LOG", Arguments: []interface{}{VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument}, // Use errors.Is

		// SIN / COS / TAN (using Pi/2 for known values)
		{Name: "SIN(Pi/2)", InputNode: FunctionCallNode{FunctionName: "SIN", Arguments: []interface{}{VariableNode{Name: "pi_o_2"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "SIN(0)", InputNode: FunctionCallNode{FunctionName: "SIN", Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "COS(Pi/2)", InputNode: FunctionCallNode{FunctionName: "COS", Arguments: []interface{}{VariableNode{Name: "pi_o_2"}}}, InitialVars: vars, Expected: math.Cos(math.Pi / 2.0), WantErr: false},
		{Name: "COS(0)", InputNode: FunctionCallNode{FunctionName: "COS", Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "TAN(0)", InputNode: FunctionCallNode{FunctionName: "TAN", Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "TAN(Pi/2)", InputNode: FunctionCallNode{FunctionName: "TAN", Arguments: []interface{}{VariableNode{Name: "pi_o_2"}}}, InitialVars: vars, Expected: math.Tan(math.Pi / 2.0), WantErr: false},
		{Name: "SIN Type Error", InputNode: FunctionCallNode{FunctionName: "SIN", Arguments: []interface{}{VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument}, // Use errors.Is

		// ASIN / ACOS
		{Name: "ASIN(1)", InputNode: FunctionCallNode{FunctionName: "ASIN", Arguments: []interface{}{VariableNode{Name: "one"}}}, InitialVars: vars, Expected: math.Asin(1.0), WantErr: false},
		{Name: "ASIN(0)", InputNode: FunctionCallNode{FunctionName: "ASIN", Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "ASIN(2)", InputNode: FunctionCallNode{FunctionName: "ASIN", Arguments: []interface{}{VariableNode{Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument}, // Use errors.Is (Domain error)
		{Name: "ACOS(1)", InputNode: FunctionCallNode{FunctionName: "ACOS", Arguments: []interface{}{VariableNode{Name: "one"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "ACOS(0)", InputNode: FunctionCallNode{FunctionName: "ACOS", Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: math.Acos(0.0), WantErr: false},
		{Name: "ACOS(-1)", InputNode: FunctionCallNode{FunctionName: "ACOS", Arguments: []interface{}{VariableNode{Name: "neg_one"}}}, InitialVars: vars, Expected: math.Acos(-1.0), WantErr: false},
		{Name: "ACOS(2)", InputNode: FunctionCallNode{FunctionName: "ACOS", Arguments: []interface{}{VariableNode{Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},             // Use errors.Is (Domain error)
		{Name: "ASIN Type Error", InputNode: FunctionCallNode{FunctionName: "ASIN", Arguments: []interface{}{VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument}, // Use errors.Is

		// ATAN
		{Name: "ATAN(0)", InputNode: FunctionCallNode{FunctionName: "ATAN", Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "ATAN(1)", InputNode: FunctionCallNode{FunctionName: "ATAN", Arguments: []interface{}{VariableNode{Name: "one"}}}, InitialVars: vars, Expected: math.Atan(1.0), WantErr: false},
		{Name: "ATAN Type Error", InputNode: FunctionCallNode{FunctionName: "ATAN", Arguments: []interface{}{VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument}, // Use errors.Is

		// Unknown function
		{Name: "Unknown Func", InputNode: FunctionCallNode{FunctionName: "SQRT", Arguments: []interface{}{NumberLiteralNode{Value: int64(4)}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrUnknownFunction}, // Use errors.Is
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

// *** REMOVED Helper Node Type definitions ***
// type FunctionCallNode struct { ... }
// type VariableNode struct { ... }
// type StringLiteralNode struct { ... }
// type NumberLiteralNode struct { ... }
