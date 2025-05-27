// NeuroScript Version: 0.3.0
// File version: 0.0.3 // Balanced IF statement context, simplified ELSE handling.
// Last Modified: 2025-05-27
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Block Handling: Enter/Exit for Control Flow Statements ---

// enterBlockContext: Pushes parent step list pointer, creates new list for block
func (l *neuroScriptListenerImpl) enterBlockContext(blockType string) {
	l.logDebugAST(">>> Enter %s Statement Context", blockType)
	if l.currentSteps == nil {
		l.logger.Warn("Entering %s block, but currentSteps is nil. Starting fresh. This might indicate an earlier error.", blockType)
		// Push nil to signify the parent context was already lost.
		// This helps in debugging but doesn't fix the root cause of currentSteps being nil here.
		l.blockStepStack = append(l.blockStepStack, nil)
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
	l.logDebugAST("--- Enter exitBlockContext for %s", blockType)

	var completedSteps []Step
	if l.currentSteps != nil {
		completedSteps = *l.currentSteps
		l.logDebugAST("      Captured %d steps from context %p for %s block", len(completedSteps), l.currentSteps, blockType)
	} else {
		l.logger.Warn("%s: Exiting block, but currentSteps was nil when capturing. Returning empty slice.", blockType)
		completedSteps = []Step{} // Return empty, but an error should have been logged if this state is unexpected.
	}

	stackSize := len(l.blockStepStack)
	if stackSize == 0 {
		l.logger.Error("%s: Cannot restore parent context, stack empty! Setting currentSteps to nil.", blockType)
		// This is a critical error point. Add to listener errors.
		l.errors = append(l.errors, NewRuntimeError(ErrorCodeInternal, "["+blockType+"] Stack empty, cannot restore parent context.", nil))
		l.currentSteps = nil // Ensure currentSteps is nil if stack is unexpectedly empty
		return completedSteps
	}

	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex] // Pop stack

	l.currentSteps = parentStepsPtr // Restore parent pointer

	if l.currentSteps == nil {
		l.logDebugAST("<<< Exit exitBlockContext for %s (Restored parent, currentSteps is now nil. Stack size: %d)", blockType, len(l.blockStepStack))
		// If parentStepsPtr was nil, this means the parent context was already lost.
		// This is a symptom of an earlier error.
		l.logger.Warn("%s: Restored parent context is nil. This is problematic for appending subsequent steps.", blockType)
	} else {
		l.logDebugAST("<<< Exit exitBlockContext for %s (Restored parent, currentSteps points to list of len %d, Stack size: %d)", blockType, len(*l.currentSteps), len(l.blockStepStack))
	}
	return completedSteps
}

func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	if l.currentSteps == nil && len(l.blockStepStack) == 0 {
		l.logDebugAST(">>> Enter Statement_list (Likely top-level of proc or unmanaged block - currentSteps: %p)", l.currentSteps)
	} else {
		l.logDebugAST(">>> Enter Statement_list (currentSteps: %p, blockStack: %d)", l.currentSteps, len(l.blockStepStack))
	}
}

func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	l.logDebugAST("<<< Exit Statement_list (%s)", ctx.GetText())
}

// --- IF Statement ---
func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST(">>> Enter IF Statement Context (Parent currentSteps: %p)", l.currentSteps)
	// This single enterBlockContext is intended for the primary body of statements within the IF.
	// In the current simplified model, this will capture 'then' statements, and if 'else' exists,
	// 'else' statements will also be appended here unless EnterStatement_list becomes context-aware.
	l.enterBlockContext("IF_BODY")
}

func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("--- ExitIf_statement: Finalizing IF step (%s)", ctx.GetText())

	// This call to exitBlockContext matches the single call in EnterIf_statement.
	// It will capture all statements parsed into the currentSteps since EnterIf_statement.
	// In the current setup, this means 'then' and 'else' (if present) are co-mingled.
	allIfSteps := l.exitBlockContext("IF_BODY")
	l.logDebugAST("         Captured all IF body steps (count: %d)", len(allIfSteps))

	// These steps are now assigned to 'Body'. 'Else' remains empty here.
	// This is a simplification. A full solution needs distinct handling for then/else bodies.
	thenSteps := allIfSteps
	var elseSteps []Step

	if ctx.KW_ELSE() != nil {
		l.logger.Warn("IF statement with an ELSE clause: The current AST building logic will place ELSE block steps into the THEN block's 'Body'. Proper separation of THEN and ELSE steps requires more sophisticated listener logic to manage distinct contexts for each branch.")
		// elseSteps remains empty with this simplified approach.
		// A more robust fix would involve splitting combinedThenElseSteps or, ideally,
		// having separate contexts pushed/popped for then and else branches.
	}

	conditionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Stack error popping condition for IF statement")
		// If currentSteps is nil here, it means exitBlockContext failed to restore a valid parent.
		if l.currentSteps == nil {
			l.addError(ctx, "Parent context for IF statement became nil after exiting IF_BODY.")
		}
		return
	}
	conditionNode, isExpr := conditionRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "IF condition is not an Expression (got %T)", conditionRaw)
		return
	}
	l.logDebugAST("         Popped IF condition: %T", conditionNode)

	if l.currentSteps == nil {
		l.addError(ctx, "Cannot append IF step: currentSteps (parent) is nil after IF block processing.")
		return
	}

	ifStep := Step{
		Pos:  tokenToPosition(ctx.KW_IF().GetSymbol()),
		Type: "if",
		Cond: conditionNode,
		Body: thenSteps,
		Else: elseSteps, // elseSteps is empty in this simplified fix
	}
	*l.currentSteps = append(*l.currentSteps, ifStep)
	l.logDebugAST("         Appended IF Step: Cond=%T, ThenSteps=%d, ElseSteps=%d to parent context %p",
		ifStep.Cond, len(ifStep.Body), len(ifStep.Else), l.currentSteps)
}

// --- WHILE Statement ---
func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.loopDepth++
	l.logDebugAST(">>> Enter WHILE Statement Context (Loop Depth: %d)", l.loopDepth)
	l.enterBlockContext("WHILE-BODY")
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	defer func() {
		l.loopDepth--
		l.logDebugAST("<<< Exit WHILE Statement Final (Loop Depth: %d)", l.loopDepth)
	}()
	l.logDebugAST("--- ExitWhile_statement: Finalizing WHILE step (%s)", ctx.GetText())

	bodySteps := l.exitBlockContext("WHILE-BODY")
	l.logDebugAST("         Captured %d WHILE body steps", len(bodySteps))

	conditionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Stack error popping condition for WHILE statement")
		if l.currentSteps == nil {
			l.addError(ctx, "Parent context for WHILE statement became nil after exiting WHILE-BODY.")
		}
		return
	}
	conditionNode, isExpr := conditionRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "WHILE condition is not an Expression (got %T)", conditionRaw)
		return
	}
	l.logDebugAST("         Popped WHILE condition: %T", conditionNode)

	if l.currentSteps == nil {
		l.addError(ctx, "Cannot append WHILE step: currentSteps (parent) is nil after block exit.")
		return
	}

	whileStep := Step{
		Pos:  tokenToPosition(ctx.KW_WHILE().GetSymbol()),
		Type: "while",
		Cond: conditionNode,
		Body: bodySteps,
	}
	*l.currentSteps = append(*l.currentSteps, whileStep)
	l.logDebugAST("         Appended WHILE Step: Cond=%T, BodySteps=%d",
		whileStep.Cond, len(whileStep.Body))
}

// --- FOR EACH Statement ---
func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	l.loopDepth++
	l.logDebugAST(">>> Enter FOR EACH Statement Context (Loop Depth: %d)", l.loopDepth)
	l.enterBlockContext("FOR-EACH-BODY")
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	defer func() {
		l.loopDepth--
		l.logDebugAST("<<< Exit FOR EACH Statement Final (Loop Depth: %d)", l.loopDepth)
	}()
	l.logDebugAST("--- ExitFor_each_statement: Finalizing FOR EACH step (%s)", ctx.GetText())

	bodySteps := l.exitBlockContext("FOR-EACH-BODY")
	l.logDebugAST("         Captured %d FOR EACH body steps", len(bodySteps))

	collectionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Stack error popping collection expression for FOR EACH statement")
		if l.currentSteps == nil {
			l.addError(ctx, "Parent context for FOR_EACH statement became nil after exiting FOR-EACH-BODY.")
		}
		return
	}
	collectionNode, isExpr := collectionRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "FOR EACH collection is not an Expression (got %T)", collectionRaw)
		return
	}
	l.logDebugAST("         Popped FOR EACH collection: %T", collectionNode)

	loopVar := ""
	if idNode := ctx.IDENTIFIER(); idNode != nil {
		loopVar = idNode.GetText()
	} else {
		l.addError(ctx, "Missing IDENTIFIER for loop variable in FOR EACH statement")
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Cannot append FOR EACH step: currentSteps (parent) is nil after block exit.")
		return
	}

	forStep := Step{
		Pos:    tokenToPosition(ctx.KW_FOR().GetSymbol()),
		Type:   "for",
		Target: loopVar,
		Cond:   collectionNode,
		Body:   bodySteps,
	}
	*l.currentSteps = append(*l.currentSteps, forStep)
	l.logDebugAST("         Appended FOR EACH Step: Var=%q, Collection=%T, BodySteps=%d",
		forStep.Target, forStep.Cond, len(forStep.Body))
}

// --- ON ERROR Statement ---
func (l *neuroScriptListenerImpl) EnterOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST(">>> Enter ON_ERROR Statement Context")
	l.enterBlockContext("ON-ERROR-HANDLER")
}

func (l *neuroScriptListenerImpl) ExitOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST("--- ExitOnErrorStmt: Finalizing ON_ERROR step (%s)", ctx.GetText())

	handlerSteps := l.exitBlockContext("ON-ERROR-HANDLER")
	l.logDebugAST("         Captured %d ON_ERROR handler steps", len(handlerSteps))

	if l.currentProc == nil {
		l.addError(ctx, "ON_ERROR statement defined outside of a procedure context.")
		return
	}

	if l.currentSteps == nil {
		// This can happen if the procedure body was empty except for on_error,
		// or if a preceding block corrupted currentSteps for the procedure.
		l.addError(ctx, "Cannot append ON_ERROR step: currentSteps (parent for procedure) is nil.")
		// Attempt to recover by directly using procedure's step list if possible,
		// but this indicates a prior issue.
		if l.currentProc != nil { // currentProc should not be nil here due to above check
			l.logger.Warn("Attempting to recover currentSteps for ON_ERROR by using procedure's main step list.")
			l.currentSteps = &l.currentProc.Steps
		} else {
			return // Cannot proceed if currentProc is also nil
		}
	}

	// This check must pass if recovery or normal flow worked.
	if l.currentSteps == nil {
		l.addError(ctx, "FATAL: currentSteps still nil before appending ON_ERROR step, even after recovery attempt.")
		return
	}

	onErrorStep := Step{
		Pos:  tokenToPosition(ctx.KW_ON_ERROR().GetSymbol()),
		Type: "on_error",
		Body: handlerSteps,
	}
	*l.currentSteps = append(*l.currentSteps, onErrorStep)
	l.logDebugAST("         Appended ON_ERROR Step: HandlerSteps=%d", len(onErrorStep.Body))
	l.logDebugAST("<<< Exit ON_ERROR Statement Final")
}
