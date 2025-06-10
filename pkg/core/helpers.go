// NeuroScript Version: 0.3.1
// File version: 0.0.7
// Purpose: Improved TestLogger to correctly handle structured logging key-value pairs.
// filename: pkg/core/helpers.go
// nlines: 255
// risk_rating: LOW
package core

import (
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	// No need to import adapters for the NoOp client if core.NewLLMClient handles it
)

// --- Internal Test Logger ---
type TestLogger struct {
	t   *testing.T
	out io.Writer
}

var _ interfaces.Logger = (*TestLogger)(nil)

// NewTestLogger creates a new logger that logs using the provided *testing.T.
// This is useful for tests within the core package itself.
func NewTestLogger(t *testing.T) interfaces.Logger {
	// If -nslog is NOT set, return a no-op logger that implements the interface
	if !*TestVerbose {
		return &coreNoOpLogger{} // tiny helper you already ship
	}
	// Otherwise return the full tracing logger
	return &TestLogger{t: t, out: os.Stderr}
}

func (l TestLogger) logStructured(level string, msg string, args ...any) {
	var sb strings.Builder
	sb.WriteString(level)
	sb.WriteString(" ")
	sb.WriteString(msg)

	if len(args) > 0 {
		if len(args)%2 != 0 {
			sb.WriteString(" (Logger Warning: odd number of key-value arguments provided)")
			// Append remaining args as best-effort
			for _, arg := range args {
				sb.WriteString(fmt.Sprintf(" %v", arg))
			}
		} else {
			for i := 0; i < len(args); i += 2 {
				key, okKey := args[i].(string)
				if !okKey {
					sb.WriteString(fmt.Sprintf(" (Logger Warning: expected string key, got %T)", args[i]))
					sb.WriteString(fmt.Sprintf(" %v=%v", args[i], args[i+1]))
					continue
				}
				sb.WriteString(fmt.Sprintf(" %s=%v", key, args[i+1]))
			}
		}
	}
	l.t.Log(sb.String())
}

func (l TestLogger) Debug(msg string, args ...any) { l.logStructured("[DEBUG]", msg, args...) }
func (l TestLogger) Info(msg string, args ...any)  { l.logStructured("[INFO]", msg, args...) }
func (l TestLogger) Warn(msg string, args ...any)  { l.logStructured("[WARN]", msg, args...) }
func (l TestLogger) Error(msg string, args ...any) { l.logStructured("[ERROR]", msg, args...) }

func (l TestLogger) SetLevel(level interfaces.LogLevel) {
	// No-op for test logger, level is effectively always Debug
	// All messages are logged via t.Logf or t.Log
}
func (l TestLogger) Debugf(format string, args ...any) { l.t.Logf("[DEBUG] "+format, args...) }
func (l TestLogger) Infof(format string, args ...any)  { l.t.Logf("[INFO] "+format, args...) }
func (l TestLogger) Warnf(format string, args ...any)  { l.t.Logf("[WARN] "+format, args...) }
func (l TestLogger) Errorf(format string, args ...any) { l.t.Logf("[ERROR] "+format, args...) }

// --- End Internal Test Logger ---

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ConvertToBool implements NeuroScript truthiness specifically for validating LLM input.
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

// ConvertToInt64E attempts to convert various numeric types (and potentially strings representing numbers) to int64
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
		// Potential overflow if uint64 > max int64, but often okay in practice if values are smaller.
		// For stricter conversion, add a check: if v > math.MaxInt64 { return 0, errors.New("overflow") }
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
	switch rv := rawValue.(type) {
	case []string:
		return rv, true, nil
	case []interface{}:
		strSlice := make([]string, len(rv))
		for i, item := range rv {
			if itemStr, ok := item.(string); ok {
				strSlice[i] = itemStr
			} else if item == nil {
				// Decide if nil should be empty string or error. Empty string is often convenient.
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
	switch rv := rawValue.(type) {
	case []interface{}:
		return rv, true, nil
	case []string: // Allow []string to be converted to []interface{}
		anySlice := make([]interface{}, len(rv))
		for i, s := range rv {
			anySlice[i] = s
		}
		return anySlice, true, nil
	// Add other common slice types if needed, e.g., []int, []float64
	default:
		// Use reflect to handle any slice type more generically, if desired,
		// but direct type assertions are safer and clearer for known types.
		return nil, false, fmt.Errorf("expected a slice (list), got %T", rawValue)
	}
}

// NewTestInterpreter creates a new interpreter instance suitable for testing.
// It initializes with a NoOpLLMClient and a temporary sandbox directory.
func NewTestInterpreter(t *testing.T, vars map[string]interface{}, lastResult interface{}) (*Interpreter, string) {
	t.Helper()
	testLogger := NewTestLogger(t) // Use the new constructor

	// Use core.NewLLMClient to get a NoOp client for testing.
	noOpLLMClient, _ := NewLLMClient("", "", testLogger)

	sandboxDir := t.TempDir()

	initialVars := vars
	if initialVars == nil { // Ensure initialVars is not nil for NewInterpreter
		initialVars = make(map[string]interface{})
	}

	interp, err := NewInterpreter(testLogger, noOpLLMClient, sandboxDir, initialVars, nil)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}

	if lastResult != nil {
		interp.lastCallResult = lastResult
	}

	err = RegisterCoreTools(interp)
	if err != nil {
		t.Fatalf("Failed to register core tools for test interpreter: %v", err)
	}

	err = interp.SetSandboxDir(sandboxDir)
	if err != nil {
		t.Fatalf("Failed to set sandbox dir for test interpreter: %v", err)
	}

	return interp, sandboxDir
}

// NewDefaultTestInterpreter provides a convenience wrapper around NewTestInterpreter.
func NewDefaultTestInterpreter(t *testing.T) (*Interpreter, string) {
	t.Helper()
	return NewTestInterpreter(t, nil, nil)
}

func IsRunningInTestMode() bool {
	return flag.Lookup("test.v") != nil
}
