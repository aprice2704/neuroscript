// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Corrected method signatures to align with core.LLMClient interface.
// Implements the core.LLMClient interface with no-operation methods.
// filename: pkg/adapters/noop_llmclient.go
// nlines: 60 // Approximate
// risk_rating: LOW

package adapters

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/core" // Correctly import core
	"github.com/google/generative-ai-go/genai"   // For genai.Client return type
)

// NoOpLLMClient is an LLM client implementation that performs no actions
// and returns minimal valid responses, satisfying the core.LLMClient interface.
type NoOpLLMClient struct {
	// No internal state needed for the no-op implementation.
}

// NewNoOpLLMClient creates a new instance of NoOpLLMClient.
func NewNoOpLLMClient() *NoOpLLMClient {
	return &NoOpLLMClient{}
}

// Ensure NoOpLLMClient implements core.LLMClient at compile time.
var _ core.LLMClient = (*NoOpLLMClient)(nil)

// Ask implements the LLMClient interface by returning a predefined response or an error.
// It now correctly accepts []*core.ConversationTurn and returns *core.ConversationTurn.
func (c *NoOpLLMClient) Ask(ctx context.Context, turns []*core.ConversationTurn) (*core.ConversationTurn, error) {
	// For a No-Op, we can return a minimal valid response or an error indicating it's a No-Op.
	// Option 1: Return a simple "No-Op response"
	// responseTurn := &core.ConversationTurn{
	// 	Role:    core.RoleAssistant,
	// 	Content: "NoOpLLMClient: No operation performed.",
	// }
	// return responseTurn, nil

	// Option 2: Return nil and a specific error (might be more informative in tests)
	return nil, fmt.Errorf("NoOpLLMClient: Ask method called, but no operation is performed")
}

// Ensure NoOpLLMClient implements the core.LLMClient interface.
var _ core.LLMClient = (*NoOpLLMClient)(nil)

// AskWithTools performs no operation.
// Returns a minimal valid ConversationTurn, nil tool calls slice, and nil error.
func (c *NoOpLLMClient) AskWithTools(ctx context.Context, turns []*core.ConversationTurn, tools []core.ToolDefinition) (*core.ConversationTurn, []*core.ToolCall, error) {
	return &core.ConversationTurn{
		Role:       core.RoleAssistant,
		Content:    "No-op AskWithTools response.",
		TokenUsage: core.TokenUsageMetrics{},
	}, nil, nil
}

// Embed performs no operation.
// Returns an empty float32 slice and nil error.
func (c *NoOpLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{}, nil
}

// Client returns nil as there is no underlying genai.Client.
func (c *NoOpLLMClient) Client() *genai.Client {
	return nil
}
