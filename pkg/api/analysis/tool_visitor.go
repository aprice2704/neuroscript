// NeuroScript Version: 0.8.0
// File version: 5
// Purpose: Corrects visitor guard logic to check struct fields directly, not compare the struct to nil.
// filename: pkg/api/analysis/tool_visitor.go
// nlines: 105+
// risk_rating: HIGH

package analysis

import (
	"reflect"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
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
		for _, expr := range n.Expressions {
			v.visit(expr)
		}
	case *ast.CommandNode:
		for i := range n.Body {
			v.visitStep(&n.Body[i])
		}
		for i := range n.ErrorHandlers {
			v.visitStep(n.ErrorHandlers[i])
		}
	case *ast.Procedure:
		for i := range n.Steps {
			v.visitStep(&n.Steps[i])
		}
		for i := range n.ErrorHandlers {
			v.visitStep(n.ErrorHandlers[i])
		}
	case *ast.OnEventDecl:
		for i := range n.Body {
			v.visitStep(&n.Body[i])
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
	for i := range step.Body {
		v.visitStep(&step.Body[i])
	}
	for i := range step.ElseBody {
		v.visitStep(&step.ElseBody[i])
	}
}

// visitExpression checks if an expression is a tool call and recurses.
func (v *toolVisitor) visitExpression(expr ast.Expression) {
	// Check for both nil interface and nil underlying value to prevent panics.
	if expr == nil || (reflect.ValueOf(expr).Kind() == reflect.Ptr && reflect.ValueOf(expr).IsNil()) {
		return
	}

	// Check if this expression IS a tool call
	if callExpr, ok := expr.(*ast.CallableExprNode); ok {
		// FIX: Check the fields of the Target struct directly. Do not compare the struct to nil.
		if callExpr.Target.IsTool && callExpr.Target.Name != "" {
			v.requiredTools[callExpr.Target.Name] = struct{}{}
		}
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
	}
}
