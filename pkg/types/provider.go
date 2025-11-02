// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Adds ProviderParams map to pass generic config to providers.
// filename: pkg/types/provider.go
// nlines: 42
// risk_rating: LOW

package types

import (
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

	// ProviderParams is a direct copy of AgentModel.Params, used to pass
	// provider-specific config (like "generic_http") to the provider.
	// This is the FIX that addresses the "req.Model undefined" error.
	ProviderParams map[string]any
}

// AIResponse encapsulates the response from an AI model provider.
type AIResponse struct {
	TextContent  string
	InputTokens  int
	OutputTokens int
	Cost         float64
	// Add other response fields here (e.g., raw response, finish reason)
}
