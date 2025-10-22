// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Corrects the Mixed_to_Int64 test to expect an error.
// filename: pkg/utils/utils_test.go
// nlines: 501
// risk_rating: LOW

package utils

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestConvertToBool(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected bool
		ok       bool
	}{
		{"nil", nil, false, true},
		{"bool true", true, true, true},
		{"bool false", false, false, true},
		{"int64 zero", int64(0), false, true},
		{"int64 non-zero", int64(1), true, true},
		{"float64 zero", 0.0, false, true},
		{"float64 non-zero", 1.5, true, true},
		{"string true lower", "true", true, true},
		{"string true upper", "TRUE", true, true},
		{"string false lower", "false", false, true},
		{"string false upper", "FALSE", false, true},
		{"string 1", "1", true, true},
		{"string 0", "0", false, true},
		{"string empty", "", false, false},
		{"string invalid", "hello", false, false},
		{"int zero", 0, false, true},
		{"int non-zero", -1, true, true},
		{"float32 zero", float32(0.0), false, true},
		{"float32 non-zero", float32(1.1), true, true},
		{"unsupported type", struct{}{}, false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ConvertToBool(tc.input)
			if ok != tc.ok {
				t.Errorf("ConvertToBool(%v): ok mismatch, got %v, want %v", tc.input, ok, tc.ok)
			}
			if got != tc.expected {
				t.Errorf("ConvertToBool(%v): value mismatch, got %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestConvertToSliceOfAny(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected []interface{}
		ok       bool
		wantErr  bool
	}{
		{"nil", nil, nil, false, true},
		{"[]interface{}", []interface{}{"a", 1}, []interface{}{"a", 1}, true, false},
		{"[]string", []string{"a", "b"}, []interface{}{"a", "b"}, true, false},
		{"[]int", []int{1, 2}, []interface{}{1, 2}, true, false},
		{"[]float64", []float64{1.1, 2.2}, []interface{}{1.1, 2.2}, true, false},
		{"empty []interface{}", []interface{}{}, []interface{}{}, true, false},
		{"empty []string", []string{}, []interface{}{}, true, false},
		{"non-slice string", "hello", nil, false, true},
		{"non-slice int", 123, nil, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok, err := ConvertToSliceOfAny(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ConvertToSliceOfAny(%v): error mismatch, got err %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if ok != tc.ok {
				t.Errorf("ConvertToSliceOfAny(%v): ok mismatch, got %v, want %v", tc.input, ok, tc.ok)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("ConvertToSliceOfAny(%v): value mismatch, got %#v, want %#v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestConvertToFloat64(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected float64
		ok       bool
	}{
		{"float64", 123.45, 123.45, true},
		{"int64", int64(123), 123.0, true},
		{"int", 456, 456.0, true},
		{"string valid float", "789.01", 789.01, true},
		{"string valid int", "22", 22.0, true},
		{"string invalid", "hello", 0, false},
		{"nil", nil, 0, false},
		{"bool", true, 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ConvertToFloat64(tc.input)
			if ok != tc.ok {
				t.Errorf("ConvertToFloat64(%v): ok mismatch, got %v, want %v", tc.input, ok, tc.ok)
			}
			if got != tc.expected {
				t.Errorf("ConvertToFloat64(%v): value mismatch, got %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestConvertToInt64E(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected int64
		wantErr  bool
	}{
		{"nil", nil, 0, true},
		{"int", 123, 123, false},
		{"int64", int64(456), 456, false},
		{"float64 whole", 789.0, 789, false},
		{"float64 fractional", 789.5, 789, false},
		{"string int", "101", 101, false},
		{"string float whole", "202.0", 202, false},
		{"string float fractional", "202.5", 202, false},
		{"string invalid", "hello", 0, true},
		{"bool true", true, 1, false},
		{"bool false", false, 0, false},
		{"other numeric int8", int8(10), 10, false},
		{"other numeric uint16", uint16(20), 20, false},
		{"other numeric float32", float32(30.5), 30, false},
		{"unsupported type", struct{}{}, 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ConvertToInt64E(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ConvertToInt64E(%v): error mismatch, got err %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if !errors.Is(err, ErrInvalidArgument) && tc.wantErr {
				// We expect our specific error type on failure
				t.Errorf("ConvertToInt64E(%v): expected error to wrap ErrInvalidArgument, but got %v", tc.input, err)
			}
			if got != tc.expected {
				t.Errorf("ConvertToInt64E(%v): value mismatch, got %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestConvertToSliceOfString(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected []string
		ok       bool
		wantErr  bool
	}{
		{"[]string", []string{"a", "b"}, []string{"a", "b"}, true, false},
		{"[]interface{} valid", []interface{}{"a", "b"}, []string{"a", "b"}, true, false},
		{"[]interface{} with nil", []interface{}{"a", nil, "c"}, []string{"a", "", "c"}, true, false},
		{"[]interface{} invalid", []interface{}{"a", 1}, nil, false, true},
		{"nil", nil, nil, false, true},
		{"non-slice", "hello", nil, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok, err := ConvertToSliceOfString(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ConvertToSliceOfString(%v): error mismatch, got err %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if ok != tc.ok {
				t.Errorf("ConvertToSliceOfString(%v): ok mismatch, got %v, want %v", tc.input, ok, tc.ok)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("ConvertToSliceOfString(%v): value mismatch, got %#v, want %#v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestConvertToSliceOfInt64(t *testing.T) {
	type CustomInt int
	testCases := []struct {
		name     string
		input    interface{}
		expected []int64
		ok       bool
		wantErr  bool
	}{
		{"[]int64", []int64{1, 2}, []int64{1, 2}, true, false},
		{"[]interface{} valid", []interface{}{int(1), int64(2), float64(3.0), "4"}, []int64{1, 2, 3, 4}, true, false},
		{"[]interface{} with nil", []interface{}{1, nil}, nil, false, true},
		{"[]interface{} invalid", []interface{}{"a", 1}, nil, false, true},
		{"[]int", []int{1, 2}, []int64{1, 2}, true, false},
		{"[]float64", []float64{1.1, 2.9}, []int64{1, 2}, true, false},
		{"[]CustomInt", []CustomInt{10, 20}, []int64{10, 20}, true, false},
		{"nil", nil, nil, false, true},
		{"non-slice", "hello", nil, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok, err := ConvertToSliceOfInt64(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ConvertToSliceOfInt64(%T): error mismatch, got err %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if ok != tc.ok {
				t.Errorf("ConvertToSliceOfInt64(%T): ok mismatch, got %v, want %v", tc.input, ok, tc.ok)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("ConvertToSliceOfInt64(%T): value mismatch, got %#v, want %#v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestConvertToSliceOfFloat64(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected []float64
		ok       bool
		wantErr  bool
	}{
		{"[]float64", []float64{1.1, 2.2}, []float64{1.1, 2.2}, true, false},
		{"[]interface{} valid", []interface{}{int(1), int64(2), float64(3.1), "4.2"}, []float64{1.0, 2.0, 3.1, 4.2}, true, false},
		{"[]interface{} with nil", []interface{}{1.1, nil}, nil, false, true},
		{"[]interface{} invalid", []interface{}{"a", 1.1}, nil, false, true},
		{"[]int", []int{1, 2}, []float64{1.0, 2.0}, true, false},
		{"[]string valid", []string{"1.1", "2"}, []float64{1.1, 2.0}, true, false},
		{"nil", nil, nil, false, true},
		{"non-slice", "hello", nil, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok, err := ConvertToSliceOfFloat64(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ConvertToSliceOfFloat64(%T): error mismatch, got err %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if ok != tc.ok {
				t.Errorf("ConvertToSliceOfFloat64(%T): ok mismatch, got %v, want %v", tc.input, ok, tc.ok)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("ConvertToSliceOfFloat64(%T): value mismatch, got %#v, want %#v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestConvertToSliceOfBool(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected []bool
		ok       bool
		wantErr  bool
	}{
		{"[]bool", []bool{true, false}, []bool{true, false}, true, false},
		{"[]interface{} valid", []interface{}{true, false, "true", "0", 1, 0.0}, []bool{true, false, true, false, true, false}, true, false},
		{"[]interface{} with nil", []interface{}{true, nil}, []bool{true, false}, true, false},
		{"[]interface{} invalid", []interface{}{"a", true}, nil, false, true},
		{"[]int", []int{1, 0, -1}, []bool{true, false, true}, true, false},
		{"[]string valid", []string{"true", "false", "1", "0"}, []bool{true, false, true, false}, true, false},
		{"nil", nil, nil, false, true},
		{"non-slice", "hello", nil, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok, err := ConvertToSliceOfBool(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ConvertToSliceOfBool(%T): error mismatch, got err %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if ok != tc.ok {
				t.Errorf("ConvertToSliceOfBool(%T): ok mismatch, got %v, want %v", tc.input, ok, tc.ok)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("ConvertToSliceOfBool(%T): value mismatch, got %#v, want %#v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestConvertToSliceOfMap(t *testing.T) {
	map1 := map[string]interface{}{"a": 1}
	map2 := map[string]interface{}{"b": "two"}
	mapString := map[string]string{"c": "three"}
	mapInt := map[string]int{"d": 4}

	expectedMapString := map[string]interface{}{"c": "three"}
	expectedMapInt := map[string]interface{}{"d": 4}

	testCases := []struct {
		name     string
		input    interface{}
		expected []map[string]interface{}
		ok       bool
		wantErr  bool
	}{
		{"[]map[string]interface{}", []map[string]interface{}{map1, map2}, []map[string]interface{}{map1, map2}, true, false},
		{"[]interface{} valid", []interface{}{map1, map2}, []map[string]interface{}{map1, map2}, true, false},
		{"[]interface{} mixed valid", []interface{}{map1, mapString, mapInt}, []map[string]interface{}{map1, expectedMapString, expectedMapInt}, true, false},
		{"[]interface{} with nil", []interface{}{map1, nil}, nil, false, true},
		{"[]interface{} invalid", []interface{}{map1, "hello"}, nil, false, true},
		{"[]map[string]string", []map[string]string{mapString}, []map[string]interface{}{expectedMapString}, true, false},
		{"[]map[string]int", []map[string]int{mapInt}, []map[string]interface{}{expectedMapInt}, true, false},
		{"nil", nil, nil, false, true},
		{"non-slice", "hello", nil, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok, err := ConvertToSliceOfMap(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ConvertToSliceOfMap(%T): error mismatch, got err %v, wantErr %v", tc.input, err, tc.wantErr)
				if err != nil {
					// DEBUG
					t.Logf("DEBUG: Error details: %s", err)
				}
			}
			if ok != tc.ok {
				t.Errorf("ConvertToSliceOfMap(%T): ok mismatch, got %v, want %v", tc.input, ok, tc.ok)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("ConvertToSliceOfMap(%T): value mismatch, got %#v, want %#v", tc.input, got, tc.expected)
			}
		})
	}
}

// Helper to check for error message content
func checkErrContains(t *testing.T, err error, wantErr bool, contains string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Errorf("expected an error containing %q, but got nil", contains)
			return
		}
		if !strings.Contains(err.Error(), contains) {
			t.Errorf("expected error message to contain %q, but got: %v", contains, err)
		}
	} else if err != nil {
		t.Errorf("expected no error, but got: %v", err)
	}
}

// Example of a combined test for multiple functions
func TestCombinedSliceConversions(t *testing.T) {
	// This input can be used for multiple slice conversion tests
	mixedSlice := []interface{}{
		"1",
		int(2),
		int64(3),
		float32(4.0),
		float64(5.2),
		"true", // This will fail Int64 conversion
		false,
		"0",
	}

	t.Run("Mixed to Int64", func(t *testing.T) {
		// FIX: This test is now expected to fail because "true" cannot be
		// converted to an int by ConvertToInt64E.
		_, ok, err := ConvertToSliceOfInt64(mixedSlice)
		if err == nil {
			t.Fatal("ConvertToSliceOfInt64 should have failed for 'true', but it passed")
		}
		if ok {
			t.Fatal("ConvertToSliceOfInt64 should have returned ok=false")
		}
		// Check that the error is the one we expect
		checkErrContains(t, err, true, "cannot convert string \"true\" to integer")
	})

	t.Run("Mixed to Float64", func(t *testing.T) {
		// Note: The behavior of string-to-float is different from string-to-int.
		// "true" is not a valid float, so this should fail.
		_, ok, err := ConvertToSliceOfFloat64(mixedSlice)
		if err == nil {
			t.Fatal("ConvertToSliceOfFloat64 should have failed for 'true'")
		}
		if ok {
			t.Fatal("ConvertToSliceOfFloat64 should have returned ok=false")
		}
		checkErrContains(t, err, true, "could not be converted to float64")
	})

	t.Run("Mixed to Bool", func(t *testing.T) {
		// "1", 2, 3, 4.0, 5.2, "true" are all truthy
		// false, "0" are falsy
		expected := []bool{true, true, true, true, true, true, false, false}
		got, ok, err := ConvertToSliceOfBool(mixedSlice)
		if err != nil {
			t.Fatalf("ConvertToSliceOfBool failed: %v", err)
		}
		if !ok {
			t.Fatal("ConvertToSliceOfBool returned ok=false")
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("value mismatch, got %#v, want %#v", got, expected)
		}
	})
}
