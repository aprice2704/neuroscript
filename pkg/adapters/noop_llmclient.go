// filename: pkg/adapters/noop_llm_client.go
package adapters

import (
	"context"

	// Import core types needed for the LLMClient interface methods
	"github.com/aprice2704/neuroscript/pkg/core"
	// Import neurogo package for the LLMClient interface definition
	"github.com/aprice2704/neuroscript/pkg/neurogo" // CORRECTED IMPORT PATH
)

// NoOpLLMClient is an LLM client implementation that performs no actions
// and returns minimal valid responses, satisfying the neurogo.LLMClient interface.
type NoOpLLMClient struct {
	// No internal state needed for the no-op implementation.
}

// NewNoOpLLMClient creates a new instance of NoOpLLMClient.
func NewNoOpLLMClient() *NoOpLLMClient {
	return &NoOpLLMClient{}
}

// Ensure NoOpLLMClient implements neurogo.LLMClient at compile time.
// This line now correctly references the imported interface from the neurogo package.
var _ neurogo.LLMClient = (*NoOpLLMClient)(nil) // CORRECTED INTERFACE CHECK

// Ask sends a request to the LLM (no-op).
// Returns a minimal valid ConversationTurn (RoleAssistant, empty content) and nil error.
// Uses types imported from pkg/core.
func (c *NoOpLLMClient) Ask(ctx context.Context, turns []*core.ConversationTurn) (*core.ConversationTurn, error) {
	// Return a minimal valid response to prevent nil pointer issues downstream.
	return &core.ConversationTurn{
		Role:    core.RoleAssistant, // Use the defined constant for the role from core
		Content: "",
		// ToolCalls will be nil by default for this struct literal
	}, nil
}

// AskWithTools sends a request to the LLM with available tools (no-op).
// Returns a minimal valid ConversationTurn, nil tool calls slice, and nil error.
// Uses types imported from pkg/core.
func (c *NoOpLLMClient) AskWithTools(ctx context.Context, turns []*core.ConversationTurn, tools []core.ToolDefinition) (*core.ConversationTurn, []*core.ToolCall, error) {
	// Return minimal valid responses.
	return &core.ConversationTurn{
		Role:    core.RoleAssistant, // Use the defined constant for the role from core
		Content: "",
	}, nil, nil // Explicitly return nil for the []*core.ToolCall slice and nil error.
}

// Embed generates embeddings for text (no-op).
// Returns an empty float32 slice and nil error.
func (c *NoOpLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	// Return an empty (but non-nil) slice to satisfy the interface contract.
	return []float32{}, nil
}
