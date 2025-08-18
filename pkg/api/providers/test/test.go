// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Implements a dummy AIProvider for testing API registration.
// filename: pkg/api/providers/test/test.go
// nlines: 28
// risk_rating: LOW

package test

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/provider"
)

// Provider implements the provider.AIProvider interface for testing purposes.
type Provider struct{}

// New creates a new instance of the test AI provider.
func New() *Provider {
	return &Provider{}
}

// Chat returns a predictable, canned response for testing purposes.
func (p *Provider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	// The response is constructed to be easily verifiable in a test.
	response := fmt.Sprintf("test_provider_ok:%s", req.Prompt)
	return &provider.AIResponse{
		TextContent: response,
	}, nil
}
