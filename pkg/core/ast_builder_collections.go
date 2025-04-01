// pkg/core/ast_builder_collections.go
package core

import (
	"strconv"

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- List Literal Handling ---

// ExitList_literal pops element nodes and pushes a ListLiteralNode
func (l *neuroScriptListenerImpl) ExitList_literal(ctx *gen.List_literalContext) {
	l.logDebugAST(">>> Exit List_literal: %q", ctx.GetText())
	numElements := 0
	if ctx.Expression_list_opt() != nil && ctx.Expression_list_opt().Expression_list() != nil {
		numElements = len(ctx.Expression_list_opt().Expression_list().AllExpression())
	}

	elements, ok := l.popNValues(numElements) // Pop element nodes
	if !ok {
		l.pushValue(nil)
		return
	} // Error handling

	listNode := ListLiteralNode{Elements: elements}
	l.pushValue(listNode) // Push the constructed list node
	l.logDebugAST("    Constructed ListLiteralNode")
}

// --- Map Literal Handling ---

// EnterMap_entry stores the parsed Key node
func (l *neuroScriptListenerImpl) EnterMap_entry(ctx *gen.Map_entryContext) {
	l.logDebugAST(">>> Enter Map_entry: %q", ctx.GetText())
	keyText := ctx.STRING_LIT().GetText()
	unquotedKey, err := strconv.Unquote(keyText)
	if err != nil {
		l.logger.Printf("[ERROR] Failed to unquote map key literal: %q - %v", keyText, err)
		l.currentMapKey = &StringLiteralNode{Value: keyText} // Store raw as fallback
	} else {
		l.currentMapKey = &StringLiteralNode{Value: unquotedKey}
	}
}

// ExitMap_entry pops the Value node, combines with stored Key node, pushes MapEntryNode
func (l *neuroScriptListenerImpl) ExitMap_entry(ctx *gen.Map_entryContext) {
	l.logDebugAST("<<< Exit Map_entry: %q", ctx.GetText())
	valueNode, ok := l.popValue()
	if !ok || l.currentMapKey == nil {
		l.logger.Printf("[ERROR] Failed to pop value or key missing for map entry: %q", ctx.GetText())
		l.currentMapKey = nil
		l.pushValue(nil) // Push error marker?
		return
	}

	entry := MapEntryNode{
		Key:   *l.currentMapKey, // Use stored StringLiteralNode
		Value: valueNode,        // Use popped value node
	}
	l.pushValue(entry)    // Push the MapEntryNode onto the stack
	l.currentMapKey = nil // Reset
	l.logDebugAST("    Constructed MapEntryNode")
}

// ExitMap_literal pops MapEntryNodes and pushes a MapLiteralNode
func (l *neuroScriptListenerImpl) ExitMap_literal(ctx *gen.Map_literalContext) {
	l.logDebugAST(">>> Exit Map_literal: %q", ctx.GetText())
	numEntries := 0
	if ctx.Map_entry_list_opt() != nil && ctx.Map_entry_list_opt().Map_entry_list() != nil {
		numEntries = len(ctx.Map_entry_list_opt().Map_entry_list().AllMap_entry())
	}

	entryNodesRaw, ok := l.popNValues(numEntries) // Pop MapEntryNodes (as interface{})
	if !ok {
		l.pushValue(nil)
		return
	} // Error handling

	// Convert []interface{} back to []MapEntryNode
	entries := make([]MapEntryNode, numEntries)
	validEntries := true
	for i := 0; i < numEntries; i++ {
		entry, ok := entryNodesRaw[i].(MapEntryNode)
		if !ok {
			l.logger.Printf("[ERROR] Expected MapEntryNode on stack, got %T for map literal entry %d: %q", entryNodesRaw[i], i, ctx.GetText())
			validEntries = false
			break
		}
		entries[i] = entry
	}
	if !validEntries {
		l.pushValue(nil)
		return
	} // Error handling

	mapNode := MapLiteralNode{Entries: entries}
	l.pushValue(mapNode) // Push the constructed map node
	l.logDebugAST("    Constructed MapLiteralNode")
}
