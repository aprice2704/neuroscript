// filename: pkg/core/evaluation_helpers.go
package core

import (
	"fmt"     // Import fmt for Sprintf
	"math"    // Required for toInt64/toFloat64 placeholders
	"reflect" // Needed for isTruthy map/slice check
	"strconv" // Required for toInt64/toFloat64 placeholders
	"strings"
	// Keep for isValidIdentifier
)

// --- Type Conversion/Checking Helpers ---

// toFloat64 attempts conversion to float64.
func toFloat64(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case string:
		// Only convert if it's a valid float string
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	// Add other base Go numeric types for robustness if they might appear
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

// toInt64 attempts conversion to int64 (only if lossless or from valid string).
func toInt64(val interface{}) (int64, bool) {
	if val == nil {
		return 0, false
	}
	switch v := val.(type) {
	case int64:
		return v, true
	case int: // Handle plain int
		return int64(v), true
	case int32: // Handle int32
		return int64(v), true
	case float64:
		// Allow conversion only if there's no fractional part
		if v == math.Trunc(v) {
			return int64(v), true
		}
		return 0, false // Don't convert float with fraction
	case float32: // Allow lossless conversion from float32
		fv64 := float64(v)
		if fv64 == math.Trunc(fv64) {
			return int64(fv64), true
		}
		return 0, false
	case string:
		// Try parsing as int first
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return i, true
		}
		// If int fails, try float then check if it's a whole number
		f, errF := strconv.ParseFloat(v, 64)
		if errF == nil && f == math.Trunc(f) {
			return int64(f), true
		}
		return 0, false
	default:
		return 0, false
	}
}

// *** ADDED toString helper ***
// toString converts a value to its string representation.
// It specifically handles nil as an empty string.
// Returns the string representation and true if the original type was string.
func toString(val interface{}) (string, bool) {
	if val == nil {
		return "", false // Represent nil as empty string, original type was not string
	}
	if s, ok := val.(string); ok {
		return s, true // Original type was string
	}
	// For other types, use default formatting, original type was not string
	return fmt.Sprintf("%v", val), false
}

// ToNumeric attempts conversion to int64 or float64.
func ToNumeric(val interface{}) (interface{}, bool) {
	if val == nil { // Explicitly handle nil
		return nil, false
	}
	// Try int64 first
	if i, ok := toInt64(val); ok {
		return i, true
	}
	// Then try float64
	if f, ok := toFloat64(val); ok {
		return f, true
	}
	// Cannot convert to either numeric type
	return nil, false
}

// isTruthy evaluates if a value is considered true in NeuroScript boolean contexts.
// UPDATED: Non-empty strings (other than "false" or "0") are now truthy.
// FIXED: Stricter string truthiness - only "true" and "1" are truthy. Others (even non-empty) are falsy.
func isTruthy(val interface{}) bool {
	if val == nil {
		return false // Nil is falsy
	}
	switch v := val.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case float64:
		return v != 0.0
	case string:
		// *** FIX: Only "true" (case-insensitive) and "1" are truthy ***
		lowerV := strings.ToLower(v)
		return lowerV == "true" || v == "1"
	// Handle other Go numeric types
	case int:
		return v != 0
	case int32:
		return v != 0
	case float32:
		return v != 0.0
	// Collections (slices/maps) are considered falsey if empty, truthy otherwise
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	case []string:
		return len(v) > 0
	default:
		// Any other type is considered truthy if it's not its zero value
		// Exception: Empty strings are handled above and are falsy.
		// Use reflect to check if the value is the zero value for its type.
		rv := reflect.ValueOf(v)
		// Check if the type is valid and has a zero value concept before calling IsZero
		if rv.IsValid() {
			// Check specifically for pointers, interfaces, maps, slices, channels, funcs being nil
			switch rv.Kind() {
			case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
				return !rv.IsNil()
			}
			// For other types, IsZero might be applicable, but defaulting to true is safer?
			// Or consider if IsZero accurately reflects "truthiness" for structs etc.
			// Let's default to TRUE for unrecognized non-zeroable types for now.
			// This maintains previous behavior for things like structs.
			// If a specific type needs falsiness check, it should be added explicitly.
			// return !rv.IsZero() // This could be used, but let's be conservative.
		}
		// If reflect couldn't determine or type is invalid, default to truthy (as before)
		return true
	}
}

// *** ADDED trimCodeFences helper ***
// trimCodeFences removes optional leading/trailing triple backticks and surrounding whitespace.
// func trimCodeFences(s string) string {
// 	trimmed := strings.TrimSpace(s)
// 	if strings.HasPrefix(trimmed, "```") && strings.HasSuffix(trimmed, "```") {
// 		trimmed = strings.TrimPrefix(trimmed, "```")
// 		trimmed = strings.TrimSuffix(trimmed, "```")
// 		// Also trim potential language identifier after opening fence and whitespace
// 		firstNewline := strings.Index(trimmed, "\n")
// 		if firstNewline != -1 {
// 			firstLine := strings.TrimSpace(trimmed[:firstNewline])
// 			// Simple check if the first line looks like just a language identifier
// 			// (e.g., "go", "python", "json"). More robust parsing could be added.
// 			// For now, if it's short and has no spaces, assume it's a language hint and remove it.
// 			if len(firstLine) > 0 && !strings.ContainsAny(firstLine, " \t(){}[];:=") {
// 				trimmed = trimmed[firstNewline:] // Keep from newline onwards
// 			}
// 		}
// 		trimmed = strings.TrimSpace(trimmed) // Trim again after removing fences/hints
// 	}
// 	return trimmed
// }
