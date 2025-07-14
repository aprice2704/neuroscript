// filename: pkg/lang/lang_additional_test.go
package lang

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/types"
)

// mockCallable is a simple mock to satisfy the Callable interface for testing.
type mockCallable struct {
	name string
}

func (m *mockCallable) Name() string { return m.name }
func (m *mockCallable) Arity() int   { return 0 }

// The first parameter is changed from *Interpreter to `any` to remove the dependency
// on the interpreter package, which lang should not have.
func (m *mockCallable) Call(i any, args []Value) (Value, error) {
	return NilValue{}, nil
}

// IsCallable is added to fully implement the Callable interface.
func (m *mockCallable) IsCallable() {}

// TestStringRepresentations verifies the exact output of the String() method for various Value types.
func TestStringRepresentations(t *testing.T) {
	testCases := []struct {
		name     string
		input    Value
		expected string
	}{
		{
			"ListValue",
			ListValue{Value: []Value{NumberValue{1}, StringValue{"a"}}},
			`[1, "a"]`,
		},
		{
			"MapValue",
			MapValue{Value: map[string]Value{"k1": NumberValue{1}, "k2": StringValue{"v2"}}},
			`{"k1": 1, "k2": "v2"}`, // Note: map order is not guaranteed, so we check for parts.
		},
		{
			"ErrorValue with message",
			ErrorValue{Value: map[string]Value{ErrorKeyMessage: StringValue{"test error"}}},
			"error: test error",
		},
		{
			"ErrorValue without message",
			ErrorValue{Value: map[string]Value{"code": NumberValue{1}}},
			"error: (unspecified)",
		},
		{
			"FunctionValue non-nil",
			FunctionValue{Value: &mockCallable{name: "my_func"}},
			"function<my_func>",
		},
		{
			"FunctionValue nil",
			FunctionValue{Value: nil},
			"<nil function>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.String()
			// Special handling for maps where order is not guaranteed
			if _, ok := tc.input.(MapValue); ok {
				// With the fix to MapValue.String(), we can now check for a more exact substring.
				if !(strings.HasPrefix(got, "{") && strings.HasSuffix(got, "}") &&
					strings.Contains(got, `"k1": 1`) && strings.Contains(got, `"k2": "v2"`)) {
					t.Errorf("Expected map string like '%s', but got '%s'", tc.expected, got)
				}
			} else if got != tc.expected {
				t.Errorf("Expected String() to be '%s', but got '%s'", tc.expected, got)
			}
		})
	}
}

// TestConversionLogicEdgeCases checks for specific behaviors in conversion utilities.
func TestConversionLogicEdgeCases(t *testing.T) {
	t.Run("ToInt64 with lossy float", func(t *testing.T) {
		// A float with a fractional part should not be convertible to int64.
		if _, ok := ToInt64(3.14); ok {
			t.Error("ToInt64(3.14) wrongfully returned ok=true")
		}
		if _, ok := ToInt64(NumberValue{Value: 3.14}); ok {
			t.Error("ToInt64(NumberValue{3.14}) wrongfully returned ok=true")
		}
	})

	t.Run("ToString for nil and empty values", func(t *testing.T) {
		// The fix in ToString handles nil explicitly to return an empty string.
		if str, _ := ToString(nil); str != "" {
			t.Errorf(`Expected ToString(nil) to be "", got %q`, str)
		}
		if str, _ := ToString(NilValue{}); str != "" {
			t.Errorf(`Expected ToString(NilValue{}) to be "", got %q`, str)
		}
	})
}

// TestWrapUnwrapNested verifies recursive wrapping and unwrapping of a complex structure.
func TestWrapUnwrapNested(t *testing.T) {
	original := map[string]any{
		"a": float64(1),
		"b": "hello",
		"c": []any{
			"nested_string",
			float64(100),
			map[string]any{"deep": true},
		},
	}

	wrapped, err := Wrap(original)
	if err != nil {
		t.Fatalf("Wrap failed: %v", err)
	}

	// Ensure it's a MapValue
	if _, ok := wrapped.(MapValue); !ok {
		t.Fatalf("Expected wrapped type to be MapValue, got %T", wrapped)
	}

	unwrapped := Unwrap(wrapped)

	if !reflect.DeepEqual(original, unwrapped) {
		t.Errorf("Unwrapped value does not match original.\nOriginal: %#v\nUnwrapped:%#v", original, unwrapped)
	}
}

// TestRuntimeErrorFormatting checks the Error() method of RuntimeError.
func TestRuntimeErrorFormatting(t *testing.T) {
	baseErr := NewRuntimeError(ErrorCodeToolExecutionFailed, "tool failed", nil)
	if !strings.Contains(baseErr.Error(), "Error 21: tool failed") {
		t.Errorf("Base error format is incorrect: %s", baseErr.Error())
	}

	withPos := NewRuntimeError(ErrorCodeToolExecutionFailed, "tool failed", nil).WithPosition(&types.Position{Line: 10, Column: 5})
	// FIX: Check for "col" instead of "column" to match the actual output.
	if !strings.Contains(withPos.Error(), "at 10:5") {
		t.Errorf("Error with position format is incorrect: %s", withPos.Error())
	}

	wrappedNative := errors.New("native cause")
	withWrapped := NewRuntimeError(ErrorCodeToolExecutionFailed, "tool failed", wrappedNative)
	if !strings.Contains(withWrapped.Error(), "(wrapped: native cause)") {
		t.Errorf("Error with wrapped error format is incorrect: %s", withWrapped.Error())
	}

	if unwrapped := errors.Unwrap(withWrapped); unwrapped != wrappedNative {
		t.Errorf("Unwrap did not return the correct wrapped error")
	}
}
