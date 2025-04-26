// pkg/core/evaluation_logical_bitwise_test.go
package core

import (
	// "math" // Not needed directly here
	// Keep for DeepEqual if runEvalExpressionTest removed it

	"testing"
)

// Assumes NewTestInterpreter( and runEvalExpressionTest (with EvalTestCase) are defined in test_helpers_test.go

func TestLogicalBitwiseOps(t *testing.T) {
	vars := map[string]interface{}{
		"trueVar":  true,
		"falseVar": false,
		"int1":     int64(1),
		"int0":     int64(0),
		"strTrue":  "true",
		"strFalse": "false",
		"str1":     "1",
		"str0":     "0",
		"strOther": "hello",
		"nilVar":   nil,
		"bit5":     int64(5), // 0101
		"bit6":     int64(6), // 0110
		"float1":   float64(1.0),
		"strABC":   "ABC", // Added for error testing
	}

	// Helper node for a variable that would cause an error if evaluated
	errorVarNode := VariableNode{Name: "errorVar"} // errorVar is not in initialVars

	// Use the named EvalTestCase struct
	testCases := []EvalTestCase{
		// --- Logical NOT ---
		{Name: "NOT True", InputNode: UnaryOpNode{Operator: "NOT", Operand: BooleanLiteralNode{Value: true}}, InitialVars: vars, Expected: false, WantErr: false},
		{Name: "NOT False", InputNode: UnaryOpNode{Operator: "NOT", Operand: BooleanLiteralNode{Value: false}}, InitialVars: vars, Expected: true, WantErr: false},
		{Name: "NOT Int 1", InputNode: UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "int1"}}, InitialVars: vars, Expected: false, WantErr: false},
		{Name: "NOT Int 0", InputNode: UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "int0"}}, InitialVars: vars, Expected: true, WantErr: false},
		{Name: "NOT Str true", InputNode: UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "strTrue"}}, InitialVars: vars, Expected: false, WantErr: false},
		{Name: "NOT Str false", InputNode: UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "strFalse"}}, InitialVars: vars, Expected: true, WantErr: false},
		{Name: "NOT Str 1", InputNode: UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "str1"}}, InitialVars: vars, Expected: false, WantErr: false},
		{Name: "NOT Str 0", InputNode: UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "str0"}}, InitialVars: vars, Expected: true, WantErr: false},
		{Name: "NOT Str Other", InputNode: UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "strOther"}}, InitialVars: vars, Expected: true, WantErr: false},
		{Name: "NOT Nil", InputNode: UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "nilVar"}}, InitialVars: vars, Expected: true, WantErr: false},

		// --- Logical AND ---
		{Name: "AND True True", InputNode: BinaryOpNode{Left: BooleanLiteralNode{Value: true}, Operator: "AND", Right: BooleanLiteralNode{Value: true}}, InitialVars: vars, Expected: true, WantErr: false},
		{Name: "AND True False", InputNode: BinaryOpNode{Left: BooleanLiteralNode{Value: true}, Operator: "AND", Right: BooleanLiteralNode{Value: false}}, InitialVars: vars, Expected: false, WantErr: false},
		{Name: "AND False True (Short Circuit)", InputNode: BinaryOpNode{Left: BooleanLiteralNode{Value: false}, Operator: "AND", Right: errorVarNode}, InitialVars: vars, Expected: false, WantErr: false},
		{Name: "AND False False", InputNode: BinaryOpNode{Left: BooleanLiteralNode{Value: false}, Operator: "AND", Right: BooleanLiteralNode{Value: false}}, InitialVars: vars, Expected: false, WantErr: false},
		{Name: "AND Int1 StrTrue", InputNode: BinaryOpNode{Left: VariableNode{Name: "int1"}, Operator: "AND", Right: VariableNode{Name: "strTrue"}}, InitialVars: vars, Expected: true, WantErr: false},
		{Name: "AND StrOther Int0", InputNode: BinaryOpNode{Left: VariableNode{Name: "strOther"}, Operator: "AND", Right: VariableNode{Name: "int0"}}, InitialVars: vars, Expected: false, WantErr: false},

		// --- Logical OR ---
		{Name: "OR True True", InputNode: BinaryOpNode{Left: BooleanLiteralNode{Value: true}, Operator: "OR", Right: BooleanLiteralNode{Value: true}}, InitialVars: vars, Expected: true, WantErr: false},
		{Name: "OR True False (Short Circuit)", InputNode: BinaryOpNode{Left: BooleanLiteralNode{Value: true}, Operator: "OR", Right: errorVarNode}, InitialVars: vars, Expected: true, WantErr: false},
		{Name: "OR False True", InputNode: BinaryOpNode{Left: BooleanLiteralNode{Value: false}, Operator: "OR", Right: BooleanLiteralNode{Value: true}}, InitialVars: vars, Expected: true, WantErr: false},
		{Name: "OR False False", InputNode: BinaryOpNode{Left: BooleanLiteralNode{Value: false}, Operator: "OR", Right: BooleanLiteralNode{Value: false}}, InitialVars: vars, Expected: false, WantErr: false},
		{Name: "OR Int0 Str1", InputNode: BinaryOpNode{Left: VariableNode{Name: "int0"}, Operator: "OR", Right: VariableNode{Name: "str1"}}, InitialVars: vars, Expected: true, WantErr: false},
		{Name: "OR StrOther Nil", InputNode: BinaryOpNode{Left: VariableNode{Name: "strOther"}, Operator: "OR", Right: VariableNode{Name: "nilVar"}}, InitialVars: vars, Expected: false, WantErr: false},

		// --- Bitwise AND (&) ---
		{Name: "Bitwise AND 5&6", InputNode: BinaryOpNode{Left: VariableNode{Name: "bit5"}, Operator: "&", Right: VariableNode{Name: "bit6"}}, InitialVars: vars, Expected: int64(4), WantErr: false},
		// *** CORRECTED ErrContains ***
		{Name: "Bitwise AND Error Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "bit5"}, Operator: "&", Right: VariableNode{Name: "float1"}}, InitialVars: vars, WantErr: true, ErrContains: "failed bitwise operation"},
		{Name: "Bitwise AND Error String", InputNode: BinaryOpNode{Left: VariableNode{Name: "bit5"}, Operator: "&", Right: VariableNode{Name: "strABC"}}, InitialVars: vars, WantErr: true, ErrContains: "failed bitwise operation"},

		// --- Bitwise OR (|) ---
		{Name: "Bitwise OR 5|6", InputNode: BinaryOpNode{Left: VariableNode{Name: "bit5"}, Operator: "|", Right: VariableNode{Name: "bit6"}}, InitialVars: vars, Expected: int64(7), WantErr: false},
		// *** CORRECTED ErrContains ***
		{Name: "Bitwise OR Error Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "bit5"}, Operator: "|", Right: VariableNode{Name: "float1"}}, InitialVars: vars, WantErr: true, ErrContains: "failed bitwise operation"},

		// --- Bitwise XOR (^) ---
		{Name: "Bitwise XOR 5^6", InputNode: BinaryOpNode{Left: VariableNode{Name: "bit5"}, Operator: "^", Right: VariableNode{Name: "bit6"}}, InitialVars: vars, Expected: int64(3), WantErr: false},
		// *** CORRECTED ErrContains ***
		{Name: "Bitwise XOR Error Float", InputNode: BinaryOpNode{Left: VariableNode{Name: "bit5"}, Operator: "^", Right: VariableNode{Name: "float1"}}, InitialVars: vars, WantErr: true, ErrContains: "failed bitwise operation"},
	}

	for _, tc := range testCases {
		// Create a new scope for t.Run
		tc := tc // Capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()                 // Optional
			runEvalExpressionTest(t, tc) // Uses helper from test_helpers_test.go
		})
	}
}
