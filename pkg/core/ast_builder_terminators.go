// pkg/core/ast_builder_terminals.go
package core

import (
	"strconv"
	"strings" // Keep for ExitLiteral

	// Keep for error logging/debugging if needed
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Exit methods for Primary Expressions, Literals, Placeholders, Access ---

// ExitExpression is just a pass-through in the listener for the top-level expression rule.
func (l *neuroScriptListenerImpl) ExitExpression(ctx *gen.ExpressionContext) {
	l.logDebugAST(">>> Exit Expression: %q (Pass through)", ctx.GetText())
}

// ExitAccessor_expr handles primary expressions potentially followed by element accessors
// (Unchanged from previous correct version)
func (l *neuroScriptListenerImpl) ExitAccessor_expr(ctx *gen.Accessor_exprContext) {
	l.logDebugAST(">>> Exit Accessor_expr: %q", ctx.GetText())
	numAccessors := len(ctx.AllLBRACK())
	if numAccessors > 0 {
		accessorNodes := make([]interface{}, numAccessors)
		for i := 0; i < numAccessors; i++ {
			node, ok := l.popValue()
			if !ok {
				l.logger.Error("AST Builder: Stack error popping accessor %d for %q", numAccessors-1-i, ctx.GetText())
				l.pushValue(nil)
				return
			}
			accessorNodes[numAccessors-1-i] = node
		}
		collectionNode, okColl := l.popValue()
		if !okColl {
			l.logger.Error("AST Builder: Stack error popping collection for %q", ctx.GetText())
			l.pushValue(nil)
			return
		}
		currentNode := collectionNode
		for _, accessorNode := range accessorNodes {
			currentNode = ElementAccessNode{Collection: currentNode, Accessor: accessorNode}
			l.logDebugAST("    Constructed intermediate ElementAccessNode: %T[%T]", currentNode.(ElementAccessNode).Collection, accessorNode)
		}
		l.pushValue(currentNode)
		l.logDebugAST("    Pushed final ElementAccessNode")
	} else {
		l.logDebugAST("    Accessor_expr is just a primary, passing through.")
	}
}

// ExitPrimary handles the base cases of expressions.
// (Unchanged from previous correct version)
func (l *neuroScriptListenerImpl) ExitPrimary(ctx *gen.PrimaryContext) {
	l.logDebugAST(">>> Exit Primary: %q", ctx.GetText())
	if ctx.Literal() != nil {
		return
	}
	if ctx.Placeholder() != nil {
		return
	}
	if ctx.Function_call() != nil {
		return
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
		return
	} // Parenthesized expression value passed through
	l.logger.Warn("ExitPrimary reached unexpected state for text: %q", ctx.GetText())
}

// ExitFunction_call builds a FunctionCallNode.
// (Unchanged from previous correct version)
func (l *neuroScriptListenerImpl) ExitFunction_call(ctx *gen.Function_callContext) {
	l.logDebugAST(">>> Exit Function_call: %q", ctx.GetText())
	var funcName string
	if ctx.KW_LN() != nil {
		funcName = "LN"
	} else if ctx.KW_LOG() != nil {
		funcName = "LOG"
	} else if ctx.KW_SIN() != nil {
		funcName = "SIN"
	} else if ctx.KW_COS() != nil {
		funcName = "COS"
	} else if ctx.KW_TAN() != nil {
		funcName = "TAN"
	} else if ctx.KW_ASIN() != nil {
		funcName = "ASIN"
	} else if ctx.KW_ACOS() != nil {
		funcName = "ACOS"
	} else if ctx.KW_ATAN() != nil {
		funcName = "ATAN"
	} else if ctx.IDENTIFIER() != nil {
		funcName = ctx.IDENTIFIER().GetText()
	} else {
		l.logger.Error("AST Builder: Could not determine function name in Function_call: %q", ctx.GetText())
		l.pushValue(nil)
		return
	}
	numArgs := 0
	if exprListOpt := ctx.Expression_list_opt(); exprListOpt != nil {
		if exprList := exprListOpt.Expression_list(); exprList != nil {
			numArgs = len(exprList.AllExpression())
		}
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
// (Unchanged from previous correct version)
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

// --- CORRECTED ExitLiteral ---
// ExitLiteral handles different types of literals. Restructured to avoid early returns.
func (l *neuroScriptListenerImpl) ExitLiteral(ctx *gen.LiteralContext) {
	l.logDebugAST(">>> Exit Literal: %q", ctx.GetText())
	var nodeToPush interface{} // Node to push at the end

	if ctx.STRING_LIT() != nil {
		strContent := ctx.STRING_LIT().GetText()
		unquoted, err := strconv.Unquote(strContent)
		valueNode := StringLiteralNode{Value: strContent, IsRaw: false} // Assume quoted string is not raw
		if err != nil {
			l.logger.Error("Failed to unquote string literal: %q - %v", strContent, err)
			// Keep raw value in node as fallback
		} else {
			valueNode.Value = unquoted
		}
		nodeToPush = valueNode
		l.logDebugAST("    Constructed StringLiteralNode (Quoted)")

	} else if ctx.TRIPLE_BACKTICK_STRING() != nil {
		rawContent := ctx.TRIPLE_BACKTICK_STRING().GetText()
		valueNode := StringLiteralNode{Value: rawContent, IsRaw: true} // Assume raw, push raw content as fallback
		// Remove the ``` delimiters
		if len(rawContent) >= 6 && strings.HasPrefix(rawContent, "```") && strings.HasSuffix(rawContent, "```") {
			valueNode.Value = rawContent[3 : len(rawContent)-3] // Use content inside backticks
			l.logDebugAST("    Constructed StringLiteralNode (Triple-Backtick/Raw)")
		} else {
			l.logger.Error("Invalid triple-backtick string format: %q", rawContent)
			// Keep rawContent in valueNode as fallback
		}
		nodeToPush = valueNode

	} else if ctx.NUMBER_LIT() != nil {
		numStr := ctx.NUMBER_LIT().GetText()
		var numValue interface{}
		var parseErr error
		if !strings.Contains(numStr, ".") {
			if iVal, err := strconv.ParseInt(numStr, 10, 64); err == nil {
				numValue = iVal
			} else {
				parseErr = err
			}
		}
		if numValue == nil { // Try float if int failed or if '.' was present
			if fVal, err := strconv.ParseFloat(numStr, 64); err == nil {
				numValue = fVal
				parseErr = nil // Clear previous int parse error if float succeeded
			} else if parseErr == nil { // Only assign float parse error if int didn't already fail
				parseErr = err
			}
		}
		if parseErr != nil {
			l.logger.Warn("Failed to parse number literal '%s': %v. Storing as string.", numStr, parseErr)
			nodeToPush = NumberLiteralNode{Value: numStr} // Fallback to storing raw string
		} else {
			nodeToPush = NumberLiteralNode{Value: numValue}
		}
		l.logDebugAST("    Constructed NumberLiteralNode")

	} else if ctx.Boolean_literal() != nil {
		// Value was pushed by ExitBoolean_literal, retrieve it
		l.logDebugAST("    Literal is Boolean (value already on stack)")
		// Pop the value pushed by ExitBoolean_literal
		val, ok := l.popValue()
		if !ok {
			l.logger.Error("AST Builder: Stack error popping value for Boolean Literal in ExitLiteral")
			nodeToPush = nil // Error marker
		} else {
			nodeToPush = val
		}
	} else if ctx.List_literal() != nil {
		// Value was pushed by ExitList_literal, retrieve it
		l.logDebugAST("    Literal is List (value already on stack)")
		val, ok := l.popValue()
		if !ok {
			l.logger.Error("AST Builder: Stack error popping value for List Literal in ExitLiteral")
			nodeToPush = nil // Error marker
		} else {
			nodeToPush = val
		}
	} else if ctx.Map_literal() != nil {
		// Value was pushed by ExitMap_literal, retrieve it
		l.logDebugAST("    Literal is Map (value already on stack)")
		val, ok := l.popValue()
		if !ok {
			l.logger.Error("AST Builder: Stack error popping value for Map Literal in ExitLiteral")
			nodeToPush = nil // Error marker
		} else {
			nodeToPush = val
		}
	} else {
		// This case should not be reachable if the grammar is correct
		l.logger.Error("ExitLiteral reached unexpected state - no known literal type found for text: %q", ctx.GetText())
		nodeToPush = nil // Push nil as an error marker
	}

	// Push the determined node (or nil if error) onto the stack
	l.pushValue(nodeToPush)
}

// ExitBoolean_literal pushes a BooleanLiteralNode.
// (Unchanged)
func (l *neuroScriptListenerImpl) ExitBoolean_literal(ctx *gen.Boolean_literalContext) {
	l.logDebugAST(">>> Exit Boolean_literal: %q", ctx.GetText())
	value := false
	if ctx.KW_TRUE() != nil {
		value = true
	}
	l.pushValue(BooleanLiteralNode{Value: value})
	l.logDebugAST("    Constructed BooleanLiteralNode: Value=%t", value)
}
