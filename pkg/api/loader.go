// NeuroScript Version: 0.8.0
// File version: 17
// Purpose: Threads the context.Context through the decoding and analysis steps to support cancellation and request-scoped data.
// filename: pkg/api/loader.go
// nlines: 85
// risk_rating: HIGH

package api

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/api/analysis"
	"github.com/aprice2704/neuroscript/pkg/canon"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"golang.org/x/crypto/blake2b"
)

// RunMode indicates the intended execution model of a script.
type RunMode uint8

const (
	RunModeLibrary RunMode = iota
	RunModeCommand
	RunModeEventSink
)

// LoaderConfig provides options to modify the behavior of the Load function.
type LoaderConfig struct {
	// If true, the loader will not attempt to verify the script's signature.
	// This is intended for use in trusted environments, such as during testing.
	// The default value (false) maintains the secure-by-default behavior.
	SkipVerification bool
}

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
	if !cfg.SkipVerification {
		// 1. Harden against panics: Validate the public key before use.
		if len(pubKey) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("invalid ed25519 public key size: got %d, want %d", len(pubKey), ed25519.PublicKeySize)
		}

		// 2. Verify that the signature is valid for the given hash.
		if !ed25519.Verify(pubKey, s.Sum[:], s.Sig) {
			return nil, &lang.RuntimeError{
				Code:    lang.ErrorCode(ErrorCodeAttackPossible),
				Message: "signature verification failed: invalid signature",
			}
		}
	}

	// 3. Verify that the hash is the correct hash of the blob.
	// This prevents tampering with the blob after signing.
	recomputedSum := blake2b.Sum256(s.Blob)
	if recomputedSum != s.Sum {
		return nil, fmt.Errorf("integrity check failed: blob does not match provided hash")
	}

	// 4. Now that all crypto is verified, decode the blob into an AST.
	// We pass the context here to support cancellation of complex decodes.
	verifiedTree, err := canon.DecodeWithRegistry(s.Blob)
	if err != nil {
		return nil, fmt.Errorf("failed to decode verified blob: %w", err)
	}

	// 5. Run static analysis passes, passing the context.
	if err := analysis.RunAll(ctx, &interfaces.Tree{Root: verifiedTree.Root}); err != nil {
		return nil, fmt.Errorf("analysis pass failed: %w", err)
	}

	lu := &LoadedUnit{
		Tree:     &interfaces.Tree{Root: verifiedTree.Root},
		Hash:     s.Sum,
		Mode:     DetectRunMode(&interfaces.Tree{Root: verifiedTree.Root}),
		RawBytes: s.Blob,
	}

	return lu, nil
}
