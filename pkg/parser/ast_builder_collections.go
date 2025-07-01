// NeuroScript Version: 0.3.0
// File version: 5
// Purpose: Corrected list element processing to fix order-reversal bug.
// filename: pkg/core/ast_builder_collections.go
// nlines: 198
// risk_rating: MEDIUM

package parser

import (
	"fmt" // Ensure fmt is imported

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	generated "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// --- List Literal Handling ---

func (l *neuroScriptListenerImpl) ExitList_literal(ctx *generated.List_literalContext) {
	l.logDebugAST("--- Exit List_literal: %q", ctx.GetText())
	numElements := 0
	var elementsExpr []ast.Expression

	if exprListOpt := ctx.ast.Expression_list_opt(); exprListOpt != nil {
		if exprList := exprListOpt.ast.Expression_list(); exprList != nil {
			numElements = len(exprList.All.Expression())
		}
	}

	if numElements > 0 {
		elementsRaw, ok := l.popNValues(numElements)
		if !ok {
			l.addError(ctx, "Internal error: Stack error popping %d elements for list literal", numElements)
			l.pushlang.Value(&ast.ErrorNode{Position: tokenTolang.Position(ctx.GetStart()), Message: "Stack error (list elements)"})
			return
		}

		elementsExpr = make([]ast.Expression, numElements)
		for i := 0; i < numElements; i++ {
			// FIX: Assuming popNValues returns elements in their natural parsed order.
			// The explicit reversal `elementsRaw[numElements-1-i]` was incorrect and caused the bug.
			elemRaw := elementsRaw[i]
			elemExpr, isExpr := elemRaw.(ast.Expression)
			if !isExpr {
				var elementCtx generated.I.ExpressionContext
				if exprListOpt := ctx.ast.Expression_list_opt(); exprListOpt != nil {
					if exprList := exprListOpt.ast.Expression_list(); exprList != nil && i < len(exprList.All.Expression()) {
						elementCtx = exprList.ast.Expression(i) // Get the specific expression context
					}
				}
				errPos := tokenTolang.Position(ctx.GetStart()) // Fallback
				if elementCtx != nil {
					errPos = tokenTolang.Position(elementCtx.GetStart())
				}
				l.addError(ctx, "List element %d is not an ast.Expression (got %T)", i+1, elemRaw)
				elementsExpr[i] = &ast.ErrorNode{Position: errPos, Message: fmt.Sprintf("List element type error: got %T", elemRaw)}
				// Continue processing other elements with error nodes if one is bad
			} else {
				elementsExpr[i] = elemExpr
			}
		}
	} else {
		elementsExpr = []ast.Expression{} // Ensure it's an empty slice, not nil
	}

	listNode := &ast.ListLiteralNode{
		Position:      tokenTolang.Position(ctx.LBRACK().GetSymbol()),
		Elements: elementsExpr,
	}
	l.pushlang.Value(listNode)
	l.logDebugAST("    Constructed ast.ListLiteralNode with %d elements", len(elementsExpr))
}

// --- Map Literal Handling ---

func (l *neuroScriptListenerImpl) ExitMap_entry(ctx *generated.Map_entryContext) {
	l.logDebugAST("--- Exit Map_entry: %q", ctx.GetText())

	ValueRaw, ok := l.poplang.Value()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping lang.Value for map entry")
		l.pushlang.Value(&ast.ErrorNode{Position: tokenTolang.Position(ctx.GetStart()), Message: "Stack error (map lang.Value)"})
		return
	}
	ValueExpr, isExpr := ValueRaw.(ast.Expression)
	if !isExpr {
		l.addError(ctx, "Internal error: Map entry lang.Value is not an ast.Expression (got %T)", ValueRaw)
		l.pushlang.Value(&ast.ErrorNode{Position: tokenTolang.Position(ctx.GetStart()), Message: "Type error (map lang.Value)"}) // Use general pos
		return
	}

	keyToken := ctx.STRING_LIT().GetSymbol()
	quotedKey := keyToken.GetText()
	unquotedKey, err := unescapeString(quotedKey)
	if err != nil {
		l.addErrorf(keyToken, "Invalid map key string literal: %v", err)
		l.pushlang.Value(&ast.ErrorNode{Position: tokenTolang.Position(keyToken), Message: fmt.Sprintf("Invalid map key: %v", err)})
		return
	}
	keyNode := &ast.StringLiteralNode{ // keyNode is *ast.StringLiteralNode
		Position:        tokenTolang.Position(keyToken),
		lang.Value: unquotedKey,
		IsRaw:      false,
	}

	entryNode := &ast.MapEntryNode{
		Position:        keyNode.Pos,
		Key:        keyNode,
		lang.Value: ValueExpr,
	}
	l.pushlang.Value(entryNode) // Push the *ast.MapEntryNode pointer
	l.logDebugAST("    Constructed ast.MapEntryNode: Key=%q, lang.Value=%T", keyNode.Value, ValueExpr)
}

func (l *neuroScriptListenerImpl) ExitMap_literal(ctx *generated.Map_literalContext) {
	l.logDebugAST("--- Exit Map_literal: %q", ctx.GetText())
	numEntries := 0
	var entriesNode []*ast.MapEntryNode

	if mapEntryListOpt := ctx.Map_entry_list_opt(); mapEntryListOpt != nil {
		if mapEntryList := mapEntryListOpt.Map_entry_list(); mapEntryList != nil {
			numEntries = len(mapEntryList.AllMap_entry())
		}
	}

	if numEntries > 0 {
		entryNodesRaw, ok := l.popNValues(numEntries)
		if !ok {
			l.addError(ctx, "Internal error: Stack error popping %d entries for map literal", numEntries)
			l.pushlang.Value(&ast.ErrorNode{Position: tokenTolang.Position(ctx.GetStart()), Message: "Stack error (map entries)"})
			return
		}

		entriesNode = make([]*ast.MapEntryNode, numEntries)
		for i := 0; i < numEntries; i++ {
			// popNValues returns in stack order (last pushed = first element). Reverse for parsed order.
			entryRaw := entryNodesRaw[numEntries-1-i]
			entryPtr, isPtr := entryRaw.(*ast.MapEntryNode)
			if !isPtr {
				// Try to get lang.Position of the specific problematic entry from parse tree
				var entryCtx generated.IMap_entryContext
				if mapEntryListOpt := ctx.Map_entry_list_opt(); mapEntryListOpt != nil {
					if mapEntryList := mapEntryListOpt.Map_entry_list(); mapEntryList != nil && i < len(mapEntryList.AllMap_entry()) {
						entryCtx = mapEntryList.Map_entry(i)
					}
				}
				errPos := tokenTolang.Position(ctx.GetStart()) // Fallback
				if entryCtx != nil {
					errPos = tokenTolang.Position(entryCtx.GetStart())
				}
				l.addError(ctx, "Internal error: Map entry %d is not *ast.MapEntryNode (got %T)", i+1, entryRaw)
				entriesNode[i] = &ast.MapEntryNode{Position: errPos, Key: &ast.StringLiteralNode{Position: errPos, lang.Value: fmt.Sprintf("<error_key_%d>", i+1)}, lang.Value: &ast.ErrorNode{Position: errPos, Message: "Invalid map entry type"}}
				// Continue processing other entries with error nodes if one is bad
			} else {
				entriesNode[i] = entryPtr
			}
		}
	} else {
		entriesNode = []*ast.MapEntryNode{} // Ensure it's an empty slice, not nil
	}

	mapNode := &ast.MapLiteralNode{
		Position:     tokenTolang.Position(ctx.LBRACE().GetSymbol()),
		Entries: entriesNode,
	}
	l.pushlang.Value(mapNode)
	l.logDebugAST("    Constructed ast.MapLiteralNode with %d entries", len(entriesNode))
}

// Exit.Expression_list implements generated.NeuroScriptListener.
func (l *neuroScriptListenerImpl) Exit.Expression_list(c *generated.ast.Expression_listContext) {
}
