// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Fixes type mismatch between []any and []*ast.Comment in registry wrappers.
// filename: pkg/api/reexport_canon.go
// nlines: 130

package api

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
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

// Re-exported errors from the canonicalization package.
var (
	// ErrInvalidMagic is returned when the byte slice to be decoded does not start
	// with the correct 'NSC' magic number.
	ErrInvalidMagic = canon.ErrInvalidMagic

	// ErrTruncatedData is returned when the decoder encounters an unexpected EOF,
	// indicating the data is incomplete.
	ErrTruncatedData = canon.ErrTruncatedData

	// ErrUnknownCodec is returned when the decoder encounters a node kind for which
	// no codec has been registered.
	ErrUnknownCodec = canon.ErrUnknownCodec
)

// CanonicaliseWithRegistry produces a deterministic binary representation of a full
// AST Tree. It is the public API inverse of api.DecodeWithRegistry.
// This function is required by host applications (like nsinterpretersvc)
// to persist a full, executable script to a graph.
func CanonicaliseWithRegistry(tree *Tree) ([]byte, [32]byte, error) {
	if tree == nil || tree.Root == nil {
		return nil, [32]byte{}, fmt.Errorf("cannot canonicalize a nil tree or a tree with a nil root")
	}

	// 1. Re-assemble the internal *ast.Tree from the public *api.Tree.
	// The internal *ast.Program (tree.Root) must hold the comments
	// for the canonicalizer to find them.
	prog, ok := tree.Root.(*ast.Program)
	if !ok {
		return nil, [32]byte{}, fmt.Errorf("tree.Root is not an *ast.Program, but %T", tree.Root)
	}

	// --- FIX: Convert public []any to internal []*ast.Comment ---
	if len(tree.Comments) > 0 {
		prog.Comments = make([]*ast.Comment, len(tree.Comments))
		for i, c := range tree.Comments {
			comment, ok := c.(*ast.Comment)
			if !ok {
				return nil, [32]byte{}, fmt.Errorf("comment at index %d is not *ast.Comment, but %T", i, c)
			}
			prog.Comments[i] = comment
		}
	} else {
		// Ensure the slice is non-nil for the canonicalizer.
		prog.Comments = make([]*ast.Comment, 0)
	}
	// --- END FIX ---

	internalTree := &ast.Tree{
		Root: prog,
	}

	// 2. Call the internal canon.CanonicaliseWithRegistry
	return canon.CanonicaliseWithRegistry(internalTree)
}

// DecodeWithRegistry reconstructs a full AST Tree from its canonical
// binary representation. It is the public API inverse of
// api.CanonicaliseWithRegistry.
//
// This function expects the provided blob to be a complete, valid
// binary AST (starting with the 'NSC' magic number) and for its
// root node to be an *ast.Program.
//
// It returns an *api.Tree wrapper, which is the required input
// for api.ExecWithInterpreter.
//
// Returns:
//
//	(*api.Tree, error)
//
// Possible Errors:
//   - ErrInvalidMagic: If the blob does not start with the 'NSC' magic number.
//   - ErrTruncatedData: If the blob is incomplete.
//   - ErrUnknownCodec: If the blob contains an invalid node kind.
//   - error: If the decoded root node is not an *ast.Program.
func DecodeWithRegistry(blob []byte) (*Tree, error) {
	// 1. Call the internal canon.DecodeWithRegistry
	internalTree, err := canon.DecodeWithRegistry(blob)
	if err != nil {
		return nil, err
	}

	// 2. Wrap the internal *ast.Tree in the public *api.Tree
	// The public api.Tree struct (aliased from interfaces.Tree)
	// separates Root from Comments, while the internal *ast.Program
	// (which is internalTree.Root) holds its own comments.
	prog, ok := internalTree.Root.(*ast.Program)
	if !ok {
		return nil, fmt.Errorf("decoded root node is not *ast.Program but %T", internalTree.Root)
	}

	// --- FIX: Convert internal []*ast.Comment to public []any ---
	var publicComments []any
	if len(prog.Comments) > 0 {
		publicComments = make([]any, len(prog.Comments))
		for i, c := range prog.Comments {
			publicComments[i] = c
		}
	} else {
		// Ensure the slice is non-nil for the public API.
		publicComments = make([]any, 0)
	}
	// --- END FIX ---

	return &Tree{Root: prog, Comments: publicComments}, nil
}
