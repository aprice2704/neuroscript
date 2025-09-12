// NeuroScript Version: 0.7.0
// File version: 20
// Purpose: Corrected the test provider to be a functional AEIOU v3 mock. It now correctly parses the USERDATA section from the envelope to return canned success responses, rather than incorrectly returning a hardcoded error. This fixes the contradictions and restores the provider's utility for testing success paths.
// filename: pkg/provider/test/test.go
// nlines: 88
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
		// Per V3 spec, UserData and Actions are required.
		UserData: `{"subject":"test-response"}`,
		Actions:  actions,
	}
	return env.Compose()
}

// Chat expects the req.Prompt to be a valid AEIOU envelope. It finds the last
// envelope in the prompt, parses it to extract the user data, and returns a
// canned response based on the content. This allows it to function as a
// predictable mock for testing success paths.
func (p *Provider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	log.Printf("[DEBUG] TestProvider received raw prompt: %q", req.Prompt)

	// The provider contract requires a valid AEIOU envelope. Find the last one.
	promptToParse := req.Prompt
	if markerPos := strings.LastIndex(req.Prompt, aeiou.Wrap(aeiou.SectionStart)); markerPos != -1 {
		promptToParse = req.Prompt[markerPos:]
	}

	env, _, err := aeiou.Parse(strings.NewReader(promptToParse))
	if err != nil {
		return nil, fmt.Errorf("test provider requires a valid AEIOU envelope, but parsing failed: %w", err)
	}

	// Simulate canned responses based on UserData content
	var responseText string
	switch {
	case strings.Contains(env.UserData, "ping"):
		responseText = "test_provider_ok:pong"
	case strings.Contains(env.UserData, "large language model"):
		responseText = "A large language model is a neural network with many parameters."
	default:
		responseText = "test_provider_ok:default_response"
	}

	// The AI's response must itself be a valid envelope.
	finalResponse, err := WrapResponseInAEIOU(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap test provider response: %w", err)
	}

	return &provider.AIResponse{
		TextContent: finalResponse,
	}, nil
}
