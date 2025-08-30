// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Defines the interface for a stateful LLM connection manager.
// filename: pkg/llmconn/interface.go
// nlines: 20
// risk_rating: LOW

package llmconn

import (
	"context"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// Connector is the interface for a stateful connection to an LLM,
// managing the lifecycle of a multi-turn "ask loop". It bridges the gap
// between the high-level agent configuration and the low-level provider.
type Connector interface {
	// Converse handles a single turn of the AEIOU loop. It takes the current
	// envelope from the host, constructs a request to the provider,
	// and returns the AI's response.
	Converse(ctx context.Context, input *aeiou.Envelope) (*provider.AIResponse, error)
}
