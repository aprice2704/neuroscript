// NeuroScript Version: 0.3.5
// File version: 1.1.0
// Purpose: Corrected test logic to properly handle the (Value, error) return signature of evaluateBinaryOp.
// filename: pkg/runtime/evaluation_comparison_test.go

package runtime

import (
	"testing"
)

func TestEvaluateComparison(t *testing.T) {
	testCases := []struct {
		name     string
		left     Value
		right    Value
		op       string
		vars     map[string]Value
		last     Value
		expected bool
		wantErr  bool
	}{
		{
			name:     "Equal Numbers",
			left:     NumberValue{Value: 5},
			right:    NumberValue{Value: 5},
			op:       "==",
			expected: true,
		},
		{
			name:     "Unequal Strings",
			left:     StringValue{Value: "hello"},
			right:    StringValue{Value: "world"},
			op:       "!=",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interpreter, err := NewTestInterpreter(t, tc.vars, tc.last)
			if err != nil {
				t.Fatalf("Failed to create test interpreter: %v", err)
			}

			// FIX: Correctly handle the (Value, error) return values.
			result, err := interpreter.evaluateBinaryOp(tc.left, tc.right, tc.op)
			if (err != nil) != tc.wantErr {
				t.Fatalf("evaluateBinaryOp() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantErr {
				return // If an error was expected, stop here.
			}

			boolResult, ok := result.(BoolValue)
			if !ok {
				t.Fatalf("Expected a BoolValue result, but got %T", result)
			}

			if boolResult.Value != tc.expected {
				t.Errorf("evaluateBinaryOp() = %v, want %v", boolResult.Value, tc.expected)
			}
		})
	}
}
