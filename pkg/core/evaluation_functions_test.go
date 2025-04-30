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
		// *** UPDATED TO USE CallableExprNode and CallTarget ***
		{Name: "LN(e)", InputNode: CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []interface{}{VariableNode{Name: "e"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "LN(1)", InputNode: CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []interface{}{VariableNode{Name: "one"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		// NOTE: Passing string to LN/LOG etc. that converts to number is NOT standard behavior.
		// Let's assume evaluateBuiltInFunction expects numeric types directly.
		// Test case "LN(10 str)" removed as it relies on implicit conversion not specified.
		{Name: "LN(0)", InputNode: CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN(-1)", InputNode: CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []interface{}{VariableNode{Name: "neg_one"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN Type Error", InputNode: CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []interface{}{VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LN Arg Count Error", InputNode: CallableExprNode{Target: CallTarget{Name: "ln"}, Arguments: []interface{}{VariableNode{Name: "one"}, VariableNode{Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrIncorrectArgCount},

		// LOG (Base 10)
		{Name: "LOG(10)", InputNode: CallableExprNode{Target: CallTarget{Name: "log"}, Arguments: []interface{}{VariableNode{Name: "ten"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "LOG(1)", InputNode: CallableExprNode{Target: CallTarget{Name: "log"}, Arguments: []interface{}{VariableNode{Name: "one"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		// Test case "LOG(100 str)" removed (see LN note).
		{Name: "LOG(0)", InputNode: CallableExprNode{Target: CallTarget{Name: "log"}, Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},
		{Name: "LOG Type Error", InputNode: CallableExprNode{Target: CallTarget{Name: "log"}, Arguments: []interface{}{VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// SIN / COS / TAN (using Pi/2 for known values)
		{Name: "SIN(Pi/2)", InputNode: CallableExprNode{Target: CallTarget{Name: "sin"}, Arguments: []interface{}{VariableNode{Name: "pi_o_2"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "SIN(0)", InputNode: CallableExprNode{Target: CallTarget{Name: "sin"}, Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "COS(Pi/2)", InputNode: CallableExprNode{Target: CallTarget{Name: "cos"}, Arguments: []interface{}{VariableNode{Name: "pi_o_2"}}}, InitialVars: vars, Expected: math.Cos(math.Pi / 2.0), WantErr: false}, // Expect near-zero float
		{Name: "COS(0)", InputNode: CallableExprNode{Target: CallTarget{Name: "cos"}, Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: float64(1.0), WantErr: false},
		{Name: "TAN(0)", InputNode: CallableExprNode{Target: CallTarget{Name: "tan"}, Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		// TAN(Pi/2) is undefined (Inf), expect error or handle appropriately if math.Inf() is allowed
		// {Name: "TAN(Pi/2)", InputNode: CallableExprNode{Target: CallTarget{Name: "tan"}, Arguments: []interface{}{VariableNode{Name: "pi_o_2"}}}, InitialVars: vars, Expected: math.Tan(math.Pi/2.0), WantErr: false}, // This will be a very large number or Inf
		{Name: "SIN Type Error", InputNode: CallableExprNode{Target: CallTarget{Name: "sin"}, Arguments: []interface{}{VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// ASIN / ACOS
		{Name: "ASIN(1)", InputNode: CallableExprNode{Target: CallTarget{Name: "asin"}, Arguments: []interface{}{VariableNode{Name: "one"}}}, InitialVars: vars, Expected: math.Asin(1.0), WantErr: false},
		{Name: "ASIN(0)", InputNode: CallableExprNode{Target: CallTarget{Name: "asin"}, Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "ASIN(2)", InputNode: CallableExprNode{Target: CallTarget{Name: "asin"}, Arguments: []interface{}{VariableNode{Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument}, // Domain error
		{Name: "ACOS(1)", InputNode: CallableExprNode{Target: CallTarget{Name: "acos"}, Arguments: []interface{}{VariableNode{Name: "one"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "ACOS(0)", InputNode: CallableExprNode{Target: CallTarget{Name: "acos"}, Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: math.Acos(0.0), WantErr: false},
		{Name: "ACOS(-1)", InputNode: CallableExprNode{Target: CallTarget{Name: "acos"}, Arguments: []interface{}{VariableNode{Name: "neg_one"}}}, InitialVars: vars, Expected: math.Acos(-1.0), WantErr: false},
		{Name: "ACOS(2)", InputNode: CallableExprNode{Target: CallTarget{Name: "acos"}, Arguments: []interface{}{VariableNode{Name: "two"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument}, // Domain error
		{Name: "ASIN Type Error", InputNode: CallableExprNode{Target: CallTarget{Name: "asin"}, Arguments: []interface{}{VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// ATAN
		{Name: "ATAN(0)", InputNode: CallableExprNode{Target: CallTarget{Name: "atan"}, Arguments: []interface{}{VariableNode{Name: "zero"}}}, InitialVars: vars, Expected: float64(0.0), WantErr: false},
		{Name: "ATAN(1)", InputNode: CallableExprNode{Target: CallTarget{Name: "atan"}, Arguments: []interface{}{VariableNode{Name: "one"}}}, InitialVars: vars, Expected: math.Atan(1.0), WantErr: false},
		{Name: "ATAN Type Error", InputNode: CallableExprNode{Target: CallTarget{Name: "atan"}, Arguments: []interface{}{VariableNode{Name: "str_abc"}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrInvalidFunctionArgument},

		// Unknown function
		{Name: "Unknown Func", InputNode: CallableExprNode{Target: CallTarget{Name: "SQRT"}, Arguments: []interface{}{NumberLiteralNode{Value: int64(4)}}}, InitialVars: vars, WantErr: true, ExpectedErrorIs: ErrProcedureNotFound},
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

// --- REMOVED Helper Node Type definitions ---
// These should be imported from the main core package or defined once in testing_helpers
// type VariableNode struct { ... }
// type StringLiteralNode struct { ... }
// type NumberLiteralNode struct { ... }
