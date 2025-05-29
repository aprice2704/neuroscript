// NeuroScript Version: 0.3.1
// File version: 0.0.5 // Refined warnings in enter/exitBlockContext for top-level nil currentSteps.
// Last Modified: 2025-05-27
package core

import (
	"fmt"

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Block Handling: Enter/Exit for Control Flow Statements ---

func (l *neuroScriptListenerImpl) enterBlockContext(blockType string) {
	l.logDebugAST(">>> Enter %s Block Context (currentSteps before: %p, stack size: %d)", blockType, l.currentSteps, len(l.blockStepStack))
	// isTopLevelContextForProcedureBody is true if currentSteps is nil AND the block stack is empty.
	// This typically means we're entering the main Statement_list of a Procedure.
	isTopLevelContextForProcedureBody := (l.currentSteps == nil && len(l.blockStepStack) == 0)

	if l.currentSteps == nil {
		if !isTopLevelContextForProcedureBody {
			// Warn only if currentSteps is nil unexpectedly (i.e., not at the very top level of a new procedure's statement list)
			l.logger.Warn(fmt.Sprintf("Entering %s block, but currentSteps is nil (and not recognized as top-level for a procedure). Starting fresh. This might indicate an earlier error.", blockType))
		} else {
			l.logDebugAST("     Entering %s block at top-level (currentSteps is nil, stack empty). This is expected for a procedure's main body.", blockType)
		}
		l.blockStepStack = append(l.blockStepStack, nil) // Push nil as the "parent context" placeholder for top-level
	} else {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps) // Push pointer to actual parent list
	}

	newSteps := make([]Step, 0)
	l.currentSteps = &newSteps // Start new list for the current block body
	l.logDebugAST("         %s: Pushed parent context. Stack size: %d. New currentSteps: %p", blockType, len(l.blockStepStack), l.currentSteps)
}

func (l *neuroScriptListenerImpl) exitBlockContext(blockType string) []Step {
	l.logDebugAST("--- Enter exitBlockContext for %s (currentSteps: %p, stack size: %d)", blockType, l.currentSteps, len(l.blockStepStack))
	var completedSteps []Step
	if l.currentSteps != nil {
		completedSteps = *l.currentSteps
		l.logDebugAST("      Captured %d steps from context %p for %s block", len(completedSteps), l.currentSteps, blockType)
	} else {
		l.logger.Warn(fmt.Sprintf("%s: Exiting block, but currentSteps (for this block) was nil when capturing. Returning empty slice. This indicates an issue within this block's lifecycle.", blockType))
		completedSteps = []Step{}
	}

	stackSize := len(l.blockStepStack)
	if stackSize == 0 {
		l.logger.Error(fmt.Sprintf("%s: Cannot restore parent context, blockStepStack is empty! Setting currentSteps to nil. This is a critical error.", blockType))
		l.errors = append(l.errors, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("[%s] Stack empty, cannot restore parent context.", blockType), nil))
		l.currentSteps = nil
		return completedSteps
	}

	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex] // Pop stack
	l.currentSteps = parentStepsPtr                 // Restore parent pointer

	if l.currentSteps == nil {
		// If, after popping, the blockStepStack is now empty, it means we've restored to the top-level context (e.g., program/procedure global scope).
		// In this specific case (restoring to top-level), currentSteps being nil is expected.
		isRestoredToTopLevel := (len(l.blockStepStack) == 0)
		if !isRestoredToTopLevel {
			// Warn only if nil is restored AND we are not at the absolute top level (meaning an intermediate parent context was nil).
			l.logger.Warn(fmt.Sprintf("%s: Restored parent context is nil (and not at program/procedure top-level). This is problematic for appending subsequent steps.", blockType))
		} else {
			l.logDebugAST("<<< Exit %s: Restored parent context to nil (top-level expected for procedure body). Stack size: %d", blockType, len(l.blockStepStack))
		}
	} else {
		l.logDebugAST("<<< Exit %s: Restored parent context %p (list len %d). Stack size: %d", blockType, l.currentSteps, len(*l.currentSteps), len(l.blockStepStack))
	}
	return completedSteps
}

// EnterStatement_list creates a new context for collecting steps for this list.
func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	blockContextName := fmt.Sprintf("SL_%p", ctx)
	l.logDebugAST(">>> Enter Statement_list (Context: %s, Parent currentSteps: %p, Stack: %d)", blockContextName, l.currentSteps, len(l.blockStepStack))
	l.enterBlockContext(blockContextName)
}

// ExitStatement_list finalizes the steps for this list and pushes them onto the valueStack.
func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	blockContextName := fmt.Sprintf("SL_%p", ctx)
	completedSteps := l.exitBlockContext(blockContextName)
	l.pushValue(completedSteps)
	l.logDebugAST("<<< Exit Statement_list (Context: %s, Pushed %d steps to value stack. Restored currentSteps: %p)", blockContextName, len(completedSteps), l.currentSteps)
}

func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST(">>> Enter IF Statement (raw text: %s). Value stack size: %d", ctx.GetText(), len(l.valueStack))
}

func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("--- ExitIf_statement: Finalizing IF step (%s). Value stack size before pops: %d", ctx.GetText(), len(l.valueStack))

	var thenSteps, elseSteps []Step

	if ctx.KW_ELSE() != nil {
		elseStepsRaw, okElse := l.popValue()
		if !okElse {
			l.addError(ctx, "Stack error popping ELSE steps for IF statement")
			return
		}
		elseStepsCasted, isElseSteps := elseStepsRaw.([]Step)
		if !isElseSteps {
			l.addError(ctx, "ELSE steps are not []Step (got %T value: %v)", elseStepsRaw, elseStepsRaw)
			l.pushValue(elseStepsRaw)
			return
		}
		elseSteps = elseStepsCasted
		l.logDebugAST("         Popped IF else_steps: Count=%d", len(elseSteps))
	}

	thenStepsRaw, okThen := l.popValue()
	if !okThen {
		l.addError(ctx, "Stack error popping THEN steps for IF statement")
		if ctx.KW_ELSE() != nil {
			l.pushValue(elseSteps) // Push back elseSteps if thenSteps pop failed
		}
		return
	}
	thenStepsCasted, isThenSteps := thenStepsRaw.([]Step)
	if !isThenSteps {
		l.addError(ctx, "THEN steps are not []Step (got %T value: %v)", thenStepsRaw, thenStepsRaw)
		l.pushValue(thenStepsRaw)
		if ctx.KW_ELSE() != nil {
			l.pushValue(elseSteps)
		}
		return
	}
	thenSteps = thenStepsCasted
	l.logDebugAST("         Popped IF then_steps: Count=%d", len(thenSteps))

	conditionRaw, okCondition := l.popValue()
	if !okCondition {
		l.addError(ctx, "Stack error popping condition for IF statement")
		l.pushValue(thenSteps)
		if ctx.KW_ELSE() != nil {
			l.pushValue(elseSteps)
		}
		return
	}
	conditionNode, isExpr := conditionRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "IF condition is not an Expression (got %T value: %v)", conditionRaw, conditionRaw)
		l.pushValue(conditionRaw)
		l.pushValue(thenSteps)
		if ctx.KW_ELSE() != nil {
			l.pushValue(elseSteps)
		}
		return
	}
	l.logDebugAST("         Popped IF condition: %T", conditionNode)

	if l.currentSteps == nil {
		if !(len(l.blockStepStack) == 0 && l.currentProc != nil) {
			l.addError(ctx, "Cannot append IF step: currentSteps (parent context) is nil unexpectedly.")
			return
		}
	}

	ifStep := Step{
		Pos:  tokenToPosition(ctx.KW_IF().GetSymbol()),
		Type: "if",
		Cond: conditionNode,
		Body: thenSteps,
		Else: elseSteps,
	}

	if l.currentSteps != nil {
		*l.currentSteps = append(*l.currentSteps, ifStep)
		l.logDebugAST("         Appended IF Step to currentSteps list (%p): ThenSteps=%d, ElseSteps=%d",
			l.currentSteps, len(ifStep.Body), len(ifStep.Else))
	} else if l.currentProc != nil {
		l.currentProc.Steps = append(l.currentProc.Steps, ifStep)
		l.logDebugAST("         Appended IF Step to currentProc.Steps: ThenSteps=%d, ElseSteps=%d",
			len(ifStep.Body), len(ifStep.Else))
	} else {
		l.addError(ctx, "Cannot append IF step: No valid parent step list (currentSteps is nil and no currentProc).")
	}
}

func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.loopDepth++
	l.logDebugAST(">>> Enter WHILE Statement (Loop Depth: %d)", l.loopDepth)
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	defer func() {
		l.loopDepth--
		l.logDebugAST("<<< Exit WHILE Statement Final (Loop Depth: %d)", l.loopDepth)
	}()
	l.logDebugAST("--- ExitWhile_statement: Finalizing WHILE step (%s)", ctx.GetText())

	bodyStepsRaw, okBody := l.popValue()
	if !okBody {
		l.addError(ctx, "Stack error popping body steps for WHILE statement")
		return
	}
	bodySteps, isBodySteps := bodyStepsRaw.([]Step)
	if !isBodySteps {
		l.addError(ctx, "WHILE body steps are not []Step (got %T)", bodyStepsRaw)
		l.pushValue(bodyStepsRaw)
		return
	}
	l.logDebugAST("         Popped WHILE body_steps: Count=%d", len(bodySteps))

	conditionRaw, okCondition := l.popValue()
	if !okCondition {
		l.addError(ctx, "Stack error popping condition for WHILE statement")
		l.pushValue(bodySteps)
		return
	}
	conditionNode, isExpr := conditionRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "WHILE condition is not an Expression (got %T)", conditionRaw)
		l.pushValue(conditionRaw)
		l.pushValue(bodySteps)
		return
	}
	l.logDebugAST("         Popped WHILE condition: %T", conditionNode)

	if l.currentSteps == nil {
		if !(len(l.blockStepStack) == 0 && l.currentProc != nil) {
			l.addError(ctx, "Cannot append WHILE step: currentSteps (parent context) is nil unexpectedly.")
			return
		}
	}

	whileStep := Step{
		Pos:  tokenToPosition(ctx.KW_WHILE().GetSymbol()),
		Type: "while",
		Cond: conditionNode,
		Body: bodySteps,
	}

	if l.currentSteps != nil {
		*l.currentSteps = append(*l.currentSteps, whileStep)
		l.logDebugAST("         Appended WHILE Step to currentSteps list (%p)", l.currentSteps)
	} else if l.currentProc != nil {
		l.currentProc.Steps = append(l.currentProc.Steps, whileStep)
		l.logDebugAST("         Appended WHILE Step to currentProc.Steps")
	} else {
		l.addError(ctx, "Cannot append WHILE step: No valid parent step list.")
	}
}

func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	l.loopDepth++
	l.logDebugAST(">>> Enter FOR EACH Statement (Loop Depth: %d)", l.loopDepth)
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	defer func() {
		l.loopDepth--
		l.logDebugAST("<<< Exit FOR EACH Statement Final (Loop Depth: %d)", l.loopDepth)
	}()
	l.logDebugAST("--- ExitFor_each_statement: Finalizing FOR EACH step (%s)", ctx.GetText())

	bodyStepsRaw, okBody := l.popValue()
	if !okBody {
		l.addError(ctx, "Stack error popping body steps for FOR EACH statement")
		return
	}
	bodySteps, isBodySteps := bodyStepsRaw.([]Step)
	if !isBodySteps {
		l.addError(ctx, "FOR EACH body steps are not []Step (got %T)", bodyStepsRaw)
		l.pushValue(bodyStepsRaw)
		return
	}
	l.logDebugAST("         Popped FOR EACH body_steps: Count=%d", len(bodySteps))

	collectionRaw, okCollection := l.popValue()
	if !okCollection {
		l.addError(ctx, "Stack error popping collection expression for FOR EACH statement")
		l.pushValue(bodySteps)
		return
	}
	collectionNode, isExpr := collectionRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "FOR EACH collection is not an Expression (got %T)", collectionRaw)
		l.pushValue(collectionRaw)
		l.pushValue(bodySteps)
		return
	}
	l.logDebugAST("         Popped FOR EACH collection: %T", collectionNode)

	loopVar := ""
	if idNode := ctx.IDENTIFIER(); idNode != nil {
		loopVar = idNode.GetText()
	} else {
		l.addError(ctx, "Missing IDENTIFIER for loop variable in FOR EACH statement")
		l.pushValue(collectionNode)
		l.pushValue(bodySteps)
		return
	}

	if l.currentSteps == nil {
		if !(len(l.blockStepStack) == 0 && l.currentProc != nil) {
			l.addError(ctx, "Cannot append FOR EACH step: currentSteps (parent context) is nil unexpectedly.")
			return
		}
	}

	forStep := Step{
		Pos:    tokenToPosition(ctx.KW_FOR().GetSymbol()),
		Type:   "for",
		Target: loopVar,
		Cond:   collectionNode,
		Body:   bodySteps,
	}

	if l.currentSteps != nil {
		*l.currentSteps = append(*l.currentSteps, forStep)
		l.logDebugAST("         Appended FOR EACH Step to currentSteps list (%p)", l.currentSteps)
	} else if l.currentProc != nil {
		l.currentProc.Steps = append(l.currentProc.Steps, forStep)
		l.logDebugAST("         Appended FOR EACH Step to currentProc.Steps")
	} else {
		l.addError(ctx, "Cannot append FOR EACH step: No valid parent step list.")
	}
}

func (l *neuroScriptListenerImpl) EnterOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST(">>> Enter ON_ERROR Statement Context")
}

func (l *neuroScriptListenerImpl) ExitOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST("--- ExitOnErrorStmt: Finalizing ON_ERROR step (%s)", ctx.GetText())

	handlerStepsRaw, okHandler := l.popValue()
	if !okHandler {
		l.addError(ctx, "Stack error popping handler steps for ON_ERROR statement")
		return
	}
	handlerSteps, isHandlerSteps := handlerStepsRaw.([]Step)
	if !isHandlerSteps {
		l.addError(ctx, "ON_ERROR handler steps are not []Step (got %T)", handlerStepsRaw)
		l.pushValue(handlerStepsRaw)
		return
	}
	l.logDebugAST("         Popped ON_ERROR handler_steps: Count=%d", len(handlerSteps))

	if l.currentSteps == nil {
		if !(len(l.blockStepStack) == 0 && l.currentProc != nil) {
			l.addError(ctx, "Cannot append ON_ERROR step: currentSteps (parent context) is nil unexpectedly.")
			return
		}
	}

	onErrorStep := Step{
		Pos:  tokenToPosition(ctx.KW_ON_ERROR().GetSymbol()),
		Type: "on_error",
		Body: handlerSteps,
	}

	if l.currentSteps != nil {
		*l.currentSteps = append(*l.currentSteps, onErrorStep)
		l.logDebugAST("         Appended ON_ERROR Step to currentSteps list (%p)", l.currentSteps)
	} else if l.currentProc != nil {
		l.currentProc.Steps = append(l.currentProc.Steps, onErrorStep)
		l.logDebugAST("         Appended ON_ERROR Step to currentProc.Steps")
	} else {
		l.addError(ctx, "Cannot append ON_ERROR step: No valid parent step list.")
	}
	l.logDebugAST("<<< Exit ON_ERROR Statement Final")
}
