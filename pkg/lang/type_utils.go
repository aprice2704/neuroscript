// NeuroScript Version: 0.4.1
// File version: 9
// Purpose: Updated unwrapValue to handle FunctionValue and ToolValue.
// Filename: pkg/core/type_utils.go

package lang

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// TypeOf returns the NeuroScriptType name for any given value, whether wrapped or raw.
func TypeOf(value interface{}) NeuroScriptType {
	if value == nil {
		return TypeNil
	}

	/* ---------- concrete non-Value types that appear unwrapped ---------- */

	switch value.(type) {
	case Callable, *Callable:
		return TypeFunction
	case Tool, *Tool:
		return TypeTool
	case time.Time, *time.Time: // native Go time value
		return TypeTimedate
	case []byte: // raw byte slice
		return TypeBytes
	case error: // plain Go error
		return TypeError
	}

	/* ---------- custom wrapper types that implement the Value interface ---------- */

	if v, ok := value.(Value); ok {
		return v.Type()
	}

	/* ---------- fallback for ordinary Go types (common in tests / tool args) ---------- */

	val := reflect.ValueOf(value)
	kind := val.Kind()

	// Dereference pointers / interfaces to see what's underneath.
	if kind == reflect.Interface || kind == reflect.Ptr {
		if val.IsNil() {
			return TypeNil
		}
		val = val.Elem()
		kind = val.Kind()
	}

	switch kind {
	case reflect.String:
		return TypeString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return TypeNumber
	case reflect.Bool:
		return TypeBoolean
	case reflect.Slice, reflect.Array:
		return TypeList
	case reflect.Map:
		return TypeMap
	case reflect.Func:
		return TypeFunction
	default:
		return TypeUnknown
	}
}

// unwrapValue recursively unwraps a Value type to its underlying native Go type.
// If the input is not a Value type, it's returned as is.
func unwrapValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case StringValue:
		return val.Value
	case NumberValue:
		return val.Value
	case BoolValue:
		return val.Value
	case NilValue:
		return nil
	case TimedateValue:
		return val.Value
	case FuzzyValue:
		// Fuzzy doesn't have a direct native equivalent, return its float value
		return val.μ
	case FunctionValue:
		return val.Value // Unwrap to Procedure struct
	case ToolValue:
		return val.Value // Unwrap to ToolImplementation struct
	case ErrorValue:
		// Unwrap the inner map
		unwrappedMap := make(map[string]interface{})
		for k, innerVal := range val.Value {
			unwrappedMap[k] = unwrapValue(innerVal)
		}
		return unwrappedMap
	case ListValue:
		// Recursively unwrap elements in the list
		unwrappedList := make([]interface{}, len(val.Value))
		for i, item := range val.Value {
			unwrappedList[i] = unwrapValue(item)
		}
		return unwrappedList
	case MapValue:
		// Recursively unwrap values in the map
		unwrappedMap := make(map[string]interface{})
		for k, innerVal := range val.Value {
			unwrappedMap[k] = unwrapValue(innerVal)
		}
		return unwrappedMap
	default:
		// It's already a native Go type or another complex type like Procedure
		return v
	}
}

// toFloat64 attempts conversion to float64 from various raw and wrapped types.
func toFloat64(val interface{}) (float64, bool) {
	nativeVal := unwrapValue(val)
	if nativeVal == nil {
		return 0, false
	}
	switch v := nativeVal.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		// Try reflection for other numeric types
		rv := reflect.ValueOf(nativeVal)
		switch rv.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32:
			return float64(rv.Int()), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(rv.Uint()), true
		case reflect.Float32:
			return rv.Float(), true
		}
		return 0, false
	}
}

// toInt64 attempts lossless conversion to int64 from various raw and wrapped types.
func toInt64(val interface{}) (int64, bool) {
	f, ok := toFloat64(val)
	if !ok {
		return 0, false
	}
	// Check if the float has a fractional part
	if f != float64(int64(f)) {
		return 0, false
	}
	return int64(f), true
}

// toString converts a value to its string representation.
// It will always produce a string, but the boolean indicates if the original type
// was naturally a string.
func toString(val interface{}) (string, bool) {
	if val == nil {
		return "", false // FIX: Return empty string for nil to correct concatenation behavior.
	}
	if s, ok := val.(string); ok {
		return s, true
	}
	if v, ok := val.(Value); ok {
		// Use the value's own String() method
		if _, isNil := v.(NilValue); isNil {
			return "", false // FIX: Also handle the NilValue type explicitly.
		}
		_, isStr := v.(StringValue)
		return v.String(), isStr
	}
	// Fallback for other native types
	return fmt.Sprintf("%v", val), false
}

// ToNumeric attempts conversion to a wrapped NumberValue.
// It returns the NumberValue and a boolean indicating success.
func ToNumeric(val interface{}) (NumberValue, bool) {
	if f, ok := toFloat64(val); ok {
		return NumberValue{Value: f}, true
	}
	return NumberValue{}, false
}

// IsTruthy determines the truthiness of a Value according to NeuroScript rules.
// This function centralizes the logic for how different types are evaluated
// in boolean contexts like 'if', 'while', and 'must' statements.
func IsTruthy(v Value) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case BoolValue:
		return val.Value
	case StringValue:
		// Per the spec, only specific string values are truthy.
		return val.Value == "true" || val.Value == "1"
	case NumberValue:
		// Any non-zero number is truthy.
		return val.Value != 0
	case ListValue:
		// A list is truthy if it is not empty.
		return len(val.Value) > 0
	case MapValue:
		// A map is truthy if it is not empty.
		return len(val.Value) > 0
	case NilValue:
		// Nil is always falsy.
		return false
	case ErrorValue, TimedateValue, FuzzyValue, EventValue:
		// Complex types are considered "truthy" if they exist,
		// similar to how objects are treated in other languages.
		return true
	default:
		// Any other unknown or unhandled type is considered falsy by default.
		return false
	}
}

// isZeroValue checks if a value is its "zero" or "empty" equivalent.
// This is used for the 'no' and 'some' unary operators.
func isZeroValue(val interface{}) bool {
	if val == nil {
		return true
	}

	// Handle Value types first
	if v, ok := val.(Value); ok {
		if _, isNil := v.(NilValue); isNil {
			return true
		}
		// For other value types, their IsTruthy method defines their non-zero state.
		// isZero is the opposite of IsTruthy for these types.
		return !v.IsTruthy()
	}

	// Fallback for native types
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		// For other types, check against the type's zero value.
		return v.IsZero()
	}
}
