package lang

import (
	"reflect"
	"testing"
	"time"
)

func TestWrap(t *testing.T) {
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
			expected: MapValue{map[string]Value{
				"a": NumberValue{1},
				"b": StringValue{"two"},
			}},
			hasError: false,
		},
		{"unsupported type", struct{}{}, nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrapped, err := Wrap(tc.input)

			if tc.hasError {
				if err == nil {
					t.Error("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				if !reflect.DeepEqual(wrapped, tc.expected) {
					t.Errorf("Expected wrapped value %+v, but got %+v", tc.expected, wrapped)
				}
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	testCases := []struct {
		name     string
		input    Value
		expected any
	}{
		{"NilValue", NilValue{}, nil},
		{"StringValue", StringValue{"hello"}, "hello"},
		{"BytesValue", BytesValue{[]byte("bytes")}, []byte("bytes")},
		{"BoolValue", BoolValue{true}, true},
		{"NumberValue", NumberValue{123}, float64(123)},
		{"TimedateValue", TimedateValue{time.Unix(0, 0)}, time.Unix(0, 0)},
		{"FuzzyValue", FuzzyValue{0.5}, 0.5},
		{"ListValue", ListValue{[]Value{StringValue{"a"}, NumberValue{1}}}, []any{"a", float64(1)}},
		{"MapValue", MapValue{map[string]Value{"a": NumberValue{1}}}, map[string]any{"a": float64(1)}},
		{"ErrorValue", ErrorValue{map[string]Value{"message": StringValue{"error"}}}, map[string]any{"message": "error"}},
		{"EventValue", EventValue{map[string]Value{"name": StringValue{"event"}}}, map[string]any{"name": "event"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unwrapped := Unwrap(tc.input)
			if !reflect.DeepEqual(unwrapped, tc.expected) {
				t.Errorf("Expected unwrapped value %#v, but got %#v", tc.expected, unwrapped)
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
			t.Errorf("Did not expect an error, but got: %v", err)
		}
		if !reflect.DeepEqual(unwrapped, expected) {
			t.Errorf("Expected unwrapped slice %#v, but got %#v", expected, unwrapped)
		}
	})
}
