// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-29

package parser

import (
	"fmt"
	"strconv"

	"github.com/antlr4-go/antlr/v4" // Correct ANTLR import is github.com/antlr4-go/antlr/v4
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
	// "github.com/aprice2704/neuroscript/pkg/logging" // Assuming logger is part of neuroScriptListenerImpl
)

// ================================================================================
// START OF LITERAL HANDLING SECTION
// ================================================================================

// ExitLiteral is called when the parser has finished processing a literal.
// This method handles terminals NUMBER_LIT, STRING_LIT, TRIPLE_BACKTICK_STRING.
// For non-terminal rules like boolean_literal, nil_literal, list_literal, map_literal,
// their respective Exit<RuleName> methods are responsible for pushing the AST node.
// So, this ExitLiteral method effectively becomes a NO-OP if the context is one of those
// non-terminal rules because the correct node is already on the ValueStack.

func (l *neuroScriptListenerImpl) ExitLiteral(ctx *gen.LiteralContext) {
	l.logDebugAST(" >> Exit Literal: %s", ctx.GetText())
	// The actual AST node for boolean_literal, nil_literal, list_literal, or map_literal
	// would have ALREADY been pushed onto the l.ValueStack by their respective
	// Exit<RuleName> methods before ExitLiteral for the overarching 'literal' rule is called.
	// So, we only need to handle the direct terminals here.

	var nodeToPush ast.Expression // All literals should result in an ast.Expression AST node

	if numNode := ctx.NUMBER_LIT(); numNode != nil {
		token := numNode.GetSymbol()
		pos := tokenTolang.Position(token)
		val, err := strconv.ParseFloat(token.GetText(), 64) // TODO: Differentiate int/float if needed by ast.NumberLiteralNode
		if err != nil {
			l.addErrorf(token, "invalid number literal: %v", err)
			nodeToPush = &ast.ErrorNode{Pos: pos, Message: fmt.Sprintf("invalid number: %v", err)}
		} else {
			// Assuming ast.NumberLiteralNode.Value is interface{} to hold float64 or potentially int64 later
			nodeToPush = &ast.NumberLiteralNode{Pos: pos, Value: val}
		}
		l.push(nodeToPush)
	} else if strNode := ctx.STRING_LIT(); strNode != nil {
		token := strNode.GetSymbol()
		pos := tokenTolang.Position(token)
		tokenText := token.GetText()

		if len(tokenText) < 2 {
			l.addErrorf(token, "malformed string literal token (too short): %s", tokenText)
			nodeToPush = &ast.ErrorNode{Pos: pos, Message: "malformed string literal"}
		} else {
			content := tokenText[1 : len(tokenText)-1]
			unescapedString, err := unescapeString(content) // From string_utils.go
			if err != nil {
				l.addErrorf(token, "invalid string literal: %v", err)
				nodeToPush = &ast.ErrorNode{Pos: pos, Message: fmt.Sprintf("invalid string: %v", err)}
			} else {
				nodeToPush = &ast.StringLiteralNode{Pos: pos, Value: unescapedString, IsRaw: false}
			}
		}
		l.push(nodeToPush)
	} else if tripleStrNode := ctx.TRIPLE_BACKTICK_STRING(); tripleStrNode != nil {
		token := tripleStrNode.GetSymbol()
		pos := tokenTolang.Position(token)
		tokenText := token.GetText()
		if len(tokenText) < 6 { // ```...```
			l.addErrorf(token, "malformed triple-backtick string literal token (too short): %s", tokenText)
			nodeToPush = &ast.ErrorNode{Pos: pos, Message: "malformed raw string"}
		} else {
			rawContent := tokenText[3 : len(tokenText)-3]
			nodeToPush = &ast.StringLiteralNode{Pos: pos, Value: rawContent, IsRaw: true}
		}
		l.push(nodeToPush)
	}
	// If ctx.Boolean_literal(), ctx.Nil_literal(), ctx.List_literal(), or ctx.Map_literal() were matched,
	// their respective Exit methods (e.g., ExitBoolean_literal) would have already pushed the correct
	// AST node onto l.ValueStack. So, ExitLiteral does nothing further for those cases.
	// The ANTLR walker calls Exit methods for children before the parent.

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
		node = &ast.BooleanLiteralNode{Pos: tokenTolang.Position(token), Value: val}
	} else if ctx.KW_FALSE() != nil {
		token = ctx.KW_FALSE().GetSymbol()
		val = false
		node = &ast.BooleanLiteralNode{Pos: tokenTolang.Position(token), Value: val}
	} else {
		// This case should ideally not be reached if the grammar ensures KW_TRUE or KW_FALSE is present.
		// However, adding robust error handling.
		startToken := ctx.GetStart() // Use the start of the context rule for position
		l.addErrorf(startToken, "malformed boolean literal: missing TRUE or FALSE keyword in rule: %s", ctx.GetText())
		node = &ast.ErrorNode{Pos: tokenTolang.Position(startToken), Message: "malformed boolean"}
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
		node = &ast.NilLiteralNode{Pos: tokenTolang.Position(token)}
	} else {
		// This case should ideally not be reached if the grammar ensures KW_NIL is present.
		startToken := ctx.GetStart()
		l.addErrorf(startToken, "malformed nil literal: missing NIL keyword in rule: %s", ctx.GetText())
		node = &ast.ErrorNode{Pos: tokenTolang.Position(startToken), Message: "malformed nil"}
	}
	l.push(node)
	l.logDebugAST("   << Exit NilLiteral, Pushed Node: %T", node)
}

// ================================================================================
// END OF LITERAL HANDLING SECTION
// ================================================================================
