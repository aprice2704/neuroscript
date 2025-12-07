// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 2
// :: description: Provides helper functions for inspecting the NeuroScript AST. Added HasDefinitions.
// :: latestChange: Added HasDefinitions to inspect AST for procedures or events.
// :: filename: pkg/api/command_helpers.go
// :: serialization: go

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

// HasDefinitions checks if the given Program node contains procedure definitions
// or event handlers.
// This is useful for distinguishing "Library Scripts" from "Command Scripts".
func HasDefinitions(prog *ast.Program) bool {
	if prog == nil {
		return false
	}
	return len(prog.Procedures) > 0 || len(prog.Events) > 0
}
