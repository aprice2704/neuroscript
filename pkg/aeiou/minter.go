// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Implements the token signing and formatting logic in the MagicMinter.
// filename: aeiou/minter.go
// nlines: 66
// risk_rating: HIGH

package aeiou

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// HostContext provides the necessary host-side information to bind a token
// to a specific session and turn.
type HostContext struct {
	SessionID string
	TurnIndex int
	TurnNonce string
	KeyID     string
	TTL       int
}

// MagicMinter creates and signs AEIOU v3 control tokens.
// It holds the host's private key and is responsible for correctly
// constructing, canonicalizing, and signing the token payload.
type MagicMinter struct {
	privateKey ed25519.PrivateKey
}

// NewMagicMinter creates a new minter with the given private key.
func NewMagicMinter(pk ed25519.PrivateKey) (*MagicMinter, error) {
	if len(pk) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid ed25519 private key size")
	}
	return &MagicMinter{privateKey: pk}, nil
}

// MintToken creates a new, signed control token.
func (m *MagicMinter) MintToken(kind ControlKind, agentPayload ControlPayload, hostCtx HostContext) (string, error) {
	// 1. Construct the full payload
	fullPayload := TokenPayload{
		Version:   3,
		Kind:      kind,
		JTI:       uuid.NewString(), // Unique ID for this token
		SessionID: hostCtx.SessionID,
		TurnIndex: hostCtx.TurnIndex,
		TurnNonce: hostCtx.TurnNonce,
		IssuedAt:  time.Now().Unix(),
		TTL:       hostCtx.TTL,
		KeyID:     hostCtx.KeyID,
		Payload:   agentPayload,
	}

	// 2. Canonicalize the payload
	canonicalPayload, err := Canonicalize(fullPayload)
	if err != nil {
		return "", fmt.Errorf("failed to canonicalize token payload: %w", err)
	}

	// 3. Sign the canonical payload
	tag := ed25519.Sign(m.privateKey, canonicalPayload)

	// 4. Format the final token string
	b64Payload := base64.RawURLEncoding.EncodeToString(canonicalPayload)
	b64Tag := base64.RawURLEncoding.EncodeToString(tag)

	token := fmt.Sprintf("%s:%s:%s.%s%s",
		TokenMarkerPrefix,
		kind,
		b64Payload,
		b64Tag,
		TokenMarkerSuffix,
	)

	return token, nil
}
