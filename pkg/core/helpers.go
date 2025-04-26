// filename: neuroscript/pkg/core/helpers.go
package core

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
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

func NewTestInterpreter(t *testing.T, vars map[string]interface{}, lastResult interface{}) (*Interpreter, string) {
	t.Helper()

	handlerOpts := &slog.HandlerOptions{
		Level:     slog.LevelWarn,
		AddSource: true, // include source file and line number
	}
	handler := slog.NewTextHandler(os.Stderr, handlerOpts) // Log to Stderr

	// Create the core slog logger
	testLogger := slog.New(handler)

	// Create a minimal LLMClient (can keep using testLogger now)
	minimalLLMClient := NewLLMClient("", "", testLogger, false)
	if minimalLLMClient == nil {
		t.Fatal("Failed to create even a minimal LLMClient for testing")
	}

	interp := NewInterpreter(testLogger, minimalLLMClient) // Pass the working logger

	testLogger.Info("Attempting to register core tools...")
	if err := RegisterCoreTools(interp.ToolRegistry()); err != nil {
		testLogger.Error("FATAL: Failed to register core tools during test setup: %v", err)
		t.Fatalf("FATAL: Failed to register core tools during test setup: %v", err)
	}
	testLogger.Info("Successfully registered core tools.")

	sandboxDirRel := t.TempDir()
	absSandboxDir, err := filepath.Abs(sandboxDirRel)
	if err != nil {
		t.Fatalf("Failed to get absolute path for sandbox %s: %v", sandboxDirRel, err)
	}

	interp.sandboxDir = absSandboxDir
	testLogger.Info("Sandbox root set in interpreter: %s", absSandboxDir)

	if vars != nil {
		for k, v := range vars {
			interp.variables[k] = v
		}
	}

	interp.lastCallResult = lastResult

	return interp, absSandboxDir
}

func NewDefaultTestInterpreter(t *testing.T) (*Interpreter, string) {
	t.Helper()
	return NewTestInterpreter(t, nil, nil)
}
