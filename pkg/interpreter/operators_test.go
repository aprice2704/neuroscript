// NeuroScript Version: 0.8.0
// File version: 2.0.0
// Purpose: Refactored to test operators via script execution, aligning with the post-refactor evaluation architecture.
// filename: pkg/interpreter/interpreter_operators_test.go
// nlines: 155
// risk_rating: LOW

package interpreter_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// runOperatorTest is a helper for this file, using the TestHarness to run a single expression.
func runOperatorTest(t *testing.T, scriptExpression string) (lang.Value, error) {
	t.Helper()
	h := NewTestHarness(t)
	script := fmt.Sprintf(`func main() returns result means return %s endfunc`, scriptExpression)
	h.T.Logf("[DEBUG] Turn 1: Harness created for script: %s", script)
	// Using ExecuteScriptString as it's a convenient wrapper for this kind of test.
	result, runErr := h.Interpreter.ExecuteScriptString("main", script, nil)
	h.T.Logf("[DEBUG] Turn 2: Run('main') completed. Result: %#v, Error: %v", result, runErr)
	return result, runErr
}

func TestPerformArithmetic(t *testing.T) {
	testCases := []struct {
		name     string
		scriptOp string
		want     lang.Value
		wantErr  error
	}{
		{"Subtract", "10 - 4", lang.NumberValue{Value: 6}, nil},
		{"Multiply", "5 * 3", lang.NumberValue{Value: 15}, nil},
		{"Divide", "20 / 4", lang.NumberValue{Value: 5}, nil},
		{"Power", "2 ** 3", lang.NumberValue{Value: 8}, nil},
		{"Modulo", "10 % 3", lang.NumberValue{Value: 1}, nil},
		{"Division by zero", "10 / 0", nil, lang.ErrDivisionByZero},
		{"Invalid type", `"a" * 1`, nil, lang.ErrInvalidOperandType},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := runOperatorTest(t, tc.scriptOp)
			if tc.wantErr != nil {
				var rtErr *lang.RuntimeError
				if errors.As(err, &rtErr) {
					if !errors.Is(rtErr.Unwrap(), tc.wantErr) {
						t.Fatalf("Expected error: %v, got: %v", tc.wantErr, rtErr.Unwrap())
					}
				} else {
					t.Fatalf("Expected a runtime error, but got %T: %v", err, err)
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tc.wantErr == nil && got != tc.want {
				t.Errorf("Expected result: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestPerformStringConcatOrNumericAdd(t *testing.T) {
	testCases := []struct {
		name     string
		scriptOp string
		want     lang.Value
	}{
		{"Add numbers", "5 + 10", lang.NumberValue{Value: 15}},
		{"Concat strings", `"hello " + "world"`, lang.StringValue{Value: "hello world"}},
		{"Concat string and number", `"age: " + 30`, lang.StringValue{Value: "age: 30"}},
		{"Concat number and string", `30 + " years"`, lang.StringValue{Value: "30 years"}},
		{"Concat with nil", `"value: " + nil`, lang.StringValue{Value: "value: nil"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := runOperatorTest(t, tc.scriptOp)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if got.String() != tc.want.String() { // Use string comparison for simplicity with nil
				t.Errorf("Expected result: %#v, got: %#v", tc.want, got)
			}
		})
	}
}

func TestPerformComparison(t *testing.T) {
	// Timedate comparisons require setting variables, so we use a different test structure.
	t.Run("Time comparison", func(t *testing.T) {
		h := NewTestHarness(t)
		t1 := time.Now()
		t2 := t1.Add(time.Second)
		h.Interpreter.SetVariable("t1", lang.TimedateValue{Value: t1})
		h.Interpreter.SetVariable("t2", lang.TimedateValue{Value: t2})
		script := `func main() means return t1 < t2, t2 > t1 endfunc`
		result, err := h.Interpreter.ExecuteScriptString("main", script, nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		list, ok := result.(lang.ListValue)
		if !ok || len(list.Value) != 2 {
			t.Fatalf("Expected list of 2 booleans, got %T", result)
		}
		if v1, _ := list.Value[0].(lang.BoolValue); !v1.Value {
			t.Error("Expected t1 < t2 to be true")
		}
		if v2, _ := list.Value[1].(lang.BoolValue); !v2.Value {
			t.Error("Expected t2 > t1 to be true")
		}
	})

	testCases := []struct {
		name     string
		scriptOp string
		want     lang.Value
		wantErr  bool
	}{
		{"Equal numbers", "5 == 5", lang.BoolValue{Value: true}, false},
		{"Not equal strings", `"a" != "b"`, lang.BoolValue{Value: true}, false},
		{"Less than", "4 < 5", lang.BoolValue{Value: true}, false},
		{"Greater than or equal", "5 >= 5", lang.BoolValue{Value: true}, false},
		{"Invalid comparison", `"a" > 1`, nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := runOperatorTest(t, tc.scriptOp)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Unexpected error state. Got err: %v, wantErr: %t", err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("Expected result: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestPerformBitwise(t *testing.T) {
	testCases := []struct {
		name     string
		scriptOp string
		want     lang.Value
		wantErr  error
	}{
		{"AND", "5 & 3", lang.NumberValue{Value: 1}, nil}, // 101 & 011 = 001
		{"OR", "5 | 3", lang.NumberValue{Value: 7}, nil},  // 101 | 011 = 111
		{"XOR", "5 ^ 3", lang.NumberValue{Value: 6}, nil}, // 101 ^ 011 = 110
		{"Invalid type (float)", "5.5 & 3", nil, lang.ErrInvalidOperandTypeInteger},
		{"Invalid type (string)", `"a" | 3`, nil, lang.ErrInvalidOperandTypeInteger},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := runOperatorTest(t, tc.scriptOp)
			if tc.wantErr != nil {
				var rtErr *lang.RuntimeError
				if errors.As(err, &rtErr) {
					if !errors.Is(rtErr.Unwrap(), tc.wantErr) {
						t.Fatalf("Expected error: %v, got: %v", tc.wantErr, rtErr.Unwrap())
					}
				} else {
					t.Fatalf("Expected a runtime error, but got %T: %v", err, err)
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tc.wantErr == nil && got != tc.want {
				t.Errorf("Expected result: %v, got: %v", tc.want, got)
			}
		})
	}
}
