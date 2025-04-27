package core

import (
	"testing"
)

// --- Test Suite for Logical and Bitwise Operations ---
func TestLogicalBitwiseOps(t *testing.T) {
	vars := map[string]interface{}{
		"trueVar":     true,
		"falseVar":    false,
		"numOne":      int64(1),
		"numZero":     int64(0),
		"floatOne":    float64(1.0),
		"floatZero":   float64(0.0),
		"strTrue":     "true",
		"strFalse":    "false",
		"strOther":    "hello", // Falsy string (now)
		"strOne":      "1",     // Truthy string
		"nilVar":      nil,
		"num5":        int64(5), // 0101
		"num3":        int64(3), // 0011
		"floatNonInt": float64(3.14),
	}
	lastResult := "LastResult" // Placeholder for LAST tests, value doesn't matter here

	tests := []EvalTestCase{
		// --- NOT Operator ---
		{"NOT True Literal", UnaryOpNode{Operator: "NOT", Operand: BooleanLiteralNode{Value: true}}, vars, lastResult, false, false, nil},
		{"NOT False Literal", UnaryOpNode{Operator: "NOT", Operand: BooleanLiteralNode{Value: false}}, vars, lastResult, true, false, nil},
		{"NOT True Var", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "trueVar"}}, vars, lastResult, false, false, nil},
		{"NOT False Var", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "falseVar"}}, vars, lastResult, true, false, nil},
		{"NOT Num NonZero", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "numOne"}}, vars, lastResult, false, false, nil}, // 1 is truthy, NOT 1 is false
		{"NOT Num Zero", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "numZero"}}, vars, lastResult, true, false, nil},    // 0 is falsy, NOT 0 is true
		// FIX: Updated expectation for NOT Str Other
		{"NOT Str Other", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "strOther"}}, vars, lastResult, true, false, nil},      // "hello" is falsy, NOT "hello" is true
		{"NOT Str True", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "strTrue"}}, vars, lastResult, false, false, nil},       // "true" is truthy, NOT "true" is false
		{"NOT Nil", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "nilVar"}}, vars, lastResult, true, false, nil},              // nil is falsy, NOT nil is true
		{"NOT Not Found", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "missing"}}, vars, lastResult, true, false, nil},       // not found (nil) is falsy, NOT is true
		{"NOT Error Operand", UnaryOpNode{Operator: "NOT", Operand: NumberLiteralNode{Value: 3.14}}, vars, lastResult, false, false, nil}, // Float 3.14 is truthy, NOT is false

		// --- AND Operator ---
		{"AND True True", BinaryOpNode{Left: VariableNode{Name: "trueVar"}, Operator: "AND", Right: BooleanLiteralNode{Value: true}}, vars, lastResult, true, false, nil},
		{"AND True False", BinaryOpNode{Left: VariableNode{Name: "trueVar"}, Operator: "AND", Right: VariableNode{Name: "falseVar"}}, vars, lastResult, false, false, nil},
		{"AND False True", BinaryOpNode{Left: VariableNode{Name: "falseVar"}, Operator: "AND", Right: VariableNode{Name: "trueVar"}}, vars, lastResult, false, false, nil},
		{"AND False False", BinaryOpNode{Left: VariableNode{Name: "falseVar"}, Operator: "AND", Right: VariableNode{Name: "falseVar"}}, vars, lastResult, false, false, nil},
		{"AND Num1 StrTrue", BinaryOpNode{Left: VariableNode{Name: "numOne"}, Operator: "AND", Right: VariableNode{Name: "strTrue"}}, vars, lastResult, true, false, nil},     // 1(T) AND "true"(T) -> T
		{"AND Num0 StrTrue", BinaryOpNode{Left: VariableNode{Name: "numZero"}, Operator: "AND", Right: VariableNode{Name: "strTrue"}}, vars, lastResult, false, false, nil},   // 0(F) AND "true"(T) -> F
		{"AND Nil True", BinaryOpNode{Left: VariableNode{Name: "nilVar"}, Operator: "AND", Right: VariableNode{Name: "trueVar"}}, vars, lastResult, false, false, nil},        // nil(F) AND true(T) -> F
		{"AND True Nil", BinaryOpNode{Left: VariableNode{Name: "trueVar"}, Operator: "AND", Right: VariableNode{Name: "nilVar"}}, vars, lastResult, false, false, nil},        // true(T) AND nil(F) -> F
		{"AND Not Found Left", BinaryOpNode{Left: VariableNode{Name: "missing"}, Operator: "AND", Right: VariableNode{Name: "trueVar"}}, vars, lastResult, false, false, nil}, // nil(F) AND true(T) -> F

		// --- OR Operator ---
		{"OR True True", BinaryOpNode{Left: VariableNode{Name: "trueVar"}, Operator: "OR", Right: BooleanLiteralNode{Value: true}}, vars, lastResult, true, false, nil},
		{"OR True False", BinaryOpNode{Left: VariableNode{Name: "trueVar"}, Operator: "OR", Right: VariableNode{Name: "falseVar"}}, vars, lastResult, true, false, nil},
		{"OR False True", BinaryOpNode{Left: VariableNode{Name: "falseVar"}, Operator: "OR", Right: VariableNode{Name: "trueVar"}}, vars, lastResult, true, false, nil},
		{"OR False False", BinaryOpNode{Left: VariableNode{Name: "falseVar"}, Operator: "OR", Right: VariableNode{Name: "falseVar"}}, vars, lastResult, false, false, nil},
		{"OR Num0 StrFalse", BinaryOpNode{Left: VariableNode{Name: "numZero"}, Operator: "OR", Right: VariableNode{Name: "strFalse"}}, vars, lastResult, false, false, nil}, // 0(F) OR "false"(F) -> F
		{"OR Num1 StrFalse", BinaryOpNode{Left: VariableNode{Name: "numOne"}, Operator: "OR", Right: VariableNode{Name: "strFalse"}}, vars, lastResult, true, false, nil},   // 1(T) OR "false"(F) -> T
		// FIX: Updated expectation for OR StrOther Nil
		{"OR StrOther Nil", BinaryOpNode{Left: VariableNode{Name: "strOther"}, Operator: "OR", Right: VariableNode{Name: "nilVar"}}, vars, lastResult, false, false, nil},     // "hello"(F) OR nil(F) -> F
		{"OR Nil False", BinaryOpNode{Left: VariableNode{Name: "nilVar"}, Operator: "OR", Right: VariableNode{Name: "falseVar"}}, vars, lastResult, false, false, nil},        // nil(F) OR false(F) -> F
		{"OR Not Found Right", BinaryOpNode{Left: VariableNode{Name: "falseVar"}, Operator: "OR", Right: VariableNode{Name: "missing"}}, vars, lastResult, false, false, nil}, // false(F) OR nil(F) -> F

		// --- Bitwise AND (&) ---
		{"Bitwise AND 5&3", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "&", Right: VariableNode{Name: "num3"}}, vars, lastResult, int64(1), false, nil}, // 0101 & 0011 = 0001
		{"Bitwise AND 5&0", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "&", Right: VariableNode{Name: "numZero"}}, vars, lastResult, int64(0), false, nil},
		{"Bitwise AND Error Float", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "&", Right: VariableNode{Name: "floatOne"}}, vars, lastResult, nil, true, ErrInvalidOperandType},
		{"Bitwise AND Error String", BinaryOpNode{Left: VariableNode{Name: "strOne"}, Operator: "&", Right: VariableNode{Name: "num3"}}, vars, lastResult, nil, true, ErrInvalidOperandType},
		{"Bitwise AND Error Nil", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "&", Right: VariableNode{Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},

		// --- Bitwise OR (|) ---
		{"Bitwise OR 5|3", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "|", Right: VariableNode{Name: "num3"}}, vars, lastResult, int64(7), false, nil}, // 0101 | 0011 = 0111
		{"Bitwise OR 5|0", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "|", Right: VariableNode{Name: "numZero"}}, vars, lastResult, int64(5), false, nil},
		{"Bitwise OR Error Float", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "|", Right: VariableNode{Name: "floatOne"}}, vars, lastResult, nil, true, ErrInvalidOperandType},
		{"Bitwise OR Error String", BinaryOpNode{Left: VariableNode{Name: "strOne"}, Operator: "|", Right: VariableNode{Name: "num3"}}, vars, lastResult, nil, true, ErrInvalidOperandType},
		{"Bitwise OR Error Nil", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "|", Right: VariableNode{Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},

		// --- Bitwise XOR (^) ---
		{"Bitwise XOR 5^3", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "^", Right: VariableNode{Name: "num3"}}, vars, lastResult, int64(6), false, nil}, // 0101 ^ 0011 = 0110
		{"Bitwise XOR 5^5", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "^", Right: VariableNode{Name: "num5"}}, vars, lastResult, int64(0), false, nil},
		{"Bitwise XOR Error Float", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "^", Right: VariableNode{Name: "floatOne"}}, vars, lastResult, nil, true, ErrInvalidOperandType},
		{"Bitwise XOR Error String", BinaryOpNode{Left: VariableNode{Name: "strOne"}, Operator: "^", Right: VariableNode{Name: "num3"}}, vars, lastResult, nil, true, ErrInvalidOperandType},
		{"Bitwise XOR Error Nil", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "^", Right: VariableNode{Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},

		// --- Bitwise NOT (~) ---
		{"Bitwise NOT ~5", UnaryOpNode{Operator: "~", Operand: VariableNode{Name: "num5"}}, vars, lastResult, int64(-6), false, nil}, // ~0...0101 = 1...1010 which is -6
		{"Bitwise NOT ~0", UnaryOpNode{Operator: "~", Operand: VariableNode{Name: "numZero"}}, vars, lastResult, int64(-1), false, nil},
		{"Bitwise NOT Error Float", UnaryOpNode{Operator: "~", Operand: VariableNode{Name: "floatNonInt"}}, vars, lastResult, nil, true, ErrInvalidOperandType},
		{"Bitwise NOT Error String", UnaryOpNode{Operator: "~", Operand: VariableNode{Name: "strOther"}}, vars, lastResult, nil, true, ErrInvalidOperandType},
		{"Bitwise NOT Error Nil", UnaryOpNode{Operator: "~", Operand: VariableNode{Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runEvalExpressionTest(t, tt) // Use the helper
		})
	}
}

// Helper Function (already provided in testing_helpers_test.go)
// Assume runEvalExpressionTest is correctly defined elsewhere and handles the EvalTestCase struct.

// EvalTestCase struct (already provided in testing_helpers_test.go)
/*
type EvalTestCase struct {
	Name            string
	InputNode       interface{}
	InitialVars     map[string]interface{}
	LastResult      interface{}
	Expected        interface{}
	WantErr         bool
	ExpectedErrorIs error // Use sentinel error or nil
}
*/

// runEvalExpressionTest (already provided in testing_helpers_test.go)
/*
func runEvalExpressionTest(t *testing.T, tt EvalTestCase) {
	t.Helper()
	interp, _ := NewTestInterpreter(t, tt.InitialVars, tt.LastResult) // Pass t

	got, err := interp.evaluateExpression(tt.InputNode)

	// Use errors.Is for specific error checking
	if tt.WantErr {
		if err == nil {
			t.Errorf("%s: Expected error, but got nil", tt.Name)
			return
		}
		if tt.ExpectedErrorIs != nil && !errors.Is(err, tt.ExpectedErrorIs) {
			t.Errorf("%s: Error mismatch.\nExpected: %v\nGot:      %v", tt.Name, tt.ExpectedErrorIs, err)
		} else if tt.ExpectedErrorIs == nil {
			// Just ensure *an* error occurred if ExpectedErrorIs is nil but WantErr is true
			t.Logf("%s: Got expected error: %v", tt.Name, err) // Log for info
		}
	} else {
		if err != nil {
			t.Errorf("%s: Unexpected error: %v", tt.Name, err)
		} else if !reflect.DeepEqual(got, tt.Expected) {
			t.Errorf("%s: Result mismatch.\nInput Node: %+v\nExpected:   %v (%T)\nGot:        %v (%T)",
				tt.Name, tt.InputNode, tt.Expected, tt.Expected, got, got)
		}
	}
}
*/
