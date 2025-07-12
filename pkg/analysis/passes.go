// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Implements the initial set of built-in static analysis passes.
// filename: pkg/analysis/passes.go
// nlines: 40
// risk_rating: MEDIUM

package analysis

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// ShapePass checks for basic structural and syntactic conventions in the AST.
type ShapePass struct{}

// Name returns the unique name of the analysis pass.
func (p *ShapePass) Name() string {
	return "shape"
}

// Analyse checks for structural issues, like empty command blocks.
func (p *ShapePass) Analyse(tree *interfaces.Tree) []interfaces.Diag {
	var diags []interfaces.Diag
	if tree == nil || tree.Root == nil {
		return nil
	}

	// ast.Walk is a hypothetical helper to visit all nodes.
	// We would build this utility to make traversal easier.
	ast.Walk(tree.Root, func(n interfaces.Node) bool {
		if cmd, ok := n.(*ast.CommandNode); ok {
			if len(cmd.Body) == 0 {
				diags = append(diags, interfaces.Diag{
					Severity: interfaces.SeverityError,
					Position: cmd.GetPos(),
					Message:  "Command block must not be empty.",
					Source:   p.Name(),
				})
			}
		}
		return true // Continue traversal
	})

	return diags
}

// init registers the built-in analysis passes.
func init() {
	RegisterPass(&ShapePass{})
}
