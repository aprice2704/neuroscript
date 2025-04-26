// pkg/core/ast_builder_terminals.go
package core

import (
	"strconv"
	"strings" // Keep for ExitLiteral

	"github.com/antlr4-go/antlr/v4" // Keep for error logging/debugging if needed
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Exit methods for Primary Expressions, Literals, Placeholders, Access ---

// ExitExpression is just a pass-through in the listener for the top-level expression rule.
func (l *neuroScriptListenerImpl) ExitExpression(ctx *gen.ExpressionContext) {
	l.logDebugAST(">>> Exit Expression: %q (Pass through)", ctx.GetText())
}

// *** NEW: ExitAccessor_expr handles primary expressions potentially followed by element accessors ***
// This replaces the old ExitTerm logic.
func (l *neuroScriptListenerImpl) ExitAccessor_expr(ctx *gen.Accessor_exprContext) {
	l.logDebugAST(">>> Exit Accessor_expr: %q", ctx.GetText())

	// Check how many accessor expressions ([expression]) followed the primary
	// Assumes ANTLR provides AllExpression() and AllLBRACK() on the Accessor_exprContext.
	// Check the generated parser code if these methods don't exist.
	numAccessors := len(ctx.AllLBRACK()) // Count based on '[' occurrences

	if numAccessors > 0 {
		// Pop the accessor expressions first (they were parsed later and pushed by ExitExpression)
		accessorNodes := make([]interface{}, numAccessors)
		for i := 0; i < numAccessors; i++ {
			node, ok := l.popValue()
			if !ok {
				l.logger.Error("AST Builder: Stack error popping accessor node %d for %q", numAccessors-1-i, ctx.GetText())
				l.pushValue(nil) // Push error marker
				return
			}
			// Arguments are popped in reverse order of parsing
			accessorNodes[numAccessors-1-i] = node // Store in correct order (0 to N-1)
		}

		// Pop the base primary expression node (was parsed first, pushed by ExitPrimary)
		collectionNode, okColl := l.popValue()
		if !okColl {
			l.logger.Error("AST Builder: Stack error popping collection node for %q", ctx.GetText())
			l.pushValue(nil) // Push error marker
			return
		}

		// Build the nested ElementAccessNode structure
		currentNode := collectionNode
		for _, accessorNode := range accessorNodes { // Iterate through accessors in parse order
			currentNode = ElementAccessNode{Collection: currentNode, Accessor: accessorNode}
			l.logDebugAST("    Constructed intermediate ElementAccessNode: %T[%T]", currentNode.(ElementAccessNode).Collection, accessorNode)
		}
		l.pushValue(currentNode) // Push the final possibly nested node
		l.logDebugAST("    Pushed final ElementAccessNode")

	} else {
		// No accessors ('[]'), the primary node is already on the stack from ExitPrimary.
		// Do nothing, let the value pass through.
		l.logDebugAST("    Accessor_expr is just a primary, passing through.")
	}
}

// ExitPrimary handles the base cases of expressions.
func (l *neuroScriptListenerImpl) ExitPrimary(ctx *gen.PrimaryContext) {
	l.logDebugAST(">>> Exit Primary: %q", ctx.GetText())
	if ctx.Literal() != nil {
		l.logDebugAST("    Primary contains Literal")
		return // Value pushed by ExitLiteral
	}
	if ctx.Placeholder() != nil {
		l.logDebugAST("    Primary contains Placeholder")
		return // Value pushed by ExitPlaceholder
	}
	if ctx.Function_call() != nil {
		l.logDebugAST("    Primary contains Function_call")
		return // Value pushed by ExitFunction_call
	}
	if ctx.IDENTIFIER() != nil {
		l.pushValue(VariableNode{Name: ctx.IDENTIFIER().GetText()})
		l.logDebugAST("    Constructed VariableNode: %s", ctx.IDENTIFIER().GetText())
		return
	}
	if ctx.KW_LAST() != nil {
		l.pushValue(LastNode{})
		l.logDebugAST("    Constructed LastNode")
		return
	}
	if ctx.KW_EVAL() != nil {
		argNode, ok := l.popValue()
		if !ok {
			l.logger.Error("AST Builder: Failed to pop argument for EVAL")
			l.pushValue(nil)
		} else {
			l.pushValue(EvalNode{Argument: argNode})
			l.logDebugAST("    Constructed EvalNode")
		}
		return
	}
	if ctx.LPAREN() != nil {
		l.logDebugAST("    Primary is Parenthesized Expression")
		return // Value pushed by inner ExitExpression
	}
	l.logger.Warn("ExitPrimary reached unexpected state for text: %q", ctx.GetText())
}

// ExitFunction_call builds a FunctionCallNode.
func (l *neuroScriptListenerImpl) ExitFunction_call(ctx *gen.Function_callContext) {
	l.logDebugAST(">>> Exit Function_call: %q", ctx.GetText())
	funcNameToken := ctx.GetChild(0).(antlr.TerminalNode)
	funcName := funcNameToken.GetText()
	numArgs := 0
	if ctx.Expression_list_opt() != nil && ctx.Expression_list_opt().Expression_list() != nil {
		numArgs = len(ctx.Expression_list_opt().Expression_list().AllExpression())
	}
	args, ok := l.popNValues(numArgs)
	if !ok {
		if numArgs > 0 {
			l.logger.Error("AST Builder: Stack error popping %d args for function %s", numArgs, funcName)
			l.pushValue(nil)
			return
		}
		args = []interface{}{}
	}
	l.pushValue(FunctionCallNode{FunctionName: funcName, Arguments: args})
	l.logDebugAST("    Constructed FunctionCallNode: %s (%d args)", funcName, len(args))
}

// ExitPlaceholder builds a PlaceholderNode (e.g., {{var}} or {{LAST}}).
func (l *neuroScriptListenerImpl) ExitPlaceholder(ctx *gen.PlaceholderContext) {
	l.logDebugAST(">>> Exit Placeholder: %q", ctx.GetText())
	name := ""
	if ctx.IDENTIFIER() != nil {
		name = ctx.IDENTIFIER().GetText()
	} else if ctx.KW_LAST() != nil {
		name = "LAST"
	} else {
		l.logger.Warn("ExitPlaceholder found unexpected content: %q", ctx.GetText())
	}
	l.pushValue(PlaceholderNode{Name: name})
	l.logDebugAST("    Constructed PlaceholderNode: Name=%s", name)
}

// ExitLiteral handles different types of literals.
func (l *neuroScriptListenerImpl) ExitLiteral(ctx *gen.LiteralContext) {
	l.logDebugAST(">>> Exit Literal: %q", ctx.GetText())
	if ctx.STRING_LIT() != nil {
		strContent := ctx.STRING_LIT().GetText()
		unquoted, err := strconv.Unquote(strContent)
		if err != nil {
			l.logger.Error("Failed to unquote string literal: %q - %v", strContent, err)
			l.pushValue(StringLiteralNode{Value: strContent}) // Fallback
		} else {
			l.pushValue(StringLiteralNode{Value: unquoted})
		}
		l.logDebugAST("    Constructed StringLiteralNode")
	} else if ctx.NUMBER_LIT() != nil {
		numStr := ctx.NUMBER_LIT().GetText()
		var numValue interface{}
		var parseErr error
		if !strings.Contains(numStr, ".") {
			iVal, err := strconv.ParseInt(numStr, 10, 64)
			if err == nil {
				numValue = iVal
			} else {
				parseErr = err
			}
		}
		if numValue == nil { // Try float if int failed or if '.' was present
			fVal, err := strconv.ParseFloat(numStr, 64)
			if err == nil {
				numValue = fVal
				parseErr = nil
			} else if parseErr == nil {
				parseErr = err
			}
		}
		if parseErr != nil {
			l.logger.Warn("Failed to parse number literal '%s': %v. Storing as string.", numStr, parseErr)
			l.pushValue(NumberLiteralNode{Value: numStr}) // Fallback
		} else {
			l.pushValue(NumberLiteralNode{Value: numValue})
		}
		l.logDebugAST("    Constructed NumberLiteralNode")
	} else if ctx.Boolean_literal() != nil {
		l.logDebugAST("    Literal is Boolean (handled by ExitBoolean_literal)")
		// Value already pushed by ExitBoolean_literal
	} else if ctx.List_literal() != nil {
		l.logDebugAST("    Literal is List (handled by ExitList_literal)")
		// Value already pushed by ExitList_literal
	} else if ctx.Map_literal() != nil {
		l.logDebugAST("    Literal is Map (handled by ExitMap_literal)")
		// Value already pushed by ExitMap_literal
	}
}

// ExitBoolean_literal pushes a BooleanLiteralNode.
func (l *neuroScriptListenerImpl) ExitBoolean_literal(ctx *gen.Boolean_literalContext) {
	l.logDebugAST(">>> Exit Boolean_literal: %q", ctx.GetText())
	value := false
	if ctx.KW_TRUE() != nil {
		value = true
	}
	l.pushValue(BooleanLiteralNode{Value: value})
	l.logDebugAST("    Constructed BooleanLiteralNode: Value=%t", value)
}

// --- REMOVED ExitTerm function ---
// func (l *neuroScriptListenerImpl) ExitTerm(ctx *gen.PrimaryContext) { ... } // REMOVED
