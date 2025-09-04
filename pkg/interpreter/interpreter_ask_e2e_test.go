// NeuroScript Version: 0.7.0
// File version: 21
// Purpose: Simplified the test script to use a plain string prompt, relying on the 'ask' statement's corrected internal envelope creation logic.
// filename: pkg/interpreter/interpreter_ask_e2e_test.go
// nlines: 230
// risk_rating: LOW

package interpreter_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/tool/account"
	"github.com/aprice2704/neuroscript/pkg/tool/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/tool/os"
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

	if m.ExpectedAPIKey != "" && req.APIKey != m.ExpectedAPIKey {
		err := fmt.Errorf("mock provider received wrong API key. got: '%s', want: '%s'", req.APIKey, m.ExpectedAPIKey)
		m.t.Error(err)
		return nil, err
	}

	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	if m.ResponseToReturn != nil {
		return m.ResponseToReturn, nil
	}

	// Default response must be a valid AEIOU envelope for the 'ask' command to parse.
	actions := `
		command
			emit "mock e2e success"
			set p = {"action":"done"}
			emit tool.aeiou.magic("LOOP", p)
		endcommand`
	env := &aeiou.Envelope{UserData: "{}", Actions: actions}
	respText, err := env.Compose()
	if err != nil {
		m.t.Fatalf("Failed to compose mock AEIOU envelope: %v", err)
	}

	return &provider.AIResponse{TextContent: respText}, nil
}

const e2eScript = `
# name: E2E AgentModel Registration and Use
# version: 1.6

func _SetupMockAgent() means
    # description: Registers the mock account and agent using tools.
    
    # 1. Register the account first.
    set key = tool.os.Getenv("MOCK_API_KEY_ENV_VAR")
    if key == nil or key == ""
        fail "MOCK_API_KEY_ENV_VAR environment variable not found by setup script"
    endif
    
    must tool.account.Register("MOCK_ACCOUNT", {\
        "kind": "llm",\
        "provider": "mock_e2e_provider",\
        "api_key": key\
    })

    # 2. Register the agent model that uses the account.
    set config = {\
        "provider": "mock_e2e_provider",\
        "model": "e2e_model",\
        "account_name": "MOCK_ACCOUNT",\
        "tool_loop_permitted": true\
    }
    must tool.agentmodel.Register("mock_e2e_agent", config)
endfunc

func TestTheAsk(returns result) means
    # description: Uses the configured agent via the 'ask' statement with a simple string prompt.
    ask "mock_e2e_agent", "Does the API key work?" into result
    return result
endfunc
`

// TestAgentModelE2E_SuccessWithPrivileges verifies the full flow works when the
// interpreter is configured with a policy that allows trusted tools to run.
func TestAgentModelE2E_SuccessWithPrivileges(t *testing.T) {
	const mockAPIKey = "secret-key-for-e2e-test"
	const mockEnvVar = "MOCK_API_KEY_ENV_VAR"
	t.Setenv(mockEnvVar, mockAPIKey)

	configPolicy := &policy.ExecPolicy{
		Context: policy.ContextConfig,
		Allow:   []string{"tool.agentmodel.*", "tool.account.*", "tool.os.Getenv", "tool.aeiou.*"},
		Grants: capability.NewGrantSet(
			[]capability.Capability{
				{Resource: "model", Verbs: []string{"admin", "use", "read"}, Scopes: []string{"*"}},
				{Resource: "account", Verbs: []string{"admin"}, Scopes: []string{"*"}},
				{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"*"}},
				{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"*"}},
			},
			capability.Limits{},
		),
	}

	interp := interpreter.NewInterpreter(
		interpreter.WithoutStandardTools(),
		interpreter.WithLogger(logging.NewTestLogger(t)),
		interpreter.WithExecPolicy(configPolicy),
	)
	mockProv := &mockE2EProvider{t: t, ExpectedAPIKey: mockAPIKey}

	// Register all necessary toolsets
	if err := tool.CreateRegistrationFunc("agentmodel", agentmodel.AgentModelToolsToRegister)(interp.ToolRegistry()); err != nil {
		t.Fatalf("Failed to register agentmodel toolset: %v", err)
	}
	if err := tool.CreateRegistrationFunc("account", account.AccountToolsToRegister)(interp.ToolRegistry()); err != nil {
		t.Fatalf("Failed to register account toolset: %v", err)
	}
	if err := tool.CreateRegistrationFunc("os", os.OsToolsToRegister)(interp.ToolRegistry()); err != nil {
		t.Fatalf("Failed to register os toolset: %v", err)
	}

	interp.RegisterProvider("mock_e2e_provider", mockProv)

	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(e2eScript)
	if pErr != nil {
		t.Fatalf("Failed to parse script: %v", pErr)
	}
	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	_, err := interp.Run("_SetupMockAgent")
	if err != nil {
		t.Fatalf("Agent setup procedure failed unexpectedly: %v", err)
	}
	_, exists := interp.AgentModels().Get("mock_e2e_agent")
	if !exists {
		t.Fatal("AgentModel 'mock_e2e_agent' was not registered by the setup script.")
	}
	_, accExists := interp.Accounts().Get("MOCK_ACCOUNT")
	if !accExists {
		t.Fatal("Account 'MOCK_ACCOUNT' was not registered by the setup script.")
	}

	resultVal, err := interp.Run("TestTheAsk")
	if err != nil {
		t.Fatalf("Main test procedure 'TestTheAsk' failed: %v", err)
	}

	if !mockProv.WasCalled {
		t.Error("Mock AI provider's Chat method was never called.")
	}
	resultStr, _ := lang.ToString(resultVal)
	expectedResponse := "mock e2e success"
	if !strings.Contains(resultStr, expectedResponse) {
		t.Errorf("Expected final result to contain '%s', but got '%s'", expectedResponse, resultStr)
	}
}
