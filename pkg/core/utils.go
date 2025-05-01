// filename: pkg/core/utils.go
package core

import (
	"fmt"
	"reflect"
	"strings"

	// Import the logging interface definition
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// --- Internal No-Op Logger ---
// Used as a fallback within the core package if no logger is provided,
// avoiding a dependency on the adapters package.

type coreNoOpLogger struct{}

// Ensure coreNoOpLogger implements logging.Logger at compile time.
var _ logging.Logger = (*coreNoOpLogger)(nil)

func (l *coreNoOpLogger) Debug(msg string, args ...any) {}
func (l *coreNoOpLogger) Info(msg string, args ...any)  {}
func (l *coreNoOpLogger) Warn(msg string, args ...any)  {}
func (l *coreNoOpLogger) Error(msg string, args ...any) {}

func (l *coreNoOpLogger) Debugf(format string, args ...any) {}
func (l *coreNoOpLogger) Infof(format string, args ...any)  {}
func (l *coreNoOpLogger) Warnf(format string, args ...any)  {}
func (l *coreNoOpLogger) Errorf(format string, args ...any) {}

// --- End Internal No-Op Logger ---

// IsTruthy determines the truthiness of a NeuroScript value according to language rules.
// nil, false, 0, 0.0, "" are false. Everything else is true.
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
	// Handle potential JSON number types
	case int:
		return v != 0
	case int32:
		return v != 0
	case float32:
		return v != 0.0
	default:
		// Slices, maps, etc., are generally considered "truthy" if non-nil
		// Check if it's a pointer type and if it's nil
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface || rv.Kind() == reflect.Map || rv.Kind() == reflect.Slice || rv.Kind() == reflect.Chan || rv.Kind() == reflect.Func {
			return !rv.IsNil()
		}
		// Other non-nil values are true
		return true
	}
}

// InterfaceToString attempts to convert an interface{} value to its string representation.
// Handles basic types and uses fmt.Sprintf for others.
func InterfaceToString(value interface{}) string {
	if value == nil {
		return "" // Represent nil as empty string in NeuroScript? Or "none"? Let's use empty for now.
	}
	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		// Format float without unnecessary trailing zeros
		// Use %g which automatically chooses %f or %e
		s := fmt.Sprintf("%g", v)
		// Ensure it looks like a float if it's a whole number, e.g., "5.0" not "5"
		// This might be too complex/opinionated for a simple conversion.
		// Let's stick with %g for now. If specific formatting needed, use a tool.
		return s
	case bool:
		if v {
			return "true"
		}
		return "false"
	case []interface{}:
		// Represent slice/list in a readable format (e.g., comma-separated)
		// This could become complex for nested structures. Using fmt is safer.
		return fmt.Sprintf("%v", v) // Default Go representation
	case map[string]interface{}:
		// Represent map in a readable format. Using fmt is safer.
		return fmt.Sprintf("%v", v) // Default Go representation
	default:
		// Use reflection for other types, potentially falling back to fmt
		return fmt.Sprintf("%v", value)
	}
}

// normalizeNewlines converts all newline variations (\r\n, \r) to a single \n.
func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// NodeToString converts an AST node (or any interface{}) to a string representation
// suitable for including in error messages or debug logs.
// It attempts to use the .String() method if available and truncates long output.
func NodeToString(node interface{}) string {
	if node == nil {
		return "<nil>"
	}

	var str string
	// Attempt to use String() method if available (common pattern for AST nodes)
	if stringer, ok := node.(fmt.Stringer); ok {
		str = stringer.String()
	} else {
		// Basic fallback using fmt.Sprintf with %#v for potentially more detail
		str = fmt.Sprintf("%#v", node)
	}

	// Truncate long representations for brevity in error messages
	maxLen := 50 // Adjust max length as needed
	if len(str) > maxLen {
		str = str[:maxLen-3] + "..."
	}
	return str
}
