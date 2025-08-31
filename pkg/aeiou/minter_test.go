// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines tests for the AEIOU v3 MagicMinter.
// filename: aeiou/minter_test.go
// nlines: 77
// risk_rating: HIGH

package aeiou

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestMagicMinter_MintToken_RoundTrip(t *testing.T) {
	// 1. Generate a new key pair for this test
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate ed25519 key pair: %v", err)
	}

	// 2. Create the minter
	minter, err := NewMagicMinter(privateKey)
	if err != nil {
		t.Fatalf("NewMagicMinter() failed: %v", err)
	}

	// 3. Define test inputs
	hostCtx := HostContext{
		SessionID: "test-sid-123",
		TurnIndex: 5,
		TurnNonce: "test-nonce-abc",
		KeyID:     "test-key-1",
		TTL:       300,
	}
	agentPayload := ControlPayload{
		Action:  ActionContinue,
		Request: json.RawMessage(`{"reason":"need_more_context"}`),
	}

	// 4. Mint the token
	token, err := minter.MintToken(KindLoop, agentPayload, hostCtx)
	if err != nil {
		t.Fatalf("MintToken() failed: %v", err)
	}

	// 5. Deconstruct and verify the token
	if !strings.HasPrefix(token, TokenMarkerPrefix) || !strings.HasSuffix(token, TokenMarkerSuffix) {
		t.Fatalf("Token has invalid prefix or suffix. Got: %s", token)
	}

	// Extract KIND:B64(PAYLOAD).B64(TAG)
	trimmedToken := strings.TrimPrefix(token, TokenMarkerPrefix+":")
	trimmedToken = strings.TrimSuffix(trimmedToken, TokenMarkerSuffix)

	parts := strings.Split(trimmedToken, ":")
	if len(parts) != 2 {
		t.Fatalf("Token has invalid structure. Expected 2 parts, got %d", len(parts))
	}
	kindStr, signedPart := parts[0], parts[1]

	if ControlKind(kindStr) != KindLoop {
		t.Errorf("Mismatched kind. Got %s, want %s", kindStr, KindLoop)
	}

	// Split payload and tag
	payloadAndTag := strings.Split(signedPart, ".")
	if len(payloadAndTag) != 2 {
		t.Fatalf("Signed part has invalid structure. Expected 2 parts, got %d", len(payloadAndTag))
	}
	b64Payload, b64Tag := payloadAndTag[0], payloadAndTag[1]

	// Decode payload and tag
	canonicalPayload, err := base64.RawURLEncoding.DecodeString(b64Payload)
	if err != nil {
		t.Fatalf("Failed to decode payload: %v", err)
	}
	tag, err := base64.RawURLEncoding.DecodeString(b64Tag)
	if err != nil {
		t.Fatalf("Failed to decode tag: %v", err)
	}

	// 6. Verify the signature
	if !ed25519.Verify(publicKey, canonicalPayload, tag) {
		t.Fatal("Signature verification failed")
	}

	// 7. Verify the payload content
	var decodedPayload TokenPayload
	if err := json.Unmarshal(canonicalPayload, &decodedPayload); err != nil {
		t.Fatalf("Failed to unmarshal canonical payload: %v", err)
	}

	if decodedPayload.SessionID != hostCtx.SessionID {
		t.Errorf("SessionID mismatch. Got %s, want %s", decodedPayload.SessionID, hostCtx.SessionID)
	}
	if decodedPayload.Payload.Action != agentPayload.Action {
		t.Errorf("Action mismatch. Got %s, want %s", decodedPayload.Payload.Action, agentPayload.Action)
	}
}
