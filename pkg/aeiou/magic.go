// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines the structures and constants for AEIOU v3 magic control tokens.
// filename: aeiou/magic.go
// nlines: 48
// risk_rating: LOW

package aeiou

import "encoding/json"

const (
	// TokenMarkerPrefix is the start of a V3 control token.
	TokenMarkerPrefix = "<<<NSMAG:V3"
	// TokenMarkerSuffix is the end of a V3 control token.
	TokenMarkerSuffix = ">>>"
)

// ControlKind defines the type of control being requested (e.g., "LOOP").
type ControlKind string

const (
	// KindLoop is the standard control kind for loop management.
	KindLoop ControlKind = "LOOP"
)

// LoopAction defines the specific action for a LOOP control token.
type LoopAction string

const (
	ActionContinue LoopAction = "continue"
	ActionDone     LoopAction = "done"
	ActionAbort    LoopAction = "abort"
)

// ControlPayload is the agent-supplied portion of the token's payload.
type ControlPayload struct {
	Action    LoopAction      `json:"action"`
	Request   json.RawMessage `json:"request,omitempty"`
	Telemetry json.RawMessage `json:"telemetry,omitempty"`
}

// TokenPayload is the full, canonical payload that gets signed.
// It includes host-stitched context and the agent's control payload.
type TokenPayload struct {
	Version   int            `json:"v"`
	Kind      ControlKind    `json:"kind"`
	JTI       string         `json:"jti"`
	SessionID string         `json:"session_id"`
	TurnIndex int            `json:"turn_index"`
	TurnNonce string         `json:"turn_nonce"`
	IssuedAt  int64          `json:"issued_at"`
	TTL       int            `json:"ttl,omitempty"`
	KeyID     string         `json:"kid"`
	Payload   ControlPayload `json:"payload"`
}
