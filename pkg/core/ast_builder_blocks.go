// NeuroScript Version: 0.3.0
// File version: 0.0.2 // Align Step creation with revised ast.go for blocks
// Last Modified: 2025-05-09 // Updated to reflect new Step struct fields
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
	l.logDebugAST("--- Enter exitBlockContext for %s", blockType)

	var completedSteps []Step
	if l.currentSteps != nil {
		completedSteps = *l.currentSteps
		l.logDebugAST("      Captured %d steps from context %p for %s block", len(completedSteps), l.currentSteps, blockType)
	} else {
		l.logger.Warn("%s: Exiting block, but currentSteps was nil when capturing. Returning empty slice.", blockType)
		completedSteps = []Step{}
	}

	stackSize := len(l.blockStepStack)
	if stackSize == 0 {
		l.logger.Error("%s: Cannot restore parent context, stack empty!", blockType)
		l.currentSteps = nil // Ensure currentSteps is nil if stack is unexpectedly empty
		return completedSteps
	}

	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex] // Pop stack

	l.currentSteps = parentStepsPtr // Restore parent pointer

	if l.currentSteps == nil {
		l.logDebugAST("<<< Exit exitBlockContext for %s (Restored parent, currentSteps is now nil, Stack size: %d)", blockType, len(l.blockStepStack))
	} else {
		l.logDebugAST("<<< Exit exitBlockContext for %s (Restored parent, currentSteps points to list of len %d, Stack size: %d)", blockType, len(*l.currentSteps), len(l.blockStepStack))
	}
	return completedSteps
}

// EnterStatement_list:
// This is a generic entry point. For specific blocks like IF, the enterBlockContext
// is typically called from a more specific rule like EnterStatement_list_within_if (if such exists)
// or directly in the Enter<BlockType>_statement.
// The main purpose here is to log or ensure currentSteps is valid if we are at the top level of a procedure.
func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	// Only initialize if currentSteps is nil AND we're not already in a block context being managed
	// (blockStepStack being empty implies we are at the procedure's top-level statement list or file header aftermath)
	if l.currentSteps == nil && len(l.blockStepStack) == 0 {
		// This scenario typically means we are starting the main body of a procedure
		// (after metadata_block and before the first actual statement of the procedure body).
		// The currentProc.Steps should be initialized in EnterProcedure_definition.
		// This function should mostly be about setting up for nested blocks if the grammar uses
		// statement_list recursively within them without a more specific enter rule.
		// For IF/WHILE/FOR/ON_ERROR, their Enter* methods handle calling enterBlockContext.
		l.logDebugAST(">>> Enter Statement_list (Potentially top-level of proc or unmanaged block - currentSteps: %p)", l.currentSteps)
	} else {
		l.logDebugAST(">>> Enter Statement_list (currentSteps: %p, blockStack: %d)", l.currentSteps, len(l.blockStepStack))
	}
}

// ExitStatement_list should generally do nothing as steps are appended directly.
func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	l.logDebugAST("<<< Exit Statement_list (%s)", ctx.GetText())
}

// --- IF Statement ---
func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST(">>> Enter IF Statement Context (Parent currentSteps: %p)", l.currentSteps)
	// The condition expression will be visited first and its result pushed onto l.valueStack.
	// Then, the statement_list for the 'then' block will be visited.
	// We push the current l.currentSteps (parent's step list) onto the blockStepStack
	// and set l.currentSteps to a new slice for the 'then' block's steps.
	l.enterBlockContext("IF-THEN")
}

// EnterStatement_list_within_if is not standard ANTLR. Logic moved to EnterIf_statement and ExitIf_statement.
// If `statement_list` is a direct child of `if_statement` for both `then` and `else` parts,
// ANTLR will call `EnterStatement_list` for them. We need to distinguish based on parent or sibling context.
// Simpler: Manage stack explicitly in ExitIf_statement. `enterBlockContext` in `EnterIf_statement` handles the "then" block.
// If there's an "else", `EnterElse_clause` (if it exists) or logic in `ExitIf_statement` would handle its block.

// For the 'else' part, we need to capture the 'then' steps, then prepare for 'else' steps.
// This is tricky with just Enter/Exit on if_statement.
// A common pattern is to handle this in ExitIf_statement.

func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("--- ExitIf_statement: Finalizing IF step (%s)", ctx.GetText())

	var elseSteps []Step
	// If there's an ELSE clause, the `exitBlockContext` for it should be called BEFORE the `exitBlockContext` for THEN.
	// This means the `enterBlockContext` for ELSE must have happened AFTER `enterBlockContext` for THEN.
	// The ANTLR walk order is: IF cond THEN_BLOCK (ELSE ELSE_BLOCK)? ENDIF
	// So when we exit IF, the last block on stack is THEN, or ELSE if present.

	if ctx.KW_ELSE() != nil {
		// The steps for the ELSE block were just completed and are in l.currentSteps.
		// Pop them using exitBlockContext. This also restores l.currentSteps to be the THEN block's steps.
		elseSteps = l.exitBlockContext("IF-ELSE")
		l.logDebugAST("         Captured %d ELSE steps", len(elseSteps))
	}

	// Now, the steps for the THEN block are in l.currentSteps (or were just restored if there was an ELSE).
	// Pop them using exitBlockContext. This restores l.currentSteps to be the parent of the IF statement.
	thenSteps := l.exitBlockContext("IF-THEN")
	l.logDebugAST("         Captured %d THEN steps", len(thenSteps))

	// The condition for the IF statement should be on the value stack.
	conditionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Stack error popping condition for IF statement")
		return
	}
	conditionNode, isExpr := conditionRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "IF condition is not an Expression (got %T)", conditionRaw)
		return
	}
	l.logDebugAST("         Popped IF condition: %T", conditionNode)

	if l.currentSteps == nil {
		l.addError(ctx, "Cannot append IF step: currentSteps (parent) is nil after block exits.")
		return
	}

	ifStep := Step{
		Pos:  tokenToPosition(ctx.KW_IF().GetSymbol()),
		Type: "if",
		Cond: conditionNode,
		Body: thenSteps, // Use Body for 'then' steps
		Else: elseSteps, // Use Else for 'else' steps
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, ifStep)
	l.logDebugAST("         Appended IF Step: Cond=%T, ThenSteps=%d, ElseSteps=%d",
		ifStep.Cond, len(ifStep.Body), len(ifStep.Else))
}

// --- WHILE Statement ---
func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.loopDepth++
	l.logDebugAST(">>> Enter WHILE Statement Context (Loop Depth: %d)", l.loopDepth)
	// Condition is visited first. Then statement_list (body).
	// Prepare for the body's steps.
	l.enterBlockContext("WHILE-BODY")
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	defer func() {
		l.loopDepth--
		l.logDebugAST("<<< Exit WHILE Statement Final (Loop Depth: %d)", l.loopDepth)
	}()
	l.logDebugAST("--- ExitWhile_statement: Finalizing WHILE step (%s)", ctx.GetText())

	bodySteps := l.exitBlockContext("WHILE-BODY") // Pop body steps, restore parent l.currentSteps
	l.logDebugAST("         Captured %d WHILE body steps", len(bodySteps))

	conditionRaw, ok := l.popValue() // Condition was pushed before body
	if !ok {
		l.addError(ctx, "Stack error popping condition for WHILE statement")
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
		Body: bodySteps, // Use Body for loop steps
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, whileStep)
	l.logDebugAST("         Appended WHILE Step: Cond=%T, BodySteps=%d",
		whileStep.Cond, len(whileStep.Body))
}

// --- FOR EACH Statement ---
func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	l.loopDepth++
	l.logDebugAST(">>> Enter FOR EACH Statement Context (Loop Depth: %d)", l.loopDepth)
	// IDENTIFIER and collection Expression are visited first. Then statement_list (body).
	// Prepare for the body's steps.
	l.enterBlockContext("FOR-EACH-BODY")
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	defer func() {
		l.loopDepth--
		l.logDebugAST("<<< Exit FOR EACH Statement Final (Loop Depth: %d)", l.loopDepth)
	}()
	l.logDebugAST("--- ExitFor_each_statement: Finalizing FOR EACH step (%s)", ctx.GetText())

	bodySteps := l.exitBlockContext("FOR-EACH-BODY") // Pop body steps
	l.logDebugAST("         Captured %d FOR EACH body steps", len(bodySteps))

	// Collection expression was pushed before the body
	collectionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Stack error popping collection expression for FOR EACH statement")
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
		// Potentially return or create an error step
	}

	if l.currentSteps == nil {
		l.addError(ctx, "Cannot append FOR EACH step: currentSteps (parent) is nil after block exit.")
		return
	}

	forStep := Step{
		Pos:    tokenToPosition(ctx.KW_FOR().GetSymbol()),
		Type:   "for",          // Or "for_each" if you want to distinguish more in Step.Type
		Target: loopVar,        // Loop iteration variable
		Cond:   collectionNode, // Collection being iterated (using Cond field as per prior use for IF/WHILE condition)
		Body:   bodySteps,      // Loop body
		// Metadata: make(map[string]string),
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

	handlerSteps := l.exitBlockContext("ON-ERROR-HANDLER") // Pop handler steps
	l.logDebugAST("         Captured %d ON_ERROR handler steps", len(handlerSteps))

	// No condition or value to pop from stack for on_error itself.
	// It's a block triggered by runtime errors.

	if l.currentProc == nil { // on_error is procedure-scoped
		l.addError(ctx, "ON_ERROR statement defined outside of a procedure context.")
		return
	}
	// The on_error block applies to the current procedure.
	// We might want to attach this to l.currentProc directly,
	// or handle it as a special kind of step if it can appear mid-procedure (current grammar implies it's a block statement).
	// For now, let's assume it's a step that the interpreter handles by associating with current procedure execution frame.

	// If it needs to be attached to the Procedure AST node:
	// l.currentProc.OnErrorHandler = handlerSteps // (Requires OnErrorHandler field in Procedure AST)
	// OR, if it's a step in the main flow that defines the handler:

	if l.currentSteps == nil {
		l.addError(ctx, "Cannot append ON_ERROR step: currentSteps (parent) is nil after block exit.")
		// This might happen if on_error is the only thing in a procedure, which is unusual but possible.
		// If currentSteps should be procedure's main step list:
		if l.currentProc != nil {
			l.currentSteps = &l.currentProc.Steps
		} else {
			return // Can't proceed
		}
	}

	onErrorStep := Step{
		Pos:  tokenToPosition(ctx.KW_ON_ERROR().GetSymbol()),
		Type: "on_error",
		Body: handlerSteps, // Use Body for handler steps
		// Metadata: make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, onErrorStep) // Appends it to the current scope's steps
	l.logDebugAST("         Appended ON_ERROR Step: HandlerSteps=%d", len(onErrorStep.Body))
	l.logDebugAST("<<< Exit ON_ERROR Statement Final")
}
