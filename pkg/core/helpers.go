// filename: pkg/core/helpers.go
package core

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	// Keep adapter import if needed elsewhere, otherwise remove
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

// --- REMOVED NewTestInterpreter ---
// func NewTestInterpreter(...) (*Interpreter, string) { ... }

// --- REMOVED NewDefaultTestInterpreter ---
// func NewDefaultTestInterpreter(...) (*Interpreter, string) { ... }
// NewTestInterpreter creates an interpreter instance for testing.
// *** KEPT: This version remains as it's test-specific ***
func NewTestInterpreter(t *testing.T, vars map[string]interface{}, lastResult interface{}) (*Interpreter, string) {
	t.Helper()

	// Setup logger based on original file's approach (using slog directly initially)
	// Using discard handler for tests unless specific output is needed
	handlerOpts := &slog.HandlerOptions{
		Level:     slog.LevelDebug, // Use Debug for test helpers maybe? Or configurable?
		AddSource: false,           // Keep source off for typical tests
	}
	// Log to io.Discard to keep test output clean by default
	handler := slog.NewTextHandler(io.Discard, handlerOpts)
	testSlogLogger := slog.New(handler) // This is the concrete slog logger

	// Adapt the slog logger to the interfaces.Logger expected by Interpreter
	// *** Use the imported adapters package ***
	testLoggerAdapter, errLog := adapters.NewSlogAdapter(testSlogLogger)
	if errLog != nil {
		t.Fatalf("Failed to create logger adapter for test: %v", errLog)
	}

	// Create a minimal LLMClient (using the adapted logger)
	minimalLLMClient := NewLLMClient("", "", testLoggerAdapter, false)
	// Check? NewLLMClient doesn't return error, assumes defaults

	// *** Create a NEW, EMPTY registry for this test interpreter ***
	testRegistry := NewToolRegistry()
	testLoggerAdapter.Info("Attempting to register core tools into NEW registry for test...")
	// Register core tools into the *new* registry instance
	if err := RegisterCoreTools(testRegistry); err != nil {
		testLoggerAdapter.Error("FATAL: Failed to register core tools during test setup", "error", err)
		t.Fatalf("FATAL: Failed to register core tools during test setup: %v", err)
	}
	testLoggerAdapter.Info("Successfully registered core tools into new registry.")
	// *** End Registry Correction ***

	// Create interpreter with the NEW registry and other components
	// Note: Ensure Interpreter struct fields match the latest definition
	interp := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16,           // Or default from constants
		toolRegistry:    testRegistry, // Assign the NEW registry
		logger:          testLoggerAdapter,
		objectCache:     make(map[string]interface{}),
		llmClient:       minimalLLMClient,
		modelName:       "gemini-1.5-flash-latest", // Default test model
		sandboxDir:      ".",                       // Default sandbox, will be replaced below
	}

	// Setup sandbox using t.TempDir()
	sandboxDirRel := t.TempDir()
	absSandboxDir, err := filepath.Abs(sandboxDirRel)
	if err != nil {
		t.Fatalf("Failed to get absolute path for sandbox %s: %v", sandboxDirRel, err)
	}
	interp.sandboxDir = absSandboxDir // Set the actual sandbox path
	testLoggerAdapter.Info("Sandbox root set in interpreter", "path", absSandboxDir)

	// Initialize variables and last result if provided
	if vars != nil {
		for k, v := range vars {
			interp.variables[k] = v
		}
	}
	interp.lastCallResult = lastResult

	return interp, absSandboxDir
}
