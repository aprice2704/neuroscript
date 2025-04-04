// pkg/core/evaluation_helpers.go
package core

import (
	"math"
	"strconv"
	"strings"
	"unicode" // Keep for isValidIdentifier
)

// --- Type Conversion/Checking Helpers ---

// toFloat64 attempts conversion to float64.
func toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case string:
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

// toInt64 attempts conversion to int64 (only if lossless).
func toInt64(val interface{}) (int64, bool) {
	switch v := val.(type) {
	case int64:
		return v, true
	case float64:
		// Allow conversion only if there's no fractional part
		if v == math.Trunc(v) {
			// Check for potential overflow if converting large floats?
			// For now, assume direct conversion is okay if Trunc matches.
			return int64(v), true
		}
		return 0, false // Don't convert float with fraction
	case string:
		// Try parsing as int first
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return i, true
		}
		// If int fails, try float then check if it's a whole number
		f, err := strconv.ParseFloat(v, 64)
		if err == nil && f == math.Trunc(f) {
			return int64(f), true
		}
		return 0, false
	// Add other base Go numeric types
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case float32: // Allow lossless conversion from float32
		if v == float32(int64(v)) {
			return int64(v), true
		}
		return 0, false
	default:
		return 0, false
	}
}

// isTruthy evaluates if a value is considered true in NeuroScript boolean contexts.
func isTruthy(val interface{}) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case float64:
		return v != 0.0
	case string:
		lowerV := strings.ToLower(v)
		// Strict check: only "true" or "1" are truthy strings
		return lowerV == "true" || v == "1"
	// Handle other potential numeric types from Go conversions if necessary
	case int:
		return v != 0
	case int32:
		return v != 0
	case float32:
		return v != 0.0
	// Collections are considered falsey
	case []interface{}, map[string]interface{}, []string:
		return false
	default:
		// Any other type is considered falsey
		return false
	}
}

// --- Other Helpers ---

// tryParseFloat attempts to parse a string as float64. (Can likely be deprecated if toFloat64 handles strings)
// Keeping for now if used elsewhere, but review its usage.
func tryParseFloat(s string) (float64, bool) {
	val, err := strconv.ParseFloat(s, 64)
	return val, err == nil
}

// isValidIdentifier checks if a string is a valid NeuroScript identifier (and not a keyword).
func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for idx, r := range name {
		if idx == 0 {
			// Must start with a letter or underscore
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			// Subsequent characters can be letters, digits, or underscores
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	// Case-insensitive keyword check
	upperName := strings.ToUpper(name)
	keywords := map[string]bool{
		"DEFINE": true, "PROCEDURE": true, "END": true, "ENDBLOCK": true,
		"COMMENT": true, "ENDCOMMENT": true, // Note: These might not be needed if COMMENT_BLOCK handled by lexer
		"SET": true, "CALL": true, "RETURN": true, "EMIT": true,
		"IF": true, "THEN": true, "ELSE": true,
		"WHILE": true, "DO": true,
		"FOR": true, "EACH": true, "IN": true,
		"TOOL": true, "LLM": true, "LAST": true, "EVAL": true,
		"TRUE": true, "FALSE": true,
		"AND": true, "OR": true, "NOT": true,
		"LN": true, "LOG": true, "SIN": true, "COS": true, "TAN": true,
		"ASIN": true, "ACOS": true, "ATAN": true,
		"FILE_VERSION": true,
		// Special variable __last_call_result treated like a keyword here
		"__LAST_CALL_RESULT": true,
	}
	// Allow __last_call_result specifically
	if name == "__last_call_result" {
		return true
	}
	// If it's in the keyword map (case-insensitive), it's not a valid identifier
	if keywords[upperName] {
		return false
	}
	return true // Passes structural and keyword checks
}
