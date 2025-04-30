// filename: pkg/core/ast_builder_collections.go
package core

import (
	// "strconv" // Handled by helpers now
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- List Literal Handling ---
// *** MODIFIED: Assert element types, set Pos, handle errors ***

// ExitList_literal pops element nodes and pushes a ListLiteralNode
func (l *neuroScriptListenerImpl) ExitList_literal(ctx *gen.List_literalContext) {
	l.logDebugAST("--- Exit List_literal: %q", ctx.GetText())
	numElements := 0
	var elementsExpr []Expression // Slice to hold asserted expressions

	// Check if the optional expression list exists and has content
	if exprListOpt := ctx.Expression_list_opt(); exprListOpt != nil {
		if exprList := exprListOpt.Expression_list(); exprList != nil {
			numElements = len(exprList.AllExpression())
		}
	}

	// Pop element nodes pushed by their respective visitors
	if numElements > 0 {
		elementsRaw, ok := l.popNValues(numElements)
		if !ok {
			l.addError(ctx, "Internal error: Stack error popping %d elements for list literal", numElements)
			l.pushValue(nil) // Push error marker
			return
		}

		// Assert each element is an Expression
		elementsExpr = make([]Expression, numElements)
		for i, elemRaw := range elementsRaw {
			elemExpr, ok := elemRaw.(Expression)
			if !ok {
				// Try to get position of the specific bad element
				var elementCtx gen.IExpressionContext
				if exprListOpt := ctx.Expression_list_opt(); exprListOpt != nil {
					if exprList := exprListOpt.Expression_list(); exprList != nil && i < len(exprList.AllExpression()) {
						elementCtx = exprList.Expression(i)
					}
				}
				if elementCtx != nil {
					l.addError(elementCtx, "Internal error: List element %d is not an Expression (got %T)", i+1, elemRaw)
				} else {
					l.addError(ctx, "Internal error: List element %d is not an Expression (got %T)", i+1, elemRaw)
				}

				l.pushValue(nil) // Push error marker
				return
			}
			elementsExpr[i] = elemExpr
		}
	} else {
		// No elements, create empty slice
		elementsExpr = []Expression{}
	}

	// Create the ListLiteralNode
	listNode := &ListLiteralNode{
		Pos:      tokenToPosition(ctx.LBRACK().GetSymbol()), // Position of '['
		Elements: elementsExpr,
	}

	l.pushValue(listNode) // Push the constructed list node pointer
	l.logDebugAST("    Constructed ListLiteralNode with %d elements", len(elementsExpr))
}

// --- Map Literal Handling ---
// *** MODIFIED: Removed EnterMap_entry, handle key/value in ExitMap_entry, assert types ***

// EnterMap_entry is no longer needed. Removed l.currentMapKey field as well.

// ExitMap_entry pops the Value node, processes the Key literal, combines them, pushes MapEntryNode
func (l *neuroScriptListenerImpl) ExitMap_entry(ctx *gen.Map_entryContext) {
	l.logDebugAST("--- Exit Map_entry: %q", ctx.GetText())

	// 1. Pop the value expression (pushed by visiting ctx.Expression())
	valueRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping value for map entry")
		l.pushValue(nil)
		return
	}
	valueExpr, ok := valueRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: Map entry value is not an Expression (got %T)", valueRaw)
		l.pushValue(nil)
		return
	}

	// 2. Process the key literal directly from the context
	keyToken := ctx.STRING_LIT().GetSymbol()
	quotedKey := keyToken.GetText()
	unquotedKey, err := unescapeString(quotedKey) // Use helper
	if err != nil {
		l.addErrorf(keyToken, "Invalid map key string literal: %v", err)
		l.pushValue(nil)
		return
	}
	keyNode := &StringLiteralNode{ // Create the key node here
		Pos:   tokenToPosition(keyToken),
		Value: unquotedKey,
		IsRaw: false, // Map keys cannot be raw strings per grammar
	}

	// 3. Create and push the MapEntryNode
	entryNode := &MapEntryNode{
		Pos:   keyNode.Pos, // Position of the entry is the position of the key
		Key:   *keyNode,    // Assign the created StringLiteralNode value
		Value: valueExpr,   // Assign the asserted Expression value
	}
	l.pushValue(entryNode) // Push the MapEntryNode pointer onto the stack
	l.logDebugAST("    Constructed MapEntryNode: Key=%q, Value=%T", keyNode.Value, valueExpr)
}

// ExitMap_literal pops MapEntryNodes and pushes a MapLiteralNode
func (l *neuroScriptListenerImpl) ExitMap_literal(ctx *gen.Map_literalContext) {
	l.logDebugAST("--- Exit Map_literal: %q", ctx.GetText())
	numEntries := 0
	var entriesNode []MapEntryNode // Slice to hold asserted MapEntryNode structs

	// Check if the optional map entry list exists and has content
	if mapEntryListOpt := ctx.Map_entry_list_opt(); mapEntryListOpt != nil {
		if mapEntryList := mapEntryListOpt.Map_entry_list(); mapEntryList != nil {
			numEntries = len(mapEntryList.AllMap_entry())
		}
	}

	// Pop MapEntryNode pointers pushed by ExitMap_entry
	if numEntries > 0 {
		entryNodesRaw, ok := l.popNValues(numEntries)
		if !ok {
			l.addError(ctx, "Internal error: Stack error popping %d entries for map literal", numEntries)
			l.pushValue(nil)
			return
		}

		// Assert each popped value is a *MapEntryNode pointer and dereference
		entriesNode = make([]MapEntryNode, numEntries)
		for i, entryRaw := range entryNodesRaw {
			entryPtr, ok := entryRaw.(*MapEntryNode) // Expecting pointer pushed by ExitMap_entry
			if !ok {
				l.addError(ctx, "Internal error: Map entry %d is not *MapEntryNode (got %T)", i+1, entryRaw)
				l.pushValue(nil)
				return
			}
			entriesNode[i] = *entryPtr // Dereference to store the struct value
		}
	} else {
		// No entries, create empty slice
		entriesNode = []MapEntryNode{}
	}

	// Create the MapLiteralNode
	mapNode := &MapLiteralNode{
		Pos:     tokenToPosition(ctx.LBRACE().GetSymbol()), // Position of '{'
		Entries: entriesNode,
	}
	l.pushValue(mapNode) // Push the constructed map node pointer
	l.logDebugAST("    Constructed MapLiteralNode with %d entries", len(entriesNode))
}
