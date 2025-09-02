// NeuroScript Version: 0.7.0
// File version: 8
// Purpose: Updated to use model.AccountName instead of the removed model.SecretRef field.
// filename: pkg/llmconn/llmconn.go
// nlines: 115
// risk_rating: MEDIUM

package llmconn

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/capsule"
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

	// On the first turn of a conversation, prepend the appropriate bootstrap
	// instructions to the entire prompt by retrieving them from the capsule registry.
	if c.turnCount == 1 {
		var capsuleID string
		if c.model.Tools.ToolLoopPermitted {
			capsuleID = "capsule/bootstrap_agentic/3"
		} else {
			capsuleID = "capsule/bootstrap_oneshot/3"
		}

		cap, ok := capsule.Get(capsuleID)
		if !ok {
			// This would be a critical setup error, as capsules should be embedded.
			return nil, fmt.Errorf("bootstrap capsule not found in registry: %s", capsuleID)
		}
		prompt = cap.Content + "\n\n" + prompt
	}

	// Correctly resolve the API key from the environment variable specified in AccountName.
	apiKey := ""
	if c.model.AccountName != "" {
		apiKey = os.Getenv(c.model.AccountName)
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
