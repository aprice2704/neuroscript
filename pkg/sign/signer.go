// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Corrected the Verify function to use the correct hashing logic and complete the AST decoding.
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
	"golang.org/x/crypto/blake2b"
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

	// 1. Re-calculate the hash of the blob using the correct (BLAKE2b) algorithm.
	freshSum := blake2b.Sum256(s.Blob)

	// 2. Compare the fresh hash with the provided hash to ensure data integrity.
	if !bytes.Equal(freshSum[:], s.Sum[:]) {
		return nil, fmt.Errorf("integrity check failed: blob hash does not match provided sum")
	}

	// 3. Verify the Ed25519 signature.
	// The message is the hash digest prepended to the data blob.
	messageToVerify := append(s.Sum[:], s.Blob...)
	if !ed25519.Verify(publicKey, messageToVerify, s.Sig) {
		return nil, fmt.Errorf("signature verification failed")
	}

	// 4. Decode the verified blob back into an AST.
	tree, err := canon.Decode(s.Blob)
	if err != nil {
		return nil, fmt.Errorf("failed to decode verified blob: %w", err)
	}

	return tree, nil
}
