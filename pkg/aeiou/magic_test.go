// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines tests for the AEIOU v3 magic control token structures.
// filename: aeiou/magic_test.go
// nlines: 40
// risk_rating: LOW

package aeiou

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTokenPayload_JSONMarshalling(t *testing.T) {
	payload := TokenPayload{
		Version:   3,
		Kind:      KindLoop,
		JTI:       "test-jti",
		SessionID: "test-sid",
		TurnIndex: 1,
		TurnNonce: "test-nonce",
		IssuedAt:  time.Now().Unix(),
		TTL:       120,
		KeyID:     "ed25519-main-2025-08",
		Payload: ControlPayload{
			Action:  ActionContinue,
			Request: json.RawMessage(`{"min_tokens":1024}`),
		},
	}

	_, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal TokenPayload to JSON: %v", err)
	}
}
