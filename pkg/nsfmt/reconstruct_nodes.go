// NeuroScript Version: 0.6.0
// File version: 18
// Purpose: Implemented precise newline calculation in the rendering loop to correctly preserve spacing between top-level blocks.
// filename: pkg/nsfmt/reconstructor_nodes.go
// nlines: 191
// risk_rating: HIGH

package nsfmt

import (
	"fmt"
	"sort"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// endNode is a synthetic node used to mark the end of a block in the timeline.
type endNode struct {
	*ast.BaseNode
	Keyword string // e.g., "endfunc", "endcommand"
}

func (e *endNode) String() string { return "endNode" }

// pos-related helpers that use the methods on the Node interface.
func getNodePos(node ast.Node) *types.Position {
	if node == nil {
		return nil
	}
	return node.GetPos()
}

func getNodeEndPos(node ast.Node) *types.Position {
	if node == nil {
		return nil
	}
	if n, ok := node.(interface{ End() *types.Position }); ok {
		return n.End()
	}
	return nil
}

// reconstructProgram builds a unified timeline of all nodes and comments.
func (r *reconstructor) reconstructProgram(prog *ast.Program) {
	var timeline []ast.Node

	// 1. Collect all nodes and comments in a single recursive pass.
	var collect func(node ast.Node)
	collect = func(node ast.Node) {
		if node == nil {
			return
		}

		switch n := node.(type) {
		case *ast.Program:
			for _, c := range n.Comments {
				timeline = append(timeline, c)
			}
			keys := make([]string, 0, len(n.Procedures))
			for k := range n.Procedures {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				collect(n.Procedures[k])
			}
			for _, event := range n.Events {
				collect(event)
			}
			for _, cmd := range n.Commands {
				collect(cmd)
			}
		case *ast.Procedure:
			timeline = append(timeline, n)
			for _, c := range n.Comments {
				timeline = append(timeline, c)
			}
			for i := range n.Steps {
				collect(&n.Steps[i])
			}
			if endPos := getNodeEndPos(n); endPos != nil {
				timeline = append(timeline, &endNode{BaseNode: &ast.BaseNode{StartPos: endPos}, Keyword: "endfunc"})
			}
		case *ast.CommandNode:
			timeline = append(timeline, n)
			for _, c := range n.Comments {
				timeline = append(timeline, c)
			}
			for i := range n.Body {
				collect(&n.Body[i])
			}
			if endPos := getNodeEndPos(n); endPos != nil {
				timeline = append(timeline, &endNode{BaseNode: &ast.BaseNode{StartPos: endPos}, Keyword: "endcommand"})
			}
		case *ast.OnEventDecl:
			timeline = append(timeline, n)
			for i := range n.Body {
				collect(&n.Body[i])
			}
			if endPos := getNodeEndPos(n); endPos != nil {
				timeline = append(timeline, &endNode{BaseNode: &ast.BaseNode{StartPos: endPos}, Keyword: "endon"})
			}
		case *ast.Step:
			timeline = append(timeline, n)
			for _, c := range n.Comments {
				timeline = append(timeline, c)
			}
		}
	}
	collect(prog)

	// 2. Sort the timeline based on position.
	sort.SliceStable(timeline, func(i, j int) bool {
		posI, posJ := getNodePos(timeline[i]), getNodePos(timeline[j])
		if posI == nil || posJ == nil {
			return false
		}
		if posI.Line != posJ.Line {
			return posI.Line < posJ.Line
		}
		return posI.Column < posJ.Column
	})

	// 3. Reconstruct from the sorted timeline.
	lastLine := 0
	for i := 0; i < len(timeline); i++ {
		node := timeline[i]
		pos := getNodePos(node)
		if pos == nil {
			continue
		}

		// *** FIX IS HERE ***
		// Precisely calculate and insert the correct number of blank lines.
		if lastLine > 0 && pos.Line > lastLine {
			for k := 0; k < pos.Line-lastLine-1; k++ {
				r.writeln("")
			}
		}

		switch n := node.(type) {
		case *ast.Comment:
			r.writeln(n.Text)
		case *ast.Procedure:
			r.writeln(fmt.Sprintf("func %s() means", n.Name()))
			r.indent++
		case *ast.CommandNode:
			r.writeln("command")
			r.indent++
		case *ast.OnEventDecl:
			r.writeln(fmt.Sprintf("on event %s", r.reconstructExpression(n.EventNameExpr)))
			r.indent++
		case *ast.Step:
			trailingComment := ""
			if i+1 < len(timeline) {
				if nextNode, ok := timeline[i+1].(*ast.Comment); ok {
					if getNodePos(nextNode).Line == pos.Line {
						trailingComment = " " + nextNode.Text
						i++ // Consume the trailing comment.
					}
				}
			}
			r.reconstructStep(n, trailingComment)
		case *endNode:
			r.indent--
			r.writeln(n.Keyword)
		}
		lastLine = pos.Line
	}
}

// reconstructStep is simplified to only render the step itself.
func (r *reconstructor) reconstructStep(step *ast.Step, trailingComment string) {
	switch step.Type {
	case "set":
		lval := r.reconstructExpression(step.LValues[0])
		rval := r.reconstructExpression(step.Values[0])
		r.writeln(fmt.Sprintf("set %s = %s%s", lval, rval, trailingComment))
	default:
		r.writeln(fmt.Sprintf("# TODO: Reconstruct step type '%s'", step.Type))
	}
}

func (r *reconstructor) reconstructExpression(expr ast.Expression) string {
	switch n := expr.(type) {
	case *ast.LValueNode:
		return n.Identifier
	case *ast.VariableNode:
		return n.Name
	case *ast.NumberLiteralNode:
		return fmt.Sprintf("%g", n.Value)
	case *ast.StringLiteralNode:
		return fmt.Sprintf("%q", n.Value)
	case *ast.NilLiteralNode:
		return "nil"
	default:
		if n == nil {
			return "<nil_expr>"
		}
		return fmt.Sprintf("TODO_EXPR(%s)", n.String())
	}
}
