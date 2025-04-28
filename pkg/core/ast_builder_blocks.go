// pkg/core/ast_builder_blocks.go
package core

import (
	// Added fmt import
	"github.com/antlr4-go/antlr/v4" // Required for ParserRuleContext
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Block Handling Helpers (Unchanged) ---
func (l *neuroScriptListenerImpl) EnterBlock(blockType string, targetVar string) {
	l.logDebugAST(">>> EnterBlock Start: %s (Target: %q)", blockType, targetVar)
	if l.currentSteps == nil {
		l.logger.Warn("EnterBlock (%s) called outside a valid step context (currentSteps is nil)", blockType)
		l.blockStepStack = append(l.blockStepStack, nil)
		return
	}
	l.blockStepStack = append(l.blockStepStack, l.currentSteps)
	newBlockSteps := make([]Step, 0)
	l.currentSteps = &newBlockSteps
	l.logDebugAST("    Started new block context for %s. Parent stack size: %d, New currentSteps pointer: %p", blockType, len(l.blockStepStack), l.currentSteps)
}

func (l *neuroScriptListenerImpl) ExitBlock(blockType string) ([]Step, bool) {
	l.logDebugAST("<<< ExitBlock Start: %s (Stack size before pop: %d)", blockType, len(l.blockStepStack))
	var finishedBlockSteps []Step
	if l.currentSteps != nil {
		finishedBlockSteps = *l.currentSteps
	} else {
		l.logger.Warn("ExitBlock (%s): currentSteps was nil when exiting.", blockType)
		finishedBlockSteps = []Step{}
	}
	l.restoreParentContext(blockType)
	l.logDebugAST("    Finished block %s. Captured %d steps. Parent context restored.", blockType, len(finishedBlockSteps))
	return finishedBlockSteps, true
}

func (l *neuroScriptListenerImpl) restoreParentContext(blockType string) {
	stackSize := len(l.blockStepStack)
	if stackSize == 0 {
		l.logger.Error("restoreParentContext (%s): Stack was empty!", blockType)
		l.currentSteps = nil
		return
	}
	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex]
	if parentStepsPtr == nil {
		l.logger.Warn("restoreParentContext (%s): Parent step list pointer from stack was nil.", blockType)
		l.currentSteps = nil // Set to nil if parent was nil
	} else {
		l.currentSteps = parentStepsPtr
	}
	l.logDebugAST("    Restored currentSteps to parent context: %p (Stack size: %d)", l.currentSteps, len(l.blockStepStack))
}

// --- Shared Statement List Logic ---
func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	parentCtx := ctx.GetParent()
	switch pCtx := parentCtx.(type) {
	// ADDED: OnErrorStmtContext
	case *gen.If_statementContext, *gen.While_statementContext, *gen.For_each_statementContext, *gen.OnErrorStmtContext:
		l.logDebugAST(">>> Enter Statement_list within structured block context (%T)", pCtx)
		if l.currentSteps == nil {
			newBlockSteps := make([]Step, 0)
			l.currentSteps = &newBlockSteps
			l.logDebugAST("    Initialized new currentSteps list (%p) for block", l.currentSteps)
		} else {
			// Store current steps in blockSteps map ONLY if it's an IF or ON_ERROR context,
			// because these blocks might have multiple statement lists (THEN/ELSE or the handler body itself)
			// whose steps need to be retrieved later using the context as a key.
			// WHILE/FOR only have one body, managed by Enter/ExitBlock stack.
			if _, isIf := pCtx.(*gen.If_statementContext); isIf {
				l.logger.Warn("EnterStatement_list (IF context): currentSteps was not nil (%p). Storing steps for potential later retrieval.", l.currentSteps)
				if l.blockSteps == nil {
					l.blockSteps = make(map[antlr.ParserRuleContext][]Step)
				}
				l.blockSteps[ctx] = *l.currentSteps // Store steps associated with this specific list context
				newSteps := make([]Step, 0)         // Start fresh for the *next* statement list in the IF
				l.currentSteps = &newSteps
			} else if _, isOnError := pCtx.(*gen.OnErrorStmtContext); isOnError {
				l.logDebugAST("    EnterStatement_list (ON_ERROR context): currentSteps (%p) likely holds steps from outer scope. Will start fresh.", l.currentSteps)
				newSteps := make([]Step, 0) // Start fresh steps for the handler body
				l.currentSteps = &newSteps
			} else {
				l.logDebugAST("    currentSteps (%p) already initialized, expected for WHILE/FOR block", l.currentSteps)
			}
		}
	case *gen.Procedure_definitionContext:
		l.logDebugAST(">>> Enter Statement_list (Procedure body)")
		if l.currentSteps == nil {
			l.logger.Error("EnterStatement_list: currentSteps is nil within Procedure context!")
			newSteps := make([]Step, 0)
			l.currentSteps = &newSteps
		}
	default:
		l.logDebugAST(">>> Enter Statement_list (Parent: %T - Unexpected?)", parentCtx)
		if l.currentSteps == nil {
			l.logger.Error("EnterStatement_list: currentSteps is nil for unexpected parent %T", parentCtx)
			newSteps := make([]Step, 0)
			l.currentSteps = &newSteps
		}
	}
}

func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	parentCtx := ctx.GetParent()
	if l.currentSteps == nil {
		l.logger.Warn("ExitStatement_list (%T context): currentSteps is nil, storing empty steps", parentCtx)
		if l.blockSteps == nil {
			l.blockSteps = make(map[antlr.ParserRuleContext][]Step)
		}
		l.blockSteps[ctx] = []Step{} // Store empty steps for this context
	} else {
		// Store the completed steps for this specific statement list context if it's within IF or ON_ERROR
		switch pCtx := parentCtx.(type) {
		case *gen.If_statementContext, *gen.OnErrorStmtContext:
			if l.blockSteps == nil {
				l.blockSteps = make(map[antlr.ParserRuleContext][]Step)
			}
			l.blockSteps[ctx] = *l.currentSteps
			l.logDebugAST("<<< Exit Statement_list (%T context). Stored %d steps for context %p", pCtx, len(*l.currentSteps), ctx)
			l.currentSteps = nil // Reset currentSteps, expect parent Exit method to handle restoration or finalization
		default:
			// For PROC, WHILE, FOR, the steps remain in l.currentSteps until the block/proc exits.
			l.logDebugAST("<<< Exit Statement_list (%T context). Current steps count: %d", parentCtx, len(*l.currentSteps))
		}
	}
}

// --- IF Statement Handling (Unchanged) ---
func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST(">>> Enter If_statement context")
	if l.currentSteps != nil {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps)
		l.logDebugAST("    Pushed parent context %p. Stack size: %d", l.currentSteps, len(l.blockStepStack))
	} else {
		l.logger.Warn("EnterIf_statement called with nil currentSteps")
		l.blockStepStack = append(l.blockStepStack, nil)
	}
	l.currentSteps = nil // Reset for block processing
	// Clear blockSteps map for this specific IF statement
	l.blockSteps = make(map[antlr.ParserRuleContext][]Step)
}

func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("<<< Exit If_statement finalization")

	conditionNode, ok := l.popValue()
	if !ok {
		l.logger.Error("AST Builder: Failed to pop condition expression for IF")
		l.restoreParentContext("IF error")
		return
	}
	l.logDebugAST("    Popped IF condition node: %T", conditionNode)

	var thenSteps []Step
	// Get THEN block steps (always the first statement list)
	if len(ctx.AllStatement_list()) > 0 {
		thenCtx := ctx.Statement_list(0)
		if steps, found := l.blockSteps[thenCtx]; found {
			thenSteps = steps
			delete(l.blockSteps, thenCtx) // Clean up map
			l.logDebugAST("    Retrieved %d THEN steps from map using key %p", len(thenSteps), thenCtx)
		} else {
			l.logger.Warn("ExitIf_statement: Did not find stored steps for THEN block context %p", thenCtx)
			thenSteps = []Step{}
		}
	} else {
		l.logger.Error("ExitIf_statement: THEN Statement_list context (Index 0) was nil (Grammar requires it)")
		thenSteps = []Step{}
	}

	var elseSteps []Step = nil // nil indicates no ELSE block steps
	if ctx.KW_ELSE() != nil {
		// Get ELSE block steps (always the second statement list if present)
		if len(ctx.AllStatement_list()) > 1 {
			elseCtx := ctx.Statement_list(1)
			if steps, found := l.blockSteps[elseCtx]; found {
				elseSteps = steps
				delete(l.blockSteps, elseCtx) // Clean up map
				l.logDebugAST("    Retrieved %d ELSE steps from map using key %p", len(elseSteps), elseCtx)
			} else {
				// This case means ExitStatement_list didn't store steps for the ELSE block correctly
				l.logger.Error("ExitIf_statement: Did not find stored steps for ELSE block context %p", elseCtx)
				elseSteps = []Step{} // Use empty slice instead of nil to indicate empty block vs no block? Let's use empty slice.
			}
		} else {
			l.logger.Error("ExitIf_statement: ELSE clause present but second Statement_list context was nil.")
			elseSteps = []Step{} // Empty slice
		}
	} else {
		l.logDebugAST("    No ELSE clause found.")
	}

	l.restoreParentContext("IF")
	if l.currentSteps == nil {
		l.logger.Error("ExitIf_statement: Parent step list nil after restore")
		// Attempt to recover by creating a new list, though this might indicate a deeper issue
		newSteps := make([]Step, 0)
		l.currentSteps = &newSteps
		// return // Early return might lose the IF step entirely. Let's try appending anyway.
	}

	// Ensure Value/ElseValue are Step slices
	ifStep := Step{Type: "if", Cond: conditionNode, Value: thenSteps, ElseValue: elseSteps}
	*l.currentSteps = append(*l.currentSteps, ifStep)
	thenLen := 0
	if thenSteps != nil {
		thenLen = len(thenSteps)
	}
	elseLen := 0
	if elseSteps != nil {
		elseLen = len(elseSteps)
	} // Check elseSteps itself for nilness
	l.logDebugAST("    Appended complete IF Step to parent: Cond=%T, THEN Steps=%d, ELSE Steps=%d", ifStep.Cond, thenLen, elseLen)

	l.blockSteps = nil // Clear map after use for this IF
}

// --- WHILE and FOR EACH (Unchanged) ---
func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST(">>> Enter While_statement")
	l.EnterBlock("WHILE", "")
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST("<<< Exit While_statement")
	steps, okSteps := l.ExitBlock("WHILE")
	if !okSteps {
		l.logger.Error("AST Builder: Failed to properly exit WHILE block context")
		if len(l.blockStepStack) > 0 {
			l.restoreParentContext("WHILE recovery")
		}
		return
	}
	l.logDebugAST("    Retrieved %d WHILE steps via ExitBlock", len(steps))

	conditionNode, okCond := l.popValue()
	if !okCond {
		l.logger.Error("AST Builder: Failed to pop condition expression for WHILE")
		return
	}
	l.logDebugAST("    Popped WHILE condition node: %T", conditionNode)

	if l.currentSteps != nil {
		whileStep := Step{Type: "while", Cond: conditionNode, Value: steps}
		*l.currentSteps = append(*l.currentSteps, whileStep)
		l.logDebugAST("    Appended WHILE Step: Cond=%T, Steps=%d", conditionNode, len(steps))
	} else {
		l.logger.Error("WHILE exit: currentSteps nil after block exit and restore")
	}
}

func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	loopVar := ""
	if id := ctx.IDENTIFIER(); id != nil {
		loopVar = id.GetText()
	}
	l.logDebugAST(">>> Enter For_each_statement: var=%s", loopVar)
	l.EnterBlock("FOR", loopVar)
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST("<<< Exit For_each_statement")
	steps, okSteps := l.ExitBlock("FOR")
	if !okSteps {
		l.logger.Error("AST Builder: Failed to properly exit FOR block context")
		if len(l.blockStepStack) > 0 {
			l.restoreParentContext("FOR recovery")
		}
		return
	}
	l.logDebugAST("    Retrieved %d FOR steps via ExitBlock", len(steps))

	collectionNode, okColl := l.popValue()
	if !okColl {
		l.logger.Error("AST Builder: Failed pop collection expression FOR")
		return
	}
	l.logDebugAST("    Popped FOR collection node: %T", collectionNode)

	loopVar := ""
	if id := ctx.IDENTIFIER(); id != nil {
		loopVar = id.GetText()
	}

	if l.currentSteps != nil {
		forStep := Step{Type: "for", Target: loopVar, Cond: collectionNode, Value: steps}
		*l.currentSteps = append(*l.currentSteps, forStep)
		l.logDebugAST("    Appended FOR Step: Var=%q, Collection:%T, Steps:%d", loopVar, collectionNode, len(steps))
	} else {
		l.logger.Error("FOR exit: currentSteps nil after block exit and restore")
	}
}

// --- REMOVED: Try/Catch/Finally Handling ---
// func (l *neuroScriptListenerImpl) EnterTry_statement(ctx *gen.Try_statementContext) { ... }
// func (l *neuroScriptListenerImpl) ExitTry_statement(ctx *gen.Try_statementContext) { ... }

// --- ADDED: OnError Handling ---
func (l *neuroScriptListenerImpl) EnterOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST(">>> Enter OnErrorStmt")
	// Push the parent step context onto the stack before starting the handler block
	if l.currentSteps != nil {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps)
		l.logDebugAST("    ON_ERROR: Pushed parent context %p. Stack size: %d", l.currentSteps, len(l.blockStepStack))
	} else {
		l.logger.Warn("EnterOnErrorStmt called with nil currentSteps")
		l.blockStepStack = append(l.blockStepStack, nil)
	}
	// Reset currentSteps; the steps for the handler body will be collected via Enter/ExitStatement_list
	l.currentSteps = nil
	// Clear blockSteps map specifically for this on_error statement
	l.blockSteps = make(map[antlr.ParserRuleContext][]Step)
}

func (l *neuroScriptListenerImpl) ExitOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST("<<< Exit OnErrorStmt")

	var handlerSteps []Step
	// Retrieve the handler body steps stored by ExitStatement_list
	if stmtListCtx := ctx.Statement_list(); stmtListCtx != nil {
		if steps, ok := l.blockSteps[stmtListCtx]; ok {
			handlerSteps = steps
			delete(l.blockSteps, stmtListCtx) // Clean up map
			l.logDebugAST("    ON_ERROR: Retrieved %d handler steps for context %p", len(handlerSteps), stmtListCtx)
		} else {
			l.logger.Warn("ExitOnErrorStmt: No steps found for handler context %p", stmtListCtx)
			handlerSteps = []Step{}
		}
	} else {
		l.logger.Error("ExitOnErrorStmt: No statement list found for handler body!")
		handlerSteps = []Step{}
	}

	// Restore the parent context
	l.restoreParentContext("ON_ERROR")
	if l.currentSteps == nil {
		l.logger.Error("ExitOnErrorStmt: Parent step list nil after restore")
		// Attempt recovery - maybe this is the top level?
		newSteps := make([]Step, 0)
		l.currentSteps = &newSteps
		// return // Avoid losing the step
	}

	// Create and append the ON_ERROR step
	// We store the handler steps in the 'Value' field for simplicity, like loops/if blocks.
	onErrorStep := Step{
		Type:  "on_error",
		Value: handlerSteps, // Handler body steps stored here
	}
	*l.currentSteps = append(*l.currentSteps, onErrorStep)
	l.logDebugAST("    Appended ON_ERROR Step: HandlerSteps=%d", len(handlerSteps))

	l.blockSteps = nil // Clear map after use
}
