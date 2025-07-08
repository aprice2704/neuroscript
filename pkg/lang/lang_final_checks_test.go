// filename: pkg/lang/lang_final_checks_test.go
package lang

import (
	"errors"
	"testing"
	"time"
)

// TestConstructorEdgeCases verifies the nil-safety of various value constructors.
func TestConstructorEdgeCases(t *testing.T) {
	t.Run("NewListValue with nil", func(t *testing.T) {
		lv := NewListValue(nil)
		if lv.Value == nil {
			t.Error("NewListValue(nil) should produce a non-nil, empty slice, but got nil")
		}
		if len(lv.Value) != 0 {
			t.Errorf("NewListValue(nil) should produce an empty slice, but len is %d", len(lv.Value))
		}
	})

	t.Run("NewMapValue with nil", func(t *testing.T) {
		mv := NewMapValue(nil)
		if mv.Value == nil {
			t.Error("NewMapValue(nil) should produce a non-nil, empty map, but got nil")
		}
	})

	t.Run("NewErrorValue with nil details", func(t *testing.T) {
		ev := NewErrorValue("code", "msg", nil)
		details, ok := ev.Value[ErrorKeyDetails]
		if !ok {
			t.Fatal("NewErrorValue should have a details key")
		}
		if _, isNil := details.(NilValue); !isNil {
			t.Errorf("Expected details to be NilValue, but got %T", details)
		}
	})

	t.Run("NewErrorValueFromRuntimeError with nil", func(t *testing.T) {
		// This test just ensures it doesn't panic and returns a valid error value.
		ev := NewErrorValueFromRuntimeError(nil)
		if ev.Type() != TypeError {
			t.Errorf("Expected an ErrorValue, but got %s", ev.Type())
		}
		msg, _ := ev.Value[ErrorKeyMessage].(StringValue)
		if msg.Value != "nil runtime error provided" {
			t.Errorf("Unexpected message for nil runtime error: %s", msg.Value)
		}
	})
}

// TestIsZeroValueForValueTypes explicitly tests IsZeroValue against wrapped Value types.
func TestIsZeroValueForValueTypes(t *testing.T) {
	testCases := []struct {
		name     string
		input    Value
		expected bool
	}{
		{"EventValue", EventValue{}, false}, // Events are always truthy, thus not zero.
		{"ErrorValue", ErrorValue{}, true},  // Errors are always falsy, thus zero.
		{"Empty ListValue", ListValue{[]Value{}}, true},
		{"Non-empty ListValue", ListValue{[]Value{NumberValue{1}}}, false},
		{"Zero TimedateValue", TimedateValue{time.Time{}}, true},
		{"Non-zero TimedateValue", TimedateValue{time.Now()}, false},
		{"Empty StringValue", StringValue{""}, true},
		{"Non-empty StringValue", StringValue{"a"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsZeroValue(tc.input); got != tc.expected {
				t.Errorf("Expected IsZeroValue for %s to be %v, but got %v", tc.name, tc.expected, got)
			}
		})
	}
}

// TestAdvancedWrappingAndTyping checks more subtle behaviors of the type system.
func TestAdvancedWrappingAndTyping(t *testing.T) {
	t.Run("Wrap should be idempotent", func(t *testing.T) {
		// Wrapping an already-wrapped value should return it unchanged.
		original := StringValue{"already wrapped"}
		wrapped, err := Wrap(original)
		if err != nil {
			t.Fatalf("Wrap failed: %v", err)
		}
		if wrapped != original {
			t.Error("Wrap was not idempotent; it changed an already-wrapped value")
		}
	})

	t.Run("TypeOf with pointer to Value", func(t *testing.T) {
		// TypeOf should be able to look through a pointer to a value.
		val := &StringValue{"a pointer to me"}
		if got := TypeOf(val); got != TypeString {
			t.Errorf("Expected TypeOf(&StringValue{}) to be %s, but got %s", TypeString, got)
		}
	})

	t.Run("Unwrap ErrorValue", func(t *testing.T) {
		// Ensure unwrapping a complex value like ErrorValue works as expected.
		ev := NewErrorValue("E_TEST", "a test", NumberValue{123})
		unwrapped := Unwrap(ev)
		unwrappedMap, ok := unwrapped.(map[string]any)
		if !ok {
			t.Fatalf("Expected Unwrap(ErrorValue) to return map[string]any, but got %T", unwrapped)
		}
		if code := unwrappedMap[ErrorKeyCode]; code != "E_TEST" {
			t.Errorf("Expected unwrapped error code to be 'E_TEST', got %v", code)
		}
	})
}

// TestErrorWrappingLogic verifies the helper function WrapErrorWithPosition.
func TestErrorWrappingLogic(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		if err := WrapErrorWithPosition(nil, &Position{Line: 1}, "ctx"); err != nil {
			t.Errorf("Expected nil for a nil error input, but got %v", err)
		}
	})

	t.Run("preserves existing position", func(t *testing.T) {
		originalPos := &Position{Line: 10, Column: 5}
		newPos := &Position{Line: 20, Column: 15}
		err := NewRuntimeError(ErrorCodeGeneric, "test", nil).WithPosition(originalPos)

		wrappedErr := WrapErrorWithPosition(err, newPos, "new context")
		re, ok := wrappedErr.(*RuntimeError)
		if !ok {
			t.Fatal("Expected a *RuntimeError")
		}

		if re.Position != originalPos {
			t.Errorf("Wrapper wrongfully overwrote an existing position. Expected %v, got %v", originalPos, re.Position)
		}
	})

	t.Run("wraps standard error", func(t *testing.T) {
		pos := &Position{Line: 1, Column: 1}
		stdErr := errors.New("standard error")
		wrappedErr := WrapErrorWithPosition(stdErr, pos, "context")

		re, ok := wrappedErr.(*RuntimeError)
		if !ok {
			t.Fatalf("Expected standard error to be wrapped in a *RuntimeError, got %T", wrappedErr)
		}
		if re.Position != pos {
			t.Errorf("Position was not attached correctly. Expected %v, got %v", pos, re.Position)
		}
		// FIX: Use errors.Is() for idiomatic error checking instead of reflect.DeepEqual.
		if !errors.Is(re, stdErr) {
			t.Errorf("Underlying error was not wrapped correctly")
		}
	})
}
