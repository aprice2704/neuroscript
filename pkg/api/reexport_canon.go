// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Re-exports internal canon functions needed for persistence.
// filename: pkg/api/reexport_canon.go
// nlines: 20

package api

import (
	"github.com/aprice2704/neuroscript/pkg/canon"
)

// This file re-exports internal canonicalization functions
// required by host applications (like FDM) to persist
// definitions to a graph.

var (
	// CanonicaliseNode serializes a minimal AST node (like *ast.Procedure
	// or *ast.StringLiteralNode) into a binary blob.
	CanonicaliseNode = canon.CanonicaliseNode

	// DecodeNode deserializes a binary blob back into its minimal AST node.
	DecodeNode = canon.DecodeNode

	// ValueToNode converts a lang.Value (like a string or map) into its
	// corresponding AST literal node (e.g., *ast.StringLiteralNode).
	ValueToNode = canon.ValueToNode

	// NodeToValue converts an AST literal node back into its lang.Value.
	NodeToValue = canon.NodeToValue
)
