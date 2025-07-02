// filename: pkg/parser/ast_builder_collections.go
package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

func (l *neuroScriptListenerImpl) ExitList_literal(c *gen.List_literalContext) {
	l.logDebugAST("<<< ExitList_literal")
	pos := tokenToPosition(c.GetStart())
	if c.Expression_list_opt() == nil || c.Expression_list_opt().Expression_list() == nil {
		l.push(&ast.ListLiteralNode{
			Pos:		&pos,
			Elements:	[]ast.Expression{},
		})
		return
	}
	numExpr := len(c.Expression_list_opt().Expression_list().AllExpression())
	values, ok := l.popN(numExpr)
	if !ok {
		l.addError(c, "stack underflow in list literal")
		return
	}

	exprs := make([]ast.Expression, len(values))
	for i, v := range values {
		expr, ok := v.(ast.Expression)
		if !ok {
			l.addError(c, "list literal expected ast.Expression, got %T", v)
			l.push(&ast.ListLiteralNode{})
			return
		}
		exprs[i] = expr
	}

	l.push(&ast.ListLiteralNode{
		Pos:		&pos,
		Elements:	exprs,
	})
}

func (l *neuroScriptListenerImpl) ExitMap_literal(c *gen.Map_literalContext) {
	l.logDebugAST("<<< ExitMap_literal")
	pos := tokenToPosition(c.GetStart())
	if c.Map_entry_list_opt() == nil || c.Map_entry_list_opt().Map_entry_list() == nil {
		l.push(&ast.MapLiteralNode{
			Pos:		&pos,
			Entries:	[]*ast.MapEntryNode{},
		})
		return
	}

	numEntries := len(c.Map_entry_list_opt().Map_entry_list().AllMap_entry())
	values, ok := l.popN(numEntries)
	if !ok {
		l.addError(c, "stack underflow in map literal")
		return
	}

	entries := make([]*ast.MapEntryNode, len(values))
	for i, v := range values {
		entry, ok := v.(*ast.MapEntryNode)
		if !ok {
			l.addError(c, "map literal expected *ast.MapEntryNode, got %T", v)
			l.push(&ast.MapLiteralNode{})
			return
		}
		entries[i] = entry
	}

	l.push(&ast.MapLiteralNode{
		Pos:		&pos,
		Entries:	entries,
	})
}

func (l *neuroScriptListenerImpl) ExitMap_entry(c *gen.Map_entryContext) {
	l.logDebugAST("<<< ExitMap_entry")
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

	key, ok := l.pop()
	if !ok {
		l.addError(c, "stack underflow in map entry key")
		return
	}
	keyExpr, ok := key.(*ast.StringLiteralNode)
	if !ok {
		l.addError(c, "map entry key is not *ast.StringLiteralNode, got %T", key)
		return
	}

	pos := tokenToPosition(c.GetStart())
	l.push(&ast.MapEntryNode{
		Pos:	&pos,
		Key:	keyExpr,
		Value:	valueExpr,
	})
}