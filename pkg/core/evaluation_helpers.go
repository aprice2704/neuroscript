// filename: pkg/core/evaluation_helpers.go
// file version: 3
package core

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// --- Type Conversion/Checking Helpers ---

// toFloat64 attempts conversion to float64 from various raw and wrapped types.
func toFloat64(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false
	}
	switch v := val.(type) {
	// Wrapped Value type
	case NumberValue:
		return v.Value, true
	// Raw Go types
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}

// toInt64 attempts lossless conversion to int64 from various raw and wrapped types.
func toInt64(val interface{}) (int64, bool) {
	if val == nil {
		return 0, false
	}
	switch v := val.(type) {
	// Wrapped Value type
	case NumberValue:
		if v.Value == math.Trunc(v.Value) {
			return int64(v.Value), true
		}
		return 0, false // Don't convert float with fraction
	// Raw Go types
	case int64:
		return v, true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case float64:
		if v == math.Trunc(v) {
			return int64(v), true
		}
		return 0, false
	case float32:
		fv64 := float64(v)
		if fv64 == math.Trunc(fv64) {
			return int64(fv64), true
		}
		return 0, false
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return i, true
		}
		f, errF := strconv.ParseFloat(v, 64)
		if errF == nil && f == math.Trunc(f) {
			return int64(f), true
		}
		return 0, false
	default:
		return 0, false
	}
}

// toString converts a value to its string representation.
// It returns the string and a boolean indicating if the original type was a string.
func toString(val interface{}) (string, bool) {
	if val == nil {
		return "", false
	}
	switch v := val.(type) {
	case StringValue:
		return v.Value, true
	case Value: // Handles other Value types like NumberValue, BoolValue etc.
		return v.String(), false
	case string:
		return v, true
	default:
		return fmt.Sprintf("%v", val), false
	}
}

// ToNumeric attempts conversion to a wrapped NumberValue.
func ToNumeric(val interface{}) (NumberValue, bool) {
	if f, ok := toFloat64(val); ok {
		return NumberValue{Value: f}, true
	}
	return NumberValue{}, false
}

// isTruthy determines the truthiness of a NeuroScript value.
func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}
	switch v := value.(type) {
	// Wrapped Value types
	case NilValue:
		return false
	case BoolValue:
		return v.Value
	case NumberValue:
		return v.Value != 0
	case StringValue:
		lowerV := strings.ToLower(v.Value)
		return lowerV == "true" || v.Value == "1"
	case ListValue:
		return len(v.Value) > 0
	case MapValue:
		return len(v.Value) > 0
	case Procedure, ToolImplementation, ErrorValue, EventValue, TimedateValue, FuzzyValue:
		return true // Complex types are truthy if they exist.

	// Raw Go types (primarily for test harness)
	case bool:
		return v
	case int, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0
	case string:
		lowerV := strings.ToLower(v)
		return lowerV == "true" || v == "1"

	default:
		val := reflect.ValueOf(value)
		kind := val.Kind()
		if kind == reflect.Slice || kind == reflect.Map {
			return !val.IsNil() && val.Len() > 0
		}
		// Default for unhandled types is false.
		return false
	}
}
