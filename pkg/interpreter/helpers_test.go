// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Corrected ExecPolicy assignment to align with the post-refactor API.
// filename: pkg/interpreter/interpreter_test_helpers.go
// nlines: 40
// risk_rating: LOW

package interpreter_test

import (
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
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

// setupAskTest configures a TestHarness with a permissive policy and a registered
// mock provider and agent for use in 'ask' statement tests.
func setupAskTest(t *testing.T) (*TestHarness, *mockAskProviderV3) {
	t.Helper()
	h := NewTestHarness(t)

	permissivePolicy := &policy.ExecPolicy{
		Context: policy.ContextConfig,
		Allow:   []string{"*"},
	}
	h.Interpreter.ExecPolicy = permissivePolicy

	mockProv := &mockAskProviderV3{}
	h.Interpreter.RegisterProvider("mock_ask_provider", mockProv)

	agentConfig := map[string]any{
		"provider": "mock_ask_provider",
		"model":    "test-model",
	}
	if err := h.Interpreter.AgentModelsAdmin().Register("test_agent", agentConfig); err != nil {
		t.Fatalf("Failed to register agent model: %v", err)
	}

	return h, mockProv
}
