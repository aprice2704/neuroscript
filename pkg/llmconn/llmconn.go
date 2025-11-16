// NeuroScript Version: 0.7.2
// File version: 16
// Purpose: Updates imports to use the neutral Emitter interface from the new ns_interfaces package.
// Latest change: Replaced capsule.DefaultRegistry() with capsule.DefaultStore().
// filename: pkg/llmconn/llmconn.go

package llmconn

import (
	"context"
	"fmt"
	"time"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/google/uuid"
)

// LLMConn represents a stateful connection to a specific AI model endpoint.
type LLMConn struct {
	model    *types.AgentModel
	provider provider.AIProvider
	emitter  interfaces.Emitter // Bus-agnostic event emitter from neutral package

	// State for the current loop
	turnCount    int
	totalTokens  int
	totalCost    float64
	lastActivity time.Time
}

// New creates and initializes a new LLMConn.
// If a non-nil emitter is provided, it will be used to send telemetry.
func New(model *types.AgentModel, provider provider.AIProvider, emitter interfaces.Emitter) (*LLMConn, error) {
	if model == nil {
		return nil, ErrModelNotSet
	}
	if provider == nil {
		return nil, ErrProviderNotSet
	}

	return &LLMConn{
		model:        model,
		provider:     provider,
		emitter:      emitter, // Store the emitter
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

	if c.turnCount == 1 {
		var capsuleName string
		if c.model.Tools.ToolLoopPermitted {
			capsuleName = "capsule/bootstrap_agentic"
		} else {
			capsuleName = "capsule/bootstrap_oneshot"
		}

		// --- THE FIX ---
		reg := capsule.DefaultStore()
		// --- END FIX ---
		cap, ok := reg.GetLatest(capsuleName)
		if !ok {
			return nil, fmt.Errorf("latest bootstrap capsule not found in default store: %s", capsuleName)
		}
		prompt = cap.Content + "\n\n" + prompt
	}

	req := provider.AIRequest{
		AgentModelName: string(c.model.Name),
		ProviderName:   c.model.Provider,
		ModelName:      c.model.Model,
		BaseURL:        c.model.BaseURL,
		APIKey:         c.model.APIKey,
		Prompt:         prompt,
		Temperature:    c.model.Generation.Temperature,
		Stream:         false,
		Timeout:        30 * time.Second,
	}

	// --- Event Emission Logic ---
	callID := uuid.NewString()
	start := time.Now()
	if c.emitter != nil {
		c.emitter.EmitLLMCallStarted(interfaces.LLMCallStartInfo{Ctx: ctx, CallID: callID, Request: req, Start: start})
	}

	resp, err := c.provider.Chat(ctx, req)
	latency := time.Since(start)

	if err != nil {
		if c.emitter != nil {
			c.emitter.EmitLLMCallFailed(interfaces.LLMCallFailureInfo{Ctx: ctx, CallID: callID, Request: req, Err: err, Latency: latency})
		}
		return nil, fmt.Errorf("provider chat failed on turn %d: %w", c.turnCount, err)
	}

	if c.emitter != nil {
		c.emitter.EmitLLMCallSucceeded(interfaces.LLMCallSuccessInfo{Ctx: ctx, CallID: callID, Request: req, Response: *resp, Latency: latency})
	}
	// --- End Event Emission Logic ---

	c.totalTokens += resp.InputTokens + resp.OutputTokens
	c.totalCost += resp.Cost

	return resp, nil
}

// TurnCount returns the number of turns completed in the current conversation.
func (c *LLMConn) TurnCount() int {
	return c.turnCount
}

// TotalTokens returns the cumulative number of tokens used in the conversation.
func (c *LLMConn) TotalTokens() int {
	return c.totalTokens
}

// TotalCost returns the cumulative cost of the conversation.
func (c *LLMConn) TotalCost() float64 {
	return c.totalCost
}
