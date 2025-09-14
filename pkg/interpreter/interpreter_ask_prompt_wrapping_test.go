// NeuroScript Version: 0.7.2
// File version: 1
// Purpose: Adds a failing test to verify that simple string prompts in 'ask' are correctly wrapped in JSON.
// filename: pkg/interpreter/interpreter_ask_prompt_wrapping_test.go
// nlines: 97
// risk_rating: LOW

package interpreter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
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

	// 1. Isolate and parse the envelope from the full prompt sent by the host.
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

	// 2. The core test: Check if UserData is the expected JSON, not a raw string.
	var userDataPayload map[string]interface{}
	if err := json.Unmarshal([]byte(env.UserData), &userDataPayload); err != nil {
		m.t.Errorf("Failing Test Assertion Failed: UserData was not a valid JSON object. Got raw string: %q", env.UserData)
		// Return a valid response to avoid cascading interpreter errors.
	} else {
		// 3. Deeper validation of the JSON structure.
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

	// Return a standard "done" response to allow the 'ask' statement to complete successfully.
	actions := `command
	   emit tool.aeiou.magic("LOOP", {"action":"done"}) 
	endcommand`
	respEnv := &aeiou.Envelope{UserData: `{"status":"ok"}`, Actions: actions}
	respText, _ := respEnv.Compose()
	return &provider.AIResponse{TextContent: respText}, nil
}

func TestAskPromptWrapping(t *testing.T) {
	const simplePrompt = "What is the capital of British Columbia?"

	interp, err := interpreter.NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}

	// Register the mock provider that performs the JSON validation.
	mockProv := &mockWrappingProvider{t: t, expectedPrompt: simplePrompt}
	interp.RegisterProvider("wrapping_checker", mockProv)

	// Register an agent that uses our special provider.
	_ = interp.RegisterAgentModel("test_agent", map[string]lang.Value{
		"provider": lang.StringValue{Value: "wrapping_checker"},
		"model":    lang.StringValue{Value: "test_model"},
	})

	// This script uses the simple string syntax that needs to be wrapped.
	script := fmt.Sprintf(`command ask "test_agent", "%s" endcommand`, simplePrompt)

	p := parser.NewParserAPI(nil)
	tree, _ := p.Parse(script)
	program, _, _ := parser.NewASTBuilder(nil).Build(tree)
	_, err = interp.Execute(program)

	if err != nil {
		t.Fatalf("Script execution failed unexpectedly: %v", err)
	}
}
