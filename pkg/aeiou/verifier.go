// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Adds token length and quote-wrapping validation.
// filename: aeiou/verifier.go
// nlines: 114
// risk_rating: HIGH

package aeiou

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	// MaxTokenLength is the maximum allowed length for a token string.
	MaxTokenLength = 1024
)

// KeyProvider is an interface for a component that can look up a public key by its ID.
type KeyProvider interface {
	PublicKey(kid string) (ed25519.PublicKey, error)
}

// MagicVerifier parses and validates AEIOU v3 control tokens.
type MagicVerifier struct {
	kp KeyProvider
}

// NewMagicVerifier creates a new verifier with the given key provider.
func NewMagicVerifier(kp KeyProvider) *MagicVerifier {
	return &MagicVerifier{kp: kp}
}

// ParseAndVerify takes a raw token string and the expected host context, performs all
// validation checks, and returns the verified token payload if successful.
func (v *MagicVerifier) ParseAndVerify(token string, hostCtx HostContext) (*TokenPayload, error) {
	kind, b64Payload, b64Tag, err := parseToken(token)
	if err != nil {
		return nil, err
	}

	canonicalPayload, err := base64.RawURLEncoding.DecodeString(b64Payload)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode payload: %v", ErrTokenInvalid, err)
	}
	tag, err := base64.RawURLEncoding.DecodeString(b64Tag)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode tag: %v", ErrTokenInvalid, err)
	}

	var payload TokenPayload
	if err := json.Unmarshal(canonicalPayload, &payload); err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal payload: %v", ErrTokenInvalid, err)
	}

	if payload.Kind != kind {
		return nil, fmt.Errorf("%w: kind mismatch in token string and payload", ErrTokenInvalid)
	}

	publicKey, err := v.kp.PublicKey(payload.KeyID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTokenUnknownKID, err)
	}

	if !ed25519.Verify(publicKey, canonicalPayload, tag) {
		return nil, ErrTokenSignature
	}

	if payload.SessionID != hostCtx.SessionID ||
		payload.TurnIndex != hostCtx.TurnIndex ||
		payload.TurnNonce != hostCtx.TurnNonce {
		return nil, ErrTokenScope
	}

	if payload.TTL != 0 {
		expiresAt := time.Unix(payload.IssuedAt, 0).Add(time.Duration(payload.TTL) * time.Second)
		if time.Now().After(expiresAt) {
			return nil, ErrTokenExpired
		}
	}

	return &payload, nil
}

// parseToken deconstructs the raw token string into its main parts.
func parseToken(token string) (ControlKind, string, string, error) {
	if len(token) > MaxTokenLength {
		return "", "", "", fmt.Errorf("%w: token exceeds max length of %d bytes", ErrTokenInvalid, MaxTokenLength)
	}

	trimmedToken := strings.TrimSpace(token)
	if (strings.HasPrefix(trimmedToken, `"`) && strings.HasSuffix(trimmedToken, `"`)) ||
		(strings.HasPrefix(trimmedToken, "`") && strings.HasSuffix(trimmedToken, "`")) {
		return "", "", "", fmt.Errorf("%w: token must not be wrapped in quotes or backticks", ErrTokenInvalid)
	}

	if !strings.HasPrefix(trimmedToken, TokenMarkerPrefix) || !strings.HasSuffix(trimmedToken, TokenMarkerSuffix) {
		return "", "", "", fmt.Errorf("%w: invalid prefix or suffix", ErrTokenInvalid)
	}

	trimmed := strings.TrimPrefix(trimmedToken, TokenMarkerPrefix+":")
	trimmed = strings.TrimSuffix(trimmed, TokenMarkerSuffix)

	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("%w: expected KIND:PAYLOAD.TAG format", ErrTokenInvalid)
	}
	kind, signedPart := ControlKind(parts[0]), parts[1]

	payloadAndTag := strings.Split(signedPart, ".")
	if len(payloadAndTag) != 2 {
		return "", "", "", fmt.Errorf("%w: expected PAYLOAD.TAG format", ErrTokenInvalid)
	}
	b64Payload, b64Tag := payloadAndTag[0], payloadAndTag[1]

	return kind, b64Payload, b64Tag, nil
}
