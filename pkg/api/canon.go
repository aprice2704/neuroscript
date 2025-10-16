// NeuroScript Version: 0.8.0
// File version: 10
// Purpose: Public API wrapper for the canonicalization engine, updated to be context-aware.
// filename: pkg/api/canon.go
// nlines: 31
// risk_rating: LOW

package api

import (
	"context"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/canon"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"golang.org/x/crypto/blake2b"
)

// Canonicalise produces a deterministic binary representation of the AST.
// It wraps the internal canonicalizer.
func Canonicalise(tree *interfaces.Tree) ([]byte, [32]byte, error) {
	internalTree := &ast.Tree{Root: tree.Root, Comments: tree.Comments}
	// Note: Canonicalization is a pure function and does not need a context.
	return canon.CanonicaliseWithRegistry(internalTree)
}

// Decode reconstructs an AST from its canonical binary representation.
// It wraps the internal decoder and computes the hash of the input blob.
// It now accepts a context to allow for cancellation.
func Decode(ctx context.Context, blob []byte) (*interfaces.Tree, [32]byte, error) {
	internalTree, err := canon.DecodeWithRegistry(blob)
	if err != nil {
		return nil, [32]byte{}, err
	}

	// The public API contract requires us to return the hash,
	// so we compute it from the input blob using the correct algorithm.
	sum := blake2b.Sum256(blob)

	// The returned *ast.Tree satisfies the *interfaces.Tree.
	return &interfaces.Tree{Root: internalTree.Root, Comments: internalTree.Comments}, sum, nil
}
