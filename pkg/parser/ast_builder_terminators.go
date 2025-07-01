// filename: pkg/core/ast_builder_terminators.go
package parser

import (
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// --- Exit methods for Primary ast.Expressions, Literals, Placeholders ---
// *** MODIFIED: Create specific AST nodes, set Pos, push nodes, add error handling ***

// Enter.Expression is required to satisfy the listener interface.
// It does not need to perform any action as the child rules handle all the logic.
func (l *neuroScriptListenerImpl) Enter.Expression(ctx *gen.ast.ExpressionContext) {
	l.logDebugAST("--- Enter ast.Expression: %q (Pass through)", ctx.GetText())
}

// Exit.Expression is just a pass-through in the listener for the top-level expression rule.
// The actual ast.Expression node will be pushed by one of its children (e.g., ExitLogical_or_expr).
func (l *neuroScriptListenerImpl) Exit.Expression(ctx *gen.ast.ExpressionContext) {
	l.logDebugAST("--- Exit ast.Expression: %q (Pass through)", ctx.GetText())
	// No lang.Value pushed here; lang.Value comes from child (logical_or_expr)
}

// --- REMOVED Duplicate ExitAccessor_expr method ---
// The implementation now resides in ast_builder_operators.go

// ExitPrimary handles the base cases of expressions.
func (l *neuroScriptListenerImpl) ExitPrimary(ctx *gen.PrimaryContext) {
	l.logDebugAST("--- Exit Primary: %q", ctx.GetText())
	if ctx.Literal() != nil || ctx.Placeholder() != nil || ctx.Callable_expr() != nil || ctx.LPAREN() != nil {
		// lang.Value already pushed by the corresponding child Exit* method (or passed through for parens).
		l.logDebugAST("    Primary is Literal, Placeholder, Call, or Parenthesized (Pass through)")
		// If it's LPAREN expression RPAREN, the lang.Value from the inner expression is already on the stack.
		return
	}

	var nodeToPush ast.Expression	// Node to push if handled directly here

	if ctx.IDENTIFIER() != nil {
		token := ctx.IDENTIFIER().GetSymbol()
		node := &ast.VariableNode{
			Position:	tokenTolang.Position(token),
			Name:	token.GetText(),
		}
		nodeToPush = node
		l.logDebugAST("    Constructed ast.VariableNode: %s", node.Name)

	} else if ctx.KW_LAST() != nil {
		token := ctx.KW_LAST().GetSymbol()
		node := &ast.EvalNode{
			Position: tokenTolang.Position(token),
		}
		nodeToPush = node
		l.logDebugAST("    Constructed ast.EvalNode")

	} else if ctx.KW_EVAL() != nil {
		token := ctx.KW_EVAL().GetSymbol()	// lang.Position of 'eval' keyword
		// Pop the argument expression pushed by visiting ctx.ast.Expression()
		argRaw, ok := l.poplang.Value()
		if !ok {
			l.addError(ctx, "Internal error: Failed to pop argument for EVAL")
			l.pushlang.Value(nil)	// Push error marker
			return
		}
		argExpr, ok := argRaw.(ast.Expression)
		if !ok {
			l.addError(ctx, "Internal error: Argument for EVAL is not an ast.Expression (got %T)", argRaw)
			l.pushlang.Value(nil)	// Push error marker
			return
		}
		node := &EvalNode{
			Position:		tokenTolang.Position(token),
			Argument:	argExpr,
		}
		nodeToPush = node
		l.logDebugAST("    Constructed EvalNode")

	} else {
		// Should not happen if grammar is correct
		l.addError(ctx, "Internal error: ExitPrimary reached unexpected state for text: %q", ctx.GetText())
		l.pushlang.Value(nil)	// Push error marker
		return
	}

	// Push the node created directly in this method
	l.pushlang.Value(nodeToPush)
}

// ExitPlaceholder builds a ast.Placeholder.Node (e.g., {{var}} or {{LAST}}).
func (l *neuroScriptListenerImpl) ExitPlaceholder(ctx *gen.PlaceholderContext) {
	l.logDebugAST("--- Exit Placeholder: %q", ctx.GetText())
	name := ""
	token := ctx.GetStart()	// lang.Position of '{{'

	if ctx.IDENTIFIER() != nil {
		name = ctx.IDENTIFIER().GetText()
	} else if ctx.KW_LAST() != nil {
		name = "LAST"	// Use canonical name
	} else {
		// Should not happen based on grammar rule
		l.addErrorf(token, "Internal error: ExitPlaceholder found unexpected content: %q", ctx.GetText())
		l.pushlang.Value(nil)	// Push error marker
		return
	}

	node := &ast.Placeholder.Node{
		Position:	tokenTolang.Position(token),
		Name:	name,
	}
	l.pushlang.Value(node)
	l.logDebugAST("    Constructed ast.Placeholder.Node: Name=%s", node.Name)
}
