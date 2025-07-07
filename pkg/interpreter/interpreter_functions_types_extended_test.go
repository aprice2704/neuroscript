// NeuroScript Version: 0.5.2
// File version: 1.0.0
// Purpose: Adds extended, dedicated tests for type-checking built-in functions like 'is_error'.
// filename: pkg/interpreter/interpreter_functions_types_extended_test.go
// nlines: 60
// risk_rating: LOW

package interpreter

import (
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestIsErrorFunction_Extended(t *testing.T) {
	testCases := []struct {
		name string
		arg  interface{}
		want bool
	}{
		{
			"is_error with a complete error map",
			lang.NewMapValue(map[string]lang.Value{
				lang.ErrorKeyCode:    lang.StringValue{Value: "E_CODE"},
				lang.ErrorKeyMessage: lang.StringValue{Value: "An error occurred"},
			}),
			true,
		},
		{
			"is_error with incomplete error map (missing message)",
			lang.NewMapValue(map[string]lang.Value{
				lang.ErrorKeyCode: lang.StringValue{Value: "E_CODE"},
			}),
			false,
		},
		{
			"is_error with incomplete error map (missing code)",
			lang.NewMapValue(map[string]lang.Value{
				lang.ErrorKeyMessage: lang.StringValue{Value: "An error occurred"},
			}),
			false,
		},
		{
			"is_error with a non-error map",
			lang.NewMapValue(map[string]lang.Value{"key": lang.StringValue{Value: "value"}}),
			false,
		},
		{
			"is_error with a timedate value",
			lang.TimedateValue{Value: time.Now()},
			false,
		},
		{
			"is_error with a fuzzy value",
			lang.NewFuzzyValue(0.5),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := evaluateBuiltInFunction("is_error", []interface{}{tc.arg})
			if err != nil {
				t.Fatalf("evaluateBuiltInFunction failed for is_error: %v", err)
			}
			got, ok := result.(lang.BoolValue)
			if !ok {
				t.Fatalf("Expected BoolValue, got %T", result)
			}
			if got.Value != tc.want {
				t.Errorf("Function 'is_error' with arg '%v': got %v, want %v", tc.arg, got.Value, tc.want)
			}
		})
	}
}
