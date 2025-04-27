// pkg/core/ast_builder_blocks.go
package core

import (
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
		l.currentSteps = nil
		return
	}
	l.currentSteps = parentStepsPtr
	l.logDebugAST("    Restored currentSteps to parent context: %p (Stack size: %d)", l.currentSteps, len(l.blockStepStack))
}

// --- Shared Statement List Logic ---

func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	parentCtx := ctx.GetParent()
	// Check if parent requires a new block context to be implicitly created
	// This now includes IF, WHILE, FOR, TRY body, CATCH body, FINALLY body
	switch pCtx := parentCtx.(type) {
	case *gen.If_statementContext, *gen.While_statementContext, *gen.For_each_statementContext, *gen.Try_statementContext:
		l.logDebugAST(">>> Enter Statement_list within structured block context (%T)", pCtx)
		// Create new step list *only* if not already set by EnterBlock/EnterIf/EnterTry etc.
		// For TRY, CATCH, FINALLY, currentSteps *should* be nil when entering the list.
		if l.currentSteps == nil {
			newBlockSteps := make([]Step, 0)
			l.currentSteps = &newBlockSteps
			l.logDebugAST("    Initialized new currentSteps list (%p) for block", l.currentSteps)
		} else {
			// This might happen legitimately for WHILE/FOR where EnterBlock already set it.
			// Log potentially unexpected state for IF/TRY parents.
			if _, isIf := pCtx.(*gen.If_statementContext); isIf {
				l.logger.Warn("EnterStatement_list (IF context): currentSteps was not nil (%p)", l.currentSteps)
			} else if _, isTry := pCtx.(*gen.Try_statementContext); isTry {
				// This condition might be tricky depending on exact ANTLR parentage for catch/finally lists
				l.logger.Warn("EnterStatement_list (TRY context - should be try_body): currentSteps was not nil (%p)", l.currentSteps)
			}
			l.logDebugAST("    currentSteps (%p) already initialized for block", l.currentSteps)
		}
	case *gen.Procedure_definitionContext:
		l.logDebugAST(">>> Enter Statement_list (Procedure body)")
		if l.currentSteps == nil {
			l.logger.Error("EnterStatement_list: currentSteps is nil within Procedure context!")
			newSteps := make([]Step, 0)
			l.currentSteps = &newSteps
		}
	default:
		// It's possible the parent is implicitly the catch or finally structure within Try_statementContext
		l.logDebugAST(">>> Enter Statement_list (Parent: %T)", parentCtx)
		if l.currentSteps == nil {
			newBlockSteps := make([]Step, 0)
			l.currentSteps = &newBlockSteps
			l.logDebugAST("    Initialized new currentSteps list (%p) likely for CATCH/FINALLY", l.currentSteps)
		} else {
			l.logger.Warn("EnterStatement_list: currentSteps (%p) was not nil for unexpected parent %T", l.currentSteps, parentCtx)
		}
	}
}

// ExitStatement_list: Store steps temporarily for multi-part blocks (IF/TRY).
func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	parentCtx := ctx.GetParent()
	if l.currentSteps == nil {
		l.logger.Warn("ExitStatement_list (%T context): currentSteps is nil, storing empty steps", parentCtx)
		l.blockSteps[ctx] = []Step{} // Store empty slice
	} else {
		// Store steps for blocks that need assembly in their parent Exit* rule
		switch parentCtx.(type) {
		// IF THEN body (if_body) OR IF ELSE body (else_body)
		case *gen.If_statementContext:
			l.blockSteps[ctx] = *l.currentSteps
			l.logDebugAST("<<< Exit Statement_list (IF context). Stored %d steps for context %p", len(*l.currentSteps), ctx)
			l.currentSteps = nil // Reset for next part of IF or after IF

		// TRY body (try_body) OR CATCH body (catch_body) OR FINALLY body (finally_body)
		case *gen.Try_statementContext:
			l.blockSteps[ctx] = *l.currentSteps
			l.logDebugAST("<<< Exit Statement_list (TRY context - could be try/catch/finally). Stored %d steps for context %p", len(*l.currentSteps), ctx)
			l.currentSteps = nil // Reset for next part of TRY structure or after TRY

		default: // Includes WHILE, FOR, Procedure body
			l.logDebugAST("<<< Exit Statement_list (%T context). Current steps count: %d", parentCtx, len(*l.currentSteps))
			// Do not store or clear for simple blocks or procedure body.
		}
	}
}

// --- IF Statement Handling (v0.2.0: uses ENDIF) ---
func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST(">>> Enter If_statement context")
	if l.currentSteps != nil {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps)
		l.logDebugAST("    Pushed parent context %p. Stack size: %d", l.currentSteps, len(l.blockStepStack))
	} else {
		l.logger.Warn("EnterIf_statement called with nil currentSteps")
		l.blockStepStack = append(l.blockStepStack, nil)
	}
	l.currentSteps = nil
}

func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("<<< Exit If_statement finalization")

	conditionNode, ok := l.popValue()
	if !ok {
		l.logger.Error("AST Builder: Failed to pop condition expression for IF")
		l.restoreParentContext("IF")
		return
	}
	l.logDebugAST("    Popped IF condition node: %T", conditionNode)

	var thenSteps []Step
	thenCtx := ctx.GetIf_body() // Use generated accessor
	if thenCtx != nil {
		if steps, found := l.blockSteps[thenCtx]; found {
			thenSteps = steps
			delete(l.blockSteps, thenCtx)
			l.logDebugAST("    Retrieved %d THEN steps from map using key %p", len(thenSteps), thenCtx)
		} else {
			l.logger.Warn("ExitIf_statement: Did not find stored steps for THEN block context %p", thenCtx)
			thenSteps = []Step{}
		}
	} else {
		l.logger.Error("ExitIf_statement: THEN Statement_list context (GetIf_body) was nil")
		thenSteps = []Step{}
	}

	var elseSteps []Step = nil
	if ctx.KW_ELSE() != nil {
		elseCtx := ctx.GetElse_body() // Use generated accessor
		if elseCtx != nil {
			if steps, found := l.blockSteps[elseCtx]; found {
				elseSteps = steps
				delete(l.blockSteps, elseCtx)
				l.logDebugAST("    Retrieved %d ELSE steps from map using key %p", len(elseSteps), elseCtx)
			} else {
				l.logger.Warn("ExitIf_statement: Did not find stored steps for ELSE block context %p", elseCtx)
			}
		} else {
			l.logger.Warn("ExitIf_statement: ELSE clause present but GetElse_body context was nil.")
		}
	}

	l.restoreParentContext("IF")
	if l.currentSteps == nil {
		l.logger.Error("ExitIf_statement: Parent step list nil after restore")
		return
	}

	ifStep := Step{Type: "if", Cond: conditionNode, Value: thenSteps, ElseValue: elseSteps}
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

// --- WHILE and FOR EACH (v0.2.0: uses ENDWHILE/ENDFOR) ---
func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST(">>> Enter While_statement")
	l.EnterBlock("WHILE", "")
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST("<<< Exit While_statement")
	conditionNode, okCond := l.popValue()
	if !okCond {
		l.logger.Error("AST Builder: Failed to pop condition expression for WHILE")
		l.ExitBlock("WHILE")
		return
	}
	l.logDebugAST("    Popped WHILE condition node: %T", conditionNode)
	steps, okSteps := l.ExitBlock("WHILE")
	if !okSteps {
		return
	}
	if l.currentSteps != nil {
		whileStep := Step{Type: "while", Cond: conditionNode, Value: steps}
		*l.currentSteps = append(*l.currentSteps, whileStep)
		l.logDebugAST("    Appended WHILE Step: Cond=%T, Steps=%d", conditionNode, len(steps))
	} else {
		l.logger.Error("WHILE exit: currentSteps nil after block exit")
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
	collectionNode, okColl := l.popValue()
	if !okColl {
		l.logger.Error("AST Builder: Failed pop collection expression FOR")
		l.ExitBlock("FOR")
		return
	}
	l.logDebugAST("    Popped FOR collection node: %T", collectionNode)
	loopVar := ""
	if id := ctx.IDENTIFIER(); id != nil {
		loopVar = id.GetText()
	}
	steps, okSteps := l.ExitBlock("FOR")
	if !okSteps {
		return
	}
	if l.currentSteps != nil {
		forStep := Step{Type: "for", Target: loopVar, Cond: collectionNode, Value: steps}
		*l.currentSteps = append(*l.currentSteps, forStep)
		l.logDebugAST("    Appended FOR Step: Var=%q, Collection:%T, Steps:%d", loopVar, collectionNode, len(steps))
	} else {
		l.logger.Error("FOR exit: currentSteps nil after block exit")
	}
}

// --- Try/Catch/Finally Handling (REVISED v0.2.0) ---

func (l *neuroScriptListenerImpl) EnterTry_statement(ctx *gen.Try_statementContext) {
	l.logDebugAST(">>> Enter Try_statement")
	if l.currentSteps != nil {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps)
		l.logDebugAST("    TRY: Pushed parent context %p. Stack size: %d", l.currentSteps, len(l.blockStepStack))
	} else {
		l.logger.Warn("EnterTry_statement called with nil currentSteps")
		l.blockStepStack = append(l.blockStepStack, nil)
	}
	l.currentSteps = nil // Reset for TRY body
	// Reset blockSteps map for this specific try/catch/finally structure
	l.blockSteps = make(map[antlr.ParserRuleContext][]Step)
}

// --- REMOVED Enter/ExitCatch_clause and Enter/ExitFinally_clause ---

func (l *neuroScriptListenerImpl) ExitTry_statement(ctx *gen.Try_statementContext) {
	l.logDebugAST("<<< Exit Try_statement")

	// Logic to retrieve steps needs to use ANTLR context accessors correctly
	allStmtLists := ctx.AllStatement_list() // Get all statement lists within the try_statement
	var trySteps, catchSteps, finallySteps []Step
	catchVar := ""

	// 1. TRY body is the first statement list
	if len(allStmtLists) > 0 {
		tryCtx := allStmtLists[0]
		if steps, ok := l.blockSteps[tryCtx]; ok {
			trySteps = steps
			delete(l.blockSteps, tryCtx)
			l.logDebugAST("    TRY: Retrieved %d try steps for context %p", len(trySteps), tryCtx)
		} else {
			l.logger.Warn("ExitTry_statement: No steps found for TRY context %p", tryCtx)
			trySteps = []Step{}
		}
	} else {
		l.logger.Error("ExitTry_statement: No statement list found for TRY body!")
		trySteps = []Step{}
	}

	// 2. CATCH clause(s) and FINALLY clause
	// Need to correlate statement lists with KW_CATCH and KW_FINALLY tokens
	catchKeywords := ctx.AllKW_CATCH()
	finallyKeyword := ctx.KW_FINALLY() // Check if FINALLY exists

	// Assuming only ONE catch block for now, matching the simplified Step struct
	if len(catchKeywords) > 0 {
		if len(catchKeywords) > 1 {
			l.logger.Warn("ExitTry_statement: Multiple catch clauses found, processing only the first.")
		}
		// The catch body statement list should be the one *after* the first KW_CATCH
		// The index depends on whether try_body was the first statement list.
		catchBodyIndex := 1 // Index in allStmtLists if try_body was [0]
		if catchBodyIndex < len(allStmtLists) {
			catchBodyCtx := allStmtLists[catchBodyIndex]
			if steps, ok := l.blockSteps[catchBodyCtx]; ok {
				catchSteps = steps
				delete(l.blockSteps, catchBodyCtx)
				l.logDebugAST("    TRY: Retrieved %d catch steps for context %p", len(catchSteps), catchBodyCtx)
			} else {
				l.logger.Warn("ExitTry_statement: No steps found for CATCH context %p", catchBodyCtx)
				catchSteps = []Step{}
			}
		} else {
			l.logger.Warn("ExitTry_statement: Could not find statement list for CATCH block.")
			catchSteps = []Step{}
		}

		// Get catch variable name if present (correlates with first CATCH)
		if catchIdent := ctx.GetCatch_param(); catchIdent != nil {
			catchVar = catchIdent.GetText()
			l.logDebugAST("    TRY: Found catch variable: %s", catchVar)
		}
	} else {
		l.logDebugAST("    TRY: No catch clause found.")
	}

	// 3. FINALLY clause
	if finallyKeyword != nil {
		// The finally body statement list is the *last* one if both catch and finally exist,
		// or the second one if only try and finally exist.
		finallyBodyIndex := 1 // If no catch
		if len(catchKeywords) > 0 {
			finallyBodyIndex = 2 // If catch exists
		}

		if finallyBodyIndex < len(allStmtLists) {
			finallyBodyCtx := allStmtLists[finallyBodyIndex]
			if steps, ok := l.blockSteps[finallyBodyCtx]; ok {
				finallySteps = steps
				delete(l.blockSteps, finallyBodyCtx)
				l.logDebugAST("    TRY: Retrieved %d finally steps for context %p", len(finallySteps), finallyBodyCtx)
			} else {
				l.logger.Warn("ExitTry_statement: No steps found for FINALLY context %p", finallyBodyCtx)
				finallySteps = []Step{}
			}
		} else {
			l.logger.Warn("ExitTry_statement: Could not find statement list for FINALLY block.")
			finallySteps = []Step{}
		}
	} else {
		l.logDebugAST("    TRY: No finally clause found.")
	}

	// 4. Restore parent context
	l.restoreParentContext("TRY")
	if l.currentSteps == nil {
		l.logger.Error("ExitTry_statement: Parent step list nil after restore")
		return
	}

	// 5. Create and append the TRY step
	tryStep := Step{
		Type:         "try",
		Value:        trySteps, // TRY body steps go into Value
		CatchVar:     catchVar,
		CatchSteps:   catchSteps,
		FinallySteps: finallySteps,
	}
	*l.currentSteps = append(*l.currentSteps, tryStep)
	l.logDebugAST("    Appended TRY Step: TrySteps=%d, CatchVar=%q, CatchSteps=%d, FinallySteps=%d",
		len(trySteps), catchVar, len(catchSteps), len(finallySteps))

	// Clean up the temporary map
	l.blockSteps = nil
}
