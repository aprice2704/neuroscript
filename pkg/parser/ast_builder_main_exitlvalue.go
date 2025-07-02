// filename: pkg/parser/ast_builder_main_exitlvalue.go
package parser

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// ExitLvalue is called when the lvalue rule is exited by the parser.
// It constructs an ast.LValueNode and pushes it onto the listener's value stack.
func (l *neuroScriptListenerImpl) ExitLvalue(ctx *gen.LvalueContext) {
	l.logDebugAST("ExitLvalue: %s", ctx.GetText())

	baseIdentifierToken := ctx.IDENTIFIER(0)	// Rule: IDENTIFIER ( LBRACK ... | DOT IDENTIFIER )*
	if baseIdentifierToken == nil {
		pos := tokenToPosition(ctx.GetStart())
		l.addErrorf(ctx.GetStart(), "AST Builder: Malformed lvalue, missing base identifier.")
		l.push(&ast.ErrorNode{Pos: &pos, Message: "Malformed lvalue: missing base identifier"})
		return
	}
	baseIdentifierName := baseIdentifierToken.GetText()
	basePos := tokenToPosition(baseIdentifierToken.GetSymbol())

	lValueNode := &ast.LValueNode{
		Position:	basePos,
		Identifier:	baseIdentifierName,
		Accessors:	make([]*ast.AccessorNode, 0),
	}

	// Expressions for bracket accessors are pushed onto the ValueStack by their Exit rules.
	// We need to pop them in the reverse order of their appearance in the lvalue.
	numBracketExpressions := len(ctx.AllExpression())
	bracketExprAsts := make([]ast.Expression, numBracketExpressions)

	// Pop expressions for bracket accessors.
	if numBracketExpressions > 0 {
		rawExprs, ok := l.popN(numBracketExpressions)
		if !ok {
			// popN already logs an error and potentially adds to l.errors
			// Ensure an ast.ErrorNode is pushed if the contract is to always push something.
			l.addErrorf(ctx.GetStart(), "AST Builder: Stack underflow or error popping %d expressions for lvalue '%s'", numBracketExpressions, baseIdentifierName)
			l.push(&ast.ErrorNode{Pos: &basePos, Message: "Lvalue stack error: issue popping bracket expressions"})
			return
		}
		for i := 0; i < numBracketExpressions; i++ {
			expr, castOk := rawExprs[i].(ast.Expression)
			if !castOk {
				// This error should ideally be caught if popN returns an error or if an ast.ErrorNode was pushed by a failing expression rule.
				l.addErrorf(ctx.GetStart(), "AST Builder: Expected ast.Expression on stack for lvalue '%s', got %T at index %d of popped values", baseIdentifierName, rawExprs[i], i)
				l.push(&ast.ErrorNode{Pos: &basePos, Message: "Lvalue stack error: invalid bracket expression type from popN"})
				return
			}
			bracketExprAsts[i] = expr	// Stored in source order
		}
	}

	// Iterate through the grammar elements that form accessors.
	accessorChildren := ctx.GetChildren()[1:]	// Skip the base IDENTIFIER

	bracketExprUsed := 0
	currentChildPtr := 0
	for currentChildPtr < len(accessorChildren) {
		child := accessorChildren[currentChildPtr]

		if term, ok := child.(antlr.TerminalNode); ok {
			tokenType := term.GetSymbol().GetTokenType()
			pos := tokenToPosition(term.GetSymbol())
			accessor := &ast.AccessorNode{}

			if tokenType == gen.NeuroScriptLexerLBRACK {
				accessor.Type = ast.BracketAccess
				if bracketExprUsed < len(bracketExprAsts) {
					accessor.Key = bracketExprAsts[bracketExprUsed]
					bracketExprUsed++
					lValueNode.Accessors = append(lValueNode.Accessors, accessor)
					currentChildPtr += 3	// Skip LBRACK, expression, RBRACK
				} else {
					l.addErrorf(term.GetSymbol(), "AST Builder: Mismatch: Found LBRACK but no corresponding expression for lvalue '%s'", baseIdentifierName)
					l.push(&ast.ErrorNode{Pos: &pos, Message: "Lvalue error: LBRACK without expression"})
					return
				}
			} else if tokenType == gen.NeuroScriptLexerDOT {
				accessor.Type = ast.DotAccess
				currentChildPtr++	// Move past DOT to the IDENTIFIER
				if currentChildPtr < len(accessorChildren) {
					fieldIdentTerm, identOk := accessorChildren[currentChildPtr].(antlr.TerminalNode)
					if identOk && fieldIdentTerm.GetSymbol().GetTokenType() == gen.NeuroScriptLexerIDENTIFIER {
						keyPos := tokenToPosition(fieldIdentTerm.GetSymbol())
						accessor.Key = &ast.StringLiteralNode{Pos: &keyPos, Value: fieldIdentTerm.GetText()}
						lValueNode.Accessors = append(lValueNode.Accessors, accessor)
						currentChildPtr++	// Skip IDENTIFIER
					} else {
						l.addErrorf(term.GetSymbol(), "AST Builder: Expected IDENTIFIER after DOT in lvalue for '%s'", baseIdentifierName)
						l.push(&ast.ErrorNode{Pos: &pos, Message: "Lvalue error: DOT not followed by IDENTIFIER"})
						return
					}
				} else {
					l.addErrorf(term.GetSymbol(), "AST Builder: DOT at end of lvalue for '%s'", baseIdentifierName)
					l.push(&ast.ErrorNode{Pos: &pos, Message: "Lvalue error: DOT at end"})
					return
				}
			} else {
				if tokenType != gen.NeuroScriptLexerRBRACK {
					l.addErrorf(term.GetSymbol(), "AST Builder: Unexpected token '%s' while parsing lvalue accessors for '%s'", term.GetText(), baseIdentifierName)
				}
				currentChildPtr++
			}
		} else {
			currentChildPtr++
		}
	}
	l.push(lValueNode)
}