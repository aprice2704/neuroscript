// NeuroScript Version: 0.3.1
// File version: 0.0.3 // Fix NewInterpreter calls in test helpers.
// filename: pkg/core/helpers.go
package core

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
	// Import "github.com/aprice2704/neuroscript/pkg/interfaces" is implicitly handled by logging
)

// --- Internal Test Logger ---
// (TestLogger implementation remains unchanged)
type TestLogger struct{ t *testing.T }

var _ logging.Logger = (*TestLogger)(nil)

func (l *TestLogger) Debug(msg string, args ...any) { l.t.Logf("[DEBUG] "+msg, args...) }
func (l *TestLogger) Info(msg string, args ...any)  { l.t.Logf("[INFO] "+msg, args...) }
func (l *TestLogger) Warn(msg string, args ...any)  { l.t.Logf("[WARN] "+msg, args...) }
func (l *TestLogger) Error(msg string, args ...any) { l.t.Logf("[ERROR] "+msg, args...) }

// Error logs an error message using the testing logger.
func (l *TestLogger) SetLevel(level logging.LogLevel) {
	// No-op for test logger, level is effectively always Debug
}
func (l *TestLogger) Debugf(format string, args ...any) { l.t.Logf("[DEBUG] "+format, args...) }
func (l *TestLogger) Infof(format string, args ...any)  { l.t.Logf("[INFO] "+format, args...) }
func (l *TestLogger) Warnf(format string, args ...any)  { l.t.Logf("[WARN] "+format, args...) }
func (l *TestLogger) Errorf(format string, args ...any) { l.t.Logf("[ERROR] "+format, args...) }

// --- End Internal Test Logger ---

// min returns the smaller of two integers.
func min(a, b int) int {
	// ... (implementation unchanged) ...
	if a < b {
		return a
	}
	return b
}

// ConvertToBool implements NeuroScript truthiness specifically for validating LLM input.
func ConvertToBool(val interface{}) (bool, bool) {
	// ... (implementation unchanged) ...
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

// ConvertToInt64E attempts to convert various numeric types (and potentially strings representing numbers) to int64
func ConvertToInt64E(value interface{}) (int64, error) {
	// ... (implementation unchanged) ...
	if value == nil {
		return 0, fmt.Errorf("%w: cannot convert nil to integer", ErrInvalidArgument)
	}
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: cannot convert string %q to integer: %w", ErrInvalidArgument, v, err)
		}
		return i, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("%w: cannot convert type %T to integer", ErrInvalidArgument, value)
	}
}

// ConvertToSliceOfString attempts to convert an interface{} value into a []string.
func ConvertToSliceOfString(rawValue interface{}) ([]string, bool, error) {
	// ... (implementation unchanged) ...
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

// convertToSliceOfAny attempts to convert an interface{} value into a []interface{}.
func convertToSliceOfAny(rawValue interface{}) ([]interface{}, bool, error) {
	// ... (implementation unchanged) ...
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

// NewTestInterpreter creates a new interpreter instance suitable for testing.
// *** UPDATED to match NewInterpreter signature (logger, llmClient, sandboxDir, initialVars) (*Interpreter, error) ***
func NewTestInterpreter(t *testing.T, vars map[string]interface{}, lastResult interface{}) (*Interpreter, string) {
	t.Helper()
	testLogger := &TestLogger{t: t}
	// Use NoOpLLMClient for testing core logic without API calls
	noOpLLMClient := NewNoOpLLMClient(testLogger) // Assuming NewNoOpLLMClient exists

	// Create sandbox directory for this test
	sandboxDir := t.TempDir() // Use testing package to manage temp dir

	// Merge provided vars with lastResult if needed (for now, assume 'vars' is the complete initial set)
	initialVars := vars
	if lastResult != nil && initialVars == nil {
		initialVars = make(map[string]interface{})
		// Decide how to handle lastResult - maybe set it *after* creation?
		// NewInterpreter doesn't directly take lastResult.
	}

	// *** FIXED: Call NewInterpreter with correct args and handle error ***
	interp, err := NewInterpreter(testLogger, noOpLLMClient, sandboxDir, initialVars)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err) // Use t.Fatalf for setup errors
	}

	// Set initial last result separately if provided
	if lastResult != nil {
		interp.lastCallResult = lastResult
	}

	// Note: Tool registration happens within NewInterpreter now via init() imports.
	// We might need to ensure test files import necessary tool packages (e.g., gosemantic).

	return interp, sandboxDir // Return the created interpreter and the sandbox path
}

// NewDefaultTestInterpreter provides a convenience wrapper around NewTestInterpreter.
// *** UPDATED call to NewTestInterpreter ***
func NewDefaultTestInterpreter(t *testing.T) (*Interpreter, string) {
	t.Helper()
	// Call the main test interpreter constructor with nil for vars and lastResult
	// The updated NewTestInterpreter handles creating the interpreter correctly.
	return NewTestInterpreter(t, nil, nil)
}
