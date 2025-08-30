// NeuroScript Version: 0.7.0
// File version: 13
// Purpose: Enforces a strict contract where incoming prompts MUST be valid AEIOU envelopes; fails immediately on non-envelope input.
// filename: pkg/api/providers/test/test.go
// nlines: 55
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
		Actions: fmt.Sprintf("command\n  emit \"%s\"\nendcommand", sanitizedResponse),
	}
	return env.Compose()
}

// Chat expects the req.Prompt to be a valid AEIOU envelope. It parses the
// envelope to extract the orchestration content and returns a canned response.
// If the prompt is not a valid envelope, it returns an error immediately.
func (p *Provider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	log.Printf("[DEBUG] TestProvider received raw prompt: %q", req.Prompt)

	// A real provider MUST receive a valid envelope. We enforce this here.
	env, err := aeiou.RobustParse(req.Prompt)
	if err != nil {
		return nil, fmt.Errorf("test provider requires a valid AEIOU envelope prompt, but parsing failed: %w", err)
	}

	actualPrompt := env.Orchestration
	log.Printf("[DEBUG] TestProvider extracted orchestration content: %q", actualPrompt)

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
