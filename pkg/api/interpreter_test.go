// NeuroScript Version: 0.6.0
// File version: 9
// Purpose: Corrected a typo in the mockLogger's Infof method receiver.
// filename: pkg/api/interpreter_test.go
// nlines: 110
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

func TestInterpreter_WithGlobals(t *testing.T) {
	src := `
func get_global_agent_id(returns string) means
	return agent_id
endfunc
`
	tree, err := api.Parse([]byte(src), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}

	// Create an interpreter with a global variable.
	globals := map[string]any{"agent_id": "agent-007"}
	interp := api.New(api.WithGlobals(globals))

	// Load the script.
	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter failed: %v", err)
	}

	// Run the procedure that accesses the global.
	result, err := api.RunProcedure(context.Background(), interp, "get_global_agent_id")
	if err != nil {
		t.Fatalf("api.RunProcedure failed: %v", err)
	}

	// Verify the result.
	unwrapped, _ := api.Unwrap(result)
	if val, ok := unwrapped.(string); !ok || val != "agent-007" {
		t.Errorf("Expected result 'agent-007', got %v (type %T)", unwrapped, unwrapped)
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
