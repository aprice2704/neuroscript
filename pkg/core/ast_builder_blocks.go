// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 12:46:14 PDT
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Block Handling: Enter/Exit for Control Flow Statements ---

// enterBlockContext: Pushes parent step list pointer, creates new list for block
func (l *neuroScriptListenerImpl) enterBlockContext(blockType string) {
	l.logDebugAST(">>> Enter %s Statement Context", blockType)
	if l.currentSteps == nil {
		l.logger.Warn("Entering %s block, but currentSteps is nil. Starting fresh.", blockType)
		l.blockStepStack = append(l.blockStepStack, nil) // Push nil parent
		newSteps := make([]Step, 0)
		l.currentSteps = &newSteps
	} else {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps) // Push pointer to parent list
		newSteps := make([]Step, 0)
		l.currentSteps = &newSteps // Start new list for the block body
		l.logDebugAST("      %s: Pushed parent context %p. Stack size: %d. New steps: %p", blockType, l.blockStepStack[len(l.blockStepStack)-1], len(l.blockStepStack), l.currentSteps)
	}
}

// exitBlockContext: Pops parent step list pointer to restore context
// Returns the step slice that was just completed for the block.
// This allows the Exit* methods for blocks to capture the steps directly.
func (l *neuroScriptListenerImpl) exitBlockContext(blockType string) []Step {
	l.logDebugAST("<<< Exit %s Statement Context - Capturing block steps & Restoring parent", blockType)

	// Capture the steps built within this block *before* restoring parent
	var completedSteps []Step
	if l.currentSteps != nil {
		completedSteps = *l.currentSteps
		l.logDebugAST("      %s: Captured %d steps from context %p", blockType, len(completedSteps), l.currentSteps)
	} else {
		l.logger.Warn("%s: Exiting block, but currentSteps is nil. Returning empty slice.", blockType)
		completedSteps = []Step{} // Return empty, not nil
	}

	// Now restore parent context
	stackSize := len(l.blockStepStack)
	if stackSize == 0 {
		l.logger.Error("%s: Cannot restore parent context, stack empty!", blockType)
		l.currentSteps = nil  // Lost context
		return completedSteps // Return what we captured anyway
	}
	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex] // Pop stack

	// Restore l.currentSteps
	l.currentSteps = parentStepsPtr

	if l.currentSteps == nil {
		l.logger.Warn("%s: Restored parent context, but it was nil (Stack size: %d)", blockType, len(l.blockStepStack))
	} else {
		l.logDebugAST("      %s: Restored parent context %p (Stack size: %d)", blockType, l.currentSteps, len(l.blockStepStack))
	}

	return completedSteps // Return the captured steps
}

// EnterStatement_list: Ensures currentSteps is initialized if needed
func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	if l.currentSteps == nil {
		l.logDebugAST(">>> Enter Statement_list: currentSteps was nil, initializing.")
		newSteps := make([]Step, 0)
		l.currentSteps = &newSteps
	} else {
		l.logDebugAST(">>> Enter Statement_list: currentSteps already exists (%p)", l.currentSteps)
	}
}

// --- CORRECTED: ExitStatement_list should do nothing with the value stack ---
func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	l.logDebugAST("<<< Exit Statement_list")
	// DO NOTHING HERE - Steps are handled by block context methods now.
}

// --- IF Statement ---
// EnterIf_statement and ExitIf_statement now handle multiple statement lists
// for THEN and potential ELSE block.

func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST(">>> Enter IF Statement Context")
	// Do NOT call enterBlockContext here yet. We handle it when entering the THEN list.
	// Also, do NOT modify loopDepth here.
}

// EnterStatement_list is now responsible for context switching within IF/ELSE
// We override the base EnterStatement_list from above temporarily when inside IF.
func (l *neuroScriptListenerImpl) EnterStatement_list_within_if(ctx *gen.Statement_listContext) {
	// Determine if this statement list belongs to the THEN or ELSE part
	// Check the parent context (which should be If_statementContext)
	if ifCtx, ok := ctx.GetParent().(*gen.If_statementContext); ok {
		// If there's an ELSE keyword and this is the second statement list
		if ifCtx.KW_ELSE() != nil && ctx == ifCtx.Statement_list(1) {
			l.logDebugAST(">>> Enter Statement_list (within IF: ELSE block)")
			l.enterBlockContext("IF-ELSE") // Prepare for ELSE block steps
		} else if ctx == ifCtx.Statement_list(0) {
			// This is the first (or only) statement list, must be the THEN block
			l.logDebugAST(">>> Enter Statement_list (within IF: THEN block)")
			l.enterBlockContext("IF-THEN") // Prepare for THEN block steps
		} else {
			// Should not happen
			l.logger.Error("EnterStatement_list_within_if: Could not determine if THEN or ELSE block.")
			l.enterBlockContext("IF-UNKNOWN") // Enter a context anyway
		}
	} else {
		// Parent wasn't If_statementContext - log error, this shouldn't occur
		l.logger.Error("EnterStatement_list_within_if: Parent context is not If_statementContext")
		l.enterBlockContext("IF-ERROR")
	}
}

func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("--- Exit If_statement Finalization")

	var elseSteps []Step = nil // Default to nil (no else block)
	var thenSteps []Step
	var conditionNode Expression

	// Capture Else Steps (if they exist) by calling exitBlockContext
	if ctx.KW_ELSE() != nil {
		// Exit the context created for the ELSE block's statement list
		elseSteps = l.exitBlockContext("IF-ELSE")
		l.logDebugAST("      Captured %d ELSE steps", len(elseSteps))
	}

	// Capture Then Steps by calling exitBlockContext
	// Exit the context created for the THEN block's statement list
	thenSteps = l.exitBlockContext("IF-THEN")
	l.logDebugAST("      Captured %d THEN steps", len(thenSteps))

	// Pop Condition Expression (still needed from value stack)
	conditionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping condition for IF")
		// Do not call exitBlockContext again here, already done for THEN/ELSE
		return
	}
	conditionAsserted, ok := conditionRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: IF condition is not an Expression (got %T)", conditionRaw)
		return
	}
	conditionNode = conditionAsserted
	l.logDebugAST("      Popped IF condition: %T", conditionNode)

	// Parent context should now be restored (by the last exitBlockContext call)
	if l.currentSteps == nil {
		l.logger.Error("ExitIf_statement: Parent step list nil after context restores. Cannot append IF step.")
		return
	}

	// Create and Append IF Step to Parent
	ifStep := Step{
		Pos:       tokenToPosition(ctx.KW_IF().GetSymbol()),
		Type:      "if",
		Cond:      conditionNode,
		Value:     thenSteps, // Assign the captured THEN []Step slice
		ElseValue: elseSteps, // Assign the captured ELSE []Step slice (or nil)
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, ifStep)
	thenLen := len(thenSteps)
	elseLen := 0
	if elseSteps != nil {
		elseLen = len(elseSteps)
	}
	l.logDebugAST("      Appended complete IF Step to parent: Cond=%T, THEN Steps=%d, ELSE Steps=%d", ifStep.Cond, thenLen, elseLen)
}

// --- WHILE Statement ---
func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.loopDepth++ // <<< Increment loop depth
	l.logDebugAST(">>> Enter WHILE Statement Context (Loop Depth: %d)", l.loopDepth)
	l.enterBlockContext("WHILE") // Push parent, prepare for body steps
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	defer func() { // <<< Decrement loop depth when function exits
		l.loopDepth--
		l.logDebugAST("<<< Exit WHILE Statement Context (Loop Depth: %d)", l.loopDepth)
	}()

	l.logDebugAST("--- Exit While_statement Finalization")

	// Capture Body Steps *before* restoring parent context
	bodySteps := l.exitBlockContext("WHILE")
	l.logDebugAST("      Captured %d WHILE body steps", len(bodySteps))

	// Pop Condition Expression
	conditionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping condition for WHILE")
		// l.exitBlockContext("WHILE - Error") // Context already exited
		return
	}
	conditionNode, ok := conditionRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: WHILE condition is not an Expression (got %T)", conditionRaw)
		// l.exitBlockContext("WHILE - Error") // Context already exited
		return
	}
	l.logDebugAST("      Popped WHILE condition: %T", conditionNode)

	// Parent context should be restored now
	if l.currentSteps == nil {
		l.logger.Error("ExitWhile_statement: Parent step list nil after context restore. Cannot append WHILE step.")
		return
	}

	// Create and Append WHILE Step to Parent
	whileStep := Step{
		Pos:       tokenToPosition(ctx.KW_WHILE().GetSymbol()),
		Type:      "while",
		Cond:      conditionNode,
		Value:     bodySteps, // Assign captured steps
		ElseValue: nil,
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, whileStep)
	l.logDebugAST("      Appended WHILE Step: Cond=%T, Steps=%d", conditionNode, len(bodySteps))
}

// --- FOR EACH Statement ---
func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	l.loopDepth++ // <<< Increment loop depth
	l.logDebugAST(">>> Enter FOR EACH Statement Context (Loop Depth: %d)", l.loopDepth)
	l.enterBlockContext("FOR") // Push parent, prepare for body steps
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	defer func() { // <<< Decrement loop depth when function exits
		l.loopDepth--
		l.logDebugAST("<<< Exit FOR EACH Statement Context (Loop Depth: %d)", l.loopDepth)
	}()

	l.logDebugAST("--- Exit For_each_statement Finalization")

	// Capture Body Steps
	bodySteps := l.exitBlockContext("FOR")
	l.logDebugAST("      Captured %d FOR body steps", len(bodySteps))

	// Pop Collection Expression
	collectionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping collection expression for FOR")
		return
	}
	collectionNode, ok := collectionRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: FOR collection is not an Expression (got %T)", collectionRaw)
		return
	}
	l.logDebugAST("      Popped FOR collection: %T", collectionNode)

	// Get Loop Variable Name
	loopVar := ""
	if id := ctx.IDENTIFIER(); id != nil {
		loopVar = id.GetText()
	} else {
		l.addError(ctx, "Internal error: Missing IDENTIFIER for loop variable in FOR statement")
		return
	}

	// Parent context should be restored
	if l.currentSteps == nil {
		l.logger.Error("ExitFor_each_statement: Parent step list nil after context restore. Cannot append FOR step.")
		return
	}

	// Create and Append FOR Step to Parent
	forStep := Step{
		Pos:       tokenToPosition(ctx.KW_FOR().GetSymbol()),
		Type:      "for", // "for" still seems appropriate for for-each
		Target:    loopVar,
		Cond:      collectionNode, // Collection expr -> Cond
		Value:     bodySteps,      // Body steps -> Value
		ElseValue: nil,
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, forStep)
	l.logDebugAST("      Appended FOR Step: Var=%q, Collection=%T, Steps=%d", loopVar, collectionNode, len(bodySteps))
}

// --- ON ERROR Statement ---
func (l *neuroScriptListenerImpl) EnterOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.enterBlockContext("ON_ERROR") // Push parent, prepare for handler steps
	// Does not affect loop depth
}

func (l *neuroScriptListenerImpl) ExitOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST("--- Exit OnErrorStmt Finalization")

	// Capture Handler Steps
	handlerSteps := l.exitBlockContext("ON_ERROR")
	l.logDebugAST("      Captured %d ON_ERROR handler steps", len(handlerSteps))

	// Parent context should be restored
	if l.currentSteps == nil {
		l.logger.Error("ExitOnErrorStmt: Parent step list nil after context restore. Cannot append ON_ERROR step.")
		return
	}

	// Create and Append ON_ERROR Step to Parent
	onErrorStep := Step{
		Pos:       tokenToPosition(ctx.KW_ON_ERROR().GetSymbol()),
		Type:      "on_error",
		Target:    "",
		Cond:      nil,
		Value:     handlerSteps, // Handler steps -> Value
		ElseValue: nil,
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, onErrorStep)
	l.logDebugAST("      Appended ON_ERROR Step: HandlerSteps=%d", len(handlerSteps))
}
