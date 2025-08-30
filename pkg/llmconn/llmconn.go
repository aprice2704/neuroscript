// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Implements the LLMConn, a stateful connection manager for agent loops.
// filename: pkg/llmconn/llmconn.go
// nlines: 95
// risk_rating: MEDIUM

package llmconn

import (
	"context"
	"fmt"
	"time"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// LLMConn represents a stateful connection to a specific AI model endpoint.
// It uses an AgentModel for configuration and an AIProvider for communication,
// managing the state and rules of a multi-turn conversation.
type LLMConn struct {
	model    *types.AgentModel
	provider provider.AIProvider

	// State for the current loop
	turnCount    int
	totalTokens  int
	totalCost    float64
	lastActivity time.Time
}

// NewLLMConn creates and initializes a new LLMConn.
// It verifies that the provided AgentModel is configured for agentic loops.
func New(model *types.AgentModel, provider provider.AIProvider) (*LLMConn, error) {
	if model == nil {
		return nil, ErrModelNotSet
	}
	if provider == nil {
		return nil, ErrProviderNotSet
	}
	// Per askloop_spec_v2, a loop is only permitted if the flag is set.
	if !model.Tools.ToolLoopPermitted {
		return nil, fmt.Errorf("%w: model '%s' does not permit tool loops", ErrLoopNotPermitted, model.Name)
	}

	return &LLMConn{
		model:    model,
		provider: provider,
		// Initialize state for a new loop
		turnCount:    0,
		lastActivity: time.Now(),
	}, nil
}

// Converse implements the Connector interface.
func (c *LLMConn) Converse(ctx context.Context, input *aeiou.Envelope) (*provider.AIResponse, error) {
	if c.turnCount >= c.model.MaxTurns && c.model.MaxTurns > 0 {
		return nil, fmt.Errorf("%w: exceeded max %d turns", ErrMaxTurnsExceeded, c.model.MaxTurns)
	}

	c.turnCount++
	c.lastActivity = time.Now()

	// 1. Construct the prompt from the AEIOU envelope.
	// The provider contract requires the entire envelope to be the prompt.
	prompt, err := input.Compose()
	if err != nil {
		return nil, fmt.Errorf("failed to compose input envelope for provider: %w", err)
	}

	// 2. Build the AIRequest from the AgentModel configuration.
	req := provider.AIRequest{
		AgentModelName: string(c.model.Name),
		ProviderName:   c.model.Provider,
		ModelName:      c.model.Model,
		BaseURL:        c.model.BaseURL,
		APIKey:         c.model.SecretRef, // Assuming SecretRef holds the key for now
		Prompt:         prompt,
		Temperature:    c.model.Generation.Temperature,
		Stream:         false, // Streaming not supported in this design yet
		Timeout:        30 * time.Second,
	}

	// 3. Call the provider.
	resp, err := c.provider.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("provider chat failed on turn %d: %w", c.turnCount, err)
	}

	// 4. Update state with response metadata.
	c.totalTokens += resp.InputTokens + resp.OutputTokens
	c.totalCost += resp.Cost

	return resp, nil
}
