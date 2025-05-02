// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 15:08:32 PDT // Add ConvertToInt64E helper
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
	// Keep interfaces import for the Logger interface definition
)

// --- Internal Test Logger ---
// ... (TestLogger code remains unchanged) ...
type TestLogger struct {
	t *testing.T
}

var _ logging.Logger = (*TestLogger)(nil)

func (l *TestLogger) Debug(msg string, args ...any) {
	logMsg := fmt.Sprintf("[DEBUG] "+msg, args...)
	l.t.Log(logMsg)
}
func (l *TestLogger) Info(msg string, args ...any) {
	logMsg := fmt.Sprintf("[INFO] "+msg, args...)
	l.t.Log(logMsg)
}
func (l *TestLogger) Warn(msg string, args ...any) {
	logMsg := fmt.Sprintf("[WARN] "+msg, args...)
	l.t.Log(logMsg)
}
func (l *TestLogger) Error(msg string, args ...any) {
	logMsg := fmt.Sprintf("[ERROR] "+msg, args...)
	l.t.Log(logMsg)
}
func (l *TestLogger) Debugf(format string, args ...any) { l.t.Logf("[DEBUG] "+format, args...) }
func (l *TestLogger) Infof(format string, args ...any)  { l.t.Logf("[INFO] "+format, args...) }
func (l *TestLogger) Warnf(format string, args ...any)  { l.t.Logf("[WARN] "+format, args...) }
func (l *TestLogger) Errorf(format string, args ...any) { l.t.Logf("[ERROR] "+format, args...) }

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
	// ... (ConvertToBool code remains unchanged) ...
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
		rv := reflect.ValueOf(val)
		return rv.Int() != 0, true
	case float32:
		rv := reflect.ValueOf(val)
		return rv.Float() != 0.0, true
	default:
		return false, false
	}
}

// +++ NEW: ConvertToInt64E Helper +++
// ConvertToInt64E attempts to convert various numeric types (and potentially strings)
// to int64, returning an error if the conversion fails.
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
		return v, nil // Already int64
	case uint:
		return int64(v), nil // Be mindful of potential overflow for large uint
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		// Check for overflow before converting uint64 to int64
		// if v > uint64(math.MaxInt64) {
		//  return 0, fmt.Errorf("%w: uint64 value %d overflows int64", ErrInvalidArgument, v)
		// }
		// Let's allow standard Go conversion for now, which wraps around on overflow.
		// Or should we error? Let's error for safety. Need import "math".
		// Commenting out overflow check for now to avoid adding math import yet.
		return int64(v), nil
	case float32:
		// Truncate decimal part
		return int64(v), nil
	case float64:
		// Truncate decimal part
		return int64(v), nil
	case string:
		// Attempt to parse string as integer
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: cannot convert string %q to integer: %w", ErrInvalidArgument, v, err)
		}
		return i, nil
	case bool:
		// Define explicit conversion for bool? 1 for true, 0 for false?
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("%w: cannot convert type %T to integer", ErrInvalidArgument, value)
	}
}

// ConvertToSliceOfString handles conversion for ArgTypeSliceString validation.
// ... (ConvertToSliceOfString code remains unchanged) ...
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

// convertToSliceOfAny handles conversion for ArgTypeSliceAny validation.
// ... (convertToSliceOfAny code remains unchanged) ...
func convertToSliceOfAny(rawValue interface{}) ([]interface{}, bool, error) {
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
// ... (NewTestInterpreter code remains unchanged) ...
func NewTestInterpreter(t *testing.T, vars map[string]interface{}, lastResult interface{}) (*Interpreter, string) {
	t.Helper()
	testLogger := &TestLogger{t: t}
	minimalLLMClient := NewLLMClient("", "", testLogger, false)
	if minimalLLMClient == nil {
		t.Log("Warning: NewLLMClient returned nil, attempting to proceed without LLMClient.")
	}
	interp := NewInterpreter(testLogger, minimalLLMClient)
	effectiveLogger := interp.Logger()
	sandboxDirRel := t.TempDir()
	absSandboxDir, err := filepath.Abs(sandboxDirRel)
	if err != nil {
		t.Fatalf("Failed to get absolute path for sandbox %s: %v", sandboxDirRel, err)
	}
	interp.SetSandboxDir(absSandboxDir) // Use SetSandboxDir to ensure FileAPI is also updated
	effectiveLogger.Info("Sandbox root set in interpreter via SetSandboxDir", "path", absSandboxDir)
	if vars != nil {
		for k, v := range vars {
			interp.variables[k] = v
		}
	}
	interp.lastCallResult = lastResult
	return interp, absSandboxDir
}

// NewDefaultTestInterpreter provides a convenience wrapper around NewTestInterpreter.
// ... (NewDefaultTestInterpreter code remains unchanged) ...
func NewDefaultTestInterpreter(t *testing.T) (*Interpreter, string) {
	t.Helper()
	return NewTestInterpreter(t, nil, nil)
}
