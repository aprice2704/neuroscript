// NeuroScript Version: 0.6.0
// File version: 11
// Purpose: Corrects a critical signature verification bug by separating cryptographic checks from AST decoding.
// filename: pkg/api/loader.go
// nlines: 65
// risk_rating: HIGH

package api

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/api/analysis"
	"github.com/aprice2704/neuroscript/pkg/canon"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"golang.org/x/crypto/blake2b"
)

// RunMode indicates the intended execution model of a script.
type RunMode uint8

const (
	RunModeLibrary RunMode = iota
	RunModeCommand
	RunModeEventSink
)

// LoaderConfig is a placeholder for future loader options.
type LoaderConfig struct{}

// LoadedUnit is the result of a successful Load operation.
type LoadedUnit struct {
	Tree     *interfaces.Tree
	Hash     [32]byte
	Mode     RunMode
	RawBytes []byte
}

// DetectRunMode determines the run mode from the AST.
func DetectRunMode(tree *interfaces.Tree) RunMode {
	// A real implementation would inspect the tree's root and top-level nodes.
	return RunModeCommand
}

// Load performs signature verification and analysis passes on a signed AST.
func Load(ctx context.Context, s *SignedAST, cfg LoaderConfig, pubKey ed25519.PublicKey) (*LoadedUnit, error) {
	// 1. Verify that the signature is valid for the given hash.
	if !ed25519.Verify(pubKey, s.Sum[:], s.Sig) {
		return nil, fmt.Errorf("signature verification failed: invalid signature")
	}

	// 2. Verify that the hash is the correct hash of the blob.
	// This prevents tampering with the blob after signing.
	recomputedSum := blake2b.Sum256(s.Blob)
	if recomputedSum != s.Sum {
		return nil, fmt.Errorf("integrity check failed: blob does not match provided hash")
	}

	// 3. Now that all crypto is verified, decode the blob into an AST.
	verifiedTree, err := canon.Decode(s.Blob)
	if err != nil {
		return nil, fmt.Errorf("failed to decode verified blob: %w", err)
	}

	// 4. Run static analysis passes.
	if err := analysis.RunAll(verifiedTree); err != nil {
		return nil, fmt.Errorf("analysis pass failed: %w", err)
	}

	lu := &LoadedUnit{
		Tree:     verifiedTree,
		Hash:     s.Sum,
		Mode:     DetectRunMode(verifiedTree),
		RawBytes: s.Blob,
	}

	return lu, nil
}
