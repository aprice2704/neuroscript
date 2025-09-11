// NeuroScript Version: 0.7.0
// File version: 18
// Purpose: Updated the mock provider to be AEIOU v3 compliant, correctly parsing the JSON object from USERDATA and composing a valid V3 response envelope.
// filename: pkg/provider/test/test.go
// nlines: 75
// risk_rating: LOW

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// Provider implements the provider.AIProvider interface for testing purposes.
type Provider struct{}

// New creates a new instance of the test AI provider.
func New() *Provider {
	return &Provider{}
}

// WrapResponseInAEIOU is a local test helper that takes a simple string and
// wraps it in a minimal, valid AEIOU envelope to be used as a mock AI response.
func WrapResponseInAEIOU(responseContent string) (string, error) {
	sanitizedResponse := strings.ReplaceAll(responseContent, "\"", "\\\"")
	// The response must be a valid V3 envelope with a control token.
	fakeDoneToken := `emit tool.aeiou.magic("LOOP", {"action":"done"})`
	actions := fmt.Sprintf("command\n  emit \"%s\"\n  %s\nendcommand", sanitizedResponse, fakeDoneToken)

	env := &aeiou.Envelope{
		UserData: "{}", // Must not be empty for a valid envelope.
		Actions:  actions,
	}
	return env.Compose()
}

// Chat expects the req.Prompt to be a valid AEIOU envelope. It parses the
// envelope to extract the user data and returns a canned response.
func (p *Provider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	log.Printf("[DEBUG] TestProvider received raw prompt: %q", req.Prompt)

	promptToParse := req.Prompt
	if markerPos := strings.LastIndex(req.Prompt, aeiou.Wrap(aeiou.SectionStart)); markerPos != -1 {
		promptToParse = req.Prompt[markerPos:]
	}

	env, _, err := aeiou.Parse(strings.NewReader(promptToParse))
	if err != nil {
		return nil, fmt.Errorf("test provider requires a valid AEIOU envelope prompt, but parsing failed: %w", err)
	}

	// V3 ask statements send USERDATA as a JSON object like: {"subject":"ask","fields":{"prompt":"..."}}
	var userData struct {
		Fields struct {
			Prompt string `json:"prompt"`
		} `json:"fields"`
	}
	var actualPrompt string
	if err := json.Unmarshal([]byte(env.UserData), &userData); err == nil {
		actualPrompt = userData.Fields.Prompt
	} else {
		// Fallback for older tests that send a raw string.
		actualPrompt = env.UserData
	}

	log.Printf("[DEBUG] TestProvider extracted user data: %q", actualPrompt)

	var responseContent string
	if strings.Contains(actualPrompt, "What is a large language model?") {
		responseContent = "A large language model is a neural network."
	} else if strings.Contains(actualPrompt, "ping") {
		responseContent = "test_provider_ok:pong"
	} else {
		responseContent = "test_provider_ok:unknown_prompt"
	}

	response, err := WrapResponseInAEIOU(responseContent)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap test response: %w", err)
	}

	return &provider.AIResponse{
		TextContent: response,
	}, nil
}
