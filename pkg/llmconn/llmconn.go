// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Corrected API key handling to resolve the secret reference from environment variables.
// filename: pkg/llmconn/llmconn.go
// nlines: 98
// risk_rating: HIGH

package llmconn

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// LLMConn represents a stateful connection to a specific AI model endpoint.
type LLMConn struct {
	model    *types.AgentModel
	provider provider.AIProvider

	// State for the current loop
	turnCount    int
	totalTokens  int
	totalCost    float64
	lastActivity time.Time
}

// New creates and initializes a new LLMConn.
func New(model *types.AgentModel, provider provider.AIProvider) (*LLMConn, error) {
	if model == nil {
		return nil, ErrModelNotSet
	}
	if provider == nil {
		return nil, ErrProviderNotSet
	}

	return &LLMConn{
		model:        model,
		provider:     provider,
		turnCount:    0,
		lastActivity: time.Now(),
	}, nil
}

// Converse implements the Connector interface.
func (c *LLMConn) Converse(ctx context.Context, input *aeiou.Envelope) (*provider.AIResponse, error) {
	if c.model.MaxTurns > 0 && c.turnCount >= c.model.MaxTurns {
		return nil, fmt.Errorf("%w: exceeded max %d turns", ErrMaxTurnsExceeded, c.model.MaxTurns)
	}

	c.turnCount++
	c.lastActivity = time.Now()

	prompt, err := input.Compose()
	if err != nil {
		return nil, fmt.Errorf("failed to compose input envelope for provider: %w", err)
	}

	// ** THE FIX IS HERE **
	// Correctly resolve the API key from the environment variable specified in SecretRef.
	apiKey := ""
	if c.model.SecretRef != "" {
		apiKey = os.Getenv(c.model.SecretRef)
	}

	req := provider.AIRequest{
		AgentModelName: string(c.model.Name),
		ProviderName:   c.model.Provider,
		ModelName:      c.model.Model,
		BaseURL:        c.model.BaseURL,
		APIKey:         apiKey,
		Prompt:         prompt,
		Temperature:    c.model.Generation.Temperature,
		Stream:         false,
		Timeout:        30 * time.Second,
	}

	resp, err := c.provider.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("provider chat failed on turn %d: %w", c.turnCount, err)
	}

	c.totalTokens += resp.InputTokens + resp.OutputTokens
	c.totalCost += resp.Cost

	return resp, nil
}
