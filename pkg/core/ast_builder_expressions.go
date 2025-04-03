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
		// Store the raw, unquoted string value. Resolution happens only via EVAL.
		unquoted, err := strconv.Unquote(strContent)
		if err != nil {
			l.logger.Printf("[ERROR] Failed to unquote string literal: %q - %v", strContent, err)
			l.pushValue(StringLiteralNode{Value: strContent}) // Push raw on error
		} else {
			l.pushValue(StringLiteralNode{Value: unquoted}) // Push unquoted raw string
		}
	} else if ctx.NUMBER_LIT() != nil {
		// ... (Number parsing remains the same) ...
		numStr := ctx.NUMBER_LIT().GetText()
		var numValue interface{}
		var parseErr error
		if strings.Contains(numStr, ".") {
			fVal, err := strconv.ParseFloat(numStr, 64)
			if err == nil {
				numValue = fVal
			} else {
				parseErr = err
			}
		} else {
			iVal, err := strconv.ParseInt(numStr, 10, 64)
			if err == nil {
				numValue = iVal
			} else {
				parseErr = err
			}
		}
		if parseErr != nil {
			l.logger.Printf("[WARN] Failed to parse number literal '%s': %v. Storing as string.", numStr, parseErr)
			numValue = numStr
		}
		l.pushValue(NumberLiteralNode{Value: numValue})
	} else if ctx.Boolean_literal() != nil { // Handle explicit boolean rule
		boolStr := ctx.Boolean_literal().GetText()
		if boolStr == "true" {
			l.pushValue(BooleanLiteralNode{Value: true})
		} else {
			l.pushValue(BooleanLiteralNode{Value: false})
		}
	}
	// List/Map literals handled by their specific exit methods
}

// ExitPlaceholder pushes a PlaceholderNode (represents the {{...}} syntax itself)
func (l *neuroScriptListenerImpl) ExitPlaceholder(ctx *gen.PlaceholderContext) {
	l.logDebugAST(">>> Exit Placeholder: %q", ctx.GetText())
	name := ""
	if ctx.IDENTIFIER() != nil {
		name = ctx.IDENTIFIER().GetText()
	} else if ctx.KW_LAST() != nil {
		name = "LAST"
	} // Assuming KW_LAST defined
	l.pushValue(PlaceholderNode{Name: name})
}

// ExitPrimary handles the base non-recursive parts of an expression
func (l *neuroScriptListenerImpl) ExitPrimary(ctx *gen.PrimaryContext) {
	l.logDebugAST(">>> Exit Primary: %q", ctx.GetText())
	if ctx.IDENTIFIER() != nil {
		ident := ctx.IDENTIFIER().GetText()
		l.pushValue(VariableNode{Name: ident}) // Treat all identifiers initially as variables
		// Booleans 'true'/'false' will be handled during evaluation if needed,
		// or potentially by adding specific KW_TRUE/KW_FALSE tokens later.
	} else if ctx.KW_LAST() != nil { // Handle KW_LAST token
		l.pushValue(LastNode{})
		// *** NEW: Handle EVAL(expression) ***
	} else if ctx.KW_EVAL() != nil && ctx.LPAREN() != nil && ctx.RPAREN() != nil && ctx.Expression() != nil {
		// Assumes grammar rule: primary: ... | KW_EVAL LPAREN expression RPAREN ;
		// The argument expression node was pushed by ExitExpression.
		argNode, ok := l.popValue()
		if !ok {
			l.logger.Printf("[ERROR] AST Builder: Failed to pop argument for EVAL")
			l.pushValue(nil)
		} else {
			l.pushValue(EvalNode{Argument: argNode}) // Push the EvalNode
			l.logDebugAST("    Constructed EvalNode")
		}
		// *** REMOVED RAWTEXT simulation ***
	}
	// If primary contains Literal, Placeholder, or (Expression),
	// the value/node is pushed by their respective Exit methods.
}

// ExitTerm handles potential element access ([...])
func (l *neuroScriptListenerImpl) ExitTerm(ctx *gen.TermContext) {
	// ... (Element access logic remains the same) ...
	l.logDebugAST(">>> Exit Term: %q", ctx.GetText())
	accessExpressions := ctx.AllExpression()
	numAccessors := len(accessExpressions)
	if numAccessors > 0 {
		l.logDebugAST("    Processing %d element accessors for term", numAccessors)
		accessorNodes := make([]interface{}, numAccessors)
		for i := numAccessors - 1; i >= 0; i-- {
			node, ok := l.popValue()
			if !ok {
				l.logger.Printf("[ERROR] AST Builder: Stack error popping accessor node %d for term: %q", i, ctx.GetText())
				l.pushValue(nil)
				return
			}
			accessorNodes[i] = node
		}
		collectionNode, okColl := l.popValue()
		if !okColl {
			l.logger.Printf("[ERROR] AST Builder: Stack error popping collection node for term: %q", ctx.GetText())
			l.pushValue(nil)
			return
		}
		currentNode := collectionNode
		for _, accessorNode := range accessorNodes {
			currentNode = ElementAccessNode{Collection: currentNode, Accessor: accessorNode}
			l.logDebugAST("    Constructed intermediate ElementAccessNode")
		}
		l.pushValue(currentNode)
		l.logDebugAST("    Pushed final ElementAccessNode for term")
	}
}

// ExitExpression handles potential concatenation
func (l *neuroScriptListenerImpl) ExitExpression(ctx *gen.ExpressionContext) {
	// ... (Concatenation node creation logic remains the same) ...
	l.logDebugAST(">>> Exit Expression: %q", ctx.GetText())
	numTerms := len(ctx.AllTerm())
	numPlus := len(ctx.AllPLUS())
	if numTerms > 1 && numPlus == numTerms-1 {
		l.logDebugAST("    Constructing ConcatenationNode")
		operands, ok := l.popNValues(numTerms)
		if !ok {
			l.pushValue(nil)
			return
		}
		l.pushValue(ConcatenationNode{Operands: operands})
	}
}
