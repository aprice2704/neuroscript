// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Corrects the "last-wins" test case by using two distinct tokens.
// filename: aeiou/loop_test.go
// nlines: 156
// risk_rating: LOW

package aeiou

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoopController_ProcessOutput(t *testing.T) {
	// Setup: Create keys, minter, verifier, and controller
	pubKey, privKey, _ := ed25519.GenerateKey(nil)
	minter, _ := NewMagicMinter(privKey)
	keyID := "test-key-1"

	// Using simpleKeyProvider for simplicity in existing tests
	kpSimple := &simpleKeyProvider{keys: map[string]ed25519.PublicKey{keyID: pubKey}}
	verifier := NewMagicVerifier(kpSimple)
	lc := NewLoopController(verifier)

	// Setup: Base context
	hostCtx := HostContext{
		SessionID: "sid-1",
		TurnIndex: 1,
		TurnNonce: "nonce-1",
		KeyID:     keyID,
		TTL:       60,
	}

	// Mint tokens for various scenarios
	continueToken, _ := minter.MintToken(KindLoop, ControlPayload{Action: ActionContinue}, hostCtx)
	continueToken2, _ := minter.MintToken(KindLoop, ControlPayload{Action: ActionContinue}, hostCtx) // A second, unique token
	doneToken, _ := minter.MintToken(KindLoop, ControlPayload{Action: ActionDone}, hostCtx)
	abortToken, _ := minter.MintToken(KindLoop, ControlPayload{Action: ActionAbort}, hostCtx)

	testCases := []struct {
		name              string
		output            string
		hostCtxOverride   *HostContext
		setupCache        func(cache *ReplayCache, token string)
		expectedAction    LoopAction
		expectedLints     []Lint
		expectNilDecision bool
	}{
		{
			name:           "One valid continue token",
			output:         continueToken,
			expectedAction: ActionContinue,
		},
		{
			name:           "One valid done token",
			output:         doneToken,
			expectedAction: ActionDone,
		},
		{
			name:              "No valid tokens",
			output:            "Just some plain text output.",
			expectNilDecision: true,
		},
		{
			name:           "Precedence: Abort wins over Done",
			output:         strings.Join([]string{doneToken, abortToken}, "\n"),
			expectedAction: ActionAbort,
		},
		{
			name:           "Precedence: Done wins over Continue",
			output:         strings.Join([]string{continueToken, doneToken}, "\n"),
			expectedAction: ActionDone,
		},
		{
			name:           "Last-wins for same precedence",
			output:         strings.Join([]string{continueToken, "some other text", continueToken2}, "\n"),
			expectedAction: ActionContinue,
			expectedLints:  nil,
		},
		{
			name:           "Post-token text creates a lint",
			output:         strings.Join([]string{continueToken, "this is extra text"}, "\n"),
			expectedAction: ActionContinue,
			expectedLints:  []Lint{{Code: LintCodePostTokenText, Message: "extraneous text found after the chosen control token"}},
		},
		{
			name:           "Post-token empty lines do not create a lint",
			output:         strings.Join([]string{doneToken, "  ", "\t"}, "\n"),
			expectedAction: ActionDone,
			expectedLints:  nil,
		},
		{
			name: "Replayed token is ignored",
			output: func() string {
				ctx := hostCtx
				ctx.TurnIndex = 99
				tok, _ := minter.MintToken(KindLoop, ControlPayload{Action: ActionAbort}, ctx)
				return tok
			}(),
			hostCtxOverride: &HostContext{
				SessionID: "sid-1",
				TurnIndex: 99,
				TurnNonce: "nonce-1",
				KeyID:     keyID,
			},
			setupCache: func(cache *ReplayCache, token string) {
				_, b64Payload, _, _ := parseToken(token)
				payloadBytes, _ := base64.RawURLEncoding.DecodeString(b64Payload)
				var p TokenPayload
				json.Unmarshal(payloadBytes, &p)
				cache.CheckAndAdd(p.JTI)
			},
			expectNilDecision: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			replayCache := NewReplayCache(10, 5*time.Minute)
			ctx := hostCtx
			if tc.hostCtxOverride != nil {
				ctx = *tc.hostCtxOverride
			}

			if tc.setupCache != nil {
				tc.setupCache(replayCache, tc.output)
			}

			decision, err := lc.ProcessOutput(tc.output, ctx, replayCache)
			if err != nil {
				t.Fatalf("ProcessOutput failed unexpectedly: %v", err)
			}

			if tc.expectNilDecision {
				if decision != nil {
					t.Fatalf("Expected a nil decision, but got: %+v", decision)
				}
				return
			}

			if decision == nil {
				t.Fatal("Expected a decision, but got nil")
			}
			if decision.Action != tc.expectedAction {
				t.Errorf("Mismatched action. Got %s, want %s", decision.Action, tc.expectedAction)
			}
			if !reflect.DeepEqual(decision.Lints, tc.expectedLints) {
				t.Errorf("Mismatched lints. Got %+v, want %+v", decision.Lints, tc.expectedLints)
			}
		})
	}
}
