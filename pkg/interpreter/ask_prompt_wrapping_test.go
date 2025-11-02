// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Fixed call to register provider using provider.NewAdmin.
// filename: pkg/interpreter/ask_prompt_wrapping_test.go
// nlines: 100

package interpreter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// mockWrappingProvider is designed specifically to test if the prompt
// was correctly wrapped into the standard AEIOU v3 JSON structure.
type mockWrappingProvider struct {
	t              *testing.T
	expectedPrompt string
}

// Chat inspects the incoming request, parses the envelope, and validates the UserData.
func (m *mockWrappingProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	m.t.Helper()
	m.t.Logf("[DEBUG] Turn X: mockWrappingProvider.Chat called.")

	startMarker := aeiou.Wrap(aeiou.SectionStart)
	envelopeStart := strings.LastIndex(req.Prompt, startMarker)
	if envelopeStart == -1 {
		return nil, fmt.Errorf("test integrity error: could not find AEIOU START marker in provider request")
	}
	envelopeText := req.Prompt[envelopeStart:]
	env, _, err := aeiou.Parse(strings.NewReader(envelopeText))
	if err != nil {
		return nil, fmt.Errorf("test integrity error: failed to parse incoming envelope: %w", err)
	}

	var userDataPayload map[string]interface{}
	if err := json.Unmarshal([]byte(env.UserData), &userDataPayload); err != nil {
		m.t.Errorf("Failing Test Assertion Failed: UserData was not a valid JSON object. Got raw string: %q", env.UserData)
	} else {
		if subject, _ := userDataPayload["subject"].(string); subject != "ask" {
			m.t.Errorf("Failing Test Assertion Failed: expected subject 'ask', got '%v'", subject)
		}
		fields, ok := userDataPayload["fields"].(map[string]interface{})
		if !ok {
			m.t.Errorf("Failing Test Assertion Failed: 'fields' key is missing or not a map")
		} else if prompt, _ := fields["prompt"].(string); prompt != m.expectedPrompt {
			m.t.Errorf("Failing Test Assertion Failed: expected prompt '%s' in fields, got '%s'", m.expectedPrompt, prompt)
		}
	}

	// THE FIX: The AI's job is just to emit a response.
	// The Go 'runAskHostLoop' handles loop termination.
	actions := `command
	   emit "ok"
	endcommand`
	respEnv := &aeiou.Envelope{UserData: `{"status":"ok"}`, Actions: actions}
	respText, _ := respEnv.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

func TestAskPromptWrapping(t *testing.T) {
	const simplePrompt = "What is the capital of British Columbia?"
	t.Logf("[DEBUG] Turn 1: Starting TestAskPromptWrapping.")

	h := NewTestHarness(t)
	mockProv := &mockWrappingProvider{t: t, expectedPrompt: simplePrompt}
	// --- FIX: Use provider.NewAdmin to register the provider ---
	if err := provider.NewAdmin(h.ProviderRegistry, h.Interpreter.GetExecPolicy()).Register("wrapping_checker", mockProv); err != nil {
		t.Fatalf("Failed to register mock provider: %v", err)
	}
	// --- End Fix ---

	_ = h.Interpreter.RegisterAgentModel("test_agent", map[string]lang.Value{
		"provider": lang.StringValue{Value: "wrapping_checker"},
		"model":    lang.StringValue{Value: "test_model"},
	})
	t.Logf("[DEBUG] Turn 2: Mock provider and agent registered.")

	script := fmt.Sprintf(`command ask "test_agent", "%s" endcommand`, simplePrompt)

	tree, _ := h.Parser.Parse(script)
	program, _, _ := h.ASTBuilder.Build(tree)
	if err := h.Interpreter.Load(&interfaces.Tree{Root: program}); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}
	t.Logf("[DEBUG] Turn 3: Script loaded. Executing.")
	_, err := h.Interpreter.Execute(program)

	if err != nil {
		t.Fatalf("Script execution failed unexpectedly: %v", err)
	}
	t.Logf("[DEBUG] Turn 4: Test completed.")
}
