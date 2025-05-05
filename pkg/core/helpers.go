// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 21:16:16 PDT // Update comments per user request
// filename: pkg/core/helpers.go
package core

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv" // Needed for string conversion in ConvertToInt64E
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
	// Import "github.com/aprice2704/neuroscript/pkg/interfaces" is implicitly handled by logging
)

// --- Internal Test Logger ---

// TestLogger provides a logging.Logger implementation that logs using the *testing.T logger.
type TestLogger struct {
	t *testing.T
}

var _ logging.Logger = (*TestLogger)(nil) // Verify TestLogger implements the interface

// Debug logs a debug message using the testing logger.
func (l *TestLogger) Debug(msg string, args ...any) {
	logMsg := fmt.Sprintf("[DEBUG] "+msg, args...)
	l.t.Log(logMsg)
}

// Info logs an info message using the testing logger.
func (l *TestLogger) Info(msg string, args ...any) {
	logMsg := fmt.Sprintf("[INFO] "+msg, args...)
	l.t.Log(logMsg)
}

// Warn logs a warning message using the testing logger.
func (l *TestLogger) Warn(msg string, args ...any) {
	logMsg := fmt.Sprintf("[WARN] "+msg, args...)
	l.t.Log(logMsg)
}

// Error logs an error message using the testing logger.
func (l *TestLogger) Error(msg string, args ...any) {
	logMsg := fmt.Sprintf("[ERROR] "+msg, args...)
	l.t.Log(logMsg)
}

// Error logs an error message using the testing logger.
func (l *TestLogger) SetLevel(level logging.LogLevel) {
	l.SetLevel(level)
}

// Debugf logs a formatted debug message using the testing logger.
func (l *TestLogger) Debugf(format string, args ...any) { l.t.Logf("[DEBUG] "+format, args...) }

// Infof logs a formatted info message using the testing logger.
func (l *TestLogger) Infof(format string, args ...any) { l.t.Logf("[INFO] "+format, args...) }

// Warnf logs a formatted warning message using the testing logger.
func (l *TestLogger) Warnf(format string, args ...any) { l.t.Logf("[WARN] "+format, args...) }

// Errorf logs a formatted error message using the testing logger.
func (l *TestLogger) Errorf(format string, args ...any) { l.t.Logf("[ERROR] "+format, args...) }

// --- End Internal Test Logger ---

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ConvertToBool implements NeuroScript truthiness specifically for validating LLM input.
// It checks if the input value represents true according to NeuroScript's rules.
//
// Args:
//
//	val (interface{}): The value to check for truthiness.
//
// Returns:
//
//	bool: The boolean representation (true or false).
//	bool: True if the conversion was valid (type was recognized), false otherwise.
func ConvertToBool(val interface{}) (bool, bool) {
	if val == nil {
		return false, true // nil is considered false and is a valid conversion
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
		return false, false // Invalid string representation for bool
	case int, int32:
		rv := reflect.ValueOf(val)
		return rv.Int() != 0, true
	case float32:
		rv := reflect.ValueOf(val)
		return rv.Float() != 0.0, true
	default:
		return false, false // Type not convertible to bool
	}
}

// ConvertToInt64E attempts to convert various numeric types (and potentially strings representing numbers)
// to int64, returning an error if the conversion fails or is ambiguous.
//
// Args:
//
//	value (interface{}): The value to convert to int64.
//
// Returns:
//
//	int64: The converted integer value.
//	error: An error (e.g., ErrInvalidArgument) if conversion is not possible.
func ConvertToInt64E(value interface{}) (int64, error) {
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
		return v, nil // No conversion needed
	case uint:
		// Note: Potential overflow converting large uint to int64 is currently ignored (wraps).
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		// Note: Potential overflow converting uint64 to int64 is currently ignored (wraps).
		// Consider adding overflow check using math.MaxInt64 if strictness is required.
		return int64(v), nil
	case float32:
		// Note: Conversion truncates the decimal part.
		return int64(v), nil
	case float64:
		// Note: Conversion truncates the decimal part.
		return int64(v), nil
	case string:
		// Attempt to parse string as base-10 integer
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: cannot convert string %q to integer: %w", ErrInvalidArgument, v, err)
		}
		return i, nil
	case bool:
		// Convert bool to integer (true=1, false=0)
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("%w: cannot convert type %T to integer", ErrInvalidArgument, value)
	}
}

// ConvertToSliceOfString attempts to convert an interface{} value (expected to be []string or []interface{})
// into a []string. Used primarily for validating tool arguments of type ArgTypeSliceString.
//
// Args:
//
//	rawValue (interface{}): The value to convert, typically from tool arguments.
//
// Returns:
//
//	[]string: The converted slice of strings.
//	bool: True if the conversion was successful, false otherwise.
//	error: An error if the input type is incompatible or contains non-string elements.
func ConvertToSliceOfString(rawValue interface{}) ([]string, bool, error) {
	switch rv := rawValue.(type) {
	case []string:
		return rv, true, nil // Already the correct type
	case []interface{}:
		// Handle slice received from script context
		strSlice := make([]string, len(rv))
		for i, item := range rv {
			if itemStr, ok := item.(string); ok {
				strSlice[i] = itemStr
			} else if item == nil {
				strSlice[i] = "" // Represent nil as empty string? Or error? Consistent approach needed.
			} else {
				// Element is not a string and not nil
				return nil, false, fmt.Errorf("expected slice of strings, but element %d has incompatible type %T", i, item)
			}
		}
		return strSlice, true, nil
	default:
		// Input was not a slice type we can handle
		return nil, false, fmt.Errorf("expected slice of strings, got %T", rawValue)
	}
}

// convertToSliceOfAny attempts to convert an interface{} value (expected to be []interface{} or []string)
// into a []interface{}. Used primarily for validating tool arguments of type ArgTypeSliceAny.
// Note: This function is unexported as it's primarily an internal helper for argument validation.
//
// Args:
//
//	rawValue (interface{}): The value to convert, typically from tool arguments.
//
// Returns:
//
//	[]interface{}: The converted slice of interfaces.
//	bool: True if the conversion was successful, false otherwise.
//	error: An error if the input type is not a slice we can handle.
func convertToSliceOfAny(rawValue interface{}) ([]interface{}, bool, error) {
	switch rv := rawValue.(type) {
	case []interface{}:
		return rv, true, nil // Already the correct type
	case []string:
		// Convert []string to []interface{}
		anySlice := make([]interface{}, len(rv))
		for i, s := range rv {
			anySlice[i] = s
		}
		return anySlice, true, nil
	default:
		// Input was not a slice type we can handle
		return nil, false, fmt.Errorf("expected a slice (list), got %T", rawValue)
	}
}

// NewTestInterpreter creates a new interpreter instance suitable for testing,
// initializing it with a test logger, a NoOp LLM client, and an optional set of initial variables
// and a last result. It also creates and sets a temporary sandbox directory.
//
// Args:
//
//	t (*testing.T): The testing framework's T object for logging and failure reporting.
//	vars (map[string]interface{}): An optional map of initial variable values to set in the interpreter's scope. Can be nil.
//	lastResult (interface{}): An optional initial value for the interpreter's last result register. Can be nil.
//
// Returns:
//
//	*Interpreter: The initialized interpreter instance.
//	string: The absolute path to the created temporary sandbox directory.
//
// Note: This function calls t.Fatalf on internal setup errors (like failing to create the sandbox path).
func NewTestInterpreter(t *testing.T, vars map[string]interface{}, lastResult interface{}) (*Interpreter, string) {
	t.Helper()
	testLogger := &TestLogger{t: t}
	// Initialize with a NoOp LLM client for testing core logic without API calls
	minimalLLMClient := NewLLMClient("", "", testLogger, false) // Assumes API key/endpoint are not needed for NoOp
	if minimalLLMClient == nil {
		// This case should ideally not happen if NewLLMClient handles NoOp correctly
		t.Log("Warning: NewLLMClient returned nil, attempting to proceed without LLMClient.")
	}
	interp := NewInterpreter(testLogger, minimalLLMClient)
	effectiveLogger := interp.Logger() // Get logger potentially wrapped by interpreter

	// Create and set the sandbox directory
	sandboxDirRel := t.TempDir()
	absSandboxDir, err := filepath.Abs(sandboxDirRel)
	if err != nil {
		// Fatal error during test setup
		t.Fatalf("Failed to get absolute path for sandbox %s: %v", sandboxDirRel, err)
	}
	interp.SetSandboxDir(absSandboxDir) // Use SetSandboxDir to ensure FileAPI is also updated
	effectiveLogger.Info("Sandbox root set in interpreter via SetSandboxDir", "path", absSandboxDir)

	// Set initial variables if provided
	if vars != nil {
		for k, v := range vars {
			interp.variables[k] = v
		}
	}
	// Set initial last result if provided
	interp.lastCallResult = lastResult

	return interp, absSandboxDir
}

// NewDefaultTestInterpreter provides a convenience wrapper around NewTestInterpreter,
// creating a test interpreter with no initial variables or last result.
//
// Args:
//
//	t (*testing.T): The testing framework's T object.
//
// Returns:
//
//	*Interpreter: The initialized interpreter instance.
//	string: The absolute path to the created temporary sandbox directory.
func NewDefaultTestInterpreter(t *testing.T) (*Interpreter, string) {
	t.Helper()
	// Call the main test interpreter constructor with nil for vars and lastResult
	return NewTestInterpreter(t, nil, nil)
}
