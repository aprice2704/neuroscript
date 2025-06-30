// NeuroScript Version: 0.3.0
// File version: 5
// Purpose: Corrected list element processing to fix order-reversal bug.
// filename: pkg/core/ast_builder_collections.go
// nlines: 198
// risk_rating: MEDIUM

package core

import (
	"fmt" // Ensure fmt is imported

	generated "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- List Literal Handling ---

func (l *neuroScriptListenerImpl) ExitList_literal(ctx *generated.List_literalContext) {
	l.logDebugAST("--- Exit List_literal: %q", ctx.GetText())
	numElements := 0
	var elementsExpr []Expression

	if exprListOpt := ctx.Expression_list_opt(); exprListOpt != nil {
		if exprList := exprListOpt.Expression_list(); exprList != nil {
			numElements = len(exprList.AllExpression())
		}
	}

	if numElements > 0 {
		elementsRaw, ok := l.popNValues(numElements)
		if !ok {
			l.addError(ctx, "Internal error: Stack error popping %d elements for list literal", numElements)
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Stack error (list elements)"})
			return
		}

		elementsExpr = make([]Expression, numElements)
		for i := 0; i < numElements; i++ {
			// FIX: Assuming popNValues returns elements in their natural parsed order.
			// The explicit reversal `elementsRaw[numElements-1-i]` was incorrect and caused the bug.
			elemRaw := elementsRaw[i]
			elemExpr, isExpr := elemRaw.(Expression)
			if !isExpr {
				var elementCtx generated.IExpressionContext
				if exprListOpt := ctx.Expression_list_opt(); exprListOpt != nil {
					if exprList := exprListOpt.Expression_list(); exprList != nil && i < len(exprList.AllExpression()) {
						elementCtx = exprList.Expression(i) // Get the specific expression context
					}
				}
				errPos := tokenToPosition(ctx.GetStart()) // Fallback
				if elementCtx != nil {
					errPos = tokenToPosition(elementCtx.GetStart())
				}
				l.addError(ctx, "List element %d is not an Expression (got %T)", i+1, elemRaw)
				elementsExpr[i] = &ErrorNode{Pos: errPos, Message: fmt.Sprintf("List element type error: got %T", elemRaw)}
				// Continue processing other elements with error nodes if one is bad
			} else {
				elementsExpr[i] = elemExpr
			}
		}
	} else {
		elementsExpr = []Expression{} // Ensure it's an empty slice, not nil
	}

	listNode := &ListLiteralNode{
		Pos:      tokenToPosition(ctx.LBRACK().GetSymbol()),
		Elements: elementsExpr,
	}
	l.pushValue(listNode)
	l.logDebugAST("    Constructed ListLiteralNode with %d elements", len(elementsExpr))
}

// --- Map Literal Handling ---

func (l *neuroScriptListenerImpl) ExitMap_entry(ctx *generated.Map_entryContext) {
	l.logDebugAST("--- Exit Map_entry: %q", ctx.GetText())

	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping value for map entry")
		l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Stack error (map value)"})
		return
	}
	valueExpr, isExpr := valueRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "Internal error: Map entry value is not an Expression (got %T)", valueRaw)
		l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Type error (map value)"}) // Use general pos
		return
	}

	keyToken := ctx.STRING_LIT().GetSymbol()
	quotedKey := keyToken.GetText()
	unquotedKey, err := unescapeString(quotedKey)
	if err != nil {
		l.addErrorf(keyToken, "Invalid map key string literal: %v", err)
		l.pushValue(&ErrorNode{Pos: tokenToPosition(keyToken), Message: fmt.Sprintf("Invalid map key: %v", err)})
		return
	}
	keyNode := &StringLiteralNode{ // keyNode is *StringLiteralNode
		Pos:   tokenToPosition(keyToken),
		Value: unquotedKey,
		IsRaw: false,
	}

	entryNode := &MapEntryNode{
		Pos:   keyNode.Pos,
		Key:   keyNode,
		Value: valueExpr,
	}
	l.pushValue(entryNode) // Push the *MapEntryNode pointer
	l.logDebugAST("    Constructed MapEntryNode: Key=%q, Value=%T", keyNode.Value, valueExpr)
}

func (l *neuroScriptListenerImpl) ExitMap_literal(ctx *generated.Map_literalContext) {
	l.logDebugAST("--- Exit Map_literal: %q", ctx.GetText())
	numEntries := 0
	var entriesNode []*MapEntryNode

	if mapEntryListOpt := ctx.Map_entry_list_opt(); mapEntryListOpt != nil {
		if mapEntryList := mapEntryListOpt.Map_entry_list(); mapEntryList != nil {
			numEntries = len(mapEntryList.AllMap_entry())
		}
	}

	if numEntries > 0 {
		entryNodesRaw, ok := l.popNValues(numEntries)
		if !ok {
			l.addError(ctx, "Internal error: Stack error popping %d entries for map literal", numEntries)
			l.pushValue(&ErrorNode{Pos: tokenToPosition(ctx.GetStart()), Message: "Stack error (map entries)"})
			return
		}

		entriesNode = make([]*MapEntryNode, numEntries)
		for i := 0; i < numEntries; i++ {
			// popNValues returns in stack order (last pushed = first element). Reverse for parsed order.
			entryRaw := entryNodesRaw[numEntries-1-i]
			entryPtr, isPtr := entryRaw.(*MapEntryNode)
			if !isPtr {
				// Try to get position of the specific problematic entry from parse tree
				var entryCtx generated.IMap_entryContext
				if mapEntryListOpt := ctx.Map_entry_list_opt(); mapEntryListOpt != nil {
					if mapEntryList := mapEntryListOpt.Map_entry_list(); mapEntryList != nil && i < len(mapEntryList.AllMap_entry()) {
						entryCtx = mapEntryList.Map_entry(i)
					}
				}
				errPos := tokenToPosition(ctx.GetStart()) // Fallback
				if entryCtx != nil {
					errPos = tokenToPosition(entryCtx.GetStart())
				}
				l.addError(ctx, "Internal error: Map entry %d is not *MapEntryNode (got %T)", i+1, entryRaw)
				entriesNode[i] = &MapEntryNode{Pos: errPos, Key: &StringLiteralNode{Pos: errPos, Value: fmt.Sprintf("<error_key_%d>", i+1)}, Value: &ErrorNode{Pos: errPos, Message: "Invalid map entry type"}}
				// Continue processing other entries with error nodes if one is bad
			} else {
				entriesNode[i] = entryPtr
			}
		}
	} else {
		entriesNode = []*MapEntryNode{} // Ensure it's an empty slice, not nil
	}

	mapNode := &MapLiteralNode{
		Pos:     tokenToPosition(ctx.LBRACE().GetSymbol()),
		Entries: entriesNode,
	}
	l.pushValue(mapNode)
	l.logDebugAST("    Constructed MapLiteralNode with %d entries", len(entriesNode))
}

// ExitExpression_list implements generated.NeuroScriptListener.
func (l *neuroScriptListenerImpl) ExitExpression_list(c *generated.Expression_listContext) {
}
