// filename: pkg/parser/ast_builder_terminators.go
// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Refactored primary expression node creation to use the newNode helper.
package parser

import (
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// --- Exit methods for Primary Expressions, Literals, Placeholders ---

// EnterExpression is required to satisfy the listener interface.
func (l *neuroScriptListenerImpl) EnterExpression(ctx *gen.ExpressionContext) {
	l.logDebugAST("--- Enter Expression: %q (Pass through)", ctx.GetText())
}

// ExitExpression is a pass-through in the listener for the top-level expression rule.
func (l *neuroScriptListenerImpl) ExitExpression(ctx *gen.ExpressionContext) {
	l.logDebugAST("--- Exit Expression: %q (Pass through)", ctx.GetText())
}

// ExitPrimary handles the base cases of expressions.
func (l *neuroScriptListenerImpl) ExitPrimary(ctx *gen.PrimaryContext) {
	l.logDebugAST("--- Exit Primary: %q", ctx.GetText())
	if ctx.Literal() != nil || ctx.Placeholder() != nil || ctx.Callable_expr() != nil || ctx.LPAREN() != nil {
		// Value already pushed by the corresponding child Exit* method.
		l.logDebugAST("    Primary is Literal, Placeholder, Call, or Parenthesized (Pass through)")
		return
	}

	var nodeToPush ast.Expression

	if ctx.IDENTIFIER() != nil {
		token := ctx.IDENTIFIER().GetSymbol()
		node := &ast.VariableNode{Name: token.GetText()}
		nodeToPush = newNode(node, token, types.KindVariable)
		l.logDebugAST("    Constructed ast.VariableNode: %s", node.Name)

	} else if ctx.KW_LAST() != nil {
		token := ctx.KW_LAST().GetSymbol()
		node := &ast.LastNode{}
		nodeToPush = newNode(node, token, types.KindLastResult)
		l.logDebugAST("    Constructed ast.LastNode")

	} else if ctx.KW_EVAL() != nil {
		token := ctx.KW_EVAL().GetSymbol()
		argRaw, ok := l.pop()
		if !ok {
			l.addError(ctx, "Internal error: Failed to pop argument for EVAL")
			l.push(newNode(&ast.ErrorNode{Message: "Stack underflow (eval)"}, token, types.KindUnknown))
			return
		}
		argExpr, ok := argRaw.(ast.Expression)
		if !ok {
			l.addError(ctx, "Internal error: Argument for EVAL is not an ast.Expression (got %T)", argRaw)
			l.push(newNode(&ast.ErrorNode{Message: "Type error (eval)"}, token, types.KindUnknown))
			return
		}
		node := &ast.EvalNode{Argument: argExpr}
		nodeToPush = newNode(node, token, types.KindEvalExpr)
		l.logDebugAST("    Constructed EvalNode")

	} else {
		l.addError(ctx, "Internal error: ExitPrimary reached unexpected state for text: %q", ctx.GetText())
		l.push(newNode(&ast.ErrorNode{Message: "Unknown primary expression"}, ctx.GetStart(), types.KindUnknown))
		return
	}

	l.push(nodeToPush)
}

// ExitPlaceholder builds a ast.PlaceholderNode (e.g., {{var}} or {{LAST}}).
func (l *neuroScriptListenerImpl) ExitPlaceholder(ctx *gen.PlaceholderContext) {
	l.logDebugAST("--- Exit Placeholder: %q", ctx.GetText())
	name := ""
	token := ctx.GetStart()

	if ctx.IDENTIFIER() != nil {
		name = ctx.IDENTIFIER().GetText()
	} else if ctx.KW_LAST() != nil {
		name = "LAST" // Use canonical name
	} else {
		l.addErrorf(token, "Internal error: ExitPlaceholder found unexpected content: %q", ctx.GetText())
		l.push(newNode(&ast.ErrorNode{Message: "Malformed placeholder"}, token, types.KindUnknown))
		return
	}

	node := &ast.PlaceholderNode{Name: name}
	l.push(newNode(node, token, types.KindPlaceholder))
	l.logDebugAST("    Constructed ast.PlaceholderNode: Name=%s", node.Name)
}
