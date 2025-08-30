// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Defines sentinel errors for the llmconn package.
// filename: pkg/llmconn/errors.go
// nlines: 15
// risk_rating: LOW

package llmconn

import "errors"

var (
	ErrModelNotSet      = errors.New("AgentModel cannot be nil")
	ErrProviderNotSet   = errors.New("AIProvider cannot be nil")
	ErrLoopNotPermitted = errors.New("agent model configuration does not permit loops")
	ErrMaxTurnsExceeded = errors.New("maximum number of turns exceeded for this loop")
)
