// pkg/core/ast_builder_blocks.go
package core

import (
	"github.com/antlr4-go/antlr/v4" // Import antlr
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

// processCondition extracts the condition node
func (l *neuroScriptListenerImpl) processCondition(ctx gen.IConditionContext) (interface{}, bool) {
	numConditionChildren := ctx.GetChildCount()
	opText := ""
	if numConditionChildren == 1 {
		condNode, ok := l.popValue()
		if !ok {
			return nil, false
		}
		l.logDebugAST("    Popped 1 node for simple condition")
		return condNode, true
	} else if numConditionChildren == 3 {
		nodeRHS, okRHS := l.popValue()
		if !okRHS {
			return nil, false
		}
		nodeLHS, okLHS := l.popValue()
		if !okLHS {
			return nil, false
		}
		opNode := ctx.GetChild(1)
		if opTerminal, ok := opNode.(antlr.TerminalNode); ok {
			opText = opTerminal.GetText()
		} else {
			l.logger.Printf("[ERROR] AST Builder: Could not get operator token text")
			return nil, false
		}
		l.logDebugAST("    Popped 2 nodes for comparison condition (LHS=%T, RHS=%T, Op=%q)", nodeLHS, nodeRHS, opText)
		return ComparisonNode{Left: nodeLHS, Operator: opText, Right: nodeRHS}, true
	} else {
		l.logger.Printf("[ERROR] AST Builder: Unexpected children (%d) in ConditionContext", numConditionChildren)
		return nil, false
	}
}

// --- IF Statement Handling (Revised using blockSteps map and indexed access) ---

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
	if _, ok := parentCtx.(*gen.If_statementContext); ok {
		l.logDebugAST(">>> Enter Statement_list within IF context")
		newBlockSteps := make([]Step, 0)
		l.currentSteps = &newBlockSteps
	} else {
		l.logDebugAST(">>> Enter Statement_list (Non-IF context)")
	}
}

// ExitStatement_list: If parent is IF, store collected steps in the map keyed by context.
func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	parentCtx := ctx.GetParent()
	if _, ok := parentCtx.(*gen.If_statementContext); ok {
		if l.currentSteps != nil {
			l.blockSteps[ctx] = *l.currentSteps
			l.logDebugAST("<<< Exit Statement_list (IF context). Stored %d steps for context %p", len(*l.currentSteps), ctx)
			l.currentSteps = nil
		} else {
			l.logger.Println("[WARN] ExitStatement_list (IF context): currentSteps was nil")
			l.blockSteps[ctx] = []Step{}
		}
	} else {
		l.logDebugAST("<<< Exit Statement_list (Non-IF context)")
	}
}

// ExitIf_statement: Retrieves steps from map using indexed access, creates Step, restores context.
func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("<<< Exit If_statement finalization")

	conditionNode, ok := l.processCondition(ctx.Condition())
	if !ok {
		l.logger.Println("[ERROR] AST Builder: Failed condition for IF")
	}

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
	if ctx.KW_ELSE() != nil {
		elseCtx := ctx.Statement_list(1) // Use standard ANTLR accessor
		if elseCtx != nil {
			steps, found := l.blockSteps[elseCtx]
			if found {
				elseSteps = steps
				l.logDebugAST("    Retrieved %d ELSE steps from map using key %p", len(elseSteps), elseCtx)
				delete(l.blockSteps, elseCtx)
			} else {
				l.logger.Println("[WARN] ExitIf_statement: Did not find stored steps for ELSE block context")
				elseSteps = nil
			} // Keep nil if map lookup fails
		} else {
			l.logger.Println("[WARN] ExitIf_statement: ELSE Statement_list(1) context was nil")
			elseSteps = nil
		}
	} else {
		l.logDebugAST("    No ELSE clause detected.")
	}

	// Restore parent context
	l.restoreParentContext("IF")
	if l.currentSteps == nil {
		l.logger.Println("[ERROR] ExitIf_statement: Parent step list nil after restore")
		return
	}

	// Create the IF step and append
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
	l.logDebugAST("    Appended complete IF Step: Cond=%T, THEN Steps=%d, ELSE Steps=%d", ifStep.Cond, thenLen, elseLen)
}

// --- WHILE and FOR EACH (No changes from previous correct version) ---
func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.EnterBlock("WHILE", "")
}
func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST("<<< Exit While_statement")
	conditionNode, ok := l.processCondition(ctx.Condition())
	if !ok {
		l.logger.Println("[ERROR] AST Builder: Failed condition for WHILE")
		l.ExitBlock("WHILE")
		return
	}
	steps, ok := l.ExitBlock("WHILE")
	if !ok {
		return
	}
	if l.currentSteps != nil {
		whileStep := newStep("WHILE", "", conditionNode, steps, nil, nil)
		*l.currentSteps = append(*l.currentSteps, whileStep)
		l.logDebugAST("    Appended WHILE Step: Cond=%T, Steps=%d", conditionNode, len(steps))
	} else {
		l.logger.Println("[ERROR] Current step list nil after exit WHILE")
	}
}
func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	loopVar := ""
	if ctx.IDENTIFIER() != nil {
		loopVar = ctx.IDENTIFIER().GetText()
	}
	l.EnterBlock("FOR", loopVar)
}
func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST("<<< Exit For_each_statement")
	collectionNode, ok := l.popValue()
	if !ok {
		l.logger.Println("[ERROR] AST Builder: Failed pop collection FOR")
		l.ExitBlock("FOR")
		return
	}
	loopVar := ""
	if ctx.IDENTIFIER() != nil {
		loopVar = ctx.IDENTIFIER().GetText()
	}
	steps, ok := l.ExitBlock("FOR")
	if !ok {
		return
	}
	if l.currentSteps != nil {
		forStep := newStep("FOR", loopVar, collectionNode, steps, nil, nil)
		*l.currentSteps = append(*l.currentSteps, forStep)
		l.logDebugAST("[DEBUG-AST] Builder: Appended FOR Step - Cond:%T Target:%q Steps:%d", collectionNode, loopVar, len(steps))
	} else {
		l.logger.Println("[ERROR] Current step list nil after exit FOR")
	}
}
