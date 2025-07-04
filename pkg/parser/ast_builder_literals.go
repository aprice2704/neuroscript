// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Corrected string literal unescaping logic to handle full quoted strings.
// filename: pkg/parser/ast_builder_literals.go
// nlines: 132
// risk_rating: HIGH

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
// This method handles terminals NUMBER_LIT, STRING_LIT, TRIPLE_BACKTICK_STRING.
// For non-terminal rules like boolean_literal, nil_literal, list_literal, map_literal,
// their respective Exit<RuleName> methods are responsible for pushing the AST node.
func (l *neuroScriptListenerImpl) ExitLiteral(ctx *gen.LiteralContext) {
	l.logDebugAST(" >> Exit Literal: %s", ctx.GetText())

	var nodeToPush ast.Expression

	if numNode := ctx.NUMBER_LIT(); numNode != nil {
		token := numNode.GetSymbol()
		pos := tokenToPosition(token)
		val, err := parseNumber(token.GetText())
		if err != nil {
			l.addErrorf(token, "invalid number literal: %v", err)
			nodeToPush = &ast.ErrorNode{Pos: &pos, Message: fmt.Sprintf("invalid number: %v", err)}
		} else {
			nodeToPush = &ast.NumberLiteralNode{Pos: &pos, Value: val}
		}
		l.push(nodeToPush)
	} else if strNode := ctx.STRING_LIT(); strNode != nil {
		token := strNode.GetSymbol()
		pos := tokenToPosition(token)
		tokenText := token.GetText()

		unescapedString, err := unescapeString(tokenText)
		if err != nil {
			l.addErrorf(token, "invalid string literal: %v", err)
			nodeToPush = &ast.ErrorNode{Pos: &pos, Message: fmt.Sprintf("invalid string: %v", err)}
		} else {
			nodeToPush = &ast.StringLiteralNode{Pos: &pos, Value: unescapedString, IsRaw: false}
		}
		l.push(nodeToPush)
	} else if tripleStrNode := ctx.TRIPLE_BACKTICK_STRING(); tripleStrNode != nil {
		token := tripleStrNode.GetSymbol()
		pos := tokenToPosition(token)
		tokenText := token.GetText()
		if len(tokenText) < 6 { // ```...```
			l.addErrorf(token, "malformed triple-backtick string literal token (too short): %s", tokenText)
			nodeToPush = &ast.ErrorNode{Pos: &pos, Message: "malformed raw string"}
		} else {
			rawContent := tokenText[3 : len(tokenText)-3]
			nodeToPush = &ast.StringLiteralNode{Pos: &pos, Value: rawContent, IsRaw: true}
		}
		l.push(nodeToPush)
	}
	// For other literal types (boolean, nil, list, map), their specific exit methods
	// will have already pushed the correct AST node onto the stack.

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
		pos := tokenToPosition(token)
		node = &ast.BooleanLiteralNode{Pos: &pos, Value: val}
	} else if ctx.KW_FALSE() != nil {
		token = ctx.KW_FALSE().GetSymbol()
		val = false
		pos := tokenToPosition(token)
		node = &ast.BooleanLiteralNode{Pos: &pos, Value: val}
	} else {
		startToken := ctx.GetStart()
		pos := tokenToPosition(startToken)
		l.addErrorf(startToken, "malformed boolean literal: missing TRUE or FALSE keyword in rule: %s", ctx.GetText())
		node = &ast.ErrorNode{Pos: &pos, Message: "malformed boolean"}
	}
	l.push(node)
	l.logDebugAST("   << Exit BooleanLiteral, Pushed Node: %T", node)
}

// ExitNil_literal handles nil literals.
func (l *neuroScriptListenerImpl) ExitNil_literal(ctx *gen.Nil_literalContext) {
	l.logDebugAST(" >> Exit NilLiteral: %s", ctx.GetText())
	var node ast.Expression
	if ctx.KW_NIL() != nil {
		token := ctx.KW_NIL().GetSymbol()
		pos := tokenToPosition(token)
		node = &ast.NilLiteralNode{Pos: &pos}
	} else {
		startToken := ctx.GetStart()
		pos := tokenToPosition(startToken)
		l.addErrorf(startToken, "malformed nil literal: missing NIL keyword in rule: %s", ctx.GetText())
		node = &ast.ErrorNode{Pos: &pos, Message: "malformed nil"}
	}
	l.push(node)
	l.logDebugAST("   << Exit NilLiteral, Pushed Node: %T", node)
}

// ================================================================================
// END OF LITERAL HANDLING SECTION
// ================================================================================
