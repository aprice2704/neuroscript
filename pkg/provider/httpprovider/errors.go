// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Defines sentinel errors for the generic HTTP provider.
// filename: pkg/provider/httpprovider/errors.go
// nlines: 15
// risk_rating: LOW

package httpprovider

import "errors"

var (
	// ErrConfigMissing is returned when the AgentModel is missing the "generic_http" block.
	ErrConfigMissing = errors.New("agentmodel missing 'generic_http' config block")
	// ErrConfigInvalid is returned when the "generic_http" block is present but invalid.
	ErrConfigInvalid = errors.New("invalid 'generic_http' config")
	// ErrInterpolation is returned when a token (e.g., {PROMPT}) cannot be processed.
	ErrInterpolation = errors.New("failed to interpolate request template")
	// ErrResponseFormat is returned when the API response cannot be parsed.
	ErrResponseFormat = errors.New("failed to parse API response")
)
