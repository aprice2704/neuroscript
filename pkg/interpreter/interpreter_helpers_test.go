// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Updated the test helper to create interpreters with 'policy.ContextConfig', granting the necessary permissions for agent registration tools to run.
// filename: pkg/interpreter/interpreter_test_helpers.go
// nlines: 55
// risk_rating: LOW

package interpreter_test

import (
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// mockAskProviderV3 is a shared mock provider for V3 tests.
type mockAskProviderV3 struct {
	ResponseToReturn *provider.AIResponse
	ErrorToReturn    error
}

func (m *mockAskProviderV3) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	if m.ResponseToReturn != nil {
		return m.ResponseToReturn, nil
	}
	// Default valid V3 response
	actions := `
	command
		emit "default mock response"
		set p = {"action": "done"}
		emit tool.aeiou.magic("LOOP", p)
	endcommand
	`
	env := &aeiou.Envelope{UserData: "{}", Actions: actions}
	respText, _ := env.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

// setupAskTestV3 creates a new interpreter instance with a permissive policy
// and a registered mock provider and agent for use in 'ask' statement tests.
func setupAskTestV3(t *testing.T) (*interpreter.Interpreter, *mockAskProviderV3) {
	t.Helper()

	permissivePolicy := &policy.ExecPolicy{
		Context: policy.ContextConfig, // FIX: Use config context for trusted operations
		Allow:   []string{"*"},
	}
	interp := interpreter.NewInterpreter(
		interpreter.WithLogger(logging.NewTestLogger(t)),
		interpreter.WithExecPolicy(permissivePolicy),
	)

	mockProv := &mockAskProviderV3{}
	interp.RegisterProvider("mock_ask_provider", mockProv)

	agentConfig := map[string]any{
		"provider": "mock_ask_provider",
		"model":    "test-model",
	}
	if err := interp.AgentModelsAdmin().Register("test_agent", agentConfig); err != nil {
		t.Fatalf("Failed to register agent model: %v", err)
	}

	return interp, mockProv
}
