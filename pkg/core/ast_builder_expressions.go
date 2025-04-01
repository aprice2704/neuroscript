// pkg/core/ast_builder_expressions.go
package core

import (
	"strconv"
	"strings"

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Handling Expression Components ---

// ExitLiteral pushes the specific literal node type onto the stack
func (l *neuroScriptListenerImpl) ExitLiteral(ctx *gen.LiteralContext) {
	l.logDebugAST(">>> Exit Literal: %q", ctx.GetText())
	if ctx.STRING_LIT() != nil {
		strContent := ctx.STRING_LIT().GetText()
		unquoted, err := strconv.Unquote(strContent)
		if err != nil {
			l.logger.Printf("[ERROR] Failed to unquote string literal: %q - %v", strContent, err)
			l.pushValue(StringLiteralNode{Value: strContent}) // Push with raw value on error
		} else {
			l.pushValue(StringLiteralNode{Value: unquoted})
		}
	} else if ctx.NUMBER_LIT() != nil {
		numStr := ctx.NUMBER_LIT().GetText()
		var numValue interface{}
		if strings.Contains(numStr, ".") { // Try float first
			fVal, err := strconv.ParseFloat(numStr, 64)
			if err == nil {
				numValue = fVal
			} else {
				numValue = numStr
			}
		} else { // Try int
			iVal, err := strconv.ParseInt(numStr, 10, 64)
			if err == nil {
				numValue = iVal
			} else {
				numValue = numStr
			}
		}
		if _, isStr := numValue.(string); isStr {
			l.logger.Printf("[WARN] Failed to parse number literal '%s', storing as string.", numStr)
		}
		l.pushValue(NumberLiteralNode{Value: numValue})
	}
	// Boolean literals handled by ExitTerm checking IDENTIFIER text if grammar used identifiers for true/false
	// List and Map literals are handled by their own Exit methods in ast_builder_collections.go
}

// ExitPlaceholder pushes a PlaceholderNode
func (l *neuroScriptListenerImpl) ExitPlaceholder(ctx *gen.PlaceholderContext) {
	l.logDebugAST(">>> Exit Placeholder: %q", ctx.GetText())
	name := ""
	if ctx.IDENTIFIER() != nil {
		name = ctx.IDENTIFIER().GetText()
	}
	l.pushValue(PlaceholderNode{Name: name})
}

// ExitTerm pushes VariableNode, LastCallResultNode, or relies on Literal/Placeholder exit methods
func (l *neuroScriptListenerImpl) ExitTerm(ctx *gen.TermContext) {
	l.logDebugAST(">>> Exit Term: %q", ctx.GetText())
	if ctx.IDENTIFIER() != nil {
		ident := ctx.IDENTIFIER().GetText()
		// Check for boolean literals represented as identifiers
		if ident == "true" {
			l.pushValue(BooleanLiteralNode{Value: true})
		} else if ident == "false" {
			l.pushValue(BooleanLiteralNode{Value: false})
		} else {
			// Assume it's a variable name
			l.pushValue(VariableNode{Name: ident})
		}
	} else if ctx.KW_LAST_CALL_RESULT() != nil {
		l.pushValue(LastCallResultNode{})
	}
	// If term contains Literal, Placeholder, or (Expression),
	// the value/node is pushed by their respective Exit methods.
}

// ExitExpression handles potential concatenation
func (l *neuroScriptListenerImpl) ExitExpression(ctx *gen.ExpressionContext) {
	l.logDebugAST(">>> Exit Expression: %q", ctx.GetText())
	numTerms := len(ctx.AllTerm())
	numPlus := len(ctx.AllPLUS())

	// If it's a concatenation (more than one term with '+' operators)
	if numTerms > 1 && numPlus == numTerms-1 {
		l.logDebugAST("    Constructing ConcatenationNode")
		operands, ok := l.popNValues(numTerms) // Pop nodes pushed by terms
		if !ok {
			l.pushValue(nil)
			return
		} // Error handling
		l.pushValue(ConcatenationNode{Operands: operands})
	}
	// If it's just a single term, its node is already on the stack, so do nothing extra.
}
