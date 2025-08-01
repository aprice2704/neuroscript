// NeuroScript Version: 0.6.0
// File version: 7
// Purpose: Adds tests for the logger option and the Unwrap function, with a complete mock logger.
// filename: pkg/api/interpreter_test.go
// nlines: 85
// risk_rating: LOW

package api_test

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// mockLogger is a simple thread-safe logger for testing.
// FIX: Added all methods required by the interfaces.Logger interface.
type mockLogger struct {
	mu     sync.Mutex
	output bytes.Buffer
}

func (m *mockLogger) Error(msg string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.output.WriteString(msg)
}
func (m *mockLogger) Info(msg string, args ...any)  {}
func (m *mockLogger) Debug(msg string, args ...any) {}
func (m *mockLogger) Warn(msg string, args ...any)  {}
func (m *mockLogger) Errorf(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.output.WriteString(format)
}
func (m *mockLogger) Debugf(format string, args ...any)  {}
func (m *mockLogger) Infof(format string, args ...any)   {}
func (m *mockLogger) Warnf(format string, args ...any)   {}
func (m *mockLogger) SetLevel(level interfaces.LogLevel) {}

func TestInterpreter_RunNonExistentProcedure(t *testing.T) {
	src := "func do_work() means\n  return\nendfunc"
	tree, err := api.Parse([]byte(src), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Setup failed: api.Parse returned an error: %v", err)
	}

	interp := api.New()
	_, err = api.ExecWithInterpreter(context.Background(), interp, tree)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	_, runErr := interp.Run("this_function_does_not_exist")
	if runErr == nil {
		t.Fatal("Expected an error when running a non-existent procedure, but got nil")
	}
}

func TestInterpreter_WithLogger(t *testing.T) {
	logger := &mockLogger{}
	invalidTool := api.ToolImplementation{
		FullName: "invalid..tool",
	}
	_ = api.New(api.WithLogger(logger), api.WithTool(invalidTool))

	logger.mu.Lock()
	defer logger.mu.Unlock()
	logOutput := logger.output.String()

	if !strings.Contains(logOutput, "failed to register tool") {
		t.Errorf("Expected logger to capture tool registration failure, but log was: %q", logOutput)
	}
}

func TestUnwrap(t *testing.T) {
	testCases := []struct {
		name     string
		input    api.Value
		expected any
	}{
		{"StringValue", lang.StringValue{Value: "hello"}, "hello"},
		{"NumberValue", lang.NumberValue{Value: 123.45}, 123.45},
		{"BoolValue (true)", lang.BoolValue{Value: true}, true},
		{"NilValue", lang.NilValue{}, nil},
		{"Native String", "already native", "already native"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unwrapped, err := api.Unwrap(tc.input)
			if err != nil {
				t.Fatalf("Unwrap failed: %v", err)
			}
			if unwrapped != tc.expected {
				t.Errorf("Expected unwrapped value to be %#v, but got %#v", tc.expected, unwrapped)
			}
		})
	}
}
