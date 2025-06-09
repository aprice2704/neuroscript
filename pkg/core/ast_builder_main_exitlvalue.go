package core

import (
	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// ExitLvalue is called when the lvalue rule is exited by the parser.
// It constructs an LValueNode and pushes it onto the listener's value stack.
func (l *neuroScriptListenerImpl) ExitLvalue(ctx *gen.LvalueContext) {
	l.logDebugAST("ExitLvalue: %s", ctx.GetText())

	baseIdentifierToken := ctx.IDENTIFIER(0) // Rule: IDENTIFIER ( LBRACK ... | DOT IDENTIFIER )*
	if baseIdentifierToken == nil {
		l.addErrorf(ctx.GetStart(), "AST Builder: Malformed lvalue, missing base identifier.")
		l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Malformed lvalue: missing base identifier"})
		return
	}
	baseIdentifierName := baseIdentifierToken.GetText()
	basePos := tokenToPosition(baseIdentifierToken.GetSymbol())

	lValueNode := &LValueNode{
		Pos:        basePos,
		Identifier: baseIdentifierName,
		Accessors:  make([]AccessorNode, 0),
	}

	// Expressions for bracket accessors are pushed onto the valueStack by their Exit rules.
	// We need to pop them in the reverse order of their appearance in the lvalue.
	numBracketExpressions := len(ctx.AllExpression())
	bracketExprAsts := make([]Expression, numBracketExpressions)

	// Pop expressions for bracket accessors.
	// Based on your popNValues: "Reverse to get them in parsing order".
	// So if source is a[expr1][expr2], stack top is expr2, then expr1.
	// popNValues(2) would return [expr1_node, expr2_node].
	if numBracketExpressions > 0 {
		rawExprs, ok := l.popNValues(numBracketExpressions)
		if !ok {
			// popNValues already logs an error and potentially adds to l.errors
			// Ensure an ErrorNode is pushed if the contract is to always push something.
			l.addErrorf(ctx.GetStart(), "AST Builder: Stack underflow or error popping %d expressions for lvalue '%s'", numBracketExpressions, baseIdentifierName)
			l.pushValue(&ErrorNode{Pos: basePos, Message: "Lvalue stack error: issue popping bracket expressions"})
			return
		}
		for i := 0; i < numBracketExpressions; i++ {
			expr, castOk := rawExprs[i].(Expression)
			if !castOk {
				// This error should ideally be caught if popNValues returns an error or if an ErrorNode was pushed by a failing expression rule.
				l.addErrorf(ctx.GetStart(), "AST Builder: Expected Expression on stack for lvalue '%s', got %T at index %d of popped values", baseIdentifierName, rawExprs[i], i)
				l.pushValue(&ErrorNode{Pos: basePos, Message: "Lvalue stack error: invalid bracket expression type from popNValues"})
				return
			}
			bracketExprAsts[i] = expr // Stored in source order
		}
	}

	// Iterate through the grammar elements that form accessors.
	// The lvalue rule structure from ANTLR: IDENTIFIER (LBRACK expression RBRACK | DOT IDENTIFIER)*
	// We need to walk through the accessor chain. ctx.children can be used, but ANTLR also provides
	// specific accessors like ctx.AllLBRACK(), ctx.AllDOT(), ctx.AllIDENTIFIER(), ctx.AllExpression().

	// Counters for elements we've used
	bracketExprUsed := 0
	dotIdentUsed := 0 // How many of ctx.IDENTIFIER(i>0) we've used

	// We determine the type of each accessor segment based on the order of LBRACK and DOT tokens.
	// This assumes that ANTLR provides these tokens in sequence corresponding to the source.
	// The children of LvalueContext will be the base IDENTIFIER, then a sequence of tokens/contexts
	// representing the accessors. E.g., for `a[e1].f[e2]`:
	// IDENTIFIER(a), LBRACK, Expression(e1), RBRACK, DOT, IDENTIFIER(f), LBRACK, Expression(e2), RBRACK

	// Iterate based on the number of LBRACKs and DOTs
	numLBracks := len(ctx.AllLBRACK())
	numDots := len(ctx.AllDOT())
	totalAccessors := numLBracks + numDots

	// We need to reconstruct the original order of accessors.
	// We can iterate through the children of the LvalueContext after the base IDENTIFIER.
	accessorChildren := ctx.GetChildren()[1:] // Skip the base IDENTIFIER

	currentChildPtr := 0
	for len(lValueNode.Accessors) < totalAccessors {
		if currentChildPtr >= len(accessorChildren) {
			break // Should have found all accessors
		}
		child := accessorChildren[currentChildPtr]

		if term, ok := child.(antlr.TerminalNode); ok {
			tokenType := term.GetSymbol().GetTokenType()
			accessor := AccessorNode{Pos: tokenToPosition(term.GetSymbol())}

			if tokenType == gen.NeuroScriptLexerLBRACK {
				accessor.Type = BracketAccess
				if bracketExprUsed < len(bracketExprAsts) {
					accessor.IndexOrKey = bracketExprAsts[bracketExprUsed]
					bracketExprUsed++
					lValueNode.Accessors = append(lValueNode.Accessors, accessor)
					currentChildPtr += 3 // Skip LBRACK, expression_rule_placeholder, RBRACK
					// The expression_rule_placeholder isn't directly a child TerminalNode here.
					// We've already popped the expression.
				} else {
					l.addErrorf(term.GetSymbol(), "AST Builder: Mismatch: Found LBRACK but no corresponding expression for lvalue '%s'", baseIdentifierName)
					l.pushValue(&ErrorNode{Pos: basePos, Message: "Lvalue error: LBRACK without expression"})
					return
				}
			} else if tokenType == gen.NeuroScriptLexerDOT {
				accessor.Type = DotAccess
				currentChildPtr++ // Move past DOT to the IDENTIFIER
				if currentChildPtr < len(accessorChildren) {
					fieldIdentTerm, identOk := accessorChildren[currentChildPtr].(antlr.TerminalNode)
					if identOk && fieldIdentTerm.GetSymbol().GetTokenType() == gen.NeuroScriptLexerIDENTIFIER {
						accessor.FieldName = fieldIdentTerm.GetText()
						// Optionally, update accessor.Pos to fieldIdentTerm.GetSymbol() if more precise
						lValueNode.Accessors = append(lValueNode.Accessors, accessor)
						dotIdentUsed++    // This counter isn't strictly necessary with child iteration
						currentChildPtr++ // Skip IDENTIFIER
					} else {
						l.addErrorf(term.GetSymbol(), "AST Builder: Expected IDENTIFIER after DOT in lvalue for '%s'", baseIdentifierName)
						l.pushValue(&ErrorNode{Pos: basePos, Message: "Lvalue error: DOT not followed by IDENTIFIER"})
						return
					}
				} else {
					l.addErrorf(term.GetSymbol(), "AST Builder: DOT at end of lvalue for '%s'", baseIdentifierName)
					l.pushValue(&ErrorNode{Pos: basePos, Message: "Lvalue error: DOT at end"})
					return
				}
			} else {
				// This might be an RBRACK or an unexpected token. RBRACKs are part of the LBRACK sequence.
				// If it's not LBRACK or DOT, we might just advance.
				if tokenType != gen.NeuroScriptLexerRBRACK { // RBRACKs are expected and skipped as part of LBRACK processing
					l.addErrorf(term.GetSymbol(), "AST Builder: Unexpected token '%s' while parsing lvalue accessors for '%s'", term.GetText(), baseIdentifierName)
				}
				currentChildPtr++
			}
		} else {
			// If child is not a TerminalNode, it might be an ExpressionContext (already handled by popping)
			// or an error node. For this simplified child iteration, we primarily expect tokens.
			// The expression part of `LBRACK expression RBRACK` is handled by popping from stack.
			// If an ExpressionContext itself is a child, it means the grammar is structured differently than assumed.
			// For `( A | B )*`, ANTLR makes direct context accessors for A and B, e.g. `ctx.A(i)` and `ctx.B(i)`.
			// The `children` based walk is an alternative if specific accessors are tricky.
			// Given the expression popping logic, we primarily care about LBRACK/DOT/IDENTIFIER tokens here.
			currentChildPtr++
		}
	}

	if len(lValueNode.Accessors) != totalAccessors {
		l.addErrorf(ctx.GetStart(), "AST Builder: Could not parse all accessors for lvalue '%s'. Expected %d, got %d.", baseIdentifierName, totalAccessors, len(lValueNode.Accessors))
		// Fallback or push error node
	}

	l.pushValue(lValueNode)
}
