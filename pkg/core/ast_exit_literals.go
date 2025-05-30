// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-29
package core

import (
	"fmt"
	"strconv"

	"github.com/antlr4-go/antlr/v4" // Correct ANTLR import is github.com/antlr4-go/antlr/v4
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	// "github.com/aprice2704/neuroscript/pkg/logging" // Assuming logger is part of neuroScriptListenerImpl
)

// NilLiteralNode represents a nil literal.
type NilLiteralNode struct {
	Pos *Position
}

// GetPos returns the position of the node.
func (n *NilLiteralNode) GetPos() *Position { return n.Pos }

// expressionNode() is a marker method to satisfy the Expression interface.
func (n *NilLiteralNode) expressionNode() {}

// ================================================================================
// START OF LITERAL HANDLING SECTION
// ================================================================================

// ExitLiteral is called when the parser has finished processing a literal.
// This method handles terminals NUMBER_LIT, STRING_LIT, TRIPLE_BACKTICK_STRING.
// For non-terminal rules like boolean_literal, nil_literal, list_literal, map_literal,
// their respective Exit<RuleName> methods are responsible for pushing the AST node.
// So, this ExitLiteral method effectively becomes a NO-OP if the context is one of those
// non-terminal rules because the correct node is already on the valueStack.
func (l *neuroScriptListenerImpl) ExitLiteral(ctx *gen.LiteralContext) {
	l.logDebugAST(" >> Exit Literal: %s", ctx.GetText())
	// The actual AST node for boolean_literal, nil_literal, list_literal, or map_literal
	// would have ALREADY been pushed onto the l.valueStack by their respective
	// Exit<RuleName> methods before ExitLiteral for the overarching 'literal' rule is called.
	// So, we only need to handle the direct terminals here.

	var nodeToPush Expression // All literals should result in an Expression AST node

	if numNode := ctx.NUMBER_LIT(); numNode != nil {
		token := numNode.GetSymbol()
		pos := tokenToPosition(token)
		value, err := strconv.ParseFloat(token.GetText(), 64) // TODO: Differentiate int/float if needed by NumberLiteralNode
		if err != nil {
			l.addErrorf(token, "invalid number literal: %v", err)
			nodeToPush = &ErrorNode{Pos: pos, Message: fmt.Sprintf("invalid number: %v", err)}
		} else {
			// Assuming NumberLiteralNode.Value is interface{} to hold float64 or potentially int64 later
			nodeToPush = &NumberLiteralNode{Pos: pos, Value: value}
		}
		l.pushValue(nodeToPush)
	} else if strNode := ctx.STRING_LIT(); strNode != nil {
		token := strNode.GetSymbol()
		pos := tokenToPosition(token)
		tokenText := token.GetText()

		if len(tokenText) < 2 {
			l.addErrorf(token, "malformed string literal token (too short): %s", tokenText)
			nodeToPush = &ErrorNode{Pos: pos, Message: "malformed string literal"}
		} else {
			content := tokenText[1 : len(tokenText)-1]
			unescapedString, err := UnescapeNeuroScriptString(content) // From string_utils.go
			if err != nil {
				l.addErrorf(token, "invalid string literal: %v", err)
				nodeToPush = &ErrorNode{Pos: pos, Message: fmt.Sprintf("invalid string: %v", err)}
			} else {
				nodeToPush = &StringLiteralNode{Pos: pos, Value: unescapedString, IsRaw: false}
			}
		}
		l.pushValue(nodeToPush)
	} else if tripleStrNode := ctx.TRIPLE_BACKTICK_STRING(); tripleStrNode != nil {
		token := tripleStrNode.GetSymbol()
		pos := tokenToPosition(token)
		tokenText := token.GetText()
		if len(tokenText) < 6 { // ```...```
			l.addErrorf(token, "malformed triple-backtick string literal token (too short): %s", tokenText)
			nodeToPush = &ErrorNode{Pos: pos, Message: "malformed raw string"}
		} else {
			rawContent := tokenText[3 : len(tokenText)-3]
			nodeToPush = &StringLiteralNode{Pos: pos, Value: rawContent, IsRaw: true}
		}
		l.pushValue(nodeToPush)
	}
	// If ctx.Boolean_literal(), ctx.Nil_literal(), ctx.List_literal(), or ctx.Map_literal() were matched,
	// their respective Exit methods (e.g., ExitBoolean_literal) would have already pushed the correct
	// AST node onto l.valueStack. So, ExitLiteral does nothing further for those cases.
	// The ANTLR walker calls Exit methods for children before the parent.

	l.logDebugAST("   << Exit Literal")
}

// ExitBoolean_literal handles boolean literals.
func (l *neuroScriptListenerImpl) ExitBoolean_literal(ctx *gen.Boolean_literalContext) {
	l.logDebugAST(" >> Exit BooleanLiteral: %s", ctx.GetText())
	var node Expression
	var token antlr.Token
	var value bool

	if ctx.KW_TRUE() != nil {
		token = ctx.KW_TRUE().GetSymbol()
		value = true
		node = &BooleanLiteralNode{Pos: tokenToPosition(token), Value: value}
	} else if ctx.KW_FALSE() != nil {
		token = ctx.KW_FALSE().GetSymbol()
		value = false
		node = &BooleanLiteralNode{Pos: tokenToPosition(token), Value: value}
	} else {
		// This case should ideally not be reached if the grammar ensures KW_TRUE or KW_FALSE is present.
		// However, adding robust error handling.
		startToken := ctx.GetStart() // Use the start of the context rule for position
		l.addErrorf(startToken, "malformed boolean literal: missing TRUE or FALSE keyword in rule: %s", ctx.GetText())
		node = &ErrorNode{Pos: tokenToPosition(startToken), Message: "malformed boolean"}
	}
	l.pushValue(node)
	l.logDebugAST("   << Exit BooleanLiteral, Pushed Node: %T", node)
}

// ExitNil_literal handles nil literals.
func (l *neuroScriptListenerImpl) ExitNil_literal(ctx *gen.Nil_literalContext) {
	l.logDebugAST(" >> Exit NilLiteral: %s", ctx.GetText())
	var node Expression
	if ctx.KW_NIL() != nil {
		token := ctx.KW_NIL().GetSymbol()
		node = &NilLiteralNode{Pos: tokenToPosition(token)}
	} else {
		// This case should ideally not be reached if the grammar ensures KW_NIL is present.
		startToken := ctx.GetStart()
		l.addErrorf(startToken, "malformed nil literal: missing NIL keyword in rule: %s", ctx.GetText())
		node = &ErrorNode{Pos: tokenToPosition(startToken), Message: "malformed nil"}
	}
	l.pushValue(node)
	l.logDebugAST("   << Exit NilLiteral, Pushed Node: %T", node)
}

// ================================================================================
// END OF LITERAL HANDLING SECTION
// ================================================================================
