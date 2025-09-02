// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements the MagicTool with primary and fallback signer logic.
// filename: aeiou/tool.go
// nlines: 54
// risk_rating: HIGH

package aeiou

import "encoding/json"

// MagicTool provides a robust interface for minting control tokens, handling
// fallback to an abort token if the primary signer is unavailable.
type MagicTool struct {
	primaryMinter  *MagicMinter
	fallbackMinter *MagicMinter
}

// NewMagicTool creates a new MagicTool. The primary minter can be nil to
// simulate a failure in loading the primary key. The fallback minter should
// be initialized with a preloaded, in-memory key.
func NewMagicTool(primary *MagicMinter, fallback *MagicMinter) *MagicTool {
	return &MagicTool{
		primaryMinter:  primary,
		fallbackMinter: fallback,
	}
}

// MintMagicToken attempts to create a token with the primary minter. If the
// primary minter is unavailable, it uses the fallback minter to issue a
// pre-defined 'abort' token. If both are unavailable, it returns an error.
func (t *MagicTool) MintMagicToken(kind ControlKind, agentPayload ControlPayload, hostCtx HostContext) (string, error) {
	// Happy path: primary signer is available.
	if t.primaryMinter != nil {
		return t.primaryMinter.MintToken(kind, agentPayload, hostCtx)
	}

	// Fallback path: primary has failed, use the fallback signer.
	if t.fallbackMinter != nil {
		// As per spec, the fallback MUST issue an abort token.
		// The agent's intended payload is ignored.
		fallbackPayload := ControlPayload{
			Action:  ActionAbort,
			Request: json.RawMessage(`{"reason":"fallback_signer_activated"}`),
		}
		// The host context's KeyID must be overridden with the fallback's KeyID.
		// This is a simplification; a real tool would get the fallback KID from its minter.
		// For now, we assume the caller provides the correct fallback KID in hostCtx.
		return t.fallbackMinter.MintToken(kind, fallbackPayload, hostCtx)
	}

	// Total failure: both signers are unavailable.
	return "", ErrMintingFailed
}
