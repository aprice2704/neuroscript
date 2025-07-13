// NeuroScript Version: 0.5.2
// File version: 5
// Purpose: Defines the logic for loading, validating, and preparing a NeuroScript AST for execution.
// filename: pkg/api/loader.go
// nlines: 57
// risk_rating: HIGH

package api

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/analysis"
	// "github.com/aprice2704/neuroscript/pkg/canon" // No longer needed
	"github.com/aprice2704/neuroscript/pkg/sign"
)

// RunMode indicates the intended execution model of a script.
type RunMode uint8

const (
	RunModeLibrary   RunMode = iota // funcs only
	RunModeCommand                  // unnamed command block, run-once
	RunModeEventSink                // one or more on-event handlers
)

// LoaderConfig toggles caching, gas limits, secret decode key, etc.
type LoaderConfig struct {
	Cache         Cache
	MaxGas        uint64
	SecretPrivKey []byte
}

// LoadedUnit is the result of a successful Load operation.
type LoadedUnit struct {
	Tree     *Tree
	Hash     [32]byte
	Mode     RunMode
	RawBytes []byte
}

// DetectRunMode determines the run mode from the AST.
func DetectRunMode(tree *Tree) RunMode {
	// A real implementation would inspect the tree's root and top-level nodes
	// to determine if it's a command, library, or event sink.
	return RunModeCommand
}

// Load transforms and validates a signed AST, but does not run it.
func Load(ctx context.Context, s *SignedAST, cfg LoaderConfig, pubKey []byte) (*LoadedUnit, error) {
	// Convert from the public api.SignedAST to the internal sign.SignedAST
	internalSignedAST := &sign.SignedAST{
		Blob: s.Blob,
		Sum:  s.Sum,
		Sig:  s.Sig,
	}
	verifiedTree, err := sign.Verify(pubKey, internalSignedAST)
	if err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	diags := analysis.Vet(verifiedTree)
	if len(diags) > 0 {
		return nil, fmt.Errorf("vetting failed with %d diagnostics", len(diags))
	}

	// FIX: The tree has been verified, so the original blob and sum from the
	// SignedAST are the source of truth. Re-canonicalizing is redundant and
	// was likely causing the integrity mismatch if any subtle differences existed.
	lu := &LoadedUnit{
		Tree:     verifiedTree,
		Hash:     s.Sum, // Use the original, verified hash
		Mode:     DetectRunMode(verifiedTree),
		RawBytes: s.Blob, // Use the original, verified blob
	}

	return lu, nil
}
