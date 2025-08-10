// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Defines the interfaces and data structures for AI model providers.
// filename: pkg/provider/provider.go
// nlines: 37
// risk_rating: HIGH

package provider

import (
	"context"
	"time"
)

// AIRequest encapsulates all parameters for a request to an AI model.
// It's constructed by the interpreter from the 'ask' statement's arguments
// and the registered AgentModel's configuration.
type AIRequest struct {
	AgentModelName string
	ProviderName   string
	ModelName      string
	BaseURL        string
	APIKey         string
	Prompt         string
	Temperature    float64
	Stream         bool
	Timeout        time.Duration
	// Add other common provider options here (e.g., JSON mode, schema) [cite: 5]
}

// AIResponse encapsulates the response from an AI model provider. [cite: 5]
type AIResponse struct {
	TextContent  string
	InputTokens  int
	OutputTokens int
	Cost         float64
	// Add other response fields here (e.g., raw response, finish reason)
}

// AIProvider is the interface for different AI model providers (e.g., OpenAI, Ollama).
// The host application is responsible for registering concrete implementations with the
// interpreter at startup. [cite: 6]
type AIProvider interface {
	// Chat sends a non-streaming request to the provider. [cite: 5]
	Chat(ctx context.Context, req AIRequest) (*AIResponse, error)
	// ChatStream sends a streaming request to the provider. (Future implementation)
	// ChatStream(ctx context.Context, req AIRequest) (<-chan StreamChunk, error) [cite: 5]
}
