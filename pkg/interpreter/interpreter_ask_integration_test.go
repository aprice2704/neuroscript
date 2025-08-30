// NeuroScript Version: 0.7.0
// File version: 15
// Purpose: Moved test to interpreter package and added 'toolLoopPermitted' to mock agent configs to fix test failures.
// filename: pkg/interpreter/interpreter_ask_integration_test.go
// nlines: 177
// risk_rating: MEDIUM

package interpreter_test

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

type mockAskProvider struct {
	LastRequest      provider.AIRequest
	ResponseToReturn *provider.AIResponse
	ErrorToReturn    error
}

func (m *mockAskProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	m.LastRequest = req
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	if m.ResponseToReturn == nil {
		// Default to a valid, simple AEIOU 'done' response.
		return &provider.AIResponse{TextContent: `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:START>>>
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:ACTIONS>>>
command
  emit "default mock response"
  emit '<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:LOOP:{"control":"done"}>>>'
endcommand
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:ACTIONS>>>
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:END>>>`}, nil
	}
	return m.ResponseToReturn, nil
}

// --- Test Setup Helper ---

func setupAskTest(t *testing.T) (*interpreter.Interpreter, *mockAskProvider) {
	t.Helper()

	interp, err := interpreter.NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create privileged test interpreter: %v", err)
	}

	mockProv := &mockAskProvider{}
	stringToolsSpec := tool.ToolSpec{Name: "Contains", Group: "str", Args: []tool.ArgSpec{{Name: "s", Type: "string"}, {Name: "substr", Type: "string"}}}
	stringToolsFunc := func(_ tool.Runtime, args []interface{}) (interface{}, error) {
		s, _ := lang.ToString(args[0])
		substr, _ := lang.ToString(args[1])
		return strings.Contains(s, substr), nil
	}
	_, _ = interp.ToolRegistry().RegisterTool(tool.ToolImplementation{Spec: stringToolsSpec, Func: stringToolsFunc})

	interp.RegisterProvider("mock_provider", mockProv)

	agentConfig := map[string]any{
		"provider":          "mock_provider",
		"model":             "mock-model-v1",
		"toolLoopPermitted": true,
	}
	if err := interp.AgentModelsAdmin().Register("default_agent", agentConfig); err != nil {
		t.Fatalf("Failed to register default agent model: %v", err)
	}
	if err := interp.SetInitialVariable("system_error_message", lang.StringValue{}); err != nil {
		t.Fatalf("Failed to set initial system variable: %v", err)
	}
	// Assuming testdata is relative to the package being tested.
	// Adjust the path if running tests from the module root.
	scriptPath := filepath.Join("ask", "testdata", "ask_scripts.ns.txt")
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
		mockProv.ResponseToReturn = &provider.AIResponse{TextContent: `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:START>>>
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:ACTIONS>>>
command
  emit "Victoria"
  emit '<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:LOOP:{"control":"done"}>>>'
endcommand
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:ACTIONS>>>
<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:END>>>`}

		finalResult, err := interp.Run("TestBasicSuccess")
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		expectedPromptContent := `What is the capital of BC?`
		if !strings.Contains(mockProv.LastRequest.Prompt, expectedPromptContent) {
			t.Errorf("Expected prompt Orchestration to contain '%s', got '%s'", expectedPromptContent, mockProv.LastRequest.Prompt)
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
		interp, _ := setupAskTest(t)
		_, err := interp.Run("TestWithOptions")
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}
		// Note: 'with' options are not yet implemented in this test's callAIProvider.
		// This test currently only verifies that the script runs without error.
		// A future change would be needed to inspect the provider request.
	})

	t.Run("Ask with non-existent AgentModel", func(t *testing.T) {
		interp, _ := setupAskTest(t)
		finalResult, err := interp.Run("TestNonExistentAgent")
		if err != nil {
			t.Fatalf("Script execution failed unexpectedly: %v", err)
		}

		resultStr, _ := lang.ToString(finalResult)
		if resultStr != "correct error caught" {
			t.Errorf("Expected to catch 'is not registered' error, but got: '%s'", resultStr)
		}
	})
}
