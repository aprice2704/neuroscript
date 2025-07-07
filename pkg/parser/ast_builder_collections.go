// filename: pkg/parser/ast_builder_collections.go
// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Corrected pointer assignments and added missing unquote helper.
// nlines: 75
// risk_rating: MEDIUM

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
	// Basic unquoting of the outer characters
	s = s[1 : len(s)-1]
	// Using Go's Unquote to handle all escape sequences correctly.
	// We need to re-add quotes for the standard library's Unquote to work.
	s, err := strconv.Unquote(`"` + s + `"`)
	if err != nil {
		// This might happen with invalid escape sequences, but the lexer should prevent this.
		// Fallback to a simple replacement for basic cases if full unquoting fails.
		return strings.ReplaceAll(s, `\"`, `"`)
	}
	return s
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

		for i := len(popped) - 1; i >= 0; i-- { // Reverse to maintain original order
			expr, ok := popped[i].(ast.Expression)
			if !ok {
				l.addError(c, "list literal expected ast.Expression, got %T", popped[i])
				continue
			}
			elements = append(elements, expr)
		}
	}

	pos := tokenToPosition(c.GetStart())
	l.push(&ast.ListLiteralNode{Pos: &pos, Elements: elements})
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

	keyPos := tokenToPosition(c.STRING_LIT().GetSymbol())
	keyNode := &ast.StringLiteralNode{
		Pos:   &keyPos,
		Value: unquote(c.STRING_LIT().GetText()),
	}

	pos := tokenToPosition(c.GetStart())
	l.push(&ast.MapEntryNode{Pos: &pos, Key: keyNode, Value: valueExpr})
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

		for i := len(popped) - 1; i >= 0; i-- { // Reverse to maintain original order
			entry, ok := popped[i].(*ast.MapEntryNode)
			if !ok {
				l.addError(c, "map literal expected *ast.MapEntryNode, got %T", popped[i])
				continue
			}
			entries = append(entries, entry)
		}
	}

	pos := tokenToPosition(c.GetStart())
	l.push(&ast.MapLiteralNode{Pos: &pos, Entries: entries})
}
