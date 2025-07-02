// NeuroScript Version: 0.3.5
// File version: 1.1.0
// Purpose: Corrected test logic to properly handle the (Value, error) return signature of evaluateBinaryOp.
// filename: pkg/interpreter/internal/eval/evaluation_comparison_test.go

package eval

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestEvaluateComparison(t *testing.T) {
	testCases := []struct {
		name		string
		left		lang.Value
		right		lang.Value
		op		string
		vars		map[string]lang.Value
		last		lang.Value
		expected	bool
		wantErr		bool
	}{
		{
			name:		"Equal Numbers",
			left:		lang.NumberValue{Value: 5},
			right:		lang.NumberValue{Value: 5},
			op:		"==",
			expected:	true,
		},
		{
			name:		"Unequal Strings",
			left:		lang.StringValue{Value: "hello"},
			right:		lang.StringValue{Value: "world"},
			op:		"!=",
			expected:	true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interpreter, err := llm.NewTestInterpreter(t, tc.vars, tc.last)
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
				return	// If an error was expected, stop here.
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