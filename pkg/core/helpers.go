package core

import (
	"fmt"
	"reflect"
	"strings"
)

// Helper for logging snippets
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ConvertToBool implements NeuroScript truthiness specifically for validating LLM input.
// Returns the bool value and true if conversion is valid, otherwise false, false.
func ConvertToBool(val interface{}) (bool, bool) {
	if val == nil {
		return false, true
	} // nil is false
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
		// Other strings are NOT considered valid booleans during strict validation
		return false, false
	// Handle other potential numeric types from JSON unmarshal
	case int, int32:
		rv := reflect.ValueOf(val)
		return rv.Int() != 0, true
	case float32:
		rv := reflect.ValueOf(val)
		return rv.Float() != 0.0, true
	default:
		// Other types (like slices, maps) are not valid booleans
		return false, false
	}
}

// convertToSliceOfString handles conversion for ArgTypeSliceString validation.
// Returns the []string, true + nil error on success.
// Returns nil, false + specific error on failure.
func convertToSliceOfString(rawValue interface{}) ([]string, bool, error) {
	switch rv := rawValue.(type) {
	case []string:
		return rv, true, nil // Already correct type
	case []interface{}:
		// Convert []interface{} ONLY IF all elements are strings or nil
		strSlice := make([]string, len(rv))
		for i, item := range rv {
			if itemStr, ok := item.(string); ok {
				strSlice[i] = itemStr
			} else if item == nil {
				strSlice[i] = "" // Treat nil as empty string
			} else {
				// Element is not a string or nil
				return nil, false, fmt.Errorf("expected slice of strings, but element %d has incompatible type %T", i, item)
			}
		}
		return strSlice, true, nil
	default:
		// Type is not []string or []interface{}
		return nil, false, fmt.Errorf("expected slice of strings, got %T", rawValue)
	}
}

// convertToSliceOfAny handles conversion for ArgTypeSliceAny validation.
// Returns the []interface{}, true + nil error on success.
// Returns nil, false + specific error on failure.
func convertToSliceOfAny(rawValue interface{}) ([]interface{}, bool, error) {
	switch rv := rawValue.(type) {
	case []interface{}:
		return rv, true, nil // Already correct type
	case []string: // Also accept []string and convert it
		anySlice := make([]interface{}, len(rv))
		for i, s := range rv {
			anySlice[i] = s
		}
		return anySlice, true, nil
	default:
		// Type is not []interface{} or []string
		return nil, false, fmt.Errorf("expected a slice (list), got %T", rawValue)
	}
}
