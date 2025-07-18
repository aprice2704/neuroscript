// NeuroScript Version: 0.6.0
// File version: 10
// Purpose: Corrects the import path for the analysis package.
// filename: pkg/api/loader.go
// nlines: 62
// risk_rating: HIGH

package api

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/api/analysis"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/sign"
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
	internalSignedAST := &sign.SignedAST{Blob: s.Blob, Sum: s.Sum, Sig: s.Sig}

	verifiedTree, err := sign.Verify(pubKey, internalSignedAST)
	if err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	if err := analysis.RunAll(verifiedTree); err != nil { // This call will now resolve correctly.
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
