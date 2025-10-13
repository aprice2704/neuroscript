// NeuroScript Version: 0.7.2
// File version: 23.0.0
// Purpose: Sets the end position on the LValueNode to ensure all nodes have valid positions.
// filename: pkg/parser/ast_builder_main_exitlvalue.go
//
// Builds an *ast.LValueNode* when the parser exits an lvalue rule.
// This version has all unrelated debug prints removed.

package parser

import (
	"fmt"
	"reflect"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// ExitLvalue is called when the lvalue rule is exited by the parser.
// It constructs an ast.LValueNode and pushes it onto the listener's value stack.
func (l *neuroScriptListenerImpl) ExitLvalue(ctx *gen.LvalueContext) {
	numBracketExpressions := len(ctx.AllExpression())

	/* ───── collect bracket expressions ───── */
	bracketExprs := make([]ast.Expression, numBracketExpressions)
	if numBracketExpressions > 0 {
		popped, ok := l.popN(numBracketExpressions)
		if !ok {
			l.addErrorf(ctx.GetStart(),
				"AST Builder: stack underflow popping bracket expressions for lvalue %q",
				ctx.GetText())
			l.push(newNode(&ast.ErrorNode{Message: "lvalue stack error"},
				ctx.GetStart(), types.KindUnknown))
			return
		}

		// popped is already left‑to‑right in source order → copy directly
		for i, n := range popped {
			expr, ok := n.(ast.Expression)
			if !ok {
				l.addErrorf(ctx.GetStart(),
					"AST Builder: expected ast.Expression, got %T", n)
				l.push(newNode(&ast.ErrorNode{Message: "lvalue type error"},
					ctx.GetStart(), types.KindUnknown))
				return
			}
			bracketExprs[i] = expr
		}
	}

	/* ───── base identifier ───── */
	baseIdentifierToken := ctx.IDENTIFIER(0)
	if baseIdentifierToken == nil {
		l.addErrorf(ctx.GetStart(), "AST Builder: malformed lvalue, missing base identifier")
		errorNode := &ast.ErrorNode{Message: "malformed lvalue"}
		// FIX: Guard against nil token to prevent panic.
		if startToken := ctx.GetStart(); startToken != nil {
			newNode(errorNode, startToken, types.KindUnknown)
		}
		l.push(errorNode)
		return
	}

	lValueNode := &ast.LValueNode{
		Identifier: baseIdentifierToken.GetText(),
		Accessors:  make([]*ast.AccessorNode, 0),
	}
	newNode(lValueNode, baseIdentifierToken.GetSymbol(), types.KindLValue)

	accessorChildren := ctx.GetChildren()[1:]
	bracketExprUsed := 0

	for i := 0; i < len(accessorChildren); {
		child := accessorChildren[i]

		term, isTerm := child.(antlr.TerminalNode)
		if !isTerm {
			i++
			continue
		}

		tokenType := term.GetSymbol().GetTokenType()

		switch tokenType {

		/* ── dot accessor ── */
		case gen.NeuroScriptLexerDOT:
			accessor := &ast.AccessorNode{Type: ast.DotAccess}
			i++ // skip '.'

			if i >= len(accessorChildren) {
				l.addErrorf(term.GetSymbol(), "AST Builder: DOT at end of lvalue %q",
					lValueNode.Identifier)
				break
			}
			fieldTerm, ok := accessorChildren[i].(antlr.TerminalNode)
			if !ok || fieldTerm.GetSymbol().GetTokenType() != gen.NeuroScriptLexerIDENTIFIER {
				l.addErrorf(term.GetSymbol(),
					"AST Builder: expected IDENTIFIER after DOT in lvalue %q",
					lValueNode.Identifier)
				break
			}

			keyTok := fieldTerm.GetSymbol()
			keyNode := &ast.StringLiteralNode{Value: keyTok.GetText()}
			accessor.Key = newNode(keyNode, keyTok, types.KindStringLiteral)
			// FIX: Use KindElementAccess instead of KindUnknown.
			lValueNode.Accessors = append(lValueNode.Accessors,
				newNode(accessor, term.GetSymbol(), types.KindElementAccess))
			i++ // skip identifier

		/* ── bracket accessor ── */
		case gen.NeuroScriptLexerLBRACK:
			if bracketExprUsed >= len(bracketExprs) {
				l.addErrorf(term.GetSymbol(),
					"AST Builder: mismatch – found '[' but no corresponding expression")
				break
			}
			accessor := &ast.AccessorNode{
				Type: ast.BracketAccess,
				Key:  bracketExprs[bracketExprUsed],
			}
			// FIX: Use KindElementAccess instead of KindUnknown.
			lValueNode.Accessors = append(lValueNode.Accessors,
				newNode(accessor, term.GetSymbol(), types.KindElementAccess))

			bracketExprUsed++
			i += 3 // '[', expression, ']'

		default:
			i++
		}
	}

	// FIX: Set the StopPos for the entire LValue node.
	SetEndPos(lValueNode, ctx.GetStop())
	l.push(lValueNode)
}

/* ───── helpers ───── */

// nodeText gives a concise printable form of common literal / variable nodes.
func nodeText(n any) string { // ← was ast.Node
	switch v := n.(type) {
	case *ast.NumberLiteralNode:
		return fmt.Sprintf("%g", v.Value)
	case *ast.StringLiteralNode:
		return v.Value
	case *ast.VariableNode:
		return v.Name
	default:
		return reflect.TypeOf(n).String()
	}
}
