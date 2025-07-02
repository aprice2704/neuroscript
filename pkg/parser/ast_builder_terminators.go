// filename: pkg/parser/ast_builder_terminators.go
package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// --- Exit methods for Primary Expressions, Literals, Placeholders ---
// *** MODIFIED: Create specific AST nodes, set Pos, push nodes, add error handling ***

// EnterExpression is required to satisfy the listener interface.
// It does not need to perform any action as the child rules handle all the logic.
func (l *neuroScriptListenerImpl) EnterExpression(ctx *gen.ExpressionContext) {
	l.logDebugAST("--- Enter Expression: %q (Pass through)", ctx.GetText())
}

// ExitExpression is just a pass-through in the listener for the top-level expression rule.
// The actual Expression node will be pushed by one of its children (e.g., ExitLogical_or_expr).
func (l *neuroScriptListenerImpl) ExitExpression(ctx *gen.ExpressionContext) {
	l.logDebugAST("--- Exit Expression: %q (Pass through)", ctx.GetText())
	// No value pushed here; value comes from child (logical_or_expr)
}

// --- REMOVED Duplicate ExitAccessor_expr method ---
// The implementation now resides in ast_builder_operators.go

// ExitPrimary handles the base cases of expressions.
func (l *neuroScriptListenerImpl) ExitPrimary(ctx *gen.PrimaryContext) {
	l.logDebugAST("--- Exit Primary: %q", ctx.GetText())
	if ctx.Literal() != nil || ctx.Placeholder() != nil || ctx.Callable_expr() != nil || ctx.LPAREN() != nil {
		// Value already pushed by the corresponding child Exit* method (or passed through for parens).
		l.logDebugAST("    Primary is Literal, Placeholder, Call, or Parenthesized (Pass through)")
		// If it's LPAREN expression RPAREN, the value from the inner expression is already on the stack.
		return
	}

	var nodeToPush ast.Expression // Node to push if handled directly here

	if ctx.IDENTIFIER() != nil {
		token := ctx.IDENTIFIER().GetSymbol()
		pos := tokenToPosition(token)
		node := &ast.VariableNode{
			Pos:  &pos,
			Name: token.GetText(),
		}
		nodeToPush = node
		l.logDebugAST("    Constructed ast.VariableNode: %s", node.Name)

	} else if ctx.KW_LAST() != nil {
		token := ctx.KW_LAST().GetSymbol()
		pos := tokenToPosition(token)
		node := &ast.LastNode{
			Pos: &pos,
		}
		nodeToPush = node
		l.logDebugAST("    Constructed ast.LastNode")

	} else if ctx.KW_EVAL() != nil {
		token := ctx.KW_EVAL().GetSymbol() // Position of 'eval' keyword
		pos := tokenToPosition(token)
		// Pop the argument expression pushed by visiting ctx.Expression()
		argRaw, ok := l.pop()
		if !ok {
			l.addError(ctx, "Internal error: Failed to pop argument for EVAL")
			l.push(nil) // Push error marker
			return
		}
		argExpr, ok := argRaw.(ast.Expression)
		if !ok {
			l.addError(ctx, "Internal error: Argument for EVAL is not an ast.Expression (got %T)", argRaw)
			l.push(nil) // Push error marker
			return
		}
		node := &ast.EvalNode{
			Pos:      &pos,
			Argument: argExpr,
		}
		nodeToPush = node
		l.logDebugAST("    Constructed EvalNode")

	} else {
		// Should not happen if grammar is correct
		l.addError(ctx, "Internal error: ExitPrimary reached unexpected state for text: %q", ctx.GetText())
		l.push(nil) // Push error marker
		return
	}

	// Push the node created directly in this method
	l.push(nodeToPush)
}

// ExitPlaceholder builds a ast.PlaceholderNode (e.g., {{var}} or {{LAST}}).
func (l *neuroScriptListenerImpl) ExitPlaceholder(ctx *gen.PlaceholderContext) {
	l.logDebugAST("--- Exit Placeholder: %q", ctx.GetText())
	name := ""
	token := ctx.GetStart() // Position of '{{'
	pos := tokenToPosition(token)

	if ctx.IDENTIFIER() != nil {
		name = ctx.IDENTIFIER().GetText()
	} else if ctx.KW_LAST() != nil {
		name = "LAST" // Use canonical name
	} else {
		// Should not happen based on grammar rule
		l.addErrorf(token, "Internal error: ExitPlaceholder found unexpected content: %q", ctx.GetText())
		l.push(nil) // Push error marker
		return
	}

	node := &ast.PlaceholderNode{
		Pos:  &pos,
		Name: name,
	}
	l.push(node)
	l.logDebugAST("    Constructed ast.PlaceholderNode: Name=%s", node.Name)
}
