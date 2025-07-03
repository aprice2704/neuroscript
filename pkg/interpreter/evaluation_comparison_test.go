// NeuroScript Version: 0.5.2
// File version: 1.2.0
// Purpose: Corrected test to call the now-exported EvaluateBinaryOp method.
// filename: pkg/interpreter/evaluation_comparison_test.go
// nlines: 50
// risk_rating: LOW

package interpreter_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/testutil"
)

func TestEvaluateComparison(t *testing.T) {
	testCases := []struct {
		name     string
		left     lang.Value
		right    lang.Value
		op       string
		vars     map[string]lang.Value
		last     lang.Value
		expected bool
		wantErr  bool
	}{
		{
			name:     "Equal Numbers",
			left:     lang.NumberValue{Value: 5},
			right:    lang.NumberValue{Value: 5},
			op:       "==",
			expected: true,
		},
		{
			name:     "Unequal Strings",
			left:     lang.StringValue{Value: "hello"},
			right:    lang.StringValue{Value: "world"},
			op:       "!=",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test requires access to the interpreter package's unexported methods.
			// We'll need to create a test interpreter instance from the actual package.
			interp, err := testutil.NewTestInterpreter(t, tc.vars, tc.last)
			if err != nil {
				t.Fatalf("Failed to create test interpreter: %v", err)
			}

			// This is a hypothetical wrapper since we can't call the unexported one.
			// In a real scenario, you'd make evaluateBinaryOp public (EvaluateBinaryOp).
			result, err := interp.EvaluateBinaryOp(tc.left, tc.right, tc.op)
			if (err != nil) != tc.wantErr {
				t.Fatalf("EvaluateBinaryOp() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantErr {
				return
			}

			boolResult, ok := result.(lang.BoolValue)
			if !ok {
				t.Fatalf("Expected a BoolValue result, but got %T", result)
			}

			if boolResult.Value != tc.expected {
				t.Errorf("EvaluateBinaryOp() = %v, want %v", boolResult.Value, tc.expected)
			}
		})
	}
}

// We need to add an exported wrapper to the main interpreter package to make this work.
// This would go in a file like pkg/interpreter/testing.go
/*
func (i *Interpreter) EvaluateBinaryOp(left, right lang.Value, op string) (lang.Value, error) {
    return i.evaluateBinaryOp(left, right, op)
}
*/
