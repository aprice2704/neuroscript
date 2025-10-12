// NeuroScript Version: 0.8.0
// File version: 28
// Purpose: Adds the missing SetLevel method to the mockLogger to conform to the Logger interface.
// filename: pkg/api/interpreter_test.go
// nlines: 237
// risk_rating: LOW

package api_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
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
func (m *mockLogger) Info(msg string, args ...any)       {}
func (m *mockLogger) Debug(msg string, args ...any)      {}
func (m *mockLogger) Warn(msg string, args ...any)       {}
func (m *mockLogger) Errorf(format string, args ...any)  { m.Error(format) }
func (m *mockLogger) Debugf(format string, args ...any)  {}
func (m *mockLogger) Infof(format string, args ...any)   {}
func (m *mockLogger) Warnf(format string, args ...any)   {}
func (m *mockLogger) SetLevel(level interfaces.LogLevel) {} // FIX: Added missing method.

// newTestHostContext creates a minimal HostContext for testing.
func newTestHostContext(logger api.Logger) *api.HostContext {
	if logger == nil {
		logger = logging.NewNoOpLogger()
	}
	hc, err := api.NewHostContextBuilder().
		WithLogger(logger).
		WithStdout(io.Discard).
		WithStdin(os.Stdin).
		WithStderr(io.Discard).
		Build()
	if err != nil {
		panic("failed to build test host context: " + err.Error())
	}
	return hc
}

func TestInterpreter_RunNonExistentProcedure(t *testing.T) {
	src := "func do_work() means\n  return\nendfunc"
	tree, err := api.Parse([]byte(src), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Setup failed: api.Parse returned an error: %v", err)
	}

	interp := api.New(api.WithHostContext(newTestHostContext(nil)))
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
	hc, err := api.NewHostContextBuilder().
		WithLogger(logger).
		WithStdout(io.Discard).
		WithStdin(os.Stdin).
		WithStderr(io.Discard).
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}

	invalidTool := api.ToolImplementation{
		Spec: api.ToolSpec{Name: ".."},
	}
	_ = api.New(api.WithHostContext(hc), api.WithTool(invalidTool))

	logger.mu.Lock()
	defer logger.mu.Unlock()
	logOutput := logger.output.String()

	if !strings.Contains(logOutput, "failed to register tool") {
		t.Errorf("Expected logger to capture tool registration failure, but log was: %q", logOutput)
	}
}

func TestInterpreter_HasEmitFunc(t *testing.T) {
	t.Run("without emit func", func(t *testing.T) {
		interp := api.New(api.WithHostContext(newTestHostContext(nil)))
		if interp.HasEmitFunc() {
			t.Error("Expected HasEmitFunc to be false, but it was true")
		}
	})

	t.Run("with emit func", func(t *testing.T) {
		hc, err := api.NewHostContextBuilder().
			WithLogger(logging.NewNoOpLogger()).
			WithStdout(io.Discard).
			WithStdin(os.Stdin).
			WithStderr(io.Discard).
			WithEmitFunc(func(v api.Value) {}).
			Build()
		if err != nil {
			t.Fatalf("Failed to build host context: %v", err)
		}
		interp := api.New(api.WithHostContext(hc))
		if !interp.HasEmitFunc() {
			t.Error("Expected HasEmitFunc to be true, but it was false")
		}
	})
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

	globals := map[string]any{"agent_id": "agent-007"}
	interp := api.New(
		api.WithHostContext(newTestHostContext(nil)),
		api.WithGlobals(globals),
	)

	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("api.ExecWithInterpreter failed: %v", err)
	}

	result, err := api.RunProcedure(context.Background(), interp, "get_global_agent_id")
	if err != nil {
		t.Fatalf("api.RunProcedure failed: %v", err)
	}

	unwrapped, _ := api.Unwrap(result)
	if val, ok := unwrapped.(string); !ok || val != "agent-007" {
		t.Errorf("Expected result 'agent-007', got %v (type %T)", unwrapped, unwrapped)
	}
}

// TestInterpreter_StatePersistence_AccountRegistration verifies that state
// created by tools (e.g., accounts) persists on the same interpreter instance.
func TestInterpreter_StatePersistence_AccountRegistration(t *testing.T) {
	src := `
func register_acct(needs name, config) means
    must tool.account.Register(name, config)
endfunc

func check_acct(needs name returns bool) means
    return tool.account.Exists(name)
endfunc
`
	tree, err := api.Parse([]byte(src), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}

	allowedTools := []string{"tool.account.Register", "tool.account.Exists"}
	requiredGrants := []api.Capability{
		api.NewWithVerbs("account", []string{api.VerbAdmin, api.VerbRead}, []string{"*"}),
	}
	// Note: NewConfigInterpreter implicitly creates a HostContext.
	interp := api.NewConfigInterpreter(allowedTools, requiredGrants)

	if _, err := api.ExecWithInterpreter(context.Background(), interp, tree); err != nil {
		t.Fatalf("ExecWithInterpreter failed: %v", err)
	}

	accountConfig := map[string]any{
		"kind":     "llm",
		"provider": "test-provider",
		"api_key":  "test-key",
	}
	_, err = api.RunProcedure(context.Background(), interp, "register_acct", "test-user", accountConfig)
	if err != nil {
		t.Fatalf("Run(register_acct) failed: %v", err)
	}

	result, err := api.RunProcedure(context.Background(), interp, "check_acct", "test-user")
	if err != nil {
		t.Fatalf("Run(check_acct) failed: %v", err)
	}
	unwrapped, _ := api.Unwrap(result)
	if val, ok := unwrapped.(bool); !ok || !val {
		t.Error("Expected account 'test-user' to exist, but it did not")
	}
}

// TestInterpreter_StateIsolation verifies that two separate interpreter
// instances have completely independent state.
func TestInterpreter_StateIsolation(t *testing.T) {
	src := `
func register_acct(needs name, config) means
    must tool.account.Register(name, config)
endfunc
func check_acct(needs name returns bool) means
    return tool.account.Exists(name)
endfunc
`
	tree, err := api.Parse([]byte(src), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}

	allowedTools := []string{"tool.account.Register", "tool.account.Exists"}
	requiredGrants := []api.Capability{
		api.NewWithVerbs("account", []string{api.VerbAdmin, api.VerbRead}, []string{"*"}),
	}
	interp1 := api.NewConfigInterpreter(allowedTools, requiredGrants)
	interp2 := api.NewConfigInterpreter(allowedTools, requiredGrants)

	if _, err := api.ExecWithInterpreter(context.Background(), interp1, tree); err != nil {
		t.Fatalf("Exec on interp1 failed: %v", err)
	}
	if _, err := api.ExecWithInterpreter(context.Background(), interp2, tree); err != nil {
		t.Fatalf("Exec on interp2 failed: %v", err)
	}

	accountConfig := map[string]any{
		"kind":     "llm",
		"provider": "test-provider",
		"api_key":  "test-key-isolate",
	}
	_, err = api.RunProcedure(context.Background(), interp1, "register_acct", "user-on-interp1", accountConfig)
	if err != nil {
		t.Fatalf("Run on interp1 failed: %v", err)
	}

	result, err := api.RunProcedure(context.Background(), interp2, "check_acct", "user-on-interp1")
	if err != nil {
		t.Fatalf("Run on interp2 failed unexpectedly: %v", err)
	}
	unwrapped, _ := api.Unwrap(result)
	if val, ok := unwrapped.(bool); !ok || val {
		t.Error("State leak detected: account from interp1 was found on interp2")
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
		{"Native String", lang.StringValue{Value: "already native"}, "already native"},
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
