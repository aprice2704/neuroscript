// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 2
// :: description: Extra tests for type_utils, including recursive truthiness depth limits.
// :: latestChange: Added TestIsTruthyDepthLimit to verify pathological cycle protection.
// :: filename: pkg/lang/type_utils_extra_test.go
// :: serialization: go

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

func TestIsTruthyDepthLimit(t *testing.T) {
	t.Run("Control: unwraps within limits", func(t *testing.T) {
		// Create an onion of ToolValues wrapping a NilValue
		// maxTruthinessDepth is 20, we use 5 here.
		var val any = NilValue{}
		for i := 0; i < 5; i++ {
			val = ToolValue{Value: val}
		}

		// Since it's < 20 layers deep and ends in NilValue, it should be falsey.
		if IsTruthy(val.(Value)) {
			t.Errorf("Expected 5-deep wrapped NilValue to be falsey, but got truthy")
		}
		// Consequently, IsZeroValue should be true.
		if !IsZeroValue(val) {
			t.Errorf("Expected 5-deep wrapped NilValue to be considered zero, but got non-zero")
		}
	})

	t.Run("Fail-Safe: aborts infinite recursion", func(t *testing.T) {
		// Create an onion of ToolValues wrapping a NilValue
		// maxTruthinessDepth is 20, we use 25 here to trip the breaker.
		var val any = NilValue{}
		for i := 0; i < 25; i++ {
			val = ToolValue{Value: val}
		}

		// Because it hits the depth limit (20), it fails safe to truthy.
		if !IsTruthy(val.(Value)) {
			t.Errorf("Expected 25-deep wrapped NilValue to trip fail-safe and return truthy, but got falsey")
		}
		// Consequently, IsZeroValue should fail safe to false.
		if IsZeroValue(val) {
			t.Errorf("Expected 25-deep wrapped NilValue to trip fail-safe and be considered non-zero, but got zero")
		}
	})
}
