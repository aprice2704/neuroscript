// NeuroScript Version: 0.6.3
// File version: 3
// Purpose: Implements signing and verification, updated to use the new registry-based decoder.
// filename: pkg/sign/signer.go
// nlines: 45
// risk_rating: HIGH

package sign

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/canon"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"golang.org/x/crypto/blake2b"
)

// SignedAST is the internal representation of a signed abstract syntax tree.
type SignedAST struct {
	Blob []byte
	Sum  [32]byte
	Sig  []byte
}

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrHashMismatch     = errors.New("blob hash does not match provided sum")
)

// NewTestKey creates a new Ed25519 key pair for testing purposes.
func NewTestKey() (ed25519.PrivateKey, ed25519.PublicKey) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(fmt.Sprintf("failed to generate test key: %v", err))
	}
	return priv, pub
}

// Sign creates a signature for the given AST blob and its hash.
func Sign(privKey ed25519.PrivateKey, blob []byte, sum [32]byte) (*SignedAST, error) {
	sig := ed25519.Sign(privKey, sum[:])
	return &SignedAST{Blob: blob, Sum: sum, Sig: sig}, nil
}

// Verify checks the hash and signature, and on success, decodes the blob.
func Verify(pubKey ed25519.PublicKey, s *SignedAST) (*interfaces.Tree, error) {
	// 1. **CRITICAL**: Verify that the blob content matches the provided hash first.
	computedSum := blake2b.Sum256(s.Blob)
	if !bytes.Equal(computedSum[:], s.Sum[:]) {
		return nil, ErrHashMismatch
	}

	// 2. Verify the signature against the (now trusted) hash.
	if !ed25519.Verify(pubKey, s.Sum[:], s.Sig) {
		return nil, ErrInvalidSignature
	}

	// 3. Decode the blob.
	tree, err := canon.DecodeWithRegistry(s.Blob)
	if err != nil {
		return nil, fmt.Errorf("failed to decode blob after verification: %w", err)
	}
	// The *ast.Tree is compatible with the *interfaces.Tree return type.
	return &interfaces.Tree{Root: tree.Root}, nil
}
