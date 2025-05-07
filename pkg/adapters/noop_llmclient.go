// filename: pkg/adapters/noop_llmclient.go
package adapters

import (
	"context"

	// Import core types needed AND the LLMClient interface definition
	"github.com/aprice2704/neuroscript/pkg/core"
	// Import genai ONLY for the Client() method's return signature
	"github.com/google/generative-ai-go/genai"
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

// Ask sends a request to the LLM (no-op).
// Returns a minimal valid ConversationTurn (RoleAssistant, empty content, zero TokenUsage) and nil error.
// Uses types imported from pkg/core.
func (c *NoOpLLMClient) Ask(ctx context.Context, turns []*core.ConversationTurn) (*core.ConversationTurn, error) {
	return &core.ConversationTurn{
		Role:       core.RoleAssistant, // Use the defined constant for the role from core
		Content:    "",
		TokenUsage: core.TokenUsageMetrics{}, // Initialize with zero values
		// ToolCalls will be nil by default for this struct literal
	}, nil
}

// AskWithTools sends a request to the LLM with available tools (no-op).
// Returns a minimal valid ConversationTurn, nil tool calls slice, and nil error.
// Uses types imported from pkg/core.
func (c *NoOpLLMClient) AskWithTools(ctx context.Context, turns []*core.ConversationTurn, tools []core.ToolDefinition) (*core.ConversationTurn, []*core.ToolCall, error) {
	return &core.ConversationTurn{
		Role:       core.RoleAssistant, // Use the defined constant for the role from core
		Content:    "",
		TokenUsage: core.TokenUsageMetrics{}, // Initialize with zero values
	}, nil, nil // Explicitly return nil for the []*core.ToolCall slice and nil error.
}

// Embed generates embeddings for text (no-op).
// Returns an empty float32 slice and nil error.
func (c *NoOpLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{}, nil
}

// Client returns the underlying *genai.Client, which is nil for the no-op implementation.
// This method is added to satisfy the core.LLMClient interface.
func (c *NoOpLLMClient) Client() *genai.Client {
	return nil // No underlying genai client
}
