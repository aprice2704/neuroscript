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
		var parseErr error                 // Explicitly define parseErr
		if strings.Contains(numStr, ".") { // Try float first
			fVal, err := strconv.ParseFloat(numStr, 64)
			if err == nil {
				numValue = fVal
			} else {
				parseErr = err // Store error
			}
		} else { // Try int
			iVal, err := strconv.ParseInt(numStr, 10, 64)
			if err == nil {
				numValue = iVal
			} else {
				parseErr = err // Store error
			}
		}
		// If parsing failed, store as string and warn
		if parseErr != nil {
			l.logger.Printf("[WARN] Failed to parse number literal '%s': %v. Storing as string.", numStr, parseErr)
			numValue = numStr
		}
		l.pushValue(NumberLiteralNode{Value: numValue})
	}
	// List and Map literals handled by ExitList_literal / ExitMap_literal
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

// ExitPrimary handles the base non-recursive parts of an expression
func (l *neuroScriptListenerImpl) ExitPrimary(ctx *gen.PrimaryContext) {
	l.logDebugAST(">>> Exit Primary: %q", ctx.GetText())
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
	// If primary contains Literal, Placeholder, or (Expression),
	// the value/node is pushed by their respective Exit methods. No action needed here.
}

// ExitTerm now handles potential element access ([...]) following a primary expression
func (l *neuroScriptListenerImpl) ExitTerm(ctx *gen.TermContext) {
	l.logDebugAST(">>> Exit Term: %q", ctx.GetText())

	// Check if element access operations ([expression]) are present
	accessExpressions := ctx.AllExpression() // Get all accessor expressions
	numAccessors := len(accessExpressions)

	if numAccessors > 0 {
		// The stack will contain [..., primary_node, accessor1_node, accessor2_node, ...]
		l.logDebugAST("    Processing %d element accessors for term", numAccessors)

		// Pop all accessor nodes first (they were pushed last)
		accessorNodes := make([]interface{}, numAccessors)
		for i := numAccessors - 1; i >= 0; i-- {
			node, ok := l.popValue()
			if !ok {
				l.logger.Printf("[ERROR] AST Builder: Stack error popping accessor node %d for term: %q", i, ctx.GetText())
				l.pushValue(nil) // Indicate error
				return
			}
			accessorNodes[i] = node
		}

		// Pop the base collection node (the primary expression)
		collectionNode, okColl := l.popValue()
		if !okColl {
			l.logger.Printf("[ERROR] AST Builder: Stack error popping collection node for term: %q", ctx.GetText())
			l.pushValue(nil) // Indicate error
			return
		}

		// Build nested ElementAccessNodes if multiple accessors exist (e.g., list[0][1])
		currentNode := collectionNode
		for _, accessorNode := range accessorNodes {
			currentNode = ElementAccessNode{
				Collection: currentNode,
				Accessor:   accessorNode,
			}
			l.logDebugAST("    Constructed intermediate ElementAccessNode")
		}

		// Push the final (potentially nested) ElementAccessNode
		l.pushValue(currentNode)
		l.logDebugAST("    Pushed final ElementAccessNode for term")

	}
	// If numAccessors is 0, it means it was just a primary expression.
	// The node for the primary expression was already pushed by ExitPrimary (or ExitLiteral, etc.)
	// and remains on the stack, so no further action is needed here in that case.
}

// ExitExpression handles potential concatenation
func (l *neuroScriptListenerImpl) ExitExpression(ctx *gen.ExpressionContext) {
	l.logDebugAST(">>> Exit Expression: %q", ctx.GetText())
	numTerms := len(ctx.AllTerm())
	numPlus := len(ctx.AllPLUS())

	// If it's a concatenation (more than one term with '+' operators)
	if numTerms > 1 && numPlus == numTerms-1 {
		l.logDebugAST("    Constructing ConcatenationNode")
		// Operands are pushed left-to-right by ExitTerm calls
		operands, ok := l.popNValues(numTerms) // Pop nodes pushed by terms
		if !ok {
			l.pushValue(nil)
			return
		} // Error handling
		l.pushValue(ConcatenationNode{Operands: operands})
	}
	// If it's just a single term, its node is already on the stack (pushed by ExitTerm),
	// so do nothing extra.
}
