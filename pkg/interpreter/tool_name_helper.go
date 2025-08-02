// NeuroScript Version: 0.6.0
// File version: 1.0.0
// Purpose: Provides a centralized helper function to resolve the canonical tool name from an AST node, as per the official parser-interpreter contract.
// filename: pkg/interpreter/evaluation_helpers.go
// nlines: 30
// risk_rating: LOW

package interpreter

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

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
		return types.FullName(n.Target.Name), nil
	}

	return types.FullName("tool." + n.Target.Name), nil
}
