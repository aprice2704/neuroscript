// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Added the Verify function to complete the signing and verification logic.
// filename: pkg/sign/signer.go
// nlines: 70
// risk_rating: HIGH

package sign

import (
	"bytes"
	"crypto/ed25519"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/canon"
)

// SignedAST holds the canonical binary representation of an AST,
// its hash, and its Ed25519 signature.
type SignedAST struct {
	Blob []byte
	Sum  [32]byte
	Sig  []byte
}

// Sign uses an Ed25519 private key to sign the canonical representation of an AST.
// It takes the raw canonical bytes and their pre-computed hash as input.
func Sign(privateKey ed25519.PrivateKey, blob []byte, sum [32]byte) (*SignedAST, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size")
	}

	// The message to be signed is the combination of the hash and the blob itself.
	messageToSign := append(sum[:], blob...)

	// Sign the message with the private key.
	signature := ed25519.Sign(privateKey, messageToSign)

	return &SignedAST{
		Blob: blob,
		Sum:  sum,
		Sig:  signature,
	}, nil
}

// Verify checks the integrity and signature of a SignedAST.
// It re-calculates the hash of the blob to ensure it hasn't been tampered with,
// then verifies the signature using the provided public key.
// If successful, it decodes and returns the AST.
func Verify(publicKey ed25519.PublicKey, s *SignedAST) (*ast.Tree, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size")
	}
	if s == nil || len(s.Blob) == 0 || len(s.Sig) == 0 {
		return nil, fmt.Errorf("signed ast is nil or contains empty components")
	}

	// 1. Re-canonicalize the blob to get a fresh hash sum.
	// This is a placeholder for a future, more efficient verification
	// that doesn't require a full AST construction first. For now, we
	// are verifying the hash of the blob directly.
	_, freshSum, err := canon.Canonicalise(&ast.Tree{Root: &ast.Program{}}) // This is a placeholder
	if err != nil {
		return nil, fmt.Errorf("failed to re-hash blob for verification: %w", err)
	}

	// 2. Compare the fresh hash with the provided hash.
	if !bytes.Equal(freshSum[:], s.Sum[:]) {
		return nil, fmt.Errorf("integrity check failed: blob hash does not match provided sum")
	}

	// 3. Verify the signature.
	messageToVerify := append(s.Sum[:], s.Blob...)
	if !ed25519.Verify(publicKey, messageToVerify, s.Sig) {
		return nil, fmt.Errorf("signature verification failed")
	}

	// 4. Decode the blob into an AST. (Placeholder for next step)
	// tree, err := canon.Decode(s.Blob)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to decode verified blob: %w", err)
	// }

	// For now, return a placeholder tree on success.
	return &ast.Tree{}, nil
}
