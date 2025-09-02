// NeuroScript Version: 0.7.0
// File version: 6
// Purpose: Replaces simpleKeyProvider with RotatingKeyProvider and adds key rotation test.
// filename: aeiou/verifier_test.go
// nlines: 204
// risk_rating: HIGH

package aeiou

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// simpleKeyProvider is a basic implementation of KeyProvider for testing.
// This is being replaced by RotatingKeyProvider but kept for reference if needed.
type simpleKeyProvider struct {
	keys map[string]ed25519.PublicKey
}

func (s *simpleKeyProvider) PublicKey(kid string) (ed25519.PublicKey, error) {
	key, ok := s.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key not found for kid: %s", kid)
	}
	return key, nil
}

func TestMagicVerifier_ParseAndVerify(t *testing.T) {
	// Setup: Create keys and minter
	pubKey, privKey, _ := ed25519.GenerateKey(nil)
	minter, _ := NewMagicMinter(privKey)
	keyID := "test-key-1"

	// Setup: Create verifier with a key provider
	kp := NewRotatingKeyProvider()
	kp.Add(keyID, pubKey)
	verifier := NewMagicVerifier(kp)

	// Setup: Base context and payload for tests
	validHostCtx := HostContext{
		SessionID: "sid-1",
		TurnIndex: 1,
		TurnNonce: "nonce-1",
		KeyID:     keyID,
		TTL:       60,
	}
	agentPayload := ControlPayload{Action: ActionDone}

	// Mint a standard valid token
	validToken, _ := minter.MintToken(KindLoop, agentPayload, validHostCtx)

	testCases := []struct {
		name        string
		token       string
		hostCtx     HostContext
		expectErrIs error
	}{
		{
			name:        "Happy Path - Valid token",
			token:       validToken,
			hostCtx:     validHostCtx,
			expectErrIs: nil,
		},
		{
			name: "Bad Signature",
			token: func() string {
				lastDotIndex := strings.LastIndex(validToken, ".")
				prefix := validToken[:lastDotIndex+1]
				b64Tag := validToken[lastDotIndex+1 : len(validToken)-len(TokenMarkerSuffix)]
				tamperedB64Tag := "A" + b64Tag[1:]
				return prefix + tamperedB64Tag + TokenMarkerSuffix
			}(),
			hostCtx:     validHostCtx,
			expectErrIs: ErrTokenSignature,
		},
		{
			name:        "Oversize Token",
			token:       validToken + strings.Repeat("a", MaxTokenLength),
			hostCtx:     validHostCtx,
			expectErrIs: ErrTokenInvalid,
		},
		{
			name:        "Quoted Token",
			token:       `"` + validToken + `"`,
			hostCtx:     validHostCtx,
			expectErrIs: ErrTokenInvalid,
		},
		{
			name:        "Backticked Token",
			token:       "`" + validToken + "`",
			hostCtx:     validHostCtx,
			expectErrIs: ErrTokenInvalid,
		},
		{
			name:        "Wrong Session ID",
			token:       validToken,
			hostCtx:     HostContext{SessionID: "sid-2", TurnIndex: 1, TurnNonce: "nonce-1"},
			expectErrIs: ErrTokenScope,
		},
		{
			name:        "Wrong Turn Index",
			token:       validToken,
			hostCtx:     HostContext{SessionID: "sid-1", TurnIndex: 2, TurnNonce: "nonce-1"},
			expectErrIs: ErrTokenScope,
		},
		{
			name:        "Wrong Turn Nonce",
			token:       validToken,
			hostCtx:     HostContext{SessionID: "sid-1", TurnIndex: 1, TurnNonce: "nonce-2"},
			expectErrIs: ErrTokenScope,
		},
		{
			name: "Expired Token",
			token: func() string {
				ctx := validHostCtx
				ctx.TTL = -1 // Expire immediately
				tok, _ := minter.MintToken(KindLoop, agentPayload, ctx)
				return tok
			}(),
			hostCtx:     validHostCtx,
			expectErrIs: ErrTokenExpired,
		},
		{
			name:        "Malformed Token String",
			token:       "this-is-not-a-valid-token",
			hostCtx:     validHostCtx,
			expectErrIs: ErrTokenInvalid,
		},
		{
			name: "Unknown Key ID",
			token: func() string {
				otherCtx := validHostCtx
				otherCtx.KeyID = "unknown-key"
				tok, _ := minter.MintToken(KindLoop, agentPayload, otherCtx)
				return tok
			}(),
			hostCtx:     validHostCtx,
			expectErrIs: ErrTokenUnknownKID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctxToUse := tc.hostCtx
			if tc.name == "Unknown Key ID" {
				var p TokenPayload
				_, b64Payload, _, _ := parseToken(tc.token)
				payloadBytes, _ := base64.RawURLEncoding.DecodeString(b64Payload)
				json.Unmarshal(payloadBytes, &p)
				ctxToUse.TurnIndex = p.TurnIndex
				ctxToUse.TurnNonce = p.TurnNonce
				ctxToUse.SessionID = p.SessionID
			}

			if tc.name == "Expired Token" {
				time.Sleep(10 * time.Millisecond)
			}

			payload, err := verifier.ParseAndVerify(tc.token, ctxToUse)

			if tc.expectErrIs != nil {
				if !errors.Is(err, tc.expectErrIs) {
					t.Fatalf("ParseAndVerify() expected error target %v, got %v", tc.expectErrIs, err)
				}
				if payload != nil {
					t.Errorf("Expected nil payload on error, but got %+v", payload)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseAndVerify() failed unexpectedly: %v", err)
			}
			if payload == nil {
				t.Fatal("Expected a valid payload, but got nil")
			}
			if payload.Payload.Action != agentPayload.Action {
				t.Errorf("Payload action mismatch. Got %s, want %s", payload.Payload.Action, agentPayload.Action)
			}
		})
	}
}

func TestMagicVerifier_KeyRotation(t *testing.T) {
	// 1. Setup initial key (v1)
	pubKey1, privKey1, _ := ed25519.GenerateKey(nil)
	minter1, _ := NewMagicMinter(privKey1)
	kid1 := "key-v1"

	// 2. Setup key provider and verifier
	kp := NewRotatingKeyProvider()
	kp.Add(kid1, pubKey1)
	verifier := NewMagicVerifier(kp)

	// 3. Mint and verify a token with the old key
	ctx1 := HostContext{SessionID: "sid-1", TurnIndex: 1, TurnNonce: "nonce-1", KeyID: kid1, TTL: 60}
	tokenV1, _ := minter1.MintToken(KindLoop, ControlPayload{Action: ActionContinue}, ctx1)

	if _, err := verifier.ParseAndVerify(tokenV1, ctx1); err != nil {
		t.Fatalf("Verification of token v1 failed before rotation: %v", err)
	}

	// 4. ROTATE KEYS: Introduce a new key (v2)
	pubKey2, privKey2, _ := ed25519.GenerateKey(nil)
	minter2, _ := NewMagicMinter(privKey2)
	kid2 := "key-v2"
	kp.Add(kid2, pubKey2) // Hot-reload the new public key

	// 5. Mint and verify a token with the NEW key
	ctx2 := HostContext{SessionID: "sid-1", TurnIndex: 2, TurnNonce: "nonce-2", KeyID: kid2, TTL: 60}
	tokenV2, _ := minter2.MintToken(KindLoop, ControlPayload{Action: ActionDone}, ctx2)
	if _, err := verifier.ParseAndVerify(tokenV2, ctx2); err != nil {
		t.Fatalf("Verification of token v2 failed after rotation: %v", err)
	}

	// 6. CRITICAL: Verify the token signed with the OLD key again. It must still pass.
	if _, err := verifier.ParseAndVerify(tokenV1, ctx1); err != nil {
		t.Fatalf("Verification of token v1 FAILED after rotation, but should have passed: %v", err)
	}
}
