// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Implements an AST visitor to find all tool calls within a script.
// filename: pkg/api/analysis/tool_visitor.go
// nlines: 105
// risk_rating: MEDIUM

package analysis

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// toolVisitor walks the AST and collects the names of all referenced tools.
type toolVisitor struct {
	requiredTools map[string]struct{} // Using a map as a set for unique names
}

// FindRequiredTools analyzes the AST and returns a set of unique tool names used.
func FindRequiredTools(tree *interfaces.Tree) map[string]struct{} {
	if tree == nil || tree.Root == nil {
		return nil
	}
	visitor := &toolVisitor{
		requiredTools: make(map[string]struct{}),
	}
	visitor.visit(tree.Root)
	return visitor.requiredTools
}

// visit recursively traverses the AST nodes.
func (v *toolVisitor) visit(node interfaces.Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.Program:
		for _, cmd := range n.Commands {
			v.visit(cmd)
		}
		for _, proc := range n.Procedures {
			v.visit(proc)
		}
		for _, event := range n.Events {
			v.visit(event)
		}
		// Also visit top-level expressions if they can contain calls (less common)
		for _, expr := range n.Expressions {
			v.visit(expr)
		}
	case *ast.CommandNode:
		for _, step := range n.Body {
			v.visitStep(&step)
		}
		for _, handler := range n.ErrorHandlers {
			v.visitStep(handler)
		}
	case *ast.Procedure:
		for _, step := range n.Steps {
			v.visitStep(&step)
		}
		for _, handler := range n.ErrorHandlers {
			v.visitStep(handler)
		}
	case *ast.OnEventDecl:
		for _, step := range n.Body {
			v.visitStep(&step)
		}
	case ast.Expression: // Handle expressions directly
		v.visitExpression(n)
	default:
		// Handle other container node types if necessary
	}
}

// visitStep specifically handles the structure within a Step.
func (v *toolVisitor) visitStep(step *ast.Step) {
	if step == nil {
		return
	}
	// Visit expressions within the step
	v.visitExpression(step.Cond)
	v.visitExpression(step.Collection)
	v.visitExpression(step.Call) // Specifically check the Call field
	if step.AskStmt != nil {
		v.visitExpression(step.AskStmt.AgentModelExpr)
		v.visitExpression(step.AskStmt.PromptExpr)
		v.visitExpression(step.AskStmt.WithOptions)
	}
	if step.PromptUserStmt != nil {
		v.visitExpression(step.PromptUserStmt.PromptExpr)
	}
	if step.WhisperStmt != nil {
		v.visitExpression(step.WhisperStmt.Handle)
		v.visitExpression(step.WhisperStmt.Value)
	}
	if step.ExpressionStmt != nil {
		v.visitExpression(step.ExpressionStmt.Expression)
	}
	for _, val := range step.Values { // RHS of set, return values, etc.
		v.visitExpression(val)
	}

	// Recurse into nested bodies
	for _, subStep := range step.Body {
		v.visitStep(&subStep)
	}
	for _, subStep := range step.ElseBody {
		v.visitStep(&subStep)
	}
}

// visitExpression checks if an expression is a tool call and recurses.
func (v *toolVisitor) visitExpression(expr ast.Expression) {
	if expr == nil {
		return
	}

	// Check if this expression IS a tool call
	if callExpr, ok := expr.(*ast.CallableExprNode); ok && callExpr.Target.IsTool {
		toolName := types.MakeFullName(callExpr.Target.Name, "") // Ensure canonical form
		v.requiredTools[string(toolName)] = struct{}{}
	}

	// Recurse into sub-expressions
	switch e := expr.(type) {
	case *ast.CallableExprNode:
		for _, arg := range e.Arguments {
			v.visitExpression(arg)
		}
	case *ast.ListLiteralNode:
		for _, elem := range e.Elements {
			v.visitExpression(elem)
		}
	case *ast.MapLiteralNode:
		for _, entry := range e.Entries {
			if entry != nil {
				v.visitExpression(entry.Value) // Key is always string literal
			}
		}
	case *ast.ElementAccessNode:
		v.visitExpression(e.Collection)
		v.visitExpression(e.Accessor)
	case *ast.UnaryOpNode:
		v.visitExpression(e.Operand)
	case *ast.BinaryOpNode:
		v.visitExpression(e.Left)
		v.visitExpression(e.Right)
	case *ast.EvalNode:
		v.visitExpression(e.Argument)
	case *ast.TypeOfNode:
		v.visitExpression(e.Argument)
		// Other node types (literals, variables, etc.) don't contain further expressions.
	}
}
