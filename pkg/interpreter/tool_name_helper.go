// NeuroScript Version: 0.6.0
// File version: 2.0.0
// Purpose: Provides a centralized, exported helper function to resolve the canonical tool name for all lookups, ensuring case-insensitivity.
// filename: pkg/interpreter/tool_name_helper.go
// nlines: 35
// risk_rating: LOW

package interpreter

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// CanonicalToolName converts a tool name string to its canonical form for lookups.
// The current canonical form is lowercase.
func CanonicalToolName(name string) types.FullName {
	return types.FullName(strings.ToLower(name))
}

// resolveToolName constructs the canonical tool name for registry lookup
// based on the contract with the AST builder.
func resolveToolName(n *ast.CallableExprNode) (types.FullName, error) {
	if !n.Target.IsTool {
		// This should not be called for non-tool functions.
		return "", fmt.Errorf("internal error: resolveToolName called on a non-tool expression")
	}

	// Per the contract, the AST provides the group-qualified name (e.g., "math.Add").
	// The interpreter must prepend "tool." to create the canonical key used in the registry.
	if n.Target.Name == "" {
		return "", fmt.Errorf("internal error: tool call expression has an empty target name")
	}

	// Ensure we don't accidentally double-prefix if the parser's behavior changes.
	if strings.HasPrefix(n.Target.Name, "tool.") {
		return CanonicalToolName(n.Target.Name), nil
	}

	return CanonicalToolName("tool." + n.Target.Name), nil
}
