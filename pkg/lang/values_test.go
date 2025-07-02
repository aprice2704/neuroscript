// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Unit tests for the new core Value types.
// filename: pkg/lang/values_test.go
// nlines: 100
// risk_rating: LOW

package lang

import (
	"testing"
	"time"
)

func TestValueTypes(t *testing.T) {
	// Helper for running tests
	expect := func(t *testing.T, name string, val Value, expectedType NeuroScriptType, expectedTruthy bool) {
		t.Helper()
		if val.Type() != expectedType {
			t.Errorf("%s: Type() got %s, want %s", name, val.Type(), expectedType)
		}
		if val.IsTruthy() != expectedTruthy {
			t.Errorf("%s: IsTruthy() got %v, want %v", name, val.IsTruthy(), expectedTruthy)
		}
		// A simple check that String() doesn't panic
		_ = val.String()
	}

	t.Run("StringValue", func(t *testing.T) {
		expect(t, "non-empty", StringValue{Value: "hello"}, TypeString, true)
		expect(t, "empty", StringValue{Value: ""}, TypeString, false)
	})

	t.Run("NumberValue", func(t *testing.T) {
		expect(t, "positive", NumberValue{Value: 123}, TypeNumber, true)
		expect(t, "zero", NumberValue{Value: 0}, TypeNumber, false)
		expect(t, "negative", NumberValue{Value: -1}, TypeNumber, true)
	})

	t.Run("BoolValue", func(t *testing.T) {
		expect(t, "true", BoolValue{Value: true}, TypeBoolean, true)
		expect(t, "false", BoolValue{Value: false}, TypeBoolean, false)
	})

	t.Run("NilValue", func(t *testing.T) {
		expect(t, "nil", NilValue{}, TypeNil, false)
	})

	t.Run("ErrorValue", func(t *testing.T) {
		errVal := ErrorValue{Value: map[string]Value{
			"code":		NumberValue{Value: 1},
			"message":	StringValue{Value: "test error"},
		}}
		expect(t, "error", errVal, TypeError, false)
	})

	t.Run("EventValue", func(t *testing.T) {
		eventVal := EventValue{Value: map[string]Value{
			"name": StringValue{Value: "test_event"},
		}}
		expect(t, "event", eventVal, TypeEvent, true)
	})

	t.Run("TimedateValue", func(t *testing.T) {
		expect(t, "non-zero", TimedateValue{Value: time.Now()}, TypeTimedate, true)
		expect(t, "zero", TimedateValue{Value: time.Time{}}, TypeTimedate, false)
	})

	t.Run("FuzzyValue", func(t *testing.T) {
		// Test IsTruthy boundaries
		expect(t, "clearly true", NewFuzzyValue(0.8), TypeFuzzy, true)
		expect(t, "clearly false", NewFuzzyValue(0.2), TypeFuzzy, false)
		expect(t, "boundary false", NewFuzzyValue(0.5), TypeFuzzy, false)
		expect(t, "boundary true", NewFuzzyValue(0.50001), TypeFuzzy, true)

		// Test clamping
		if f := NewFuzzyValue(1.5); f.μ != 1.0 {
			t.Errorf("FuzzyValue clamping failed for high value: got %f, want 1.0", f.μ)
		}
		if f := NewFuzzyValue(-0.5); f.μ != 0.0 {
			t.Errorf("FuzzyValue clamping failed for low value: got %f, want 0.0", f.μ)
		}
	})
}