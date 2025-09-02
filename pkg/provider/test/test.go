// NeuroScript Version: 0.7.0
// File version: 17
// Purpose: Updated the mock provider to find the LAST envelope in the prompt, correctly skipping over examples in the bootstrap text.
// filename: pkg/provider/test/test.go
// nlines: 63
// risk_rating: LOW

package test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// Provider implements the provider.AIProvider interface for testing purposes.
// It strictly expects all incoming prompts to be valid AEIOU envelopes.
type Provider struct{}

// New creates a new instance of the test AI provider.
func New() *Provider {
	return &Provider{}
}

// WrapResponseInAEIOU is a local test helper that takes a simple string and
// wraps it in a minimal, valid AEIOU envelope to be used as a mock AI response.
func WrapResponseInAEIOU(responseContent string) (string, error) {
	sanitizedResponse := strings.ReplaceAll(responseContent, "\"", "\\\"")
	env := &aeiou.Envelope{
		UserData: "{}", // V3 requires a UserData section.
		Actions:  fmt.Sprintf("command\n  emit \"%s\"\nendcommand", sanitizedResponse),
	}
	return env.Compose()
}

// Chat expects the req.Prompt to be a valid AEIOU envelope. It parses the
// envelope to extract the user data and returns a canned response.
// If the prompt is not a valid envelope, it returns an error immediately.
func (p *Provider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	log.Printf("[DEBUG] TestProvider received raw prompt: %q", req.Prompt)

	// ** THE FIX IS HERE **
	// The prompt may contain bootstrap text with example envelopes. The mock provider
	// must find the LAST occurrence of the start marker to parse the real envelope.
	promptToParse := req.Prompt
	if markerPos := strings.LastIndex(req.Prompt, aeiou.Wrap(aeiou.SectionStart)); markerPos != -1 {
		promptToParse = req.Prompt[markerPos:]
	}

	env, _, err := aeiou.Parse(strings.NewReader(promptToParse))
	if err != nil {
		return nil, fmt.Errorf("test provider requires a valid AEIOU envelope prompt, but parsing failed: %w", err)
	}

	actualPrompt := env.UserData
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
