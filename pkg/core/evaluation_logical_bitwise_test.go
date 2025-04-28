package core

import (
	"testing"
)

// --- Test Suite for Logical and Bitwise Operations ---
func TestLogicalBitwiseOps(t *testing.T) {
	// FIX: Added list/map variables
	vars := map[string]interface{}{
		"trueVar":      true,
		"falseVar":     false,
		"numOne":       int64(1),
		"numZero":      int64(0),
		"floatOne":     float64(1.0),
		"floatZero":    float64(0.0),
		"strTrue":      "true",   // Truthy string
		"strFalse":     "false",  // Falsy string
		"strEmpty":     "",       // Falsy string (Zero Value)
		"strOther":     "hello",  // Falsy string (but non-zero value)
		"strOne":       "1",      // Truthy string
		"nilVar":       nil,      // Falsy (Zero Value)
		"num5":         int64(5), // 0101
		"num3":         int64(3), // 0011
		"floatNonInt":  float64(3.14),
		"emptyListVar": []interface{}{},                      // Falsy (Zero Value)
		"listVar":      []interface{}{int64(1), "a"},         // Truthy (non-zero value)
		"emptyMapVar":  map[string]interface{}{},             // Falsy (Zero Value)
		"mapVar":       map[string]interface{}{"key": "val"}, // Truthy (non-zero value)
	}
	lastResult := "LastResult" // Placeholder for LAST tests, value doesn't matter here

	tests := []EvalTestCase{
		// --- NOT Operator ---
		{"NOT True Literal", UnaryOpNode{Operator: "NOT", Operand: BooleanLiteralNode{Value: true}}, vars, lastResult, false, false, nil},
		{"NOT False Literal", UnaryOpNode{Operator: "NOT", Operand: BooleanLiteralNode{Value: false}}, vars, lastResult, true, false, nil},
		{"NOT True Var", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "trueVar"}}, vars, lastResult, false, false, nil},
		{"NOT False Var", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "falseVar"}}, vars, lastResult, true, false, nil},
		{"NOT Num NonZero", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "numOne"}}, vars, lastResult, false, false, nil},             // 1 is truthy, NOT 1 is false
		{"NOT Num Zero", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "numZero"}}, vars, lastResult, true, false, nil},                // 0 is falsy, NOT 0 is true
		{"NOT Str Empty", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "strEmpty"}}, vars, lastResult, true, false, nil},              // "" is falsy, NOT "" is true
		{"NOT Str Other", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "strOther"}}, vars, lastResult, true, false, nil},              // "hello" is falsy, NOT "hello" is true
		{"NOT Str True", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "strTrue"}}, vars, lastResult, false, false, nil},               // "true" is truthy, NOT "true" is false
		{"NOT Nil", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "nilVar"}}, vars, lastResult, true, false, nil},                      // nil is falsy, NOT nil is true
		{"NOT Empty List", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "emptyListVar"}}, vars, lastResult, true, false, nil},         // [] is falsy, NOT [] is true
		{"NOT List", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "listVar"}}, vars, lastResult, false, false, nil},                   // [1,"a"] is truthy, NOT [...] is false
		{"NOT Empty Map", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "emptyMapVar"}}, vars, lastResult, true, false, nil},           // {} is falsy, NOT {} is true
		{"NOT Map", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "mapVar"}}, vars, lastResult, false, false, nil},                     // {k:v} is truthy, NOT {...} is false
		{"NOT Not Found", UnaryOpNode{Operator: "NOT", Operand: VariableNode{Name: "missing"}}, vars, lastResult, nil, true, ErrVariableNotFound}, // Expect error

		// --- AND Operator ---
		{"AND True True", BinaryOpNode{Left: VariableNode{Name: "trueVar"}, Operator: "AND", Right: BooleanLiteralNode{Value: true}}, vars, lastResult, true, false, nil},
		{"AND True False", BinaryOpNode{Left: VariableNode{Name: "trueVar"}, Operator: "AND", Right: VariableNode{Name: "falseVar"}}, vars, lastResult, false, false, nil},
		{"AND False True", BinaryOpNode{Left: VariableNode{Name: "falseVar"}, Operator: "AND", Right: VariableNode{Name: "trueVar"}}, vars, lastResult, false, false, nil},                 // Short-circuits
		{"AND False False", BinaryOpNode{Left: VariableNode{Name: "falseVar"}, Operator: "AND", Right: VariableNode{Name: "falseVar"}}, vars, lastResult, false, false, nil},               // Short-circuits
		{"AND Num1 StrTrue", BinaryOpNode{Left: VariableNode{Name: "numOne"}, Operator: "AND", Right: VariableNode{Name: "strTrue"}}, vars, lastResult, true, false, nil},                  // 1(T) AND "true"(T) -> T
		{"AND Num0 StrTrue", BinaryOpNode{Left: VariableNode{Name: "numZero"}, Operator: "AND", Right: VariableNode{Name: "strTrue"}}, vars, lastResult, false, false, nil},                // 0(F) AND "true"(T) -> F (Short-circuits)
		{"AND Nil True", BinaryOpNode{Left: VariableNode{Name: "nilVar"}, Operator: "AND", Right: VariableNode{Name: "trueVar"}}, vars, lastResult, false, false, nil},                     // nil(F) AND true(T) -> F (Short-circuits)
		{"AND True Nil", BinaryOpNode{Left: VariableNode{Name: "trueVar"}, Operator: "AND", Right: VariableNode{Name: "nilVar"}}, vars, lastResult, false, false, nil},                     // true(T) AND nil(F) -> F
		{"AND Not Found Left", BinaryOpNode{Left: VariableNode{Name: "missing"}, Operator: "AND", Right: VariableNode{Name: "trueVar"}}, vars, lastResult, nil, true, ErrVariableNotFound}, // Expect error

		// --- OR Operator ---
		{"OR True True", BinaryOpNode{Left: VariableNode{Name: "trueVar"}, Operator: "OR", Right: BooleanLiteralNode{Value: true}}, vars, lastResult, true, false, nil}, // Short-circuits
		{"OR True False", BinaryOpNode{Left: VariableNode{Name: "trueVar"}, Operator: "OR", Right: VariableNode{Name: "falseVar"}}, vars, lastResult, true, false, nil}, // Short-circuits
		{"OR False True", BinaryOpNode{Left: VariableNode{Name: "falseVar"}, Operator: "OR", Right: VariableNode{Name: "trueVar"}}, vars, lastResult, true, false, nil},
		{"OR False False", BinaryOpNode{Left: VariableNode{Name: "falseVar"}, Operator: "OR", Right: VariableNode{Name: "falseVar"}}, vars, lastResult, false, false, nil},
		{"OR Num0 StrFalse", BinaryOpNode{Left: VariableNode{Name: "numZero"}, Operator: "OR", Right: VariableNode{Name: "strFalse"}}, vars, lastResult, false, false, nil},                // 0(F) OR "false"(F) -> F
		{"OR Num1 StrFalse", BinaryOpNode{Left: VariableNode{Name: "numOne"}, Operator: "OR", Right: VariableNode{Name: "strFalse"}}, vars, lastResult, true, false, nil},                  // 1(T) OR "false"(F) -> T (Short-circuits)
		{"OR StrOther Nil", BinaryOpNode{Left: VariableNode{Name: "strOther"}, Operator: "OR", Right: VariableNode{Name: "nilVar"}}, vars, lastResult, false, false, nil},                  // "hello"(F) OR nil(F) -> F
		{"OR Nil False", BinaryOpNode{Left: VariableNode{Name: "nilVar"}, Operator: "OR", Right: VariableNode{Name: "falseVar"}}, vars, lastResult, false, false, nil},                     // nil(F) OR false(F) -> F
		{"OR Not Found Right", BinaryOpNode{Left: VariableNode{Name: "falseVar"}, Operator: "OR", Right: VariableNode{Name: "missing"}}, vars, lastResult, nil, true, ErrVariableNotFound}, // Expect error

		// --- Bitwise AND (&) ---
		{"Bitwise AND 5&3", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "&", Right: VariableNode{Name: "num3"}}, vars, lastResult, int64(1), false, nil}, // 0101 & 0011 = 0001
		{"Bitwise AND 5&0", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "&", Right: VariableNode{Name: "numZero"}}, vars, lastResult, int64(0), false, nil},
		{"Bitwise AND Error Float", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "&", Right: VariableNode{Name: "floatOne"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise AND Error String", BinaryOpNode{Left: VariableNode{Name: "strOne"}, Operator: "&", Right: VariableNode{Name: "num3"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise AND Error Nil", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "&", Right: VariableNode{Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},

		// --- Bitwise OR (|) ---
		{"Bitwise OR 5|3", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "|", Right: VariableNode{Name: "num3"}}, vars, lastResult, int64(7), false, nil}, // 0101 | 0011 = 0111
		{"Bitwise OR 5|0", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "|", Right: VariableNode{Name: "numZero"}}, vars, lastResult, int64(5), false, nil},
		{"Bitwise OR Error Float", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "|", Right: VariableNode{Name: "floatOne"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise OR Error String", BinaryOpNode{Left: VariableNode{Name: "strOne"}, Operator: "|", Right: VariableNode{Name: "num3"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise OR Error Nil", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "|", Right: VariableNode{Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},

		// --- Bitwise XOR (^) ---
		{"Bitwise XOR 5^3", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "^", Right: VariableNode{Name: "num3"}}, vars, lastResult, int64(6), false, nil}, // 0101 ^ 0011 = 0110
		{"Bitwise XOR 5^5", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "^", Right: VariableNode{Name: "num5"}}, vars, lastResult, int64(0), false, nil},
		{"Bitwise XOR Error Float", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "^", Right: VariableNode{Name: "floatOne"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise XOR Error String", BinaryOpNode{Left: VariableNode{Name: "strOne"}, Operator: "^", Right: VariableNode{Name: "num3"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise XOR Error Nil", BinaryOpNode{Left: VariableNode{Name: "num5"}, Operator: "^", Right: VariableNode{Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},

		// --- Bitwise NOT (~) ---
		{"Bitwise NOT ~5", UnaryOpNode{Operator: "~", Operand: VariableNode{Name: "num5"}}, vars, lastResult, int64(-6), false, nil}, // ~0...0101 = 1...1010 which is -6
		{"Bitwise NOT ~0", UnaryOpNode{Operator: "~", Operand: VariableNode{Name: "numZero"}}, vars, lastResult, int64(-1), false, nil},
		{"Bitwise NOT Error Float", UnaryOpNode{Operator: "~", Operand: VariableNode{Name: "floatNonInt"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise NOT Error String", UnaryOpNode{Operator: "~", Operand: VariableNode{Name: "strOther"}}, vars, lastResult, nil, true, ErrInvalidOperandTypeInteger},
		{"Bitwise NOT Error Nil", UnaryOpNode{Operator: "~", Operand: VariableNode{Name: "nilVar"}}, vars, lastResult, nil, true, ErrNilOperand},

		// +++ Tests for 'no' and 'some' +++
		// --- no Operator (Zero Value Check) ---
		{"no nilVar", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "nilVar"}}, vars, lastResult, true, false, nil},
		{"no strEmpty", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "strEmpty"}}, vars, lastResult, true, false, nil},
		{"no numZero", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "numZero"}}, vars, lastResult, true, false, nil},
		{"no floatZero", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "floatZero"}}, vars, lastResult, true, false, nil},
		{"no falseVar", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "falseVar"}}, vars, lastResult, true, false, nil},
		{"no emptyListVar", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "emptyListVar"}}, vars, lastResult, true, false, nil},
		{"no emptyMapVar", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "emptyMapVar"}}, vars, lastResult, true, false, nil},
		// FIX: Expect error for variable not found
		{"no notFoundVar", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "notFoundVar"}}, vars, lastResult, nil, true, ErrVariableNotFound},
		{"no strOther", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "strOther"}}, vars, lastResult, false, false, nil},
		{"no numOne", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "numOne"}}, vars, lastResult, false, false, nil},
		{"no floatOne", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "floatOne"}}, vars, lastResult, false, false, nil},
		{"no trueVar", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "trueVar"}}, vars, lastResult, false, false, nil},
		{"no listVar", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "listVar"}}, vars, lastResult, false, false, nil},
		{"no mapVar", UnaryOpNode{Operator: "no", Operand: VariableNode{Name: "mapVar"}}, vars, lastResult, false, false, nil},

		// --- some Operator (Non-Zero Value Check) ---
		{"some nilVar", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "nilVar"}}, vars, lastResult, false, false, nil},
		{"some strEmpty", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "strEmpty"}}, vars, lastResult, false, false, nil},
		{"some numZero", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "numZero"}}, vars, lastResult, false, false, nil},
		{"some floatZero", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "floatZero"}}, vars, lastResult, false, false, nil},
		{"some falseVar", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "falseVar"}}, vars, lastResult, false, false, nil},
		{"some emptyListVar", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "emptyListVar"}}, vars, lastResult, false, false, nil},
		{"some emptyMapVar", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "emptyMapVar"}}, vars, lastResult, false, false, nil},
		// FIX: Expect error for variable not found
		{"some notFoundVar", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "notFoundVar"}}, vars, lastResult, nil, true, ErrVariableNotFound},
		{"some strOther", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "strOther"}}, vars, lastResult, true, false, nil},
		{"some numOne", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "numOne"}}, vars, lastResult, true, false, nil},
		{"some floatOne", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "floatOne"}}, vars, lastResult, true, false, nil},
		{"some trueVar", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "trueVar"}}, vars, lastResult, true, false, nil},
		{"some listVar", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "listVar"}}, vars, lastResult, true, false, nil},
		{"some mapVar", UnaryOpNode{Operator: "some", Operand: VariableNode{Name: "mapVar"}}, vars, lastResult, true, false, nil},
		// +++ END Tests +++
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Ensure errors is imported if runEvalExpressionTest is copied here
			runEvalExpressionTest(t, tt) // Use the helper from testing_helpers_test.go
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
	// Ensure errors is imported if runEvalExpressionTest is copied here
	// import "errors"
	interp, _ := NewDefaultTestInterpreter(t) // Use default interpreter
	if tt.InitialVars != nil {
		for k, v := range tt.InitialVars {
			interp.variables[k] = v
		}
	}
	interp.lastCallResult = tt.LastResult

	got, err := interp.evaluateExpression(tt.InputNode)

	// Use errors.Is for specific error checking
	if tt.WantErr {
		if err == nil {
			t.Errorf("%s: Expected error, but got nil", tt.Name)
			return
		}
		if tt.ExpectedErrorIs != nil && !errors.Is(err, tt.ExpectedErrorIs) {
			t.Errorf("%s: Error mismatch.\nExpected error wrapping [%v]\nGot:                 [%v]", tt.Name, tt.ExpectedErrorIs, err)
		} else if tt.ExpectedErrorIs == nil {
			// Just ensure *an* error occurred if ExpectedErrorIs is nil but WantErr is true
			t.Logf("%s: Got expected error: %v", tt.Name, err) // Log for info
		}
	} else { // No error wanted
		if err != nil {
			t.Errorf("%s: Unexpected error: %v", tt.Name, err)
		} else if !reflect.DeepEqual(got, tt.Expected) {
			t.Errorf("%s: Result mismatch.\nInput Node: %+v\nExpected:   %v (%T)\nGot:        %v (%T)",
				tt.Name, tt.InputNode, tt.Expected, tt.Expected, got, got)
		}
	}
}
*/
