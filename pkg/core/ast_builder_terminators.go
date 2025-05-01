// filename: pkg/core/ast_builder_terminators.go
package core

import (
	// "strconv" // Now handled in helpers
	"strings"

	"github.com/antlr4-go/antlr/v4" // Ensure antlr is imported
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Exit methods for Primary Expressions, Literals, Placeholders ---
// *** MODIFIED: Create specific AST nodes, set Pos, push nodes, add error handling ***

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

	var nodeToPush Expression // Node to push if handled directly here

	if ctx.IDENTIFIER() != nil {
		token := ctx.IDENTIFIER().GetSymbol()
		node := &VariableNode{
			Pos:  tokenToPosition(token),
			Name: token.GetText(),
		}
		nodeToPush = node
		l.logDebugAST("    Constructed VariableNode: %s", node.Name)

	} else if ctx.KW_LAST() != nil {
		token := ctx.KW_LAST().GetSymbol()
		node := &LastNode{
			Pos: tokenToPosition(token),
		}
		nodeToPush = node
		l.logDebugAST("    Constructed LastNode")

	} else if ctx.KW_EVAL() != nil {
		token := ctx.KW_EVAL().GetSymbol() // Position of 'eval' keyword
		// Pop the argument expression pushed by visiting ctx.Expression()
		argRaw, ok := l.popValue()
		if !ok {
			l.addError(ctx, "Internal error: Failed to pop argument for EVAL")
			l.pushValue(nil) // Push error marker
			return
		}
		argExpr, ok := argRaw.(Expression)
		if !ok {
			l.addError(ctx, "Internal error: Argument for EVAL is not an Expression (got %T)", argRaw)
			l.pushValue(nil) // Push error marker
			return
		}
		node := &EvalNode{
			Pos:      tokenToPosition(token),
			Argument: argExpr,
		}
		nodeToPush = node
		l.logDebugAST("    Constructed EvalNode")

	} else {
		// Should not happen if grammar is correct
		l.addError(ctx, "Internal error: ExitPrimary reached unexpected state for text: %q", ctx.GetText())
		l.pushValue(nil) // Push error marker
		return
	}

	// Push the node created directly in this method
	l.pushValue(nodeToPush)
}

// ExitPlaceholder builds a PlaceholderNode (e.g., {{var}} or {{LAST}}).
func (l *neuroScriptListenerImpl) ExitPlaceholder(ctx *gen.PlaceholderContext) {
	l.logDebugAST("--- Exit Placeholder: %q", ctx.GetText())
	name := ""
	token := ctx.GetStart() // Position of '{{'

	if ctx.IDENTIFIER() != nil {
		name = ctx.IDENTIFIER().GetText()
	} else if ctx.KW_LAST() != nil {
		name = "LAST" // Use canonical name
	} else {
		// Should not happen based on grammar rule
		l.addErrorf(token, "Internal error: ExitPlaceholder found unexpected content: %q", ctx.GetText())
		l.pushValue(nil) // Push error marker
		return
	}

	node := &PlaceholderNode{
		Pos:  tokenToPosition(token),
		Name: name,
	}
	l.pushValue(node)
	l.logDebugAST("    Constructed PlaceholderNode: Name=%s", node.Name)
}

// ExitLiteral handles different types of literals.
func (l *neuroScriptListenerImpl) ExitLiteral(ctx *gen.LiteralContext) {
	l.logDebugAST("--- Exit Literal: %q", ctx.GetText())

	// Handle specific literal types first
	if ctx.STRING_LIT() != nil {
		token := ctx.STRING_LIT().GetSymbol()
		quotedStr := token.GetText()
		unquotedVal, err := unescapeString(quotedStr) // Use helper
		if err != nil {
			l.addErrorf(token, "Invalid string literal: %v", err)
			l.pushValue(nil) // Push error marker
			return
		}
		node := &StringLiteralNode{
			Pos:   tokenToPosition(token),
			Value: unquotedVal,
			IsRaw: false,
		}
		l.pushValue(node)
		l.logDebugAST("    Constructed StringLiteralNode (Quoted)")
		return

	} else if ctx.TRIPLE_BACKTICK_STRING() != nil {
		token := ctx.TRIPLE_BACKTICK_STRING().GetSymbol()
		rawContent := token.GetText()
		// Remove the ``` delimiters
		var actualContent string
		if len(rawContent) >= 6 && strings.HasPrefix(rawContent, "```") && strings.HasSuffix(rawContent, "```") {
			actualContent = rawContent[3 : len(rawContent)-3]
		} else {
			l.addErrorf(token, "Invalid triple-backtick string format: %q", rawContent)
			l.pushValue(nil) // Push error marker
			return
		}
		node := &StringLiteralNode{
			Pos:   tokenToPosition(token),
			Value: actualContent,
			IsRaw: true,
		}
		l.pushValue(node)
		l.logDebugAST("    Constructed StringLiteralNode (Triple-Backtick/Raw)")
		return

	} else if ctx.NUMBER_LIT() != nil {
		token := ctx.NUMBER_LIT().GetSymbol()
		numStr := token.GetText()
		numValue, err := parseNumber(numStr) // Use helper
		if err != nil {
			l.addErrorf(token, "Invalid number literal: %v", err)
			l.pushValue(nil) // Push error marker
			return
		}
		node := &NumberLiteralNode{
			Pos:   tokenToPosition(token),
			Value: numValue, // Holds int64 or float64
		}
		l.pushValue(node)
		l.logDebugAST("    Constructed NumberLiteralNode: Value=%v (%T)", node.Value, node.Value)
		return
	}

	// If it's not one of the above, it must be a boolean, list, or map literal.
	// The values for these should have already been pushed onto the stack by their
	// respective Exit* methods (ExitBoolean_literal, ExitList_literal, ExitMap_literal).
	// So, here we just act as a pass-through.
	if ctx.Boolean_literal() != nil || ctx.List_literal() != nil || ctx.Map_literal() != nil {
		l.logDebugAST("    Literal is Boolean, List, or Map (Pass through)")
		// Value already pushed by child Exit* method.
		return
	}

	// If none of the above matched, something is wrong.
	l.addError(ctx, "Internal error: ExitLiteral reached unexpected state - no known literal type found for text: %q", ctx.GetText())
	l.pushValue(nil) // Push nil as an error marker
}

// ExitBoolean_literal pushes a BooleanLiteralNode.
func (l *neuroScriptListenerImpl) ExitBoolean_literal(ctx *gen.Boolean_literalContext) {
	l.logDebugAST("--- Exit Boolean_literal: %q", ctx.GetText())
	value := false
	var token antlr.Token
	if ctx.KW_TRUE() != nil {
		value = true
		token = ctx.KW_TRUE().GetSymbol()
	} else {
		token = ctx.KW_FALSE().GetSymbol()
	}

	node := &BooleanLiteralNode{
		Pos:   tokenToPosition(token),
		Value: value,
	}
	l.pushValue(node)
	l.logDebugAST("    Constructed BooleanLiteralNode: Value=%t", node.Value)
}
