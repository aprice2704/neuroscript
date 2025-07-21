// NeuroScript Version: 0.6.0
// File version: 4
// Purpose: Provides the main entry point and core types for the reconstructor.
// filename: pkg/nsfmt/reconstructor.go
// nlines: 65
// risk_rating: LOW

package nsfmt

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// Reconstruct takes an AST and rebuilds the equivalent source code string.
func Reconstruct(tree *ast.Tree) (string, error) {
	if tree == nil || tree.Root == nil {
		return "", fmt.Errorf("cannot reconstruct from a nil tree or root")
	}

	r := &reconstructor{
		builder: &strings.Builder{},
		indent:  0,
	}

	program, ok := tree.Root.(*ast.Program)
	if !ok {
		return "", fmt.Errorf("tree root is not an *ast.Program node")
	}

	r.reconstructProgram(program)

	return r.builder.String(), nil
}

type reconstructor struct {
	builder        *strings.Builder
	indent         int
	lineHasContent bool
}

// write handles writing strings and managing indentation.
func (r *reconstructor) write(s string) {
	if !r.lineHasContent && r.indent > 0 && s != "" {
		r.builder.WriteString(strings.Repeat("\t", r.indent))
	}
	r.builder.WriteString(s)
	if s != "" {
		r.lineHasContent = true
	}
}

// writeln handles writing a string followed by a newline.
func (r *reconstructor) writeln(s string) {
	if s != "" || r.lineHasContent {
		r.write(s)
		r.builder.WriteString("\n")
		r.lineHasContent = false
	} else {
		r.builder.WriteString("\n")
	}
}
