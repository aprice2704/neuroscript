// filename: pkg/core/helpers.go
package core

import (
	"fmt"
	"path/filepath"
	"reflect"

	// "sort" // No longer needed after removing diagnostic logging
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
	// Keep interfaces import for the Logger interface definition
)

// --- Internal Test Logger ---

// TestLogger implements logging.Logger and writes output using t.Logf.
// This avoids an import cycle between core and
type TestLogger struct {
	t *testing.T
}

// Ensure TestLogger implements logging.Logger at compile time.
var _ logging.Logger = (*TestLogger)(nil)

// Debug logs a debug message using t.Logf.
func (l *TestLogger) Debug(msg string, args ...any) {
	logMsg := fmt.Sprintf("[DEBUG] "+msg, args...)
	l.t.Log(logMsg) // Use t.Log or t.Logf as appropriate
}

// Info logs an informational message using t.Logf.
func (l *TestLogger) Info(msg string, args ...any) {
	logMsg := fmt.Sprintf("[INFO] "+msg, args...)
	l.t.Log(logMsg)
}

// Warn logs a warning message using t.Logf.
func (l *TestLogger) Warn(msg string, args ...any) {
	logMsg := fmt.Sprintf("[WARN] "+msg, args...)
	l.t.Log(logMsg)
}

// Error logs an error message using t.Logf.
func (l *TestLogger) Error(msg string, args ...any) {
	logMsg := fmt.Sprintf("[ERROR] "+msg, args...)
	l.t.Log(logMsg)
}

// Debugf logs a formatted debug message using t.Logf.
func (l *TestLogger) Debugf(format string, args ...any) {
	l.t.Logf("[DEBUG] "+format, args...)
}

// Infof logs a formatted informational message using t.Logf.
func (l *TestLogger) Infof(format string, args ...any) {
	l.t.Logf("[INFO] "+format, args...)
}

// Warnf logs a formatted warning message using t.Logf.
func (l *TestLogger) Warnf(format string, args ...any) {
	l.t.Logf("[WARN] "+format, args...)
}

// Errorf logs a formatted error message using t.Logf.
func (l *TestLogger) Errorf(format string, args ...any) {
	l.t.Logf("[ERROR] "+format, args...)
}

// --- End Internal Test Logger ---

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

// ConvertToSliceOfString handles conversion for ArgTypeSliceString validation.
// Renamed from convertToSliceOfString to be exported.
// Returns the []string, true + nil error on success.
// Returns nil, false + specific error on failure.
func ConvertToSliceOfString(rawValue interface{}) ([]string, bool, error) {
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

// NewTestInterpreter creates a new interpreter instance suitable for testing.
// It uses the internal core.TestLogger, registers core tools (via NewInterpreter),
// and sets up a temporary sandbox.
func NewTestInterpreter(t *testing.T, vars map[string]interface{}, lastResult interface{}) (*Interpreter, string) {
	t.Helper()
	// Use the internal TestLogger defined above
	testLogger := &TestLogger{t: t}

	// Create a minimal LLMClient using the TestLogger
	// Note: NewLLMClient is defined in pkg/core/llm.go and takes logging.Logger
	minimalLLMClient := NewLLMClient("", "", testLogger, false) // Assuming false disables actual LLM calls
	if minimalLLMClient == nil {
		t.Log("Warning: NewLLMClient returned nil, attempting to proceed without LLMClient.")
	}

	// Create interpreter - NOTE: NewInterpreter ALREADY registers core tools
	// NewInterpreter is defined in pkg/core/interpreter_new.go and takes logging.Logger, core.LLMClient
	interp := NewInterpreter(testLogger, minimalLLMClient)
	effectiveLogger := interp.Logger() // Use the logger attached to the interpreter

	// Setup sandbox directory
	sandboxDirRel := t.TempDir()
	absSandboxDir, err := filepath.Abs(sandboxDirRel)
	if err != nil {
		t.Fatalf("Failed to get absolute path for sandbox %s: %v", sandboxDirRel, err)
	}

	interp.sandboxDir = absSandboxDir                                         // Set the sandbox path on the interpreter
	effectiveLogger.Info("Sandbox root set in interpreter: " + absSandboxDir) // Now logs to test output

	// Initialize variables if provided
	if vars != nil {
		for k, v := range vars {
			interp.variables[k] = v
		}
	}

	// Set last result if provided
	interp.lastCallResult = lastResult

	return interp, absSandboxDir
}

// NewDefaultTestInterpreter provides a convenience wrapper around NewTestInterpreter.
func NewDefaultTestInterpreter(t *testing.T) (*Interpreter, string) {
	t.Helper()
	return NewTestInterpreter(t, nil, nil)
}
