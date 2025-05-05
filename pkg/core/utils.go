// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 22:35:18 PDT // Add ConvertToIntE helper
// filename: pkg/core/utils.go
package core

import (
	"fmt"
	"math" // Needed for ConvertToIntE
	"reflect"
	"strings"

	// Import the logging interface definition
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// --- Internal No-Op Logger ---
// --- (coreNoOpLogger struct and methods remain unchanged) ---
type coreNoOpLogger struct{}

var _ logging.Logger = (*coreNoOpLogger)(nil)

func (l *coreNoOpLogger) Debug(msg string, args ...any)     {}
func (l *coreNoOpLogger) Info(msg string, args ...any)      {}
func (l *coreNoOpLogger) Warn(msg string, args ...any)      {}
func (l *coreNoOpLogger) Error(msg string, args ...any)     {}
func (l *coreNoOpLogger) Debugf(format string, args ...any) {}
func (l *coreNoOpLogger) Infof(format string, args ...any)  {}
func (l *coreNoOpLogger) Warnf(format string, args ...any)  {}
func (l *coreNoOpLogger) Errorf(format string, args ...any) {}
func (l *coreNoOpLogger) SetLevel(level logging.LogLevel)   {}

// --- Type Conversion / Checking Utilities ---

// ConvertToIntE attempts to convert an interface{} value to an int.
// It handles int, int64, and float64 (if the float has no fractional part).
// It returns the int value and true on success, or 0 and false on failure.
func ConvertToIntE(value interface{}) (int, bool) {
	if value == nil {
		return 0, false
	}
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		// Check for potential overflow if converting int64 to int
		// (though unlikely in typical use cases like line/col numbers)
		if v < math.MinInt || v > math.MaxInt {
			return 0, false // Overflow
		}
		return int(v), true
	case float64:
		// Check if float is actually a whole number
		if v == math.Trunc(v) {
			// Check for potential overflow
			if v < math.MinInt || v > math.MaxInt {
				return 0, false
			}
			return int(v), true
		}
		return 0, false // Has fractional part
		// Add cases for other potential numeric types if needed (e.g., float32, int32)
	case float32:
		if v == float32(math.Trunc(float64(v))) {
			if v < math.MinInt || v > math.MaxInt {
				return 0, false
			}
			return int(v), true
		}
		return 0, false
	case int32:
		// int32 always fits in int on 32-bit and 64-bit systems
		return int(v), true
	default:
		return 0, false // Not a convertible numeric type
	}
}

// IsTruthy determines the truthiness of a NeuroScript value.
// --- (IsTruthy remains unchanged) ---
func IsTruthy(val interface{}) bool {
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
		return v != ""
	case int:
		return v != 0
	case int32:
		return v != 0
	case float32:
		return v != 0.0
	default:
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface || rv.Kind() == reflect.Map || rv.Kind() == reflect.Slice || rv.Kind() == reflect.Chan || rv.Kind() == reflect.Func {
			return !rv.IsNil()
		}
		return true
	}
}

// InterfaceToString attempts to convert an interface{} value to its string representation.
// --- (InterfaceToString remains unchanged) ---
func InterfaceToString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		s := fmt.Sprintf("%g", v)
		return s
	case bool:
		if v {
			return "true"
		} else {
			return "false"
		}
	case []interface{}, map[string]interface{}:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// normalizeNewlines converts all newline variations (\r\n, \r) to a single \n.
// --- (normalizeNewlines remains unchanged) ---
func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// NodeToString converts an AST node (or any interface{}) to a string representation.
// --- (NodeToString remains unchanged) ---
func NodeToString(node interface{}) string {
	if node == nil {
		return "<nil>"
	}
	var str string
	if stringer, ok := node.(fmt.Stringer); ok {
		str = stringer.String()
	} else {
		str = fmt.Sprintf("%#v", node)
	}
	maxLen := 50
	if len(str) > maxLen {
		str = str[:maxLen-3] + "..."
	}
	return str
}
