// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Contains tests for unexported functions in evaluation_functions.go.
// filename: pkg/interpreter/evaluation_functions_internal_test.go
package interpreter

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestEvaluateBuiltInFunction_Len_Internal(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"string", "hello", 5},
		{"unicode string", "你好", 2},
		{"empty string", "", 0},
		{"list", []interface{}{1, true, "three"}, 3},
		{"empty list", []interface{}{}, 0},
		{"map", map[string]interface{}{"a": 1, "b": 2}, 2},
		{"empty map", map[string]interface{}{}, 0},
		{"number", 123.45, 1},
		{"boolean", true, 1},
		{"nil", nil, 0},
		{"error type", errors.New("an error"), 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := []interface{}{tc.input}
			result, err := evaluateBuiltInFunction("len", args)
			if err != nil {
				t.Fatalf("evaluateBuiltInFunction failed: %v", err)
			}

			numResult, ok := result.(lang.NumberValue)
			if !ok {
				t.Fatalf("Expected NumberValue, got %T", result)
			}

			if numResult.Value != tc.expected {
				t.Errorf("Expected length %f, got %f", tc.expected, numResult.Value)
			}
		})
	}

	t.Run("incorrect argument count", func(t *testing.T) {
		_, err := evaluateBuiltInFunction("len", []interface{}{})
		if !errors.Is(err, lang.ErrIncorrectArgCount) {
			t.Errorf("Expected ErrIncorrectArgCount, got %v", err)
		}

		_, err = evaluateBuiltInFunction("len", []interface{}{"a", "b"})
		if !errors.Is(err, lang.ErrIncorrectArgCount) {
			t.Errorf("Expected ErrIncorrectArgCount, got %v", err)
		}
	})
}
