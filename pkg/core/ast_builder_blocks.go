// pkg/core/ast_builder_blocks.go
package core

import (
	"github.com/antlr4-go/antlr/v4" // Required for ParserRuleContext
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Block Handling Helpers (Unchanged) ---
// ... (EnterBlock, ExitBlock, restoreParentContext remain the same) ...
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
// (This logic seems okay, relies on parent context identification)
func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	parentCtx := ctx.GetParent()
	switch pCtx := parentCtx.(type) {
	case *gen.If_statementContext, *gen.While_statementContext, *gen.For_each_statementContext, *gen.Try_statementContext:
		l.logDebugAST(">>> Enter Statement_list within structured block context (%T)", pCtx)
		if l.currentSteps == nil {
			newBlockSteps := make([]Step, 0)
			l.currentSteps = &newBlockSteps
			l.logDebugAST("    Initialized new currentSteps list (%p) for block", l.currentSteps)
		} else {
			if _, isIf := pCtx.(*gen.If_statementContext); isIf {
				l.logger.Warn("EnterStatement_list (IF context): currentSteps was not nil (%p). May be expected for THEN block, check if ELSE intended.", l.currentSteps)
			} else if _, isTry := pCtx.(*gen.Try_statementContext); isTry {
				l.logger.Warn("EnterStatement_list (TRY context): currentSteps was not nil (%p). Check if expected for TRY vs CATCH/FINALLY.", l.currentSteps)
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
		l.blockSteps[ctx] = []Step{}
	} else {
		switch parentCtx.(type) {
		case *gen.If_statementContext:
			if l.blockSteps == nil {
				l.blockSteps = make(map[antlr.ParserRuleContext][]Step)
			}
			l.blockSteps[ctx] = *l.currentSteps
			l.logDebugAST("<<< Exit Statement_list (IF context). Stored %d steps for context %p", len(*l.currentSteps), ctx)
			l.currentSteps = nil

		case *gen.Try_statementContext:
			if l.blockSteps == nil {
				l.blockSteps = make(map[antlr.ParserRuleContext][]Step)
			}
			l.blockSteps[ctx] = *l.currentSteps
			l.logDebugAST("<<< Exit Statement_list (TRY context). Stored %d steps for context %p", len(*l.currentSteps), ctx)
			l.currentSteps = nil

		default:
			l.logDebugAST("<<< Exit Statement_list (%T context). Current steps count: %d", parentCtx, len(*l.currentSteps))
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
	l.blockSteps = make(map[antlr.ParserRuleContext][]Step)
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
	// FIX: Use indexed Statement_list(0) accessor based on user's confirmed generated code structure
	if len(ctx.AllStatement_list()) > 0 {
		thenCtx := ctx.Statement_list(0) // First statement list is the THEN block
		if steps, found := l.blockSteps[thenCtx]; found {
			thenSteps = steps
			delete(l.blockSteps, thenCtx)
			l.logDebugAST("    Retrieved %d THEN steps from map using key %p", len(thenSteps), thenCtx)
		} else {
			l.logger.Warn("ExitIf_statement: Did not find stored steps for THEN block context %p", thenCtx)
			thenSteps = []Step{}
		}
	} else {
		l.logger.Error("ExitIf_statement: THEN Statement_list context (Index 0) was nil (Grammar requires it)")
		thenSteps = []Step{}
	}

	var elseSteps []Step = nil
	if ctx.KW_ELSE() != nil {
		// FIX: Use indexed Statement_list(1) accessor based on user's confirmed generated code structure
		if len(ctx.AllStatement_list()) > 1 {
			elseCtx := ctx.Statement_list(1) // Second statement list is the ELSE block
			if steps, found := l.blockSteps[elseCtx]; found {
				elseSteps = steps
				delete(l.blockSteps, elseCtx)
				l.logDebugAST("    Retrieved %d ELSE steps from map using key %p", len(elseSteps), elseCtx)
			} else {
				l.logger.Error("ExitIf_statement: Did not find stored steps for ELSE block context %p", elseCtx)
				elseSteps = []Step{}
			}
		} else {
			l.logger.Error("ExitIf_statement: ELSE clause present but second Statement_list context was nil.")
			elseSteps = []Step{}
		}
	} else {
		l.logDebugAST("    No ELSE clause found.")
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

	l.blockSteps = nil
}

// --- WHILE and FOR EACH ---
// (No changes needed here as they use ExitBlock which uses l.currentSteps)
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

// --- Try/Catch/Finally Handling ---
// FIX: Use indexed Statement_list(i) and check indices carefully
func (l *neuroScriptListenerImpl) EnterTry_statement(ctx *gen.Try_statementContext) {
	l.logDebugAST(">>> Enter Try_statement")
	if l.currentSteps != nil {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps)
		l.logDebugAST("    TRY: Pushed parent context %p. Stack size: %d", l.currentSteps, len(l.blockStepStack))
	} else {
		l.logger.Warn("EnterTry_statement called with nil currentSteps")
		l.blockStepStack = append(l.blockStepStack, nil)
	}
	l.currentSteps = nil
	l.blockSteps = make(map[antlr.ParserRuleContext][]Step)
}

func (l *neuroScriptListenerImpl) ExitTry_statement(ctx *gen.Try_statementContext) {
	l.logDebugAST("<<< Exit Try_statement")

	var trySteps, catchSteps, finallySteps []Step
	catchVar := ""
	stmtLists := ctx.AllStatement_list() // Get all statement lists

	// TRY Body is always the first statement list
	if len(stmtLists) > 0 {
		tryCtx := stmtLists[0]
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

	// CATCH Body (assuming max 1 catch block for now)
	if len(ctx.AllKW_CATCH()) > 0 {
		// Catch body is the second statement list IF it exists
		if len(stmtLists) > 1 {
			catchBodyCtx := stmtLists[1]
			if steps, ok := l.blockSteps[catchBodyCtx]; ok {
				catchSteps = steps
				delete(l.blockSteps, catchBodyCtx)
				l.logDebugAST("    TRY: Retrieved %d catch steps for context %p", len(catchSteps), catchBodyCtx)
			} else {
				l.logger.Warn("ExitTry_statement: No steps found for CATCH context %p", catchBodyCtx)
				catchSteps = []Step{}
			}
			// Get corresponding catch parameter if it exists
			if catchParamToken := ctx.GetCatch_param(); catchParamToken != nil { // Use specific accessor
				catchVar = catchParamToken.GetText()
				l.logDebugAST("    TRY: Found catch variable: %s", catchVar)
			} else {
				l.logDebugAST("    TRY: Catch clause found but no variable specified.")
			}
		} else {
			l.logger.Error("ExitTry_statement: KW_CATCH present but only one statement list found!")
			catchSteps = []Step{}
		}
		if len(ctx.AllKW_CATCH()) > 1 {
			l.logger.Warn("ExitTry_statement: Multiple catch clauses found in grammar, AST builder currently only handles the first.")
		}
	} else {
		l.logDebugAST("    TRY: No catch clause found.")
	}

	// FINALLY Body
	if ctx.KW_FINALLY() != nil {
		// Finally body is the last statement list
		finallyIndex := len(stmtLists) - 1                     // Calculate expected index
		if finallyIndex > 0 && finallyIndex < len(stmtLists) { // Index must be valid and after try block
			finallyBodyCtx := stmtLists[finallyIndex]
			if steps, ok := l.blockSteps[finallyBodyCtx]; ok {
				finallySteps = steps
				delete(l.blockSteps, finallyBodyCtx)
				l.logDebugAST("    TRY: Retrieved %d finally steps for context %p", len(finallySteps), finallyBodyCtx)
			} else {
				l.logger.Warn("ExitTry_statement: No steps found for FINALLY context %p", finallyBodyCtx)
				finallySteps = []Step{}
			}
		} else {
			l.logger.Error("ExitTry_statement: KW_FINALLY present but could not determine valid index for its statement list (Index: %d, Count: %d)", finallyIndex, len(stmtLists))
			finallySteps = []Step{}
		}
	} else {
		l.logDebugAST("    TRY: No finally clause found.")
	}

	// Restore parent context
	l.restoreParentContext("TRY")
	if l.currentSteps == nil {
		l.logger.Error("ExitTry_statement: Parent step list nil after restore")
		return
	}

	// Create and append the TRY step
	tryStep := Step{
		Type:         "try",
		Value:        trySteps, // TRY body steps
		CatchVar:     catchVar,
		CatchSteps:   catchSteps,
		FinallySteps: finallySteps,
	}
	*l.currentSteps = append(*l.currentSteps, tryStep)
	l.logDebugAST("    Appended TRY Step: TrySteps=%d, CatchVar=%q, CatchSteps=%d, FinallySteps=%d",
		len(trySteps), catchVar, len(catchSteps), len(finallySteps))

	l.blockSteps = nil
}
