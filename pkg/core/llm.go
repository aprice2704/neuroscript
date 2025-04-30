// filename: pkg/core/llm.go
package core

import (
	"context" // Ensure context is imported
	"fmt"

	// Standard library imports needed for actual LLM client implementation (e.g., net/http, encoding/json)

	// Import the logging interface
	"github.com/aprice2704/neuroscript/pkg/logging"
	// Import genai client and option types
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option" // <<< ADDED IMPORT
)

// LLMClient interface definition is in llm_types.go

// --- Concrete LLM Client Implementation ---

// ConcreteLLMClient represents the actual implementation that talks to an LLM API.
type concreteLLMClient struct {
	apiKey      string
	apiHost     string
	logger      logging.Logger
	enabled     bool          // Tracks if real calls should be made
	modelID     string        // Added modelID field
	genaiClient *genai.Client // Store the actual client
}

// Ensure concreteLLMClient implements the LLMClient interface.
var _ LLMClient = (*concreteLLMClient)(nil)

// NewLLMClient creates a new instance of the concrete LLM client.
// It acts as a factory, returning an LLMClient interface type.
// If enabled is false, it returns the internal no-op client.
func NewLLMClient(apiKey, apiHost string, logger logging.Logger, enabled bool) LLMClient {
	if logger == nil {
		logger = &coreNoOpLogger{} // Use internal no-op logger if nil
	}
	if !enabled {
		logger.Info("LLM client created but disabled. Using NoOp behavior.")
		return NewNoOpLLMClient(logger) // Return the NoOp client
	}

	logger.Info("Creating concrete LLM client.", "host", apiHost, "enabled", enabled)
	if apiKey == "" {
		logger.Error("API Key is missing for enabled LLM client.")
		logger.Warn("API Key missing, falling back to NoOpLLMClient.")
		return NewNoOpLLMClient(logger)
	}

	// --- Initialize actual GenAI Client ---
	ctx := context.Background()
	// *** CORRECTED: Use option.WithAPIKey ***
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey)) // <<< CORRECTED
	if err != nil {
		logger.Error("Failed to initialize GenAI client", "error", err)
		logger.Warn("GenAI client init failed, falling back to NoOpLLMClient.")
		return NewNoOpLLMClient(logger)
	}
	// --- End GenAI Client Init ---

	return &concreteLLMClient{
		apiKey:      apiKey,
		apiHost:     apiHost,
		logger:      logger,
		enabled:     true,
		genaiClient: client, // Store the initialized client
	}
}

// Ask sends a request to the actual LLM API.
func (c *concreteLLMClient) Ask(ctx context.Context, turns []*ConversationTurn) (*ConversationTurn, error) {
	c.logger.Debug("ConcreteLLMClient Ask called", "turn_count", len(turns))
	if !c.enabled || c.genaiClient == nil {
		c.logger.Warn("Ask called on disabled or uninitialized concrete LLM client. Returning empty response.")
		return &ConversationTurn{Role: RoleAssistant, Content: ""}, nil
	}
	// TODO: Implement actual API call logic
	c.logger.Warn("ConcreteLLMClient Ask not fully implemented")
	return nil, fmt.Errorf("concrete LLM Ask method not implemented")
}

// AskWithTools sends a request with tools to the actual LLM API.
func (c *concreteLLMClient) AskWithTools(ctx context.Context, turns []*ConversationTurn, tools []ToolDefinition) (*ConversationTurn, []*ToolCall, error) {
	c.logger.Debug("ConcreteLLMClient AskWithTools called", "turn_count", len(turns), "tool_count", len(tools))
	if !c.enabled || c.genaiClient == nil {
		c.logger.Warn("AskWithTools called on disabled or uninitialized concrete LLM client. Returning empty response.")
		return &ConversationTurn{Role: RoleAssistant, Content: ""}, nil, nil
	}
	// TODO: Implement actual API call logic with tools
	c.logger.Warn("ConcreteLLMClient AskWithTools not fully implemented")
	return nil, nil, fmt.Errorf("concrete LLM AskWithTools method not implemented")
}

// Embed generates embeddings using the actual LLM API.
func (c *concreteLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	c.logger.Debug("ConcreteLLMClient Embed called", "text_length", len(text))
	if !c.enabled || c.genaiClient == nil {
		c.logger.Warn("Embed called on disabled or uninitialized concrete LLM client. Returning empty slice.")
		return []float32{}, nil
	}
	// TODO: Implement actual embedding API call logic
	c.logger.Warn("ConcreteLLMClient Embed not implemented")
	return nil, fmt.Errorf("concrete LLM Embed method not implemented")
}

// Client returns the underlying *genai.Client.
func (c *concreteLLMClient) Client() *genai.Client {
	return c.genaiClient
}

// --- No-Op LLM Client Implementation (Internal to Core) ---

type coreNoOpLLMClient struct {
	logger logging.Logger
}

var _ LLMClient = (*coreNoOpLLMClient)(nil)

func NewNoOpLLMClient(logger logging.Logger) LLMClient {
	if logger == nil {
		logger = &coreNoOpLogger{}
	}
	logger.Info("Creating internal coreNoOpLLMClient.")
	return &coreNoOpLLMClient{logger: logger}
}
func (c *coreNoOpLLMClient) Ask(ctx context.Context, turns []*ConversationTurn) (*ConversationTurn, error) {
	c.logger.Debug("coreNoOpLLMClient Ask called")
	return &ConversationTurn{Role: RoleAssistant, Content: ""}, nil
}
func (c *coreNoOpLLMClient) AskWithTools(ctx context.Context, turns []*ConversationTurn, tools []ToolDefinition) (*ConversationTurn, []*ToolCall, error) {
	c.logger.Debug("coreNoOpLLMClient AskWithTools called")
	return &ConversationTurn{Role: RoleAssistant, Content: ""}, nil, nil
}
func (c *coreNoOpLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	c.logger.Debug("coreNoOpLLMClient Embed called")
	return []float32{}, nil
}
func (c *coreNoOpLLMClient) Client() *genai.Client {
	return nil // No-op client doesn't have an underlying client
}
