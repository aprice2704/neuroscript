// NeuroScript Version: 0.3.1
// File version: 0.0.7 // AST Block context and procedure body step handling refinement
// Last Modified: 2025-06-02
// Purpose: Refines AST block context management for procedures, if/else, loops, and on_error handlers to ensure correct valueStack operations.
// filename: pkg/core/ast_builder_blocks.go
// nlines: 290
// risk_rating: HIGH
package core

import (
	"fmt"

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Block Handling: Enter/Exit for Control Flow Statements ---

func (l *neuroScriptListenerImpl) enterBlockContext(blockType string) {
	l.logDebugAST(">>> Enter %s Block Context (currentSteps before: %p, stack size: %d)", blockType, l.currentSteps, len(l.blockStepStack))

	isTopLevelContextForProcedureBody := (l.currentSteps == nil && len(l.blockStepStack) == 0 && l.currentProc != nil)

	if l.currentSteps != nil {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps)
		l.logDebugAST("     Pushed old currentSteps (%p) to stack. New stack size: %d", l.currentSteps, len(l.blockStepStack))
	} else if !isTopLevelContextForProcedureBody {
		l.logger.Warn(fmt.Sprintf("Entering %s block, but currentSteps is nil unexpectedly (not top-level procedure body). This might indicate an earlier error or listener state issue.", blockType))
	} else {
		l.logDebugAST("     Entering %s block at top-level of procedure body (currentSteps is nil, stack empty). This is expected.", blockType)
	}

	newSteps := make([]Step, 0)
	l.currentSteps = &newSteps
	l.logDebugAST("     New currentSteps initialized for %s: %p", blockType, l.currentSteps)
}

func (l *neuroScriptListenerImpl) exitBlockContext(blockType string) []Step {
	l.logDebugAST("<<< Exit %s Block Context (currentSteps for block: %p, items: %d, stack size before pop: %d)", blockType, l.currentSteps, len(*l.currentSteps), len(l.blockStepStack))

	if l.currentSteps == nil {
		l.logger.Error(fmt.Sprintf("Exiting %s block, but currentSteps is unexpectedly nil. Returning empty steps.", blockType))
		emptySteps := make([]Step, 0)
		l.pushValue(emptySteps)
		return emptySteps
	}

	completedBlockSteps := *l.currentSteps
	l.pushValue(completedBlockSteps)
	l.logDebugAST("     Pushed completed %s block steps (%d items) to value stack. Value stack size: %d", blockType, len(completedBlockSteps), len(l.valueStack))

	if len(l.blockStepStack) > 0 {
		l.currentSteps = l.blockStepStack[len(l.blockStepStack)-1]
		l.blockStepStack = l.blockStepStack[:len(l.blockStepStack)-1]
		l.logDebugAST("     Popped parent's steps (%p) from blockStepStack. Restored currentSteps. New stack size: %d", l.currentSteps, len(l.blockStepStack))
	} else {
		l.logDebugAST("     Exiting %s block, blockStepStack is empty. currentSteps becomes nil (expected if top-level procedure block or error).", blockType)
		l.currentSteps = nil
	}
	return completedBlockSteps
}

// --- Statement List ---
func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	parentRuleContext := ctx.GetParent()
	parentCtxType := fmt.Sprintf("%T", parentRuleContext)
	l.logDebugAST(">>> Enter Statement_list (Parent: %s)", parentCtxType)

	isProcedureBody := false
	if _, ok := parentRuleContext.(*gen.Procedure_definitionContext); ok {
		isProcedureBody = true
	}

	if isProcedureBody {
		if l.currentProc == nil {
			l.addError(ctx, "Internal error: Entering statement_list for a procedure, but currentProc is nil.")
			dummySteps := make([]Step, 0) // Prevent panic
			l.currentSteps = &dummySteps
			return
		}
		l.logDebugAST("     Statement_list is for Procedure: %s main body", l.currentProc.Name)
		if l.currentSteps != nil || len(l.blockStepStack) != 0 {
			l.logger.Warn(fmt.Sprintf("EnterStatement_list for procedure body '%s': currentSteps (%p) or blockStepStack (len %d) not in expected initial state.", l.currentProc.Name, l.currentSteps, len(l.blockStepStack)))
		}
		l.currentSteps = &l.currentProc.Steps
	} else if ifCtx, ok := parentRuleContext.(*gen.If_statementContext); ok {
		// Check if this Statement_listContext is the first child (then block) or second (else block)
		if len(ifCtx.AllStatement_list()) > 0 && ifCtx.Statement_list(0) == ctx {
			l.logDebugAST("     Statement_list is for IF-THEN body")
			l.enterBlockContext("IF_THEN_BODY")
		} else if len(ifCtx.AllStatement_list()) > 1 && ifCtx.Statement_list(1) == ctx {
			l.logDebugAST("     Statement_list is for IF-ELSE body")
			l.enterBlockContext("IF_ELSE_BODY")
		}
	}
	// For WHILE, FOR_EACH, ON_ERROR, their respective Enter<Block>_statement methods
	// are responsible for calling enterBlockContext.
}

func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	numItems := 0
	if l.currentSteps != nil {
		numItems = len(*l.currentSteps)
	}
	parentRuleContext := ctx.GetParent()
	parentCtxType := fmt.Sprintf("%T", parentRuleContext)
	l.logDebugAST("<<< Exit Statement_list (currentSteps has %d items, Parent: %s)", numItems, parentCtxType)

	if _, ok := parentRuleContext.(*gen.Procedure_definitionContext); ok {
		if l.currentProc != nil && l.currentSteps == &l.currentProc.Steps {
			l.logDebugAST("     Procedure body's Statement_list exiting. Pushing its steps (%d items) to value stack.", len(l.currentProc.Steps))
			stepsToPush := *l.currentSteps
			l.pushValue(stepsToPush)
		} else if l.currentProc != nil {
			l.logger.Error(fmt.Sprintf("ExitStatement_list: Exiting procedure body for '%s', but currentSteps (%p) does not point to l.currentProc.Steps (%p). Incorrect state.", l.currentProc.Name, l.currentSteps, &l.currentProc.Steps))
		} else { // currentProc is nil
			l.logger.Error("ExitStatement_list: Exiting procedure body, but currentProc is nil. Cannot push steps.")
		}
	} else if ifCtx, ok := parentRuleContext.(*gen.If_statementContext); ok {
		if len(ifCtx.AllStatement_list()) > 0 && ifCtx.Statement_list(0) == ctx {
			l.logDebugAST("     Exiting Statement_list for IF-THEN body, calling exitBlockContext.")
			l.exitBlockContext("IF_THEN_BODY") // This pushes the then-steps to valueStack
		} else if len(ifCtx.AllStatement_list()) > 1 && ifCtx.Statement_list(1) == ctx {
			l.logDebugAST("     Exiting Statement_list for IF-ELSE body, calling exitBlockContext.")
			l.exitBlockContext("IF_ELSE_BODY") // This pushes the else-steps to valueStack
		}
	}
	// For WHILE, FOR_EACH, ON_ERROR, their respective Exit<Block>_statement methods
	// will call exitBlockContext to push their steps.
}

// --- Control Flow Statements ---

func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST(">>> Enter IF Statement Context (%s)", getRuleText(ctx))
	// enterBlockContext is now handled by EnterStatement_list for THEN/ELSE bodies
}

func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("--- ExitIf_statement: Finalizing IF step (%s)", getRuleText(ctx))
	var elseSteps []Step
	if ctx.KW_ELSE() != nil {
		elseStepsRaw, okElse := l.popValue() // Pop ELSE steps (pushed by ExitStatement_list for else body)
		if !okElse {
			l.addError(ctx, "Stack error popping else_steps for IF statement")
			return
		}
		var castOk bool
		elseSteps, castOk = elseStepsRaw.([]Step)
		if !castOk {
			l.addError(ctx, "Else steps are not []Step (got %T)", elseStepsRaw)
			l.pushValue(elseStepsRaw) // Push back
			return
		}
		l.logDebugAST("         Popped IF else_steps: Count=%d", len(elseSteps))
	}

	thenStepsRaw, okThen := l.popValue() // Pop THEN steps (pushed by ExitStatement_list for then body)
	if !okThen {
		l.addError(ctx, "Stack error popping then_steps for IF statement")
		return
	}
	thenSteps, castOkThen := thenStepsRaw.([]Step)
	if !castOkThen {
		l.addError(ctx, "Then steps are not []Step (got %T)", thenStepsRaw)
		l.pushValue(thenStepsRaw) // Push back
		return
	}
	l.logDebugAST("         Popped IF then_steps: Count=%d", len(thenSteps))

	conditionRaw, okCond := l.popValue() // Pop condition (pushed by expression rule)
	if !okCond {
		l.addError(ctx, "Stack error popping condition for IF statement")
		return
	}
	conditionNode, isExpr := conditionRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "Condition for IF is not an Expression (got %T)", conditionRaw)
		l.pushValue(conditionRaw) // Push back
		return
	}
	l.logDebugAST("         Popped IF condition: %T", conditionNode)

	if l.currentSteps == nil {
		l.addError(ctx, "Cannot append IF step: currentSteps (parent context) is nil unexpectedly.")
		return
	}

	ifStep := Step{
		Pos:  tokenToPosition(ctx.KW_IF().GetSymbol()),
		Type: "if",
		Cond: conditionNode,
		Body: thenSteps,
		Else: elseSteps,
	}
	*l.currentSteps = append(*l.currentSteps, ifStep)
	l.logDebugAST("         Appended IF Step to currentSteps list (%p)", l.currentSteps)
}

func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST(">>> Enter WHILE Statement Context")
	l.enterBlockContext("WHILE_BODY") // Enter context for the while loop's body
	l.loopDepth++
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST("--- ExitWhile_statement: Finalizing WHILE step (%s)", getRuleText(ctx))
	l.loopDepth--

	// Exit context for WHILE_BODY, this pushes its steps to valueStack and restores parent currentSteps.
	_ = l.exitBlockContext("WHILE_BODY")

	bodyStepsRaw, okBody := l.popValue() // Pop body steps
	if !okBody {
		l.addError(ctx, "Stack error popping body_steps for WHILE statement (after exitBlockContext)")
		return
	}
	bodySteps, castOkBody := bodyStepsRaw.([]Step)
	if !castOkBody {
		l.addError(ctx, "WHILE body steps are not []Step (got %T)", bodyStepsRaw)
		l.pushValue(bodyStepsRaw) // Push back if wrong type
		return
	}
	l.logDebugAST("         Popped WHILE body_steps: Count=%d", len(bodySteps))

	conditionRaw, okCond := l.popValue() // Pop condition
	if !okCond {
		l.addError(ctx, "Stack error popping condition for WHILE statement")
		return
	}
	conditionNode, isExpr := conditionRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "Condition for WHILE is not an Expression (got %T)", conditionRaw)
		l.pushValue(conditionRaw) // Push back if wrong type
		return
	}
	l.logDebugAST("         Popped WHILE condition: %T", conditionNode)

	if l.currentSteps == nil {
		l.addError(ctx, "Cannot append WHILE step: currentSteps (parent context) is nil.")
		return
	}
	whileStep := Step{
		Pos:  tokenToPosition(ctx.KW_WHILE().GetSymbol()),
		Type: "while",
		Cond: conditionNode,
		Body: bodySteps,
	}
	*l.currentSteps = append(*l.currentSteps, whileStep)
	l.logDebugAST("         Appended WHILE Step to currentSteps list (%p)", l.currentSteps)
}

func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST(">>> Enter FOR_EACH Statement Context")
	l.enterBlockContext("FOR_EACH_BODY") // Enter context for the for_each loop's body
	l.loopDepth++
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST("--- ExitFor_each_statement: Finalizing FOR_EACH step (%s)", getRuleText(ctx))
	l.loopDepth--

	// Exit context for FOR_EACH_BODY, this pushes its steps to valueStack.
	_ = l.exitBlockContext("FOR_EACH_BODY")

	bodyStepsRaw, okBody := l.popValue() // Pop body steps
	if !okBody {
		l.addError(ctx, "Stack error popping body_steps for FOR_EACH statement (after exitBlockContext)")
		return
	}
	bodySteps, castOkBody := bodyStepsRaw.([]Step)
	if !castOkBody {
		l.addError(ctx, "FOR_EACH body steps are not []Step (got %T)", bodyStepsRaw)
		l.pushValue(bodyStepsRaw) // Push back if wrong type
		return
	}
	l.logDebugAST("         Popped FOR_EACH body_steps: Count=%d", len(bodySteps))

	collectionRaw, okColl := l.popValue() // Pop collection expression
	if !okColl {
		l.addError(ctx, "Stack error popping collection for FOR_EACH statement")
		return
	}
	collectionNode, isExpr := collectionRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "Collection for FOR_EACH is not an Expression (got %T)", collectionRaw)
		l.pushValue(collectionRaw) // Push back if wrong type
		return
	}
	l.logDebugAST("         Popped FOR_EACH collection: %T", collectionNode)

	loopVarName := ""
	if identNode := ctx.IDENTIFIER(); identNode != nil {
		loopVarName = identNode.GetText()
	} else {
		l.addError(ctx, "Missing loop variable identifier in FOR_EACH statement")
		loopVarName = "_invalidLoopVar_" // To prevent nil issues, though it's a parse error
	}
	l.logDebugAST("         FOR_EACH loop variable: %s", loopVarName)

	if l.currentSteps == nil {
		l.addError(ctx, "Cannot append FOR_EACH step: currentSteps (parent context) is nil.")
		return
	}
	forStep := Step{
		Pos:         tokenToPosition(ctx.KW_FOR().GetSymbol()),
		Type:        "for_each",
		LoopVarName: loopVarName,
		Collection:  collectionNode,
		Body:        bodySteps,
	}
	*l.currentSteps = append(*l.currentSteps, forStep)
	l.logDebugAST("         Appended FOR_EACH Step to currentSteps list (%p)", l.currentSteps)
}

func (l *neuroScriptListenerImpl) EnterOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST(">>> Enter ON_ERROR Statement Context")
	l.enterBlockContext("ON_ERROR_BODY") // Enter context for the on_error handler's body
}

func (l *neuroScriptListenerImpl) ExitOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST("--- ExitOnErrorStmt: Finalizing ON_ERROR step (%s)", getRuleText(ctx))

	// Exit context for ON_ERROR_BODY, this pushes its steps to valueStack.
	_ = l.exitBlockContext("ON_ERROR_BODY")

	handlerStepsRaw, okHandler := l.popValue() // Pop handler steps
	if !okHandler {
		l.addError(ctx, "Stack error popping handler steps for ON_ERROR statement (after exitBlockContext)")
		return
	}
	handlerSteps, isHandlerSteps := handlerStepsRaw.([]Step)
	if !isHandlerSteps {
		l.addError(ctx, "ON_ERROR handler steps are not []Step (got %T)", handlerStepsRaw)
		l.pushValue(handlerStepsRaw) // Push back if wrong type
		return
	}
	l.logDebugAST("         Popped ON_ERROR handler_steps: Count=%d", len(handlerSteps))

	if l.currentSteps == nil {
		l.addError(ctx, "Cannot append ON_ERROR step: currentSteps (parent context) is nil.")
		return
	}
	onErrorStep := Step{
		Pos:  tokenToPosition(ctx.KW_ON_ERROR().GetSymbol()),
		Type: "on_error",
		Body: handlerSteps,
	}
	*l.currentSteps = append(*l.currentSteps, onErrorStep)
	l.logDebugAST("         Appended ON_ERROR Step to currentSteps list (%p)", l.currentSteps)
}
