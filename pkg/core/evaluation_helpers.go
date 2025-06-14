// NeuroScript Version: 0.4.1
// File version: 9
// Purpose: Updated unwrapValue to handle FunctionValue and ToolValue.
// filename: pkg/core/evaluation_helpers.go
// nlines: 234
// risk_rating: LOW

package core

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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

// TypeOf returns the NeuroScriptType name for any given value, whether wrapped or raw.
func TypeOf(value interface{}) NeuroScriptType {
	if value == nil {
		return TypeNil
	}

	// FIX: Check for concrete complex types like Procedure and ToolImplementation first,
	// as they do not implement the Value interface and would be missed by later checks.
	switch value.(type) {
	case Procedure, *Procedure:
		return TypeFunction
	case ToolImplementation, *ToolImplementation:
		return TypeTool
	}

	// Then, check for our custom Value wrapper types.
	if v, ok := value.(Value); ok {
		return v.Type()
	}

	// Fallback for native Go types (often used in tests or tool args)
	val := reflect.ValueOf(value)
	kind := val.Kind()

	// Dereference pointers and interfaces to get the underlying kind
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

// isTruthy determines the truthiness of a NeuroScript value.
func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}

	// Prioritize our custom Value types, which have a defined IsTruthy method.
	if v, ok := value.(Value); ok {
		// FIX: String truthiness is specific, delegate to specialized logic.
		if sv, isString := v.(StringValue); isString {
			lowerV := strings.ToLower(sv.Value)
			return lowerV == "true" || lowerV == "1"
		}
		return v.IsTruthy()
	}

	// Fallback for raw Go types, mirroring the logic in the Value types.
	switch v := value.(type) {
	case bool:
		return v
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0
	case string:
		// FIX: Raw strings also follow specific truthiness rules.
		lowerV := strings.ToLower(v)
		return lowerV == "true" || lowerV == "1"
	case time.Time:
		return !v.IsZero()
	default:
		// For other complex types like slices or maps, check if they are non-nil and have length > 0
		valOf := reflect.ValueOf(value)
		if valOf.Kind() == reflect.Slice || valOf.Kind() == reflect.Map {
			return !valOf.IsNil() && valOf.Len() > 0
		}
		// All other unhandled types are considered false.
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
