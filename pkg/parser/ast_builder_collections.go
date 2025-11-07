// NeuroScript Version: 0.8.0
// File version: 7
// Purpose: Corrected list and map literal parsing to iterate 'popN' results in forward (FIFO) order, fixing an element-reversal bug.
// filename: pkg/parser/ast_builder_collections.go
// nlines: 105

package parser

import (
	"strconv"
	"strings"

	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
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

		// FIX: popN returns elements in source order (FIFO).
		// Iterate forwards, not backwards.
		for _, val := range popped {
			expr, ok := val.(ast.Expression)
			if !ok {
				l.addError(c, "list literal expected ast.Expression, got %T", val)
				continue
			}
			elements = append(elements, expr)
		}
	}

	node := &ast.ListLiteralNode{Elements: elements}
	l.push(newNode(node, c.GetStart(), types.KindListLiteral))
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
	newNode(keyNode, keyToken, types.KindStringLiteral)

	entry := &ast.MapEntryNode{Key: keyNode, Value: valueExpr}
	node := newNode(entry, keyToken, types.KindMapEntry)
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

		// FIX: popN returns elements in source order (FIFO).
		// Iterate forwards, not backwards.
		for _, val := range popped {
			entry, ok := val.(*ast.MapEntryNode)
			if !ok {
				l.addError(c, "map literal expected *ast.MapEntryNode, got %T", val)
				continue
			}
			entries = append(entries, entry)
		}
	}

	node := &ast.MapLiteralNode{Entries: entries}
	l.push(newNode(node, c.GetStart(), types.KindMapLiteral))
}
