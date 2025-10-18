// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Adds a tiger test to ensure Wrap() fails on structs.
// filename: pkg/lang/tiger_test.go
// nlines: 180
// risk_rating: HIGH

package lang

import (
	"errors"
	"math"
	"testing"
)

// TestConversionTiger_Whitespace attacks ToFloat64 with various whitespace.
func TestConversionTiger_Whitespace(t *testing.T) {
	if _, ok := ToFloat64(" 5.5 "); !ok {
		t.Error(`ToFloat64(" 5.5 ") failed, expected ok=true`)
	}
	if _, ok := ToFloat64("\t5\n"); !ok {
		t.Error(`ToFloat64("\t5\n") failed, expected ok=true`)
	}
}

// TestConversionTiger_InvalidStrings attacks ToFloat64 with non-numeric strings.
func TestConversionTiger_InvalidStrings(t *testing.T) {
	if _, ok := ToFloat64("five"); ok {
		t.Error(`ToFloat64("five") passed, expected ok=false`)
	}
	if _, ok := ToFloat64(""); ok {
		t.Error(`ToFloat64("") passed, expected ok=false`)
	}
}

// TestConversionTiger_LossyInts attacks ToInt64 with lossy float values.
func TestConversionTiger_LossyInts(t *testing.T) {
	if _, ok := ToInt64("5.1"); ok {
		t.Error(`ToInt64("5.1") passed, expected ok=false`)
	}
	if _, ok := ToInt64(NumberValue{Value: 5.1}); ok {
		t.Error(`ToInt64(NumberValue{5.1}) passed, expected ok=false`)
	}
	if val, ok := ToInt64("5.0"); !ok || val != 5 {
		t.Error(`ToInt64("5.0") failed or returned wrong value, expected 5, ok=true`)
	}
}

// TestConversionTiger_NilToString attacks ToString with nil inputs.
func TestConversionTiger_NilToString(t *testing.T) {
	if str, _ := ToString(nil); str != "" {
		t.Errorf(`ToString(nil) expected "", got %q`, str)
	}
	if str, _ := ToString(NilValue{}); str != "" {
		t.Errorf(`ToString(NilValue{}) expected "", got %q`, str)
	}
	if str, _ := ToString(&NilValue{}); str != "" {
		t.Errorf(`ToString(&NilValue{}) expected "", got %q`, str)
	}
}

// TestOperatorTiger_NaNAndInf attacks operations with NaN/Inf.
func TestOperatorTiger_NaNAndInf(t *testing.T) {
	nan := NumberValue{Value: math.NaN()}
	inf := NumberValue{Value: math.Inf(1)}
	negInf := NumberValue{Value: math.Inf(-1)}

	if res, _ := PerformBinaryOperation("==", nan, nan); IsTruthy(res) {
		t.Error(`NaN == NaN should be false`)
	}
	res, err := PerformBinaryOperation("+", inf, negInf)
	if err != nil {
		t.Fatalf("Inf + -Inf failed: %v", err)
	}
	if f, ok := ToFloat64(res); !ok || !math.IsNaN(f) {
		t.Error(`Inf + -Inf should be NaN`)
	}
}

// TestOperatorTiger_InvalidCollections attacks operations on collections.
func TestOperatorTiger_InvalidCollections(t *testing.T) {
	listA := ListValue{Value: []Value{}}
	listB := ListValue{Value: []Value{}}
	if _, err := PerformBinaryOperation("+", listA, listB); err == nil {
		t.Error(`ListValue + ListValue should produce an error`)
	}

	mapA := MapValue{Value: map[string]Value{}}
	mapB := MapValue{Value: map[string]Value{}}
	if _, err := PerformBinaryOperation("+", mapA, mapB); err == nil {
		t.Error(`MapValue + MapValue should produce an error`)
	}
}

// TestOperatorTiger_IncompatibleTypes attacks ops on completely wrong types.
func TestOperatorTiger_IncompatibleTypes(t *testing.T) {
	if _, err := PerformBinaryOperation(">", StringValue{"a"}, BoolValue{true}); err == nil {
		t.Error(`string > bool should produce an error`)
	}
	if _, err := PerformBinaryOperation("<", ErrorValue{}, NumberValue{1}); err == nil {
		t.Error(`error < number should produce an error`)
	}
	if _, err := PerformBinaryOperation("*", StringValue{"a"}, StringValue{"b"}); err == nil {
		t.Error(`string * string should produce an error`)
	}
}

// TestOperatorTiger_EqualityOnComplexTypes attacks equality checks on collections.
func TestOperatorTiger_EqualityOnComplexTypes(t *testing.T) {
	// FIX: The original test was wrong. In a value-based language,
	// [1] == [1] should be TRUE. We assert this is the case.
	listA := ListValue{Value: []Value{StringValue{"a"}}}
	listB := ListValue{Value: []Value{StringValue{"a"}}}
	res, err := PerformBinaryOperation("==", listA, listB)
	if err != nil {
		t.Fatalf("ListA == ListB failed: %v", err)
	}
	if !IsTruthy(res) {
		t.Error(`ListA == ListB should be true (DeepEqual check)`)
	}

	// FIX: Same for maps.
	mapA := MapValue{Value: map[string]Value{"a": NumberValue{1}}}
	mapB := MapValue{Value: map[string]Value{"a": NumberValue{1}}}
	res, err = PerformBinaryOperation("==", mapA, mapB)
	if err != nil {
		t.Fatalf("MapA == MapB failed: %v", err)
	}
	if !IsTruthy(res) {
		t.Error(`MapA == MapB should be true (DeepEqual check)`)
	}
}

// TestWrapTiger_UnsupportedTypes attacks Wrap with invalid Go types.
func TestWrapTiger_UnsupportedTypes(t *testing.T) {
	if _, err := Wrap(map[int]string{1: "a"}); err == nil {
		t.Error(`Wrap(map[int]string) should have failed (invalid key type)`)
	}
	if _, err := Wrap(make(chan int)); err == nil {
		t.Error(`Wrap(chan int) should have failed (unsupported type)`)
	}
	if _, err := Wrap(errors.New("raw error")); err == nil {
		t.Error(`Wrap(errors.New(...)) should have failed (use NewErrorValue)`)
	}
}

// TestWrapTiger_Structs asserts that Wrap() fails on structs.
func TestWrapTiger_Structs(t *testing.T) {
	type MyTestStruct struct {
		Name string
	}
	myStruct := MyTestStruct{Name: "test"}

	if _, err := Wrap(myStruct); err == nil {
		t.Error("Wrap(myStruct) should have failed, but it succeeded")
	}

	myStructPtr := &MyTestStruct{Name: "test"}
	if _, err := Wrap(myStructPtr); err == nil {
		t.Error("Wrap(&myStruct) should have failed, but it succeeded")
	}
}

// TestWrapTiger_NestedUnsupported attacks Wrap with nested invalid Go types.
func TestWrapTiger_NestedUnsupported(t *testing.T) {
	m := map[string]any{
		"a": "good",
		"b": make(chan int),
	}
	if _, err := Wrap(m); err == nil {
		t.Error(`Wrap() should have failed due to nested channel`)
	}

	l := []any{
		"good",
		make(chan int),
	}
	if _, err := Wrap(l); err == nil {
		t.Error(`Wrap() should have failed due to nested channel in list`)
	}
}

// TestWrapTiger_PointerDeref attacks the Wrap function's idempotency logic.
func TestWrapTiger_PointerDeref(t *testing.T) {
	// This tests the explicit fix in value_helpers.go to dereference *MapValue
	// when Wrap() is called on an already-wrapped value.
	mvPtr := NewMapValue(map[string]Value{"a": NumberValue{1}})

	wrapped, err := Wrap(mvPtr)
	if err != nil {
		t.Fatalf("Wrap(mvPtr) failed: %v", err)
	}

	if _, ok := wrapped.(MapValue); !ok {
		t.Errorf("Wrap(mvPtr) should have returned MapValue (by value), but got %T", wrapped)
	}

	if _, ok := wrapped.(*MapValue); ok {
		t.Error("Wrap(mvPtr) returned *MapValue (by pointer), which is incorrect")
	}
}

// TestLangTiger_NilDeref attacks operations that might panic on nil.
func TestLangTiger_NilDeref(t *testing.T) {
	var nilMapPtr *MapValue

	// FIX: This now tests the isNilLike logic in areValuesEqual
	res, err := PerformBinaryOperation("==", nilMapPtr, NilValue{})
	if err != nil {
		t.Fatalf("nilMapPtr == NilValue failed: %v", err)
	}
	if !IsTruthy(res) {
		t.Error("nilMapPtr == NilValue should be true, but was false")
	}

	// FIX: This now tests the nil-pointer check in TypeOf
	if TypeOf(nilMapPtr) != TypeNil {
		t.Errorf("TypeOf(nilMapPtr) should be TypeNil, got %s", TypeOf(nilMapPtr))
	}

	// Unwrap should handle nil pointers
	if unwrapRes := Unwrap(nilMapPtr); unwrapRes != nil {
		t.Errorf("Unwrap(nilMapPtr) should be nil, got %v", unwrapRes)
	}
}
