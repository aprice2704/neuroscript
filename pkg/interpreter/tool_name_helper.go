// NeuroScript Version: 0.8.0
// File version: 3.0.0
// Purpose: Provides a centralized helper to resolve canonical tool names.
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
// NOTE: This is a candidate for moving to the 'eval' package.
func resolveToolName(n *ast.CallableExprNode) (types.FullName, error) {
	if !n.Target.IsTool {
		return "", fmt.Errorf("internal error: resolveToolName called on a non-tool expression")
	}
	if n.Target.Name == "" {
		return "", fmt.Errorf("internal error: tool call expression has an empty target name")
	}
	if strings.HasPrefix(n.Target.Name, "tool.") {
		return CanonicalToolName(n.Target.Name), nil
	}
	return CanonicalToolName("tool." + n.Target.Name), nil
}
