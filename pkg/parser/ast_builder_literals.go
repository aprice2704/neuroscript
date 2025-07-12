// filename: pkg/parser/ast_builder_literals.go
// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Refactored node creation to use the newNode helper function.

package parser

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// ================================================================================
// START OF LITERAL HANDLING SECTION
// ================================================================================

// ExitLiteral is called when the parser has finished processing a literal.
func (l *neuroScriptListenerImpl) ExitLiteral(ctx *gen.LiteralContext) {
	l.logDebugAST(" >> Exit Literal: %s", ctx.GetText())

	var nodeToPush ast.Expression

	if numNode := ctx.NUMBER_LIT(); numNode != nil {
		token := numNode.GetSymbol()
		val, err := parseNumber(token.GetText())
		if err != nil {
			l.addErrorf(token, "invalid number literal: %v", err)
			errorNode := &ast.ErrorNode{Message: fmt.Sprintf("invalid number: %v", err)}
			nodeToPush = newNode(errorNode, token, ast.KindUnknown)
		} else {
			node := &ast.NumberLiteralNode{Value: val}
			nodeToPush = newNode(node, token, ast.KindNumberLiteral)
		}
		l.push(nodeToPush)
	} else if strNode := ctx.STRING_LIT(); strNode != nil {
		token := strNode.GetSymbol()
		unescapedString, err := unescapeString(token.GetText())
		if err != nil {
			l.addErrorf(token, "invalid string literal: %v", err)
			errorNode := &ast.ErrorNode{Message: fmt.Sprintf("invalid string: %v", err)}
			nodeToPush = newNode(errorNode, token, ast.KindUnknown)
		} else {
			node := &ast.StringLiteralNode{Value: unescapedString, IsRaw: false}
			nodeToPush = newNode(node, token, ast.KindStringLiteral)
		}
		l.push(nodeToPush)
	} else if tripleStrNode := ctx.TRIPLE_BACKTICK_STRING(); tripleStrNode != nil {
		token := tripleStrNode.GetSymbol()
		tokenText := token.GetText()
		if len(tokenText) < 6 { // ```...```
			l.addErrorf(token, "malformed triple-backtick string literal token (too short): %s", tokenText)
			errorNode := &ast.ErrorNode{Message: "malformed raw string"}
			nodeToPush = newNode(errorNode, token, ast.KindUnknown)
		} else {
			rawContent := tokenText[3 : len(tokenText)-3]
			node := &ast.StringLiteralNode{Value: rawContent, IsRaw: true}
			nodeToPush = newNode(node, token, ast.KindStringLiteral)
		}
		l.push(nodeToPush)
	}
	// For other literal types, their specific exit methods handle pushing to the stack.

	l.logDebugAST("   << Exit Literal")
}

// ExitBoolean_literal handles boolean literals.
func (l *neuroScriptListenerImpl) ExitBoolean_literal(ctx *gen.Boolean_literalContext) {
	l.logDebugAST(" >> Exit BooleanLiteral: %s", ctx.GetText())
	var node ast.Expression
	var token antlr.Token
	var val bool

	if ctx.KW_TRUE() != nil {
		token = ctx.KW_TRUE().GetSymbol()
		val = true
		boolNode := &ast.BooleanLiteralNode{Value: val}
		node = newNode(boolNode, token, ast.KindBooleanLiteral)
	} else if ctx.KW_FALSE() != nil {
		token = ctx.KW_FALSE().GetSymbol()
		val = false
		boolNode := &ast.BooleanLiteralNode{Value: val}
		node = newNode(boolNode, token, ast.KindBooleanLiteral)
	} else {
		token = ctx.GetStart()
		l.addErrorf(token, "malformed boolean literal: missing TRUE or FALSE keyword in rule: %s", ctx.GetText())
		errorNode := &ast.ErrorNode{Message: "malformed boolean"}
		node = newNode(errorNode, token, ast.KindUnknown)
	}
	l.push(node)
	l.logDebugAST("   << Exit BooleanLiteral, Pushed Node: %T", node)
}

// ExitNil_literal handles nil literals.
func (l *neuroScriptListenerImpl) ExitNil_literal(ctx *gen.Nil_literalContext) {
	l.logDebugAST(" >> Exit NilLiteral: %s", ctx.GetText())
	var node ast.Expression
	var token antlr.Token

	if ctx.KW_NIL() != nil {
		token = ctx.KW_NIL().GetSymbol()
		node = newNode(&ast.NilLiteralNode{}, token, ast.KindNilLiteral)
	} else {
		token = ctx.GetStart()
		l.addErrorf(token, "malformed nil literal: missing NIL keyword in rule: %s", ctx.GetText())
		errorNode := &ast.ErrorNode{Message: "malformed nil"}
		node = newNode(errorNode, token, ast.KindUnknown)
	}
	l.push(node)
	l.logDebugAST("   << Exit NilLiteral, Pushed Node: %T", node)
}

// ================================================================================
// END OF LITERAL HANDLING SECTION
// ================================================================================
