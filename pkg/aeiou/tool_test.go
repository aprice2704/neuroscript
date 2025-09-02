// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Corrects the test logic to check the KeyID on the decoded payload.
// filename: aeiou/tool_test.go
// nlines: 83
// risk_rating: HIGH

package aeiou

import (
	"crypto/ed25519"
	"errors"
	"testing"
)

func TestMagicTool(t *testing.T) {
	// --- Setup ---
	// Primary key and minter
	pubPrimary, privPrimary, _ := ed25519.GenerateKey(nil)
	minterPrimary, _ := NewMagicMinter(privPrimary)
	kidPrimary := "key-primary-1"

	// Fallback key and minter
	pubFallback, privFallback, _ := ed25519.GenerateKey(nil)
	minterFallback, _ := NewMagicMinter(privFallback)
	kidFallback := "key-fallback-1"

	// Verifier that knows about both keys
	kp := NewRotatingKeyProvider()
	kp.Add(kidPrimary, pubPrimary)
	kp.Add(kidFallback, pubFallback)
	verifier := NewMagicVerifier(kp)

	hostCtx := HostContext{SessionID: "sid-tool", TurnIndex: 1, TurnNonce: "nonce-tool"}
	agentPayload := ControlPayload{Action: ActionContinue}

	// --- Test Cases ---
	t.Run("Primary signer works", func(t *testing.T) {
		tool := NewMagicTool(minterPrimary, minterFallback)
		ctx := hostCtx
		ctx.KeyID = kidPrimary

		token, err := tool.MintMagicToken(KindLoop, agentPayload, ctx)
		if err != nil {
			t.Fatalf("MintMagicToken failed unexpectedly: %v", err)
		}

		// Verify the token has the agent's intended action and correct KID
		payload, err := verifier.ParseAndVerify(token, ctx)
		if err != nil {
			t.Fatalf("Token verification failed: %v", err)
		}
		if payload.KeyID != kidPrimary {
			t.Errorf("Expected KeyID to be '%s', but got '%s'", kidPrimary, payload.KeyID)
		}
		if payload.Payload.Action != ActionContinue {
			t.Errorf("Expected action to be 'continue', but got '%s'", payload.Payload.Action)
		}
	})

	t.Run("Fallback signer is used when primary fails", func(t *testing.T) {
		tool := NewMagicTool(nil, minterFallback) // Simulate primary failure
		ctx := hostCtx
		ctx.KeyID = kidFallback // Context must point to the fallback key

		token, err := tool.MintMagicToken(KindLoop, agentPayload, ctx)
		if err != nil {
			t.Fatalf("MintMagicToken failed unexpectedly: %v", err)
		}

		// Verify the token was forced to be an 'abort' action with the fallback KID
		payload, err := verifier.ParseAndVerify(token, ctx)
		if err != nil {
			t.Fatalf("Token verification failed: %v", err)
		}
		if payload.KeyID != kidFallback {
			t.Errorf("Expected KeyID to be '%s', but got '%s'", kidFallback, payload.KeyID)
		}
		if payload.Payload.Action != ActionAbort {
			t.Errorf("Expected action to be forced to 'abort', but got '%s'", payload.Payload.Action)
		}
	})

	t.Run("Returns error when both signers fail", func(t *testing.T) {
		tool := NewMagicTool(nil, nil) // Simulate total failure
		ctx := hostCtx
		ctx.KeyID = "any-key"

		_, err := tool.MintMagicToken(KindLoop, agentPayload, ctx)
		if !errors.Is(err, ErrMintingFailed) {
			t.Fatalf("Expected ErrMintingFailed, but got: %v", err)
		}
	})
}
