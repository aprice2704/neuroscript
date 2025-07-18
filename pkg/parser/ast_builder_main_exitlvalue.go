// filename: pkg/parser/ast_builder_main_exitlvalue.go
// NeuroScript Version: 0.5.2
// File version: 18.0.0
//
// Builds an *ast.LValueNode* when the parser exits an lvalue rule.
// Reverted to always produce an LValueNode to satisfy consumers
// like the 'set' statement and canonicalizer. Ambiguity with r-values
// is now handled by the specific statement listeners (e.g., for 'return').

package parser

import (
	"fmt"
	"reflect"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// ExitLvalue is called when the lvalue rule is exited by the parser.
// It constructs an ast.LValueNode and pushes it onto the listener's value stack.
func (l *neuroScriptListenerImpl) ExitLvalue(ctx *gen.LvalueContext) {
	/* ───── banner ───── */
	fmt.Println("\n=========================================================")
	fmt.Printf(">>> Enter ExitLvalue for context: %s\n", ctx.GetText())
	fmt.Printf("    Initial ValueStack size: %d\n", len(l.ValueStack))

	numBracketExpressions := len(ctx.AllExpression())
	fmt.Printf("    Detected %d bracket expressions in the context.\n", numBracketExpressions)

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

		fmt.Printf("    Popped %d items from value stack for bracket expressions.\n", len(popped))
		for i, p := range popped {
			fmt.Printf("      - Popped item %d: Type=%T, Text=%s\n",
				i, p, nodeText(p))
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

		fmt.Println("    Bracket expression AST nodes (source order):")
		for i, expr := range bracketExprs {
			fmt.Printf("      - bracketExprs[%d]: Type=%T, Text=%s\n",
				i, expr, nodeText(expr))
		}
	}

	/* ───── base identifier ───── */
	baseIdentifierToken := ctx.IDENTIFIER(0)
	if baseIdentifierToken == nil {
		l.addErrorf(ctx.GetStart(), "AST Builder: malformed lvalue, missing base identifier")
		l.push(newNode(&ast.ErrorNode{Message: "malformed lvalue"},
			ctx.GetStart(), types.KindUnknown))
		return
	}

	lValueNode := &ast.LValueNode{
		Identifier: baseIdentifierToken.GetText(),
		Accessors:  make([]*ast.AccessorNode, 0),
	}
	newNode(lValueNode, baseIdentifierToken.GetSymbol(), types.KindLValue)

	fmt.Printf("    Base Identifier: %s\n", lValueNode.Identifier)
	fmt.Println("    Walking children of LvalueContext to build accessor chain:")

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
		fmt.Printf("      - Child %d: TerminalNode, Type: %d, Text: '%s'\n",
			i, tokenType, term.GetText())

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
			lValueNode.Accessors = append(lValueNode.Accessors,
				newNode(accessor, term.GetSymbol(), types.KindUnknown))

			fmt.Printf("        → Created DOT accessor for key: %q\n", keyNode.Value)
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
			lValueNode.Accessors = append(lValueNode.Accessors,
				newNode(accessor, term.GetSymbol(), types.KindUnknown))

			fmt.Printf("        → Created BRACKET accessor with key [%d]: %s\n",
				bracketExprUsed, nodeText(accessor.Key))
			bracketExprUsed++
			i += 3 // '[', expression, ']'

		default:
			i++
		}
	}

	/* ───── final dump ───── */
	fmt.Println("    Final constructed LValueNode before pushing to stack:")
	fmt.Printf("      Identifier: %s\n", lValueNode.Identifier)
	for j, acc := range lValueNode.Accessors {
		accType := "Dot"
		if acc.Type == ast.BracketAccess {
			accType = "Bracket"
		}
		fmt.Printf("      Accessor %d: Type=%s, Key=%s\n",
			j, accType, nodeText(acc.Key))
	}
	fmt.Printf("<<< Exit ExitLvalue, pushing node to stack. Final stack size will be: %d\n",
		len(l.ValueStack)+1)
	fmt.Println("=========================================================")

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
