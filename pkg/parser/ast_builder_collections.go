// filename: pkg/parser/ast_builder_collections.go
// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Refactored collection literal creation to use the newNode helper.

package parser

import (
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// unquote removes the surrounding quotes from a string literal and processes escape sequences.
func unquote(s string) string {
	if len(s) < 2 {
		return ""
	}
	// Using Go's Unquote to handle all escape sequences correctly.
	unquoted, err := strconv.Unquote(s)
	if err != nil {
		// This might happen with invalid escape sequences, but the lexer should prevent this.
		// Fallback to a simple replacement for basic cases if full unquoting fails.
		return strings.ReplaceAll(s[1:len(s)-1], `\"`, `"`)
	}
	return unquoted
}

func (l *neuroScriptListenerImpl) ExitList_literal(c *gen.List_literalContext) {
	numElements := 0
	if c.Expression_list_opt() != nil && c.Expression_list_opt().Expression_list() != nil {
		numElements = len(c.Expression_list_opt().Expression_list().AllExpression())
	}

	elements := make([]ast.Expression, 0, numElements)
	if numElements > 0 {
		popped, ok := l.popN(numElements)
		if !ok {
			l.addError(c, "stack underflow in list literal")
			return
		}

		// Reverse to maintain original source order
		for i := len(popped) - 1; i >= 0; i-- {
			expr, ok := popped[i].(ast.Expression)
			if !ok {
				l.addError(c, "list literal expected ast.Expression, got %T", popped[i])
				continue
			}
			elements = append(elements, expr)
		}
	}

	node := &ast.ListLiteralNode{Elements: elements}
	l.push(newNode(node, c.GetStart(), ast.KindListLiteral))
}

func (l *neuroScriptListenerImpl) ExitMap_entry(c *gen.Map_entryContext) {
	val, ok := l.pop()
	if !ok {
		l.addError(c, "stack underflow in map entry value")
		return
	}
	valueExpr, ok := val.(ast.Expression)
	if !ok {
		l.addError(c, "map entry value is not ast.Expression, got %T", val)
		return
	}

	keyToken := c.STRING_LIT().GetSymbol()
	keyNode := &ast.StringLiteralNode{
		Value: unquote(keyToken.GetText()),
	}
	newNode(keyNode, keyToken, ast.KindStringLiteral)

	node := &ast.MapEntryNode{Key: keyNode, Value: valueExpr}
	// A MapEntry isn't a standalone expression, so it doesn't get a kind itself,
	// it's part of a MapLiteralNode. We can give it a position from its key.
	node.Pos = keyNode.Pos
	node.BaseNode.StartPos = keyNode.BaseNode.StartPos

	l.push(node)
}

func (l *neuroScriptListenerImpl) ExitMap_literal(c *gen.Map_literalContext) {
	numEntries := 0
	if c.Map_entry_list_opt() != nil && c.Map_entry_list_opt().Map_entry_list() != nil {
		numEntries = len(c.Map_entry_list_opt().Map_entry_list().AllMap_entry())
	}

	entries := make([]*ast.MapEntryNode, 0, numEntries)
	if numEntries > 0 {
		popped, ok := l.popN(numEntries)
		if !ok {
			l.addError(c, "stack underflow in map literal")
			return
		}

		// Reverse to maintain original source order
		for i := len(popped) - 1; i >= 0; i-- {
			entry, ok := popped[i].(*ast.MapEntryNode)
			if !ok {
				l.addError(c, "map literal expected *ast.MapEntryNode, got %T", popped[i])
				continue
			}
			entries = append(entries, entry)
		}
	}

	node := &ast.MapLiteralNode{Entries: entries}
	l.push(newNode(node, c.GetStart(), ast.KindMapLiteral))
}
