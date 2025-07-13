// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Updated to use the centralized types.Diag struct and corrected formatting.
// filename: pkg/api/analysis/pass.go
// nlines: 75
// risk_rating: HIGH

package analysis

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Pass is the interface that all static analysis passes must implement.
type Pass interface {
	Name() string
	Analyse(tree *interfaces.Tree) []types.Diag
}

var registeredPasses []Pass

// RegisterPass adds a new analysis pass to the global registry.
func RegisterPass(p Pass) {
	registeredPasses = append(registeredPasses, p)
}

// Vet runs all registered analysis passes on the given AST.
func Vet(tree *interfaces.Tree) []types.Diag {
	var allDiags []types.Diag
	for _, pass := range registeredPasses {
		diags := pass.Analyse(tree)
		allDiags = append(allDiags, diags...)
	}
	return allDiags
}

// --- ShapeValidator Pass Implementation ---

// ShapeValidatorPass checks for basic structural invariants in the AST.
type ShapeValidatorPass struct{}

func (p *ShapeValidatorPass) Name() string { return "shape-validator" }

func (p *ShapeValidatorPass) Analyse(tree *interfaces.Tree) []types.Diag {
	if tree == nil || tree.Root == nil {
		return nil
	}
	visitor := &shapeVisitor{diags: []types.Diag{}}
	visitor.visit(tree.Root, false)
	return visitor.diags
}

// shapeVisitor recursively walks the AST.
type shapeVisitor struct {
	diags []types.Diag
}

func (v *shapeVisitor) visit(node interfaces.Node, inCommand bool) {
	if node == nil {
		return
	}

	// Rule: Disallow nested command blocks.
	if node.Kind() == types.KindCommandBlock {
		if inCommand {
			v.diags = append(v.diags, types.Diag{
				Position: node.GetPos(),
				Severity: types.SeverityError,
				Source:   "shape-validator",
				Message:  "nested command blocks are not allowed",
			})
		}
		inCommand = true
	}

	// Recurse into children. This is a simplified example.
	// A full implementation would iterate over all relevant child nodes.
	if program, ok := node.(*ast.Program); ok {
		for _, cmd := range program.Commands {
			v.visit(cmd, inCommand)
		}
		for _, proc := range program.Procedures {
			v.visit(proc, inCommand)
		}
	}
}

// Automatically register the built-in passes.
func init() {
	RegisterPass(&ShapeValidatorPass{})
}
