// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Corrects expected values for slice and float-to-int coercion tests.
// filename: pkg/tool/tools_coerce_test.go
// nlines: 264
// risk_rating: LOW

package tool

import (
	"reflect"
	"strings"
	"testing"
)

func TestCoerceArg(t *testing.T) {
	// --- Test Fixtures ---
	mapIn := map[string]interface{}{"a": 1}
	sliceAnyIn := []interface{}{"a", 1}
	sliceStringIn := []string{"a", "b"}
	sliceIntIn := []int{1, 2}
	sliceMapIn := []map[string]interface{}{{"a": 1}, {"b": 2}}

	// From lang package, as they would be after lang.Unwrap()
	unwrappedNum := float64(123)
	unwrappedStrNum := "456"
	unwrappedBool := true
	unwrappedList := []interface{}{"a", float64(1)}
	unwrappedMap := map[string]interface{}{"key": "value"}
	unwrappedStringList := []interface{}{"a", "b", "c"}
	unwrappedIntList := []interface{}{float64(1), "2", int64(3)}
	unwrappedFloatList := []interface{}{float64(1.1), "2.2", int(3)}
	unwrappedBoolList := []interface{}{true, "false", float64(1), float64(0)}
	unwrappedMapList := []interface{}{
		map[string]interface{}{"a": 1},
		map[string]interface{}{"b": 2},
	}
	unwrappedMixedList := []interface{}{"a", 1} // Will fail type-specific slice coercion

	// --- Expected Outputs ---
	expectedStringList := []string{"a", "b", "c"} // FIX
	expectedIntList := []int64{1, 2, 3}
	expectedFloatList := []float64{1.1, 2.2, 3.0}
	expectedBoolList := []bool{true, false, true, false}
	expectedMapList := []map[string]interface{}{ // FIX
		{"a": 1},
		{"b": 2},
	}

	testCases := []struct {
		name     string
		input    interface{} // The raw value after lang.Unwrap()
		target   ArgType     // The ArgType from the tool spec
		expected interface{} // The expected Go type
		wantErr  bool
	}{
		// --- Nil Handling ---
		{"nil input", nil, ArgTypeString, nil, false},
		{"nil input to int", nil, ArgTypeInt, nil, false},
		{"nil input to slice", nil, ArgTypeSliceString, nil, false},
		{"nil input to any", nil, ArgTypeAny, nil, false},
		{"nil input to nil", nil, ArgTypeNil, nil, false},

		// --- ArgTypeString ---
		{"string to string", "hello", ArgTypeString, "hello", false},
		{"int to string (fail)", 123, ArgTypeString, nil, true},
		{"numval to string (fail)", unwrappedNum, ArgTypeString, nil, true},

		// --- ArgTypeInt ---
		{"numval to int", unwrappedNum, ArgTypeInt, int64(123), false},
		{"strNum to int", unwrappedStrNum, ArgTypeInt, int64(456), false},
		// FIX: lang.ToInt64 is lossless and rejects fractional floats. This MUST error.
		{"float to int (fail)", 123.45, ArgTypeInt, nil, true},
		{"string to int (fail)", "hello", ArgTypeInt, nil, true},
		{"bool to int (fail)", true, ArgTypeInt, nil, true}, // lang.ToInt64 returns false for bools

		// --- ArgTypeFloat ---
		{"numval to float", unwrappedNum, ArgTypeFloat, 123.0, false},
		{"strNum to float", unwrappedStrNum, ArgTypeFloat, 456.0, false},
		{"string to float (fail)", "hello", ArgTypeFloat, nil, true},
		{"bool to float (fail)", true, ArgTypeFloat, nil, true}, // lang.ToFloat64 returns false for bools

		// --- ArgTypeBool ---
		{"bool to bool", unwrappedBool, ArgTypeBool, true, false},
		{"string true to bool", "true", ArgTypeBool, true, false},
		{"string 1 to bool", "1", ArgTypeBool, true, false},
		{"string false to bool", "false", ArgTypeBool, false, false},
		{"string 0 to bool", "0", ArgTypeBool, false, false},
		{"num 1 to bool", 1.0, ArgTypeBool, true, false},
		{"num 0 to bool", 0.0, ArgTypeBool, false, false},
		{"string to bool (fail)", "hello", ArgTypeBool, nil, true},

		// --- ArgTypeMap ---
		{"map to map", unwrappedMap, ArgTypeMap, unwrappedMap, false},
		{"string to map (fail)", "hello", ArgTypeMap, nil, true},
		{"list to map (fail)", unwrappedList, ArgTypeMap, nil, true},

		// --- ArgTypeNil ---
		{"string to nil", "hello", ArgTypeNil, nil, false},
		{"int to nil", 123, ArgTypeNil, nil, false},

		// --- ArgTypeAny ---
		{"any string to any", "hello", ArgTypeAny, "hello", false},
		{"any int to any", 123, ArgTypeAny, 123, false},
		{"any numval to any", unwrappedNum, ArgTypeAny, unwrappedNum, false},
		{"any list to any", unwrappedList, ArgTypeAny, unwrappedList, false},
		{"any map to any", unwrappedMap, ArgTypeAny, unwrappedMap, false},

		// --- ArgTypeSlice / ArgTypeSliceAny ---
		{"list to slice_any", unwrappedList, ArgTypeSliceAny, unwrappedList, false},
		{"[]string to slice_any", sliceStringIn, ArgTypeSliceAny, []interface{}{"a", "b"}, false},
		{"[]int to slice_any", sliceIntIn, ArgTypeSliceAny, []interface{}{1, 2}, false},
		{"list to slice", unwrappedList, ArgTypeSlice, unwrappedList, false}, // Test alias
		{"string to slice_any (fail)", "hello", ArgTypeSliceAny, nil, true},

		// --- ArgTypeSliceString ---
		{"[]string to slice_string", sliceStringIn, ArgTypeSliceString, sliceStringIn, false},
		// FIX: Expected value is []string, not []interface{}
		{"unwrapped string list to slice_string", unwrappedStringList, ArgTypeSliceString, expectedStringList, false},
		{"unwrapped mixed list to slice_string (fail)", unwrappedMixedList, ArgTypeSliceString, nil, true},
		{"string to slice_string (fail)", "hello", ArgTypeSliceString, nil, true},
		{"[]int to slice_string (fail)", sliceIntIn, ArgTypeSliceString, nil, true},

		// --- ArgTypeSliceInt ---
		{"[]int to slice_int", sliceIntIn, ArgTypeSliceInt, []int64{1, 2}, false},
		{"unwrapped int list to slice_int", unwrappedIntList, ArgTypeSliceInt, expectedIntList, false},
		{"unwrapped mixed list to slice_int (fail)", unwrappedMixedList, ArgTypeSliceInt, nil, true},
		{"string to slice_int (fail)", "hello", ArgTypeSliceInt, nil, true},
		{"[]string to slice_int (fail)", sliceStringIn, ArgTypeSliceInt, nil, true},

		// --- ArgTypeSliceFloat ---
		{"unwrapped float list to slice_float", unwrappedFloatList, ArgTypeSliceFloat, expectedFloatList, false},
		{"[]int to slice_float", sliceIntIn, ArgTypeSliceFloat, []float64{1.0, 2.0}, false},
		{"unwrapped mixed list to slice_float (fail)", unwrappedMixedList, ArgTypeSliceFloat, nil, true},
		{"string to slice_float (fail)", "hello", ArgTypeSliceFloat, nil, true},
		{"[]string to slice_float (fail)", sliceStringIn, ArgTypeSliceFloat, nil, true},

		// --- ArgTypeSliceBool ---
		{"unwrapped bool list to slice_bool", unwrappedBoolList, ArgTypeSliceBool, expectedBoolList, false},
		{"[]int to slice_bool", sliceIntIn, ArgTypeSliceBool, []bool{true, true}, false},
		{"unwrapped mixed list to slice_bool (fail)", unwrappedMixedList, ArgTypeSliceBool, nil, true},
		{"string to slice_bool (fail)", "hello", ArgTypeSliceBool, nil, true},

		// --- ArgTypeSliceMap ---
		{"[]map to slice_map", sliceMapIn, ArgTypeSliceMap, sliceMapIn, false},
		// FIX: Expected value is []map[string]interface{}, not []interface{}
		{"unwrapped map list to slice_map", unwrappedMapList, ArgTypeSliceMap, expectedMapList, false},
		{"unwrapped mixed list to slice_map (fail)", unwrappedMixedList, ArgTypeSliceMap, nil, true},
		{"string to slice_map (fail)", "hello", ArgTypeSliceMap, nil, true},

		// --- Unknown Type ---
		{"unknown type", "hello", ArgType("bad_type"), nil, true},
		{"unknown type with int", 123, ArgType("bad_type"), nil, true},

		// --- Edge cases from previous bugs ---
		{"real: map[string]any to map", mapIn, ArgTypeMap, mapIn, false},
		{"real: []any to slice_any", sliceAnyIn, ArgTypeSliceAny, sliceAnyIn, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// DEBUG
			// t.Logf("--- RUNNING: %s ---", tc.name)
			// t.Logf("Input: (%T) %#v", tc.input, tc.input)
			// t.Logf("Target: %s", tc.target)

			got, err := coerceArg(tc.input, tc.target)

			// Check error presence
			if (err != nil) != tc.wantErr {
				t.Fatalf("coerceArg() error = %v, wantErr %v", err, tc.wantErr)
			}

			// If we expected an error, check its content
			if tc.wantErr {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
				// Check for the "unknown argument type" error specifically
				if tc.target == ArgType("bad_type") {
					if !strings.Contains(err.Error(), "unknown argument type") {
						t.Errorf("Expected 'unknown argument type' error, but got: %v", err)
					}
				}
				return // Test is done
			}

			// If we didn't expect an error, check the coerced value
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("coerceArg() got = (%T) %#v, want (%T) %#v", got, got, tc.expected, tc.expected)
			}
		})
	}
}
