// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Implements the initial set of built-in static analysis passes with a direct AST traversal.
// filename: pkg/analysis/passes.go
// nlines: 50
// risk_rating: MEDIUM

package analysis

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// ShapePass checks for basic structural and syntactic conventions in the AST.
type ShapePass struct{}

// Name returns the unique name of the analysis pass.
func (p *ShapePass) Name() string {
	return "shape"
}

// Analyse checks for structural issues, like empty command blocks.
func (p *ShapePass) Analyse(tree *interfaces.Tree) []types.Diag {
	var diags []types.Diag
	if tree == nil || tree.Root == nil {
		return nil
	}

	// This is a direct, manual traversal of the AST.
	// We can replace this with a generic ast.Walk utility later.
	if program, ok := tree.Root.(*ast.Program); ok {
		for _, cmd := range program.Commands {
			if len(cmd.Body) == 0 {
				diags = append(diags, types.Diag{
					Severity: types.SeverityError,
					Position: cmd.GetPos(),
					Message:  "Command block must not be empty.",
					Source:   p.Name(),
				})
			}
		}
	}

	return diags
}

// init registers the built-in analysis passes.
func init() {
	RegisterPass(&ShapePass{})
}
