// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Adds a statement to the command block to prevent parser errors for empty blocks.
// filename: pkg/api/loader_test.go
// nlines: 43
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

	// **FIX:** Add an emit statement to create a valid, non-empty command block.
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
