// filename: pkg/lang/type_utils_extra_test.go
package lang

import (
	"testing"
)

func TestNumericConversionsWithNonNumerics(t *testing.T) {
	t.Run("ToFloat64 with booleans", func(t *testing.T) {
		// Booleans should not be convertible to float.
		if val, ok := ToFloat64(true); ok {
			t.Errorf("ToFloat64(true) wrongfully returned ok=true with value %f", val)
		}

		if val, ok := ToFloat64(false); ok {
			t.Errorf("ToFloat64(false) wrongfully returned ok=true with value %f", val)
		}

		// Test with wrapped boolean value
		if val, ok := ToFloat64(BoolValue{Value: true}); ok {
			t.Errorf("ToFloat64(BoolValue{true}) wrongfully returned ok=true with value %f", val)
		}
	})

	t.Run("ToInt64 with booleans", func(t *testing.T) {
		// This test depends on ToFloat64 behaving correctly.
		if _, ok := ToInt64(true); ok {
			t.Errorf("ToInt64(true) wrongfully returned ok=true")
		}

		if _, ok := ToInt64(BoolValue{Value: false}); ok {
			t.Errorf("ToInt64(BoolValue{false}) wrongfully returned ok=true")
		}
	})

	t.Run("ToNumeric with booleans", func(t *testing.T) {
		// This test also depends on ToFloat64.
		if _, ok := ToNumeric(true); ok {
			t.Errorf("ToNumeric(true) wrongfully returned ok=true")
		}
		if _, ok := ToNumeric(BoolValue{Value: false}); ok {
			t.Errorf("ToNumeric(BoolValue{false}) wrongfully returned ok=true")
		}
	})

	t.Run("ToFloat64 with nil", func(t *testing.T) {
		// Nil values should not be convertible to float.
		if val, ok := ToFloat64(nil); ok {
			t.Errorf("ToFloat64(nil) wrongfully returned ok=true with value %f", val)
		}

		if val, ok := ToFloat64(NilValue{}); ok {
			t.Errorf("ToFloat64(NilValue{}) wrongfully returned ok=true with value %f", val)
		}
	})
}
