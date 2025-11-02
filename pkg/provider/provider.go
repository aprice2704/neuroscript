// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Added type aliases for types.AIRequest/AIResponse to fix compiler errors during refactor.
// filename: pkg/provider/provider.go
// nlines: 26
// risk_rating: HIGH

package provider

import (
	"context"

	"github.com/aprice2704/neuroscript/pkg/types"
)

// --- Type Aliases for Refactoring ---
// These aliases are used to fix compiler errors in other packages
// that were importing from provider.AIRequest directly.
// New code should use types.AIRequest and types.AIResponse.

type AIRequest = types.AIRequest
type AIResponse = types.AIResponse

// --- End Type Aliases ---

// AIProvider is the interface for different AI model providers (e.g., OpenAI, Ollama).
// The host application is responsible for registering concrete implementations with the
// interpreter at startup.
type AIProvider interface {
	// Chat sends a non-streaming request to the provider.
	Chat(ctx context.Context, req types.AIRequest) (*types.AIResponse, error)
	// ChatStream sends a streaming request to the provider. (Future implementation)
	// ChatStream(ctx context.Context, req types.AIRequest) (<-chan StreamChunk, error)
}
