// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Provides helper functions for inspecting the NeuroScript AST.
// filename: pkg/api/command_helpers.go
// nlines: 21
// risk_rating: LOW

package api

import "github.com/aprice2704/neuroscript/pkg/ast"

// HasCommandBlock checks if the given Program node contains one or more command blocks.
// This is useful for determining if a script is a "Command Script" intended for direct
// execution.
func HasCommandBlock(prog *ast.Program) bool {
	if prog == nil {
		return false
	}
	return len(prog.Commands) > 0
}
