// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Fixes ConvertToInt64E to handle custom types based on integer kinds.
// filename: pkg/utils/utils.go
// nlines: 257
// risk_rating: MEDIUM

package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ErrInvalidArgument is used when a function receives an argument of the wrong type or format.
var ErrInvalidArgument = errors.New("invalid argument")

// ConvertToBool implements NeuroScript truthiness. Exported for package use.
func ConvertToBool(val interface{}) (bool, bool) {
	if val == nil {
		return false, true
	}
	switch v := val.(type) {
	case bool:
		return v, true
	case int64:
		return v != 0, true
	case float64:
		return v != 0.0, true
	case string:
		lowerV := strings.ToLower(v)
		if lowerV == "true" || v == "1" {
			return true, true
		}
		if lowerV == "false" || v == "0" {
			return false, true
		}
		return false, false
	case int, int32:
		return reflect.ValueOf(val).Int() != 0, true
	case float32:
		return reflect.ValueOf(val).Float() != 0.0, true
	default:
		return false, false
	}
}

// ConvertToSliceOfAny attempts to convert an interface{} value into a []interface{}.
// This now robustly handles any slice type (e.g., []string, []int) using reflection.
func ConvertToSliceOfAny(rawValue interface{}) ([]interface{}, bool, error) {
	if rawValue == nil {
		return nil, false, fmt.Errorf("cannot convert nil to slice")
	}

	val := reflect.ValueOf(rawValue)
	if val.Kind() != reflect.Slice {
		return nil, false, fmt.Errorf("expected a slice (list), got %T", rawValue)
	}

	length := val.Len()
	anySlice := make([]interface{}, length)
	for i := 0; i < length; i++ {
		anySlice[i] = val.Index(i).Interface()
	}
	return anySlice, true, nil
}

// ConvertToFloat64 is a helper to handle potential int64/float64 from map[string]interface{}
func ConvertToFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case string:
		fVal, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return fVal, true
		}
	}
	return 0, false
}

// ConvertToInt64E attempts to convert various numeric types (and potentially strings representing numbers) to int64.
func ConvertToInt64E(value interface{}) (int64, error) {
	if value == nil {
		return 0, fmt.Errorf("%w: cannot convert nil to integer", ErrInvalidArgument)
	}
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			// Try parsing as float first for cases like "10.0"
			f, fErr := strconv.ParseFloat(v, 64)
			if fErr != nil {
				return 0, fmt.Errorf("%w: cannot convert string %q to integer: %w", ErrInvalidArgument, v, err)
			}
			return int64(f), nil
		}
		return i, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		// Fallback for other numeric types AND custom types based on them
		valOf := reflect.ValueOf(value)
		switch valOf.Kind() {
		// FIX: Added reflect.Int, reflect.Int64
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return valOf.Int(), nil
		// FIX: Added reflect.Uint, reflect.Uintptr
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return int64(valOf.Uint()), nil
		// FIX: Added reflect.Float64
		case reflect.Float32, reflect.Float64:
			return int64(valOf.Float()), nil
		}
		return 0, fmt.Errorf("%w: cannot convert type %T to integer", ErrInvalidArgument, value)
	}
}

// ConvertToSliceOfString attempts to convert an interface{} value into a []string.
func ConvertToSliceOfString(rawValue interface{}) ([]string, bool, error) {
	switch rv := rawValue.(type) {
	case []string:
		return rv, true, nil
	case []interface{}:
		strSlice := make([]string, len(rv))
		for i, item := range rv {
			if itemStr, ok := item.(string); ok {
				strSlice[i] = itemStr
			} else if item == nil {
				strSlice[i] = ""
			} else {
				return nil, false, fmt.Errorf("expected slice of strings, but element %d has incompatible type %T", i, item)
			}
		}
		return strSlice, true, nil
	default:
		return nil, false, fmt.Errorf("expected slice of strings, got %T", rawValue)
	}
}

// ConvertToSliceOfInt64 attempts to convert an interface{} value into a []int64.
func ConvertToSliceOfInt64(rawValue interface{}) ([]int64, bool, error) {
	anySlice, ok, err := ConvertToSliceOfAny(rawValue)
	if !ok {
		return nil, false, err
	}

	intSlice := make([]int64, len(anySlice))
	for i, item := range anySlice {
		i64, err := ConvertToInt64E(item) // Use existing helper
		if err != nil {
			return nil, false, fmt.Errorf("element %d (%T) could not be converted to int64: %w", i, item, err)
		}
		intSlice[i] = i64
	}
	return intSlice, true, nil
}

// ConvertToSliceOfFloat64 attempts to convert an interface{} value into a []float64.
func ConvertToSliceOfFloat64(rawValue interface{}) ([]float64, bool, error) {
	anySlice, ok, err := ConvertToSliceOfAny(rawValue)
	if !ok {
		return nil, false, err
	}

	floatSlice := make([]float64, len(anySlice))
	for i, item := range anySlice {
		f64, ok := ConvertToFloat64(item) // Use existing helper
		if !ok {
			return nil, false, fmt.Errorf("element %d (%T) could not be converted to float64", i, item)
		}
		floatSlice[i] = f64
	}
	return floatSlice, true, nil
}

// ConvertToSliceOfBool attempts to convert an interface{} value into a []bool.
func ConvertToSliceOfBool(rawValue interface{}) ([]bool, bool, error) {
	anySlice, ok, err := ConvertToSliceOfAny(rawValue)
	if !ok {
		return nil, false, err
	}

	boolSlice := make([]bool, len(anySlice))
	for i, item := range anySlice {
		b, ok := ConvertToBool(item) // Use existing helper
		if !ok {
			return nil, false, fmt.Errorf("element %d (%T) could not be converted to bool", i, item)
		}
		boolSlice[i] = b
	}
	return boolSlice, true, nil
}

// ConvertToSliceOfMap attempts to convert an interface{} value into a []map[string]interface{}.
func ConvertToSliceOfMap(rawValue interface{}) ([]map[string]interface{}, bool, error) {
	anySlice, ok, err := ConvertToSliceOfAny(rawValue)
	if !ok {
		return nil, false, err
	}

	mapSlice := make([]map[string]interface{}, len(anySlice))
	for i, item := range anySlice {
		m, ok := item.(map[string]interface{})
		if !ok {
			// Also check for typed maps
			rv := reflect.ValueOf(item)
			if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
				// Convert map[string]SomeType to map[string]interface{}
				m = make(map[string]interface{}, rv.Len())
				iter := rv.MapRange()
				for iter.Next() {
					m[iter.Key().String()] = iter.Value().Interface()
				}
				ok = true
			}
		}

		if !ok {
			return nil, false, fmt.Errorf("element %d expected map[string]any, got %T", i, item)
		}
		mapSlice[i] = m
	}
	return mapSlice, true, nil
}
