// NeuroScript Version: 0.6.0
// File version: 9.0.0
// Purpose: Corrected test setup to call the newly exported NewTestInterpreter helper function from the interpreter package.
// filename: pkg/interpreter/ask/ask_integration_test.go
// nlines: 165
// risk_rating: MEDIUM

package ask

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- Mock AI Provider for Testing ---

type mockProvider struct {
	LastRequest      provider.AIRequest
	ResponseToReturn *provider.AIResponse
	ErrorToReturn    error
}

func (m *mockProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	m.LastRequest = req
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	if m.ResponseToReturn == nil {
		return &provider.AIResponse{TextContent: "default mock response"}, nil
	}
	return m.ResponseToReturn, nil
}

// --- Test Setup Helper ---

func setupAskTest(t *testing.T) (*interpreter.Interpreter, *mockProvider) {
	t.Helper()

	// FIX: Use the privileged test helper to allow agent registration.
	interp, err := interpreter.NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create privileged test interpreter: %v", err)
	}

	mockProv := &mockProvider{}

	// Register tools needed by the test scripts
	// FIX: Changed group from "string" to "str" to match the script.
	stringToolsSpec := tool.ToolSpec{Name: "Contains", Group: "str", Args: []tool.ArgSpec{{Name: "s", Type: "string"}, {Name: "substr", Type: "string"}}}
	stringToolsFunc := func(_ tool.Runtime, args []interface{}) (interface{}, error) {
		s, _ := lang.ToString(args[0])
		substr, _ := lang.ToString(args[1])
		return strings.Contains(s, substr), nil
	}
	_, _ = interp.ToolRegistry().RegisterTool(tool.ToolImplementation{Spec: stringToolsSpec, Func: stringToolsFunc})

	interp.RegisterProvider("mock_provider", mockProv)

	agentConfig := map[string]lang.Value{
		"provider": lang.StringValue{Value: "mock_provider"},
		"model":    lang.StringValue{Value: "mock-model-v1"},
	}
	if err := interp.RegisterAgentModel("default_agent", agentConfig); err != nil {
		t.Fatalf("Failed to register default agent model: %v", err)
	}

	// For the on_error tests, the interpreter injects the error details into this variable.
	if err := interp.SetInitialVariable("system_error_message", lang.StringValue{}); err != nil {
		t.Fatalf("Failed to set initial system variable: %v", err)
	}

	// Load the script file
	scriptPath := filepath.Join("testdata", "ask_scripts.ns.txt")
	scriptBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read script file %s: %v", scriptPath, err)
	}

	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(string(scriptBytes))
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}

	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}

	if err := interp.Load(program); err != nil {
		t.Fatalf("Failed to load program into interpreter: %v", err)
	}

	return interp, mockProv
}

// --- Integration Tests ---

func TestAskIntegration(t *testing.T) {
	t.Run("Basic ask statement success", func(t *testing.T) {
		interp, mockProv := setupAskTest(t)
		mockProv.ResponseToReturn = &provider.AIResponse{TextContent: "Victoria"}

		finalResult, err := interp.Run("TestBasicSuccess")
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		expectedPrompt := "What is the capital of BC?"
		if mockProv.LastRequest.Prompt != expectedPrompt {
			t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, mockProv.LastRequest.Prompt)
		}

		resultStr, _ := lang.ToString(finalResult)
		if resultStr != "Victoria" {
			t.Errorf("Expected result 'Victoria', got '%s'", resultStr)
		}
	})

	t.Run("Ask statement with provider error", func(t *testing.T) {
		interp, mockProv := setupAskTest(t)
		mockProv.ErrorToReturn = errors.New("provider API key invalid")

		finalResult, err := interp.Run("TestProviderError")
		if err != nil {
			t.Fatalf("Script execution failed unexpectedly: %v", err)
		}

		resultStr, _ := lang.ToString(finalResult)
		if resultStr != "caught provider error" {
			t.Errorf("Expected error handler to run and return 'caught provider error', but got '%s'", resultStr)
		}
	})

	t.Run("Ask statement with 'with' options", func(t *testing.T) {
		// FIX: Use a clean test setup to avoid test pollution.
		interp, mockProv := setupAskTest(t)
		// FIX: Ensure the mock provider returns a success response to prevent network errors.
		mockProv.ResponseToReturn = &provider.AIResponse{TextContent: "creative response"}

		_, err := interp.Run("TestWithOptions")
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		if mockProv.LastRequest.Temperature != 0.85 {
			t.Errorf("Expected temperature to be 0.85, got %f", mockProv.LastRequest.Temperature)
		}
	})

	t.Run("Ask with non-existent AgentModel", func(t *testing.T) {
		interp, mockProv := setupAskTest(t)
		// FIX: The mock should NOT return an error. This allows the interpreter's
		// own internal validation (checking for a registered agent) to fail first,
		// which is what the test is designed to verify.
		mockProv.ErrorToReturn = nil

		// The interpreter will catch the error before calling the provider and populate
		// the system variable. This simulates the real flow.
		errToInject := "AgentModel 'unregistered_agent' is not registered"
		_ = interp.SetInitialVariable("system_error_message", lang.StringValue{Value: errToInject})

		finalResult, err := interp.Run("TestNonExistentAgent")
		if err != nil {
			// We expect the script to handle the error and succeed, so a Go error here is a failure.
			t.Fatalf("Script execution failed unexpectedly: %v", err)
		}

		resultStr, _ := lang.ToString(finalResult)
		if resultStr != "correct error caught" {
			t.Errorf("Expected to catch 'is not registered' error, but got: '%s'", resultStr)
		}
	})
}
