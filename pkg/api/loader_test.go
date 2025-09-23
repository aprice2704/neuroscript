// NeuroScript Version: 0.6.0
// File version: 6
// Purpose: Adds a statement to the command block to prevent parser errors for empty blocks. Adds a test for skipping signature verification.
// filename: pkg/api/loader_test.go
// nlines: 101
// risk_rating: MEDIUM

package api_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// TestLoad_InvalidPublicKey confirms that Load returns an error, not a panic,
// when given a nil or incorrectly sized public key.
func TestLoad_InvalidPublicKey(t *testing.T) {
	signedAST := &api.SignedAST{}
	config := api.LoaderConfig{} // Default: verification is ON

	testCases := []struct {
		name    string
		pubKey  ed25519.PublicKey
		wantErr string
	}{
		{
			name:    "nil public key",
			pubKey:  nil,
			wantErr: "invalid ed25519 public key size",
		},
		{
			name:    "short public key",
			pubKey:  []byte{0x01, 0x02, 0x03},
			wantErr: "invalid ed25519 public key size",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := api.Load(context.Background(), signedAST, config, tc.pubKey)
			if err == nil {
				t.Fatalf("Expected an error for an invalid public key, but got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Error message should indicate invalid key size, but got: %v", err)
			}
		})
	}
}

// TestLoad_InvalidSignature verifies that the loader rejects an AST whose
// signature does not match the public key.
func TestLoad_InvalidSignature(t *testing.T) {
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate public key: %v", err)
	}
	_, wrongPrivKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	validSrc := "command\n  emit \"ok\"\nendcommand"
	tree, err := api.Parse([]byte(validSrc), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Setup failed: api.Parse returned an error: %v", err)
	}

	blob, sum, err := api.Canonicalise(tree)
	if err != nil {
		t.Fatalf("Setup failed: api.Canonicalise returned an error: %v", err)
	}

	sig := ed25519.Sign(wrongPrivKey, sum[:])
	signedAST := &api.SignedAST{Blob: blob, Sum: sum, Sig: sig}

	_, err = api.Load(context.Background(), signedAST, api.LoaderConfig{}, pubKey)
	if err == nil {
		t.Fatal("Expected an error for an invalid signature, but got nil")
	}
	if !strings.Contains(err.Error(), "signature verification failed") {
		t.Errorf("Error message should indicate signature failure, but got: %v", err)
	}
}

// TestLoad_SkipVerification verifies that the loader successfully loads a script
// without a valid signature if the SkipVerification flag is set.
func TestLoad_SkipVerification(t *testing.T) {
	validSrc := "command\n  emit \"ok\"\nendcommand"
	tree, err := api.Parse([]byte(validSrc), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Setup failed: api.Parse returned an error: %v", err)
	}

	blob, sum, err := api.Canonicalise(tree)
	if err != nil {
		t.Fatalf("Setup failed: api.Canonicalise returned an error: %v", err)
	}

	// Create a signedAST with no actual signature.
	signedAST := &api.SignedAST{Blob: blob, Sum: sum, Sig: nil}
	config := api.LoaderConfig{SkipVerification: true}

	// Pass nil for the public key, which would otherwise cause an error.
	_, err = api.Load(context.Background(), signedAST, config, nil)
	if err != nil {
		t.Fatalf("Expected Load to succeed with SkipVerification=true, but it failed: %v", err)
	}
}
