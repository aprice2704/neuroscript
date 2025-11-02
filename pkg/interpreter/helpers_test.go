// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Fixed call to register provider using provider.NewAdmin.
// filename: pkg/interpreter/helpers_test.go
// nlines: 68

package interpreter_test

import (
	"context"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
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

// setupAskTest configures a TestHarness with a registered mock provider and
// agent for use in 'ask' statement tests. It relies on the default privileged
// policy from NewTestHarness.
func setupAskTest(t *testing.T) (*TestHarness, *mockAskProviderV3) {
	t.Helper()
	h := NewTestHarness(t)

	// The local policy override has been removed.

	mockProv := &mockAskProviderV3{}
	// --- FIX: Use provider.NewAdmin to register the provider ---
	if err := provider.NewAdmin(h.ProviderRegistry, h.Interpreter.GetExecPolicy()).Register("mock_ask_provider", mockProv); err != nil {
		t.Fatalf("Failed to register mock provider: %v", err)
	}
	// --- End Fix ---

	agentConfig := map[string]any{
		"provider": "mock_ask_provider",
		"model":    "test-model",
	}
	if err := h.Interpreter.AgentModelsAdmin().Register("test_agent", agentConfig); err != nil {
		t.Fatalf("Failed to register agent model: %v", err)
	}

	return h, mockProv
}
