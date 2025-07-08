// filename: pkg/tool/internal/tools_helpers_test.go
package internal

import (
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- Mocks for testing toolExec ---

// mockLogger implements the interfaces.Logger interface.
type mockLogger struct {
	debugMessages []string
	errorMessages []string
}

// slog-style methods
func (m *mockLogger) Debug(msg string, args ...any) { m.debugMessages = append(m.debugMessages, msg) }
func (m *mockLogger) Info(msg string, args ...any)  {}
func (m *mockLogger) Warn(msg string, args ...any)  {}
func (m *mockLogger) Error(msg string, args ...any) { m.errorMessages = append(m.errorMessages, msg) }

// Printf-style methods
func (m *mockLogger) Debugf(format string, args ...any) {
	m.debugMessages = append(m.debugMessages, fmt.Sprintf(format, args...))
}
func (m *mockLogger) Infof(format string, args ...any) {}
func (m *mockLogger) Warnf(format string, args ...any) {}
func (m *mockLogger) Errorf(format string, args ...any) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(format, args...))
}

// Other methods
func (m *mockLogger) SetLevel(level interfaces.LogLevel) {}

// Ensure mockLogger satisfies the interface.
var _ interfaces.Logger = (*mockLogger)(nil)

// mockRuntime correctly implements the real tool.Runtime interface.
type mockRuntime struct {
	logger interfaces.Logger
}

func (m *mockRuntime) Println(...any)                                {}
func (m *mockRuntime) Ask(prompt string) string                      { return "" }
func (m *mockRuntime) GetVar(name string) (any, bool)                { return nil, false }
func (m *mockRuntime) SetVar(name string, val any)                   {}
func (m *mockRuntime) CallTool(name string, args []any) (any, error) { return nil, nil }
func (m *mockRuntime) GetLogger() interfaces.Logger                  { return m.logger }
func (m *mockRuntime) SandboxDir() string                            { return "/tmp/ns_test_sandbox" }
func (m *mockRuntime) ToolRegistry() tool.ToolRegistry               { return nil }
func (m *mockRuntime) LLM() interfaces.LLMClient                     { return nil }

// --- Tests ---

func TestMakeArgs(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		args := MakeArgs(nil)
		if len(args) != 1 || args[0] != nil {
			t.Errorf(`Expected [nil], got %v`, args)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		args := MakeArgs()
		if len(args) != 0 {
			t.Errorf(`Expected [], got %v`, args)
		}
	})

	t.Run("multiple args", func(t *testing.T) {
		args := MakeArgs("a", 1, true)
		expected := []interface{}{"a", 1, true}
		if !reflect.DeepEqual(args, expected) {
			t.Errorf("Expected %v, got %v", expected, args)
		}
	})
}

func TestToolExec(t *testing.T) {
	// Find the go executable to have a reliable command for testing
	goExe, err := exec.LookPath("go")
	if err != nil {
		t.Skip("Skipping TestToolExec: 'go' command not found in PATH")
	}

	t.Run("successful execution", func(t *testing.T) {
		mockRt := &mockRuntime{logger: &mockLogger{}}
		output, err := toolExec(mockRt, goExe, "version")
		if err != nil {
			t.Fatalf("Expected successful execution, but got error: %v", err)
		}
		if !strings.Contains(output, "go version") {
			t.Errorf("Expected output to contain 'go version', but got: %s", output)
		}
		// Check logs
		logger := mockRt.logger.(*mockLogger)
		if len(logger.debugMessages) != 2 {
			t.Errorf("Expected 2 debug log messages, got %d", len(logger.debugMessages))
		}
		if len(logger.errorMessages) != 0 {
			t.Errorf("Expected 0 error log messages, got %d", len(logger.errorMessages))
		}
	})

	t.Run("failed execution (non-zero exit)", func(t *testing.T) {
		mockRt := &mockRuntime{logger: &mockLogger{}}
		// 'go help unknowncommand' exits with status 1
		_, err := toolExec(mockRt, goExe, "help", "unknowncommand")
		if err == nil {
			t.Fatal("Expected an error for non-zero exit, but got nil")
		}
		if !errors.Is(err, lang.ErrInternalTool) {
			t.Errorf("Expected error to wrap lang.ErrInternalTool, but it didn't")
		}
		// Check logs
		logger := mockRt.logger.(*mockLogger)
		if len(logger.debugMessages) != 1 {
			t.Errorf("Expected 1 debug log message, got %d", len(logger.debugMessages))
		}
		if len(logger.errorMessages) != 1 {
			t.Errorf("Expected 1 error log message, got %d", len(logger.errorMessages))
		}
	})

	t.Run("security block path traversal", func(t *testing.T) {
		mockRt := &mockRuntime{} // No logger needed
		_, err := toolExec(mockRt, "../bin/go")
		if err == nil {
			t.Fatal("Expected an error for path traversal, but got nil")
		}
		if !strings.Contains(err.Error(), "blocked suspicious command path") {
			t.Errorf("Expected error message about blocked path, got: %v", err)
		}
	})

	t.Run("security block special characters", func(t *testing.T) {
		mockRt := &mockRuntime{} // No logger needed
		_, err := toolExec(mockRt, "go; ls")
		if err == nil {
			t.Fatal("Expected an error for special characters, but got nil")
		}
		if !strings.Contains(err.Error(), "blocked suspicious command path") {
			t.Errorf("Expected error message about blocked path, got: %v", err)
		}
	})

	t.Run("no command provided", func(t *testing.T) {
		mockRt := &mockRuntime{}
		_, err := toolExec(mockRt)
		if err == nil {
			t.Fatal("Expected an error when no command is provided, but got nil")
		}
		if err.Error() != "toolExec requires at least a command" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}

func TestGetStringArg(t *testing.T) {
	args := map[string]interface{}{
		"strKey":  "hello",
		"intKey":  123,
		"boolKey": true,
	}

	t.Run("successful retrieval", func(t *testing.T) {
		val, err := getStringArg(args, "strKey")
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		if val != "hello" {
			t.Errorf("Expected 'hello', but got '%s'", val)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		_, err := getStringArg(args, "missingKey")
		if err == nil {
			t.Fatal("Expected an error for missing key, but got nil")
		}
		expected := "missing required argument 'missingKey'"
		if err.Error() != expected {
			t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		_, err := getStringArg(args, "intKey")
		if err == nil {
			t.Fatal("Expected an error for wrong type, but got nil")
		}
		expected := "invalid type for argument 'intKey': expected string, got int"
		if err.Error() != expected {
			t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
		}
	})
}

func TestMakeArgMap(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		m, err := makeArgMap("key1", "value1", "key2", 100)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		expected := map[string]interface{}{"key1": "value1", "key2": 100}
		if !reflect.DeepEqual(m, expected) {
			t.Errorf("Expected map %v, got %v", expected, m)
		}
	})

	t.Run("odd number of arguments", func(t *testing.T) {
		_, err := makeArgMap("key1", "value1", "key2")
		if err == nil {
			t.Fatal("Expected an error for odd number of arguments, but got nil")
		}
		expected := "makeArgMap requires an even number of arguments (key-value pairs)"
		if err.Error() != expected {
			t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("non-string key", func(t *testing.T) {
		_, err := makeArgMap("key1", "value1", 2, "value2")
		if err == nil {
			t.Fatal("Expected an error for non-string key, but got nil")
		}
		expected := "makeArgMap requires string keys, got int at index 2"
		if err.Error() != expected {
			t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("no arguments", func(t *testing.T) {
		m, err := makeArgMap()
		if err != nil {
			t.Fatalf("Expected no error for no arguments, but got: %v", err)
		}
		if len(m) != 0 {
			t.Errorf("Expected empty map, but got %v", m)
		}
	})
}
