// :: product: NS
// :: majorVersion: 1
// :: fileVersion: 5
// :: description: Unit tests for tool argument coercion.
// :: latestChange: Added test cases for NSEntity map coercion (EntityID pass-through and NodeID extraction).
// :: filename: pkg/tool/tools_coerce_test.go
// :: serialization: go

package tool

import (
	"reflect"
	"testing"
)

func TestCoerceArg(t *testing.T) {
	// --- Test Fixtures ---
	mapIn := map[string]interface{}{"a": 1}
	sliceAnyIn := []interface{}{"a", 1}
	sliceStringIn := []string{"a", "b"}
	sliceIntIn := []int{1, 2}
	sliceMapIn := []map[string]interface{}{{"a": 1}, {"b": 2}}

	// NSEntity Fixtures
	validEntityID := "E_01KDVGEDWRZC0EBS566QMM90GR"
	validNodeID := "N_01KDVGEDX830JQB09F9CTRYF0W"
	nsEntityMap := map[string]interface{}{
		"id":       validEntityID,
		"_version": validNodeID,
		"fields":   map[string]interface{}{"foo": "bar"},
	}
	invalidEntityMap := map[string]interface{}{
		"id": "bad_prefix",
	}
	missingIDMap := map[string]interface{}{
		"foo": "bar",
	}

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
	unwrappedMixedList := []interface{}{"a", 1}

	// --- Expected Outputs ---
	expectedStringList := []string{"a", "b", "c"}
	expectedIntList := []int64{1, 2, 3}
	expectedFloatList := []float64{1.1, 2.2, 3.0}
	expectedBoolList := []bool{true, false, true, false}
	expectedMapList := []map[string]interface{}{
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

		// --- ArgTypeInt ---
		{"numval to int", unwrappedNum, ArgTypeInt, int64(123), false},
		{"strNum to int", unwrappedStrNum, ArgTypeInt, int64(456), false},
		{"float to int (fail)", 123.45, ArgTypeInt, nil, true},

		// --- ArgTypeFloat ---
		{"numval to float", unwrappedNum, ArgTypeFloat, 123.0, false},
		{"strNum to float", unwrappedStrNum, ArgTypeFloat, 456.0, false},

		// --- ArgTypeBool ---
		{"bool to bool", unwrappedBool, ArgTypeBool, true, false},
		{"string true to bool", "true", ArgTypeBool, true, false},
		{"num 1 to bool", 1.0, ArgTypeBool, true, false},
		{"num 0 to bool", 0.0, ArgTypeBool, false, false},

		// --- ArgTypeHandle ---
		{"valid NS Handle", "user.profile_123", ArgTypeHandle, "user.profile_123", false},
		{"invalid Handle (spaces)", "user profile", ArgTypeHandle, nil, true},

		// --- ArgTypeNodeID ---
		{"valid NodeID (string)", validNodeID, ArgTypeNodeID, validNodeID, false},
		{"invalid NodeID (wrong prefix)", "E_01KDVGEDX830JQB09F9CTRYF0W", ArgTypeNodeID, nil, true},
		{"NodeID extraction from NSEntity", nsEntityMap, ArgTypeNodeID, validNodeID, false}, // Should extract _version

		// --- ArgTypeEntityID ---
		{"valid EntityID (string)", validEntityID, ArgTypeEntityID, validEntityID, false},
		{"invalid EntityID (wrong prefix)", validNodeID, ArgTypeEntityID, nil, true},
		// New Pass-through behavior:
		{"NSEntity Map (pass-through)", nsEntityMap, ArgTypeEntityID, nsEntityMap, false},
		{"NSEntity Map (bad ID)", invalidEntityMap, ArgTypeEntityID, nil, true},
		{"NSEntity Map (missing ID)", missingIDMap, ArgTypeEntityID, nil, true},

		// --- ArgTypeMap ---
		{"map to map", unwrappedMap, ArgTypeMap, unwrappedMap, false},
		{"real: map[string]any to map", mapIn, ArgTypeMap, mapIn, false},

		// --- ArgTypeAny ---
		{"any string to any", "hello", ArgTypeAny, "hello", false},

		// --- ArgTypeSlice / ArgTypeSliceAny ---
		{"list to slice_any", unwrappedList, ArgTypeSliceAny, unwrappedList, false},
		{"[]string to slice_any", sliceStringIn, ArgTypeSliceAny, []interface{}{"a", "b"}, false},
		{"real: []any to slice_any", sliceAnyIn, ArgTypeSliceAny, sliceAnyIn, false},

		// --- ArgTypeSliceString ---
		{"[]string to slice_string", sliceStringIn, ArgTypeSliceString, sliceStringIn, false},
		{"unwrapped string list to slice_string", unwrappedStringList, ArgTypeSliceString, expectedStringList, false},
		{"unwrapped mixed list to slice_string (fail)", unwrappedMixedList, ArgTypeSliceString, nil, true},

		// --- ArgTypeSliceInt ---
		{"[]int to slice_int", sliceIntIn, ArgTypeSliceInt, []int64{1, 2}, false},
		{"unwrapped int list to slice_int", unwrappedIntList, ArgTypeSliceInt, expectedIntList, false},

		// --- ArgTypeSliceFloat ---
		{"unwrapped float list to slice_float", unwrappedFloatList, ArgTypeSliceFloat, expectedFloatList, false},

		// --- ArgTypeSliceBool ---
		{"unwrapped bool list to slice_bool", unwrappedBoolList, ArgTypeSliceBool, expectedBoolList, false},

		// --- ArgTypeSliceMap ---
		{"[]map to slice_map", sliceMapIn, ArgTypeSliceMap, sliceMapIn, false},
		{"unwrapped map list to slice_map", unwrappedMapList, ArgTypeSliceMap, expectedMapList, false},

		// --- Unknown Type ---
		{"unknown type", "hello", ArgType("bad_type"), nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := coerceArg(tc.input, tc.target)

			// Check error presence
			if (err != nil) != tc.wantErr {
				t.Fatalf("coerceArg() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr {
				return
			}

			// If we didn't expect an error, check the coerced value
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("coerceArg() got = (%T) %#v, want (%T) %#v", got, got, tc.expected, tc.expected)
			}
		})
	}
}
