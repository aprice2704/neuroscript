// NeuroScript Version: 0.6.0
// File version: 3
// Purpose: Re-exports core types from foundational packages for a clean public API. Aligned with contract v0.6.
// filename: pkg/api/reexport.go
// nlines: 23
// risk_rating: LOW

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Re-exported types for the public API, as per the v0.6 contract.
// These types provide the stable, public-facing surface for all interactions
// with the NeuroScript AST and its components.
type (
	// Foundational types from pkg/types, ensuring a stable AST contract.
	Kind     = types.Kind
	Position = types.Position
	Node     = interfaces.Node
	Tree     = interfaces.Tree

	// SignedAST is the transport wrapper for a canonicalized and signed tree.
	SignedAST struct {
		Blob []byte   // The canonicalized AST, produced by Canonicalise.
		Sum  [32]byte // The BLAKE2b-256 digest of the Blob.
		Sig  []byte   // The Ed25519 signature of the Sum.
	}

	// Value represents the result of an execution.
	// This is a placeholder for the actual Value type from the interpreter.
	Value any
)
