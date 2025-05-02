// NeuroScript Version: 0.3.0 // Keep user's version marker
// Last Modified: 2025-05-01 12:46:14 PDT // Keep user's timestamp
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
		l.logDebugAST("         %s: Pushed parent context %p. Stack size: %d. New steps: %p", blockType, l.blockStepStack[len(l.blockStepStack)-1], len(l.blockStepStack), l.currentSteps)
	}
}

// exitBlockContext: Pops parent step list pointer to restore context
// Returns the step slice that was just completed for the block.
func (l *neuroScriptListenerImpl) exitBlockContext(blockType string) []Step {
	l.logDebugAST("--- Enter exitBlockContext for %s", blockType) // Log Entry

	var completedSteps []Step
	if l.currentSteps != nil {
		completedSteps = *l.currentSteps
		l.logDebugAST("      Captured %d steps from context %p", len(completedSteps), l.currentSteps)
	} else {
		l.logger.Warn("%s: Exiting block, but currentSteps was nil when capturing. Returning empty slice.", blockType)
		completedSteps = []Step{}
	}

	stackSize := len(l.blockStepStack)
	if stackSize == 0 {
		l.logger.Error("%s: Cannot restore parent context, stack empty!", blockType)
		// +++ Add Debug Logging +++
		l.logger.Debug("exitBlockContext Restore Check", "blockType", blockType, "stackEmpty", true, "settingCurrentStepsToNil", true)
		// +++ End Debug Logging +++
		l.currentSteps = nil // Set currentSteps to nil as stack is empty
		return completedSteps
	}

	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex] // Pop stack

	// +++ Add Debug Logging +++
	restoringToNil := parentStepsPtr == nil
	l.logger.Debug("exitBlockContext Restore Check", "blockType", blockType, "stackEmpty", false, "restoringToNil", restoringToNil, "newStackSize", len(l.blockStepStack))
	if parentStepsPtr != nil {
		l.logger.Debug("exitBlockContext Restore Check", "restoredListLen", len(*parentStepsPtr))
	}
	// +++ End Debug Logging +++

	l.currentSteps = parentStepsPtr // Restore parent pointer

	l.logDebugAST("<<< Exit exitBlockContext for %s (currentSteps is now nil: %v)", blockType, l.currentSteps == nil) // Log Exit
	return completedSteps
}

// EnterStatement_list: Ensures currentSteps is initialized if needed
func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	// This basic version might conflict with the IF statement's specific handling.
	// Ensure this doesn't get called when inside an IF statement context where
	// EnterStatement_list_within_if should handle things.
	// ANTLR listener dispatch should handle this correctly based on context.
	if _, isIfParent := ctx.GetParent().(*gen.If_statementContext); !isIfParent {
		if l.currentSteps == nil {
			l.logDebugAST(">>> Enter Statement_list (Top level or non-IF block): currentSteps was nil, initializing.")
			// If we are not inside IF and currentSteps is nil, it might indicate an issue
			// This initialization might be problematic if it overrides a context expecting nil.
			// For now, let's keep it but be wary.
			// newSteps := make([]Step, 0)
			// l.currentSteps = &newSteps
		} else {
			l.logDebugAST(">>> Enter Statement_list (Top level or non-IF block): currentSteps already exists (%p)", l.currentSteps)
		}
	} else {
		// Let EnterStatement_list_within_if handle it
		l.logDebugAST(">>> Enter Statement_list (Within IF): Deferring to EnterStatement_list_within_if")
		l.EnterStatement_list_within_if(ctx)
	}

}

// ExitStatement_list should do nothing with the value stack or step context
func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	if _, isIfParent := ctx.GetParent().(*gen.If_statementContext); !isIfParent {
		l.logDebugAST("<<< Exit Statement_list (Top level or non-IF block)")
	} else {
		l.logDebugAST("<<< Exit Statement_list (Within IF)")
		// No action needed here, context popped by ExitIf_statement calling exitBlockContext
	}
}

// --- IF Statement ---
func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST(">>> Enter IF Statement Context")
	// Context pushed when entering THEN/ELSE via EnterStatement_list_within_if
}

// EnterStatement_list_within_if handles context for THEN/ELSE blocks
func (l *neuroScriptListenerImpl) EnterStatement_list_within_if(ctx *gen.Statement_listContext) {
	if ifCtx, ok := ctx.GetParent().(*gen.If_statementContext); ok {
		if ifCtx.KW_ELSE() != nil && ctx == ifCtx.Statement_list(1) {
			l.logDebugAST("      >>> Enter Statement_list (within IF: ELSE block)")
			l.enterBlockContext("IF-ELSE") // Prepare for ELSE block steps
		} else if ctx == ifCtx.Statement_list(0) {
			l.logDebugAST("      >>> Enter Statement_list (within IF: THEN block)")
			l.enterBlockContext("IF-THEN") // Prepare for THEN block steps
		} else {
			l.logger.Error("EnterStatement_list_within_if: Could not determine if THEN or ELSE block.")
			l.enterBlockContext("IF-UNKNOWN")
		}
	} else {
		l.logger.Error("EnterStatement_list_within_if: Parent context is not If_statementContext")
		l.enterBlockContext("IF-ERROR")
	}
}

func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("--- Exit If_statement Finalization ---")
	var elseSteps []Step = nil
	var thenSteps []Step
	var conditionNode Expression
	if ctx.KW_ELSE() != nil {
		elseSteps = l.exitBlockContext("IF-ELSE")
		l.logDebugAST("         Captured %d ELSE steps", len(elseSteps))
	}
	thenSteps = l.exitBlockContext("IF-THEN")
	l.logDebugAST("         Captured %d THEN steps", len(thenSteps))
	conditionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping condition for IF")
		return
	}
	conditionAsserted, ok := conditionRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: IF condition is not an Expression (got %T)", conditionRaw)
		return
	}
	conditionNode = conditionAsserted
	l.logDebugAST("         Popped IF condition: %T", conditionNode)
	if l.currentSteps == nil {
		l.logger.Error("ExitIf_statement: Parent step list nil after context restores. Cannot append IF step.")
		return
	}
	ifStep := Step{Pos: tokenToPosition(ctx.KW_IF().GetSymbol()), Type: "if", Cond: conditionNode, Value: thenSteps, ElseValue: elseSteps, Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, ifStep)
	thenLen := len(thenSteps)
	elseLen := 0
	if elseSteps != nil {
		elseLen = len(elseSteps)
	}
	l.logDebugAST("         Appended complete IF Step to parent: Cond=%T, THEN Steps=%d, ELSE Steps=%d", ifStep.Cond, thenLen, elseLen)
}

// --- WHILE Statement ---
func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.loopDepth++
	l.logDebugAST(">>> Enter WHILE Statement Context (Loop Depth: %d)", l.loopDepth)
	l.enterBlockContext("WHILE")
}
func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	defer func() { l.loopDepth--; l.logDebugAST("<<< Exit WHILE Statement Final (Loop Depth: %d)", l.loopDepth) }() // Changed log
	l.logDebugAST("--- Exit While_statement Context Capture ---")                                                   // Changed log
	bodySteps := l.exitBlockContext("WHILE")
	l.logDebugAST("         Captured %d WHILE body steps", len(bodySteps))
	conditionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping condition for WHILE")
		return
	}
	conditionNode, ok := conditionRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: WHILE condition is not an Expression (got %T)", conditionRaw)
		return
	}
	l.logDebugAST("         Popped WHILE condition: %T", conditionNode)
	if l.currentSteps == nil {
		l.logger.Error("ExitWhile_statement: Parent step list nil after context restore. Cannot append WHILE step.")
		return
	}
	whileStep := Step{Pos: tokenToPosition(ctx.KW_WHILE().GetSymbol()), Type: "while", Cond: conditionNode, Value: bodySteps, Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, whileStep)
	l.logDebugAST("         Appended WHILE Step: Cond=%T, Steps=%d", conditionNode, len(bodySteps))
}

// --- FOR EACH Statement ---
func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	l.loopDepth++
	l.logDebugAST(">>> Enter FOR EACH Statement Context (Loop Depth: %d)", l.loopDepth)
	l.enterBlockContext("FOR")
}
func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	defer func() {
		l.loopDepth--
		l.logDebugAST("<<< Exit FOR EACH Statement Final (Loop Depth: %d)", l.loopDepth)
	}() // Changed log
	l.logDebugAST("--- Exit For_each_statement Context Capture ---") // Changed log
	bodySteps := l.exitBlockContext("FOR")
	l.logDebugAST("         Captured %d FOR body steps", len(bodySteps))
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
	l.logDebugAST("         Popped FOR collection: %T", collectionNode)
	loopVar := ""
	if id := ctx.IDENTIFIER(); id != nil {
		loopVar = id.GetText()
	} else {
		l.addError(ctx, "Internal error: Missing IDENTIFIER for loop variable in FOR statement")
		return
	}
	if l.currentSteps == nil {
		l.logger.Error("ExitFor_each_statement: Parent step list nil after context restore. Cannot append FOR step.")
		return
	}
	forStep := Step{Pos: tokenToPosition(ctx.KW_FOR().GetSymbol()), Type: "for", Target: loopVar, Cond: collectionNode, Value: bodySteps, Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, forStep)
	l.logDebugAST("         Appended FOR Step: Var=%q, Collection=%T, Steps=%d", loopVar, collectionNode, len(bodySteps))
}

// --- ON ERROR Statement ---
func (l *neuroScriptListenerImpl) EnterOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST(">>> Enter ON_ERROR Statement Context") // Changed log
	l.enterBlockContext("ON_ERROR")
}
func (l *neuroScriptListenerImpl) ExitOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST("--- Exit OnErrorStmt Context Capture ---") // Changed log
	handlerSteps := l.exitBlockContext("ON_ERROR")
	l.logDebugAST("         Captured %d ON_ERROR handler steps", len(handlerSteps))
	if l.currentSteps == nil {
		l.logger.Error("ExitOnErrorStmt: Parent step list nil after context restore. Cannot append ON_ERROR step.")
		return
	}
	onErrorStep := Step{Pos: tokenToPosition(ctx.KW_ON_ERROR().GetSymbol()), Type: "on_error", Value: handlerSteps, Metadata: make(map[string]string)}
	*l.currentSteps = append(*l.currentSteps, onErrorStep)
	l.logDebugAST("         Appended ON_ERROR Step: HandlerSteps=%d", len(handlerSteps))
	l.logDebugAST("<<< Exit ON_ERROR Statement Final") // Changed log
}
