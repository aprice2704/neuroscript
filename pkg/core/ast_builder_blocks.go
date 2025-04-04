// pkg/core/ast_builder_blocks.go
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Block Handling ---

// EnterBlock starts a new step list context for structured blocks (WHILE, FOR).
func (l *neuroScriptListenerImpl) EnterBlock(blockType string, targetVar string) {
	l.logDebugAST(">>> EnterBlock Start: %s (Target: %q)", blockType, targetVar)
	if l.currentSteps == nil {
		l.logger.Printf("[WARN] EnterBlock (%s) called outside a valid step context", blockType)
		return
	}
	l.blockStepStack = append(l.blockStepStack, l.currentSteps) // Push current (parent) step list
	newBlockSteps := make([]Step, 0)
	l.currentSteps = &newBlockSteps // Point to new list for block's steps
	l.logDebugAST("    Started new block context for %s. Parent stack size: %d, New currentSteps pointer: %p", blockType, len(l.blockStepStack), l.currentSteps)
}

// ExitBlock finalizes the *current* block's steps, restores parent context, and returns the steps.
func (l *neuroScriptListenerImpl) ExitBlock(blockType string) ([]Step, bool) {
	l.logDebugAST("<<< ExitBlock Start: %s (Stack size before pop: %d)", blockType, len(l.blockStepStack))
	if l.currentSteps == nil {
		l.logger.Printf("[ERROR] ExitBlock (%s): currentSteps is nil!", blockType)
		if len(l.blockStepStack) > 0 {
			l.restoreParentContext(blockType)
		}
		return nil, false
	}
	finishedBlockSteps := *l.currentSteps // Capture the steps collected
	l.restoreParentContext(blockType)     // Restore parent context (pops stack, sets l.currentSteps)
	l.logDebugAST("    Finished block %s. Captured %d steps. Parent context restored.", blockType, len(finishedBlockSteps))
	return finishedBlockSteps, true
}

// restoreParentContext pops the stack and sets l.currentSteps back to the parent.
func (l *neuroScriptListenerImpl) restoreParentContext(blockType string) {
	stackSize := len(l.blockStepStack)
	if stackSize == 0 {
		l.logger.Printf("[ERROR] restoreParentContext (%s): Stack was empty!", blockType)
		l.currentSteps = nil
		return
	}
	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex] // Pop stack
	if parentStepsPtr == nil {
		l.logger.Printf("[ERROR] restoreParentContext (%s): Parent step list pointer was nil", blockType)
		l.currentSteps = nil
		return
	}
	l.currentSteps = parentStepsPtr // Restore pointer
	l.logDebugAST("    Restored currentSteps to parent context: %p (Stack size: %d)", l.currentSteps, len(l.blockStepStack))
}

// --- REMOVED processCondition helper function ---
// func (l *neuroScriptListenerImpl) processCondition(...) { ... }

// --- IF Statement Handling ---

// EnterIf_statement: Pushes parent context.
func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST(">>> Enter If_statement context")
	if l.currentSteps != nil {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps)
		l.logDebugAST("    Pushed parent context %p. Stack size: %d", l.currentSteps, len(l.blockStepStack))
	} else {
		l.logger.Println("[WARN] EnterIf_statement called with nil currentSteps")
		l.blockStepStack = append(l.blockStepStack, nil)
	}
	l.currentSteps = nil // Ready for EnterStatement_list
}

// EnterStatement_list: If parent is IF, create new step list.
func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	parentCtx := ctx.GetParent()
	// Check if parent is If_statementContext OR While_statementContext OR For_each_statementContext
	switch parentCtx.(type) {
	case *gen.If_statementContext, *gen.While_statementContext, *gen.For_each_statementContext:
		l.logDebugAST(">>> Enter Statement_list within structured block context")
		// Only create new step list if currentSteps is nil (set by EnterIf_statement)
		// For WHILE/FOR, currentSteps is set by EnterBlock
		if l.currentSteps == nil {
			newBlockSteps := make([]Step, 0)
			l.currentSteps = &newBlockSteps
		}
	default:
		l.logDebugAST(">>> Enter Statement_list (Non-block context or already handled)")
	}
}

// ExitStatement_list: If parent is IF, store collected steps in the map keyed by context.
func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	parentCtx := ctx.GetParent()
	if _, ok := parentCtx.(*gen.If_statementContext); ok {
		if l.currentSteps != nil {
			l.blockSteps[ctx] = *l.currentSteps
			l.logDebugAST("<<< Exit Statement_list (IF context). Stored %d steps for context %p", len(*l.currentSteps), ctx)
			l.currentSteps = nil // Clear current steps after storing for IF body/else body
		} else {
			l.logger.Println("[WARN] ExitStatement_list (IF context): currentSteps was nil")
			l.blockSteps[ctx] = []Step{} // Store empty slice if nil
		}
	} else if _, ok := parentCtx.(*gen.While_statementContext); ok {
		l.logDebugAST("<<< Exit Statement_list (WHILE context). Current steps count: %d", len(*l.currentSteps))
		// No special handling needed for WHILE/FOR here, steps collected directly by Enter/ExitBlock
	} else if _, ok := parentCtx.(*gen.For_each_statementContext); ok {
		l.logDebugAST("<<< Exit Statement_list (FOR context). Current steps count: %d", len(*l.currentSteps))
		// No special handling needed for WHILE/FOR here, steps collected directly by Enter/ExitBlock
	} else {
		l.logDebugAST("<<< Exit Statement_list (Non-IF/WHILE/FOR context)")
	}
}

// ExitIf_statement: Pops condition node, retrieves steps from map, creates Step, restores context.
func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("<<< Exit If_statement finalization")

	// Pop the condition expression node from the stack.
	// It was pushed by the relevant Exit*expr method (e.g., ExitLogical_or_expr).
	conditionNode, ok := l.popValue()
	if !ok {
		l.logger.Println("[ERROR] AST Builder: Failed to pop condition expression for IF")
		// Attempt to restore context anyway
		l.restoreParentContext("IF")
		return
	}
	l.logDebugAST("    Popped IF condition node: %T", conditionNode)

	// Retrieve THEN steps using ctx.Statement_list(0)
	var thenSteps []Step
	thenCtx := ctx.Statement_list(0) // Use standard ANTLR accessor
	if thenCtx != nil {
		steps, found := l.blockSteps[thenCtx]
		if found {
			thenSteps = steps
			l.logDebugAST("    Retrieved %d THEN steps from map using key %p", len(thenSteps), thenCtx)
			delete(l.blockSteps, thenCtx)
		} else {
			l.logger.Println("[WARN] ExitIf_statement: Did not find stored steps for THEN block context")
			thenSteps = []Step{}
		}
	} else {
		l.logger.Println("[WARN] ExitIf_statement: THEN Statement_list(0) context was nil")
		thenSteps = []Step{}
	}

	// Retrieve ELSE steps using ctx.Statement_list(1) if ELSE exists
	var elseSteps []Step = nil // Default to nil
	if ctx.KW_ELSE() != nil {  // Use generated method to check if ELSE keyword exists
		elseCtx := ctx.Statement_list(1) // Use standard ANTLR accessor
		if elseCtx != nil {
			steps, found := l.blockSteps[elseCtx]
			if found {
				elseSteps = steps
				l.logDebugAST("    Retrieved %d ELSE steps from map using key %p", len(elseSteps), elseCtx)
				delete(l.blockSteps, elseCtx)
			} else {
				l.logger.Println("[WARN] ExitIf_statement: Did not find stored steps for ELSE block context")
				// Keep nil if map lookup fails (might happen if ELSE block is empty and steps weren't stored)
			}
		} else {
			l.logger.Println("[WARN] ExitIf_statement: ELSE Statement_list(1) context was nil")
			// Keep nil if context doesn't exist
		}
	} else {
		l.logDebugAST("    No ELSE clause detected.")
	}

	// Restore parent context BEFORE appending the IF step
	l.restoreParentContext("IF")
	if l.currentSteps == nil {
		l.logger.Println("[ERROR] ExitIf_statement: Parent step list nil after restore")
		return // Cannot append step if parent list is gone
	}

	// Create the IF step and append to the PARENT step list
	ifStep := newStep("IF", "", conditionNode, thenSteps, elseSteps, nil)
	*l.currentSteps = append(*l.currentSteps, ifStep)

	thenLen := 0
	if thenSteps != nil {
		thenLen = len(thenSteps)
	}
	elseLen := 0
	if elseSteps != nil {
		elseLen = len(elseSteps)
	}
	l.logDebugAST("    Appended complete IF Step to parent: Cond=%T, THEN Steps=%d, ELSE Steps=%d", ifStep.Cond, thenLen, elseLen)
}

// --- WHILE and FOR EACH ---
func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.EnterBlock("WHILE", "") // EnterBlock sets l.currentSteps for the block body
}
func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST("<<< Exit While_statement")

	// Pop the condition expression node first
	conditionNode, okCond := l.popValue()
	if !okCond {
		l.logger.Println("[ERROR] AST Builder: Failed to pop condition expression for WHILE")
		l.ExitBlock("WHILE") // Still attempt cleanup
		return
	}
	l.logDebugAST("    Popped WHILE condition node: %T", conditionNode)

	// Exit the block context (captures steps collected in l.currentSteps)
	steps, okSteps := l.ExitBlock("WHILE")
	if !okSteps {
		return // Error during block exit
	}

	// Append the WHILE step to the parent context
	if l.currentSteps != nil {
		whileStep := newStep("WHILE", "", conditionNode, steps, nil, nil)
		*l.currentSteps = append(*l.currentSteps, whileStep)
		l.logDebugAST("    Appended WHILE Step to parent: Cond=%T, Steps=%d", conditionNode, len(steps))
	} else {
		l.logger.Println("[ERROR] AST Builder: Current step list nil after exiting WHILE block")
	}
}
func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	loopVar := ""
	if ctx.IDENTIFIER() != nil { // Use generated method
		loopVar = ctx.IDENTIFIER().GetText()
	}
	l.EnterBlock("FOR", loopVar) // EnterBlock sets l.currentSteps for the block body
}
func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST("<<< Exit For_each_statement")

	// Pop the collection expression node first
	collectionNode, okColl := l.popValue()
	if !okColl {
		l.logger.Println("[ERROR] AST Builder: Failed pop collection expression FOR")
		l.ExitBlock("FOR") // Attempt cleanup
		return
	}
	l.logDebugAST("    Popped FOR collection node: %T", collectionNode)

	// Determine loop variable name
	loopVar := ""
	if ctx.IDENTIFIER() != nil { // Use generated method
		loopVar = ctx.IDENTIFIER().GetText()
	}

	// Exit block context (captures steps)
	steps, okSteps := l.ExitBlock("FOR")
	if !okSteps {
		return
	}

	// Append FOR step to parent context
	if l.currentSteps != nil {
		// Note: collectionNode is passed as the 'Cond' field for FOR steps
		forStep := newStep("FOR", loopVar, collectionNode, steps, nil, nil)
		*l.currentSteps = append(*l.currentSteps, forStep)
		l.logDebugAST("    Appended FOR Step to parent: Collection:%T Target:%q Steps:%d", collectionNode, loopVar, len(steps))
	} else {
		l.logger.Println("[ERROR] AST Builder: Current step list nil after exiting FOR block")
	}
}
