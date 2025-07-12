// filename: pkg/lang/type_utils_test.go
package lang

import (
	"fmt"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestTypeOf(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected NeuroScriptType
	}{
		{"nil", nil, TypeNil},
		{"*string", new(string), TypeString},
		{"string", "hello", TypeString},
		{"int", 123, TypeNumber},
		{"float64", 3.14, TypeNumber},
		{"bool", true, TypeBoolean},
		{"[]byte", []byte("bytes"), TypeBytes},
		{"time.Time", time.Time{}, TypeTimedate},
		{"[]any", []any{}, TypeList},
		{"map[string]any", map[string]any{}, TypeMap},
		{"Callable", new(Callable), TypeFunction},
		{"Tool", new(interfaces.Tool), TypeTool},
		{"error", fmt.Errorf("error"), TypeError},
		{"StringValue", StringValue{}, TypeString},
		{"NumberValue", NumberValue{}, TypeNumber},
		{"BoolValue", BoolValue{}, TypeBoolean},
		{"BytesValue", BytesValue{}, TypeBytes},
		{"ListValue", ListValue{}, TypeList},
		{"MapValue", MapValue{}, TypeMap},
		{"NilValue", NilValue{}, TypeNil},
		{"FunctionValue", FunctionValue{}, TypeFunction},
		{"ToolValue", ToolValue{}, TypeTool},
		{"ErrorValue", ErrorValue{}, TypeError},
		{"EventValue", EventValue{}, TypeEvent},
		{"TimedateValue", TimedateValue{}, TypeTimedate},
		{"FuzzyValue", FuzzyValue{}, TypeFuzzy},
		{"unknown", struct{}{}, TypeUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := TypeOf(tc.input); got != tc.expected {
				t.Errorf("Expected TypeOf(%T) to be %s, but got %s", tc.input, tc.expected, got)
			}
		})
	}
}

func TestIsTruthy(t *testing.T) {
	testCases := []struct {
		name     string
		input    Value
		expected bool
	}{
		{"true", BoolValue{true}, true},
		{"false", BoolValue{false}, false},
		{"non-empty string", StringValue{"hello"}, true},
		{"empty string", StringValue{""}, false},
		{"zero", NumberValue{0}, false},
		{"non-zero", NumberValue{1}, true},
		{"empty list", ListValue{[]Value{}}, false},
		{"non-empty list", ListValue{[]Value{NumberValue{1}}}, true},
		{"empty map", MapValue{map[string]Value{}}, false},
		{"non-empty map", MapValue{map[string]Value{"a": NumberValue{1}}}, true},
		{"nil", NilValue{}, false},
		{"error", ErrorValue{}, false},
		{"event", EventValue{}, true},
		{"timedate", TimedateValue{time.Now()}, true},
		{"fuzzy", FuzzyValue{0.5}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsTruthy(tc.input); got != tc.expected {
				t.Errorf("Expected IsTruthy for %s to be %v, but got %v", tc.name, tc.expected, got)
			}
		})
	}
}

func TestIsZeroValue(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"non-empty string", "a", false},
		{"zero int", 0, true},
		{"non-zero int", 1, false},
		{"zero float", 0.0, true},
		{"non-zero float", 1.0, false},
		{"false", false, true},
		{"true", true, false},
		{"empty slice", []int{}, true},
		{"non-empty slice", []int{1}, false},
		{"empty map", map[string]int{}, true},
		{"non-empty map", map[string]int{"a": 1}, false},
		{"nil pointer", (*int)(nil), true},
		{"non-nil pointer", new(int), false},
		{"zero struct", struct{}{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsZeroValue(tc.input); got != tc.expected {
				t.Errorf("Expected isZeroValue for %s to be %v, but got %v", tc.name, tc.expected, got)
			}
		})
	}
}

func TestPositionString(t *testing.T) {
	t.Run("nil position", func(t *testing.T) {
		var pos *types.Position
		expected := "<nil position>"
		if pos.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, pos.String())
		}
	})
}
