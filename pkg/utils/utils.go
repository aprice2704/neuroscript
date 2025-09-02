// NeuroScript Version: 0.3.1
// File version: 2
// Purpose: Removed imports of 'interfaces' and 'lang' to break a package dependency cycle.
// filename: pkg/utils/utils.go
// nlines: 150
// risk_rating: LOW

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

// ConvertToSliceOfAny attempts to convert an interface{} value into a []interface{}. Exported for package use.
func ConvertToSliceOfAny(rawValue interface{}) ([]interface{}, bool, error) {
	switch rv := rawValue.(type) {
	case []interface{}:
		return rv, true, nil
	case []string:
		anySlice := make([]interface{}, len(rv))
		for i, s := range rv {
			anySlice[i] = s
		}
		return anySlice, true, nil
	default:
		return nil, false, fmt.Errorf("expected a slice (list), got %T", rawValue)
	}
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
		// Fallback for other numeric types
		valOf := reflect.ValueOf(value)
		switch valOf.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32:
			return valOf.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int64(valOf.Uint()), nil
		case reflect.Float32:
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
