// filename: pkg/lang/values_helpers_test.go
// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Added tests for numeric slice wrapping and to verify the lang.Wrap error message.
// nlines: 177
// risk_rating: MEDIUM
package lang

import (
	"reflect"
	"strings" // Added import
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestWrap(t *testing.T) {
	// Define a custom string type for testing the generic slice wrapping.
	type CustomString string
	customSlice := []CustomString{"custom1", "custom2"}
	expectedCustomSlice := ListValue{Value: []Value{StringValue{"custom1"}, StringValue{"custom2"}}}

	testCases := []struct {
		name     string
		input    any
		expected Value
		hasError bool
	}{
		{"nil", nil, NilValue{}, false},
		{"string", "hello", StringValue{"hello"}, false},
		{"int", 123, NumberValue{123}, false},
		{"float64", 3.14, NumberValue{3.14}, false},
		{"bool", true, BoolValue{true}, false},
		{"[]byte", []byte("bytes"), BytesValue{[]byte("bytes")}, false},
		{"time.Time", time.Unix(0, 0).UTC(), TimedateValue{time.Unix(0, 0).UTC()}, false},
		{
			name:     "[]string",
			input:    []string{"a", "b"},
			expected: ListValue{[]Value{StringValue{"a"}, StringValue{"b"}}},
			hasError: false,
		},
		{
			name:     "[]types.AgentModelName",
			input:    []types.AgentModelName{"model-a", "model-b"},
			expected: ListValue{[]Value{StringValue{"model-a"}, StringValue{"model-b"}}},
			hasError: false,
		},
		{
			name:     "[]CustomString",
			input:    customSlice,
			expected: expectedCustomSlice,
			hasError: false,
		},
		{
			name:     "[]any",
			input:    []any{"a", 1},
			expected: ListValue{[]Value{StringValue{"a"}, NumberValue{1}}},
			hasError: false,
		},
		{
			name: "map[string]any",
			input: map[string]any{
				"a": 1,
				"b": "two",
			},
			expected: MapValue{Value: map[string]Value{
				"a": NumberValue{1},
				"b": StringValue{"two"},
			}},
			hasError: false,
		},
		// --- MODIFIED & ADDED TESTS ---
		{
			name:     "[]int", // Was "unsupported type", now supported
			input:    []int{1, 2},
			expected: ListValue{[]Value{NumberValue{1}, NumberValue{2}}},
			hasError: false,
		},
		{
			name:     "[]int64",
			input:    []int64{10, 20},
			expected: ListValue{[]Value{NumberValue{10}, NumberValue{20}}},
			hasError: false,
		},
		{
			name:     "[]float64",
			input:    []float64{1.1, 2.2},
			expected: ListValue{[]Value{NumberValue{1.1}, NumberValue{2.2}}},
			hasError: false,
		},
		{
			name:     "[]uint",
			input:    []uint{5, 6},
			expected: ListValue{[]Value{NumberValue{5}, NumberValue{6}}},
			hasError: false,
		},
		{
			name:     "unsupported struct", // New test for unsupported type
			input:    struct{ F string }{"test"},
			expected: nil,
			hasError: true,
		},
		// --- END MODIFIED & ADDED TESTS ---
		{"already wrapped NilValue", NilValue{}, NilValue{}, false},
		{"already wrapped *NilValue", &NilValue{}, NilValue{}, false}, // Wrap should handle pointer
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrapped, err := Wrap(tc.input)

			if tc.hasError {
				if err == nil {
					t.Error("Expected an error, but got nil")
				}
				// Check for the specific error message from lang.Wrap
				if tc.name == "unsupported struct" {
					expectedErr := "lang.Wrap: unsupported type"
					if !strings.Contains(err.Error(), expectedErr) {
						t.Errorf("Expected error to contain %q, but got: %v", expectedErr, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				// Use reflect.DeepEqual for comparison
				if !reflect.DeepEqual(wrapped, tc.expected) {
					t.Errorf("Expected wrapped value %#v (%T), but got %#v (%T)", tc.expected, tc.expected, wrapped, wrapped)
				}
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	// Nested structure for testing recursion
	nestedMapVal := MapValue{Value: map[string]Value{
		"list": ListValue{[]Value{StringValue{"nested"}, NumberValue{99}}},
		"num":  NumberValue{123},
	}}
	nestedMapRaw := map[string]any{
		"list": []any{"nested", float64(99)},
		"num":  float64(123),
	}

	testCases := []struct {
		name     string
		input    Value // Input must be a Value type now
		expected any
	}{
		{"NilValue", NilValue{}, nil},
		{"*NilValue", &NilValue{}, nil}, // Test the specific fix
		{"StringValue", StringValue{"hello"}, "hello"},
		{"BytesValue", BytesValue{[]byte("bytes")}, []byte("bytes")},
		{"BoolValue", BoolValue{true}, true},
		{"NumberValue", NumberValue{123}, float64(123)},
		{"TimedateValue", TimedateValue{time.Unix(0, 0)}, time.Unix(0, 0)},
		{"FuzzyValue", FuzzyValue{0.5}, 0.5},
		{"ListValue", ListValue{[]Value{StringValue{"a"}, NumberValue{1}}}, []any{"a", float64(1)}},
		{"MapValue", MapValue{Value: map[string]Value{"a": NumberValue{1}}}, map[string]any{"a": float64(1)}},
		{"*MapValue", &MapValue{Value: map[string]Value{"a": NumberValue{1}}}, map[string]any{"a": float64(1)}},
		{"*MapValue (nil ptr)", (*MapValue)(nil), nil}, // Test nil pointer case
		{"ErrorValue", ErrorValue{map[string]Value{"message": StringValue{"error"}}}, map[string]any{"message": "error"}},
		{"EventValue", EventValue{map[string]Value{"name": StringValue{"event"}}}, map[string]any{"name": "event"}},
		{"Nested MapValue", nestedMapVal, nestedMapRaw}, // Test nested structure
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unwrapped := Unwrap(tc.input)
			if !reflect.DeepEqual(unwrapped, tc.expected) {
				t.Errorf("Expected unwrapped value %#v (%T), but got %#v (%T)", tc.expected, tc.expected, unwrapped, unwrapped)
			}
		})
	}
}

func TestUnwrapSlice(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		unwrapped, err := UnwrapSlice(nil)
		if err != nil {
			t.Errorf("Did not expect an error, but got: %v", err)
		}
		if unwrapped != nil {
			t.Errorf("Expected nil, but got: %v", unwrapped)
		}
	})

	t.Run("slice with values", func(t *testing.T) {
		input := []Value{StringValue{"a"}, NumberValue{1}}
		expected := []any{"a", float64(1)}
		unwrapped, err := UnwrapSlice(input)
		if err != nil {
			t.Errorf("Did not expect an error, but got: %V", err)
		}
		if !reflect.DeepEqual(unwrapped, expected) {
			t.Errorf("Expected unwrapped slice %#v, but got %#v", expected, unwrapped)
		}
	})

	t.Run("slice with nil value", func(t *testing.T) {
		input := []Value{StringValue{"a"}, NilValue{}, NumberValue{1}}
		expected := []any{"a", nil, float64(1)}
		unwrapped, err := UnwrapSlice(input)
		if err != nil {
			t.Errorf("Did not expect an error, but got: %v", err)
		}
		if !reflect.DeepEqual(unwrapped, expected) {
			t.Errorf("Expected unwrapped slice %#v, but got %#v", expected, unwrapped)
		}
	})

	t.Run("slice with pointer nil value", func(t *testing.T) {
		input := []Value{StringValue{"a"}, &NilValue{}, NumberValue{1}}
		expected := []any{"a", nil, float64(1)}
		unwrapped, err := UnwrapSlice(input)
		if err != nil {
			t.Errorf("Did not expect an error, but got: %v", err)
		}
		if !reflect.DeepEqual(unwrapped, expected) {
			t.Errorf("Expected unwrapped slice %#v, but got %#v", expected, unwrapped)
		}
	})
}
