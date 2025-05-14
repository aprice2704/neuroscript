// NeuroScript Version: 0.3.0
// File version: 0.0.1 // Assuming this is the base version you have that should work
// Implements the core.LLMClient interface with no-operation methods.
// filename: pkg/adapters/noop_llmclient.go
// nlines: 55
// risk_rating: LOW
package adapters

import (
	"context"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/generative-ai-go/genai" // For genai.Client return type
)

// NoOpLLMClient is an LLM client implementation that performs no actions
// and returns minimal valid responses, satisfying the core.LLMClient interface.
type NoOpLLMClient struct {
	// No internal state needed for the no-op implementation.
}

// NewNoOpLLMClient creates a new instance of NoOpLLMClient.
// It does not require a logger as it performs no operations that would need logging.
func NewNoOpLLMClient() *NoOpLLMClient {
	return &NoOpLLMClient{}
}

// Ensure NoOpLLMClient implements core.LLMClient at compile time.
var _ core.LLMClient = (*NoOpLLMClient)(nil)

// Ask performs no operation.
// Returns a minimal valid ConversationTurn (RoleAssistant, empty content, zero TokenUsage) and nil error.
func (c *NoOpLLMClient) Ask(ctx context.Context, turns []*core.ConversationTurn) (*core.ConversationTurn, error) {
	return &core.ConversationTurn{
		Role:       core.RoleAssistant,       // Use the defined constant for the role from core
		Content:    "No-op Ask response.",    // Provide some minimal content
		TokenUsage: core.TokenUsageMetrics{}, // Initialize with zero values
		// ToolCalls and ToolResults will be nil by default for this struct literal
	}, nil
}

// AskWithTools performs no operation.
// Returns a minimal valid ConversationTurn, nil tool calls slice, and nil error.
func (c *NoOpLLMClient) AskWithTools(ctx context.Context, turns []*core.ConversationTurn, tools []core.ToolDefinition) (*core.ConversationTurn, []*core.ToolCall, error) {
	return &core.ConversationTurn{
		Role:       core.RoleAssistant,             // Use the defined constant for the role from core
		Content:    "No-op AskWithTools response.", // Provide some minimal content
		TokenUsage: core.TokenUsageMetrics{},       // Initialize with zero values
	}, nil, nil // Explicitly return nil for the []*core.ToolCall slice and nil error.
}

// Embed performs no operation.
// Returns an empty float32 slice and nil error.
func (c *NoOpLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{}, nil
}

// Client returns nil as there is no underlying genai.Client.
func (c *NoOpLLMClient) Client() *genai.Client {
	return nil // No underlying genai client
}
