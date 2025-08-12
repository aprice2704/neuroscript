// NeuroScript Version: 0.6.0
// File version: 2.0.0
// Purpose: Corrected the test script to include line continuations for the multi-line map literal, fixing the parser error.
// filename: pkg/interpreter/ask/agentmodel_e2e_test.go
// nlines: 150
// risk_rating: MEDIUM

package ask

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/tool/agentmodel"
)

// --- Mock E2E AI Provider ---

type mockE2EProvider struct {
	t                *testing.T
	ExpectedAPIKey   string
	WasCalled        bool
	ResponseToReturn *provider.AIResponse
	ErrorToReturn    error
}

func (m *mockE2EProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	m.t.Helper()
	m.WasCalled = true

	if req.APIKey != m.ExpectedAPIKey {
		err := fmt.Errorf("mock provider received wrong API key. got: '%s', want: '%s'", req.APIKey, m.ExpectedAPIKey)
		m.t.Error(err)
		return nil, err
	}

	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	if m.ResponseToReturn == nil {
		return &provider.AIResponse{TextContent: "mock e2e success"}, nil
	}
	return m.ResponseToReturn, nil
}

// --- Test ---

func TestAgentModelE2E(t *testing.T) {
	// 1. Define the script that uses the tools to configure the system.
	const script = `
:: name: E2E AgentModel Registration and Use
:: version: 1.0

func _SetupMockAgent() means
    :: description: Registers the mock agent using the tool.
    set config = {\
        "provider": "mock_e2e_provider",\
        "model": "e2e_model",\
        "api_key_ref": "MOCK_API_KEY_ENV_VAR"\
    }
    must tool.agentmodel.Register("mock_e2e_agent", config)
endfunc

func TestTheAsk(returns result) means
    :: description: Uses the configured agent via the 'ask' statement.
    ask "mock_e2e_agent", "Does the API key work?" into result
    return result
endfunc
`
	const mockAPIKey = "secret-key-for-e2e-test"
	const mockEnvVar = "MOCK_API_KEY_ENV_VAR"

	// 2. Set up the environment (mock API key)
	t.Setenv(mockEnvVar, mockAPIKey)
	defer os.Unsetenv(mockEnvVar)

	// 3. Setup Interpreter with tools and mock provider
	interp := interpreter.NewInterpreter(interpreter.WithoutStandardTools(), interpreter.WithLogger(logging.NewTestLogger(t)))
	mockProv := &mockE2EProvider{t: t, ExpectedAPIKey: mockAPIKey}

	// Register the agentmodel tools so the script can use them
	regFunc := tool.CreateRegistrationFunc("agentmodel", agentmodel.AgentModelToolsToRegister)
	if err := regFunc(interp.ToolRegistry()); err != nil {
		t.Fatalf("Failed to register agentmodel toolset: %v", err)
	}

	// Register the mock provider
	interp.RegisterProvider("mock_e2e_provider", mockProv)

	// 4. Load the script
	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(script)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(program); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// 5. Run the setup procedure from the script
	_, err := interp.Run("_SetupMockAgent")
	if err != nil {
		var rtErr *lang.RuntimeError
		if errors.As(err, &rtErr) {
			t.Fatalf("Agent setup procedure failed: %s (wrapped: %v)", rtErr.Message, rtErr.Unwrap())
		}
		t.Fatalf("Agent setup procedure failed with non-runtime error: %v", err)
	}

	// Verify the agent was registered
	_, exists := interp.GetAgentModel("mock_e2e_agent")
	if !exists {
		t.Fatal("AgentModel 'mock_e2e_agent' was not registered by the setup script.")
	}

	// 6. Run the main test procedure
	resultVal, err := interp.Run("TestTheAsk")
	if err != nil {
		t.Fatalf("Main test procedure 'TestTheAsk' failed: %v", err)
	}

	// 7. Assertions
	if !mockProv.WasCalled {
		t.Error("Mock AI provider's Chat method was never called.")
	}

	resultStr, _ := lang.ToString(resultVal)
	expectedResponse := "mock e2e success"
	if resultStr != expectedResponse {
		t.Errorf("Expected final result to be '%s', but got '%s'", expectedResponse, resultStr)
	}

	t.Log("Successfully completed E2E test: script registered an AgentModel and 'ask' used it correctly.")
}
