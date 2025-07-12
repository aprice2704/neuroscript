// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Re-exports core types from foundational packages for a clean public API.
// filename: pkg/api/reexport.go
// nlines: 20
// risk_rating: LOW

package api

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Re-exported types for the public API.
type (
	// Foundational types
	Kind     = types.Kind
	Position = types.Position

	// Core AST interfaces and structs
	Node = interfaces.Node
	Tree = interfaces.Tree

	// Specific AST nodes that might be useful externally
	Program = ast.Program
)
