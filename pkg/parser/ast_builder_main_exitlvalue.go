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

	baseIdentifierToken := ctx.IDENTIFIER(0)
	if baseIdentifierToken == nil {
		l.addErrorf(ctx.GetStart(), "AST Builder: Malformed lvalue, missing base identifier.")
		l.push(newNode(&ast.ErrorNode{Message: "Malformed lvalue: missing base identifier"}, ctx.GetStart(), ast.KindUnknown))
		return
	}
	baseIdentifierName := baseIdentifierToken.GetText()

	lValueNode := &ast.LValueNode{
		Identifier: baseIdentifierName,
		Accessors:  make([]*ast.AccessorNode, 0),
	}
	newNode(lValueNode, baseIdentifierToken.GetSymbol(), ast.KindLValue)

	// Pop expressions for bracket accessors.
	numBracketExpressions := len(ctx.AllExpression())
	bracketExprAsts := make([]ast.Expression, numBracketExpressions)
	if numBracketExpressions > 0 {
		rawExprs, ok := l.popN(numBracketExpressions)
		if !ok {
			l.addErrorf(ctx.GetStart(), "AST Builder: Stack underflow or error popping %d expressions for lvalue '%s'", numBracketExpressions, baseIdentifierName)
			l.push(newNode(&ast.ErrorNode{Message: "Lvalue stack error: issue popping bracket expressions"}, ctx.GetStart(), ast.KindUnknown))
			return
		}
		for i := 0; i < numBracketExpressions; i++ {
			// Popped in reverse order, so we iterate backwards to restore source order.
			expr, castOk := rawExprs[len(rawExprs)-1-i].(ast.Expression)
			if !castOk {
				l.addErrorf(ctx.GetStart(), "AST Builder: Expected ast.Expression on stack for lvalue '%s', got %T at index %d of popped values", baseIdentifierName, rawExprs[i], i)
				l.push(newNode(&ast.ErrorNode{Message: "Lvalue stack error: invalid bracket expression type from popN"}, ctx.GetStart(), ast.KindUnknown))
				return
			}
			bracketExprAsts[i] = expr
		}
	}

	// Iterate through the grammar elements that form accessors.
	accessorChildren := ctx.GetChildren()[1:] // Skip the base IDENTIFIER

	bracketExprUsed := 0
	currentChildPtr := 0
	for currentChildPtr < len(accessorChildren) {
		child := accessorChildren[currentChildPtr]

		if term, ok := child.(antlr.TerminalNode); ok {
			tokenType := term.GetSymbol().GetTokenType()
			accessor := &ast.AccessorNode{}

			if tokenType == gen.NeuroScriptLexerLBRACK {
				accessor.Type = ast.BracketAccess
				if bracketExprUsed < len(bracketExprAsts) {
					accessor.Key = bracketExprAsts[bracketExprUsed]
					bracketExprUsed++
					lValueNode.Accessors = append(lValueNode.Accessors, newNode(accessor, term.GetSymbol(), ast.KindUnknown))
					currentChildPtr += 3 // Skip LBRACK, expression, RBRACK
				} else {
					l.addErrorf(term.GetSymbol(), "AST Builder: Mismatch: Found LBRACK but no corresponding expression for lvalue '%s'", baseIdentifierName)
					l.push(newNode(&ast.ErrorNode{Message: "Lvalue error: LBRACK without expression"}, term.GetSymbol(), ast.KindUnknown))
					return
				}
			} else if tokenType == gen.NeuroScriptLexerDOT {
				accessor.Type = ast.DotAccess
				currentChildPtr++ // Move past DOT to the IDENTIFIER
				if currentChildPtr < len(accessorChildren) {
					fieldIdentTerm, identOk := accessorChildren[currentChildPtr].(antlr.TerminalNode)
					if identOk && fieldIdentTerm.GetSymbol().GetTokenType() == gen.NeuroScriptLexerIDENTIFIER {
						keyToken := fieldIdentTerm.GetSymbol()
						keyNode := &ast.StringLiteralNode{Value: fieldIdentTerm.GetText()}
						accessor.Key = newNode(keyNode, keyToken, ast.KindStringLiteral)
						lValueNode.Accessors = append(lValueNode.Accessors, newNode(accessor, term.GetSymbol(), ast.KindUnknown))
						currentChildPtr++ // Skip IDENTIFIER
					} else {
						l.addErrorf(term.GetSymbol(), "AST Builder: Expected IDENTIFIER after DOT in lvalue for '%s'", baseIdentifierName)
						l.push(newNode(&ast.ErrorNode{Message: "Lvalue error: DOT not followed by IDENTIFIER"}, term.GetSymbol(), ast.KindUnknown))
						return
					}
				} else {
					l.addErrorf(term.GetSymbol(), "AST Builder: DOT at end of lvalue for '%s'", baseIdentifierName)
					l.push(newNode(&ast.ErrorNode{Message: "Lvalue error: DOT at end"}, term.GetSymbol(), ast.KindUnknown))
					return
				}
			} else {
				currentChildPtr++
			}
		} else {
			currentChildPtr++
		}
	}
	l.push(lValueNode)
}
