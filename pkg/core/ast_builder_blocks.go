// filename: pkg/core/ast_builder_blocks.go
package core

import (
	// Import fmt

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Block Handling: Enter/Exit for Control Flow Statements ---
// *** MODIFIED: Use stack push/pop for step lists, assert conditions ***

// Enter block statement: Pushes parent step list pointer, creates new list for block
func (l *neuroScriptListenerImpl) enterBlockContext(blockType string) {
	l.logDebugAST(">>> Enter %s Statement Context", blockType)
	if l.currentSteps == nil {
		// This might happen if a block is somehow the very first thing in a procedure,
		// though the grammar likely requires a statement_list container.
		// Or if a previous error occurred.
		l.logger.Warn("Entering %s block, but currentSteps is nil. Starting fresh.", blockType)
		l.blockStepStack = append(l.blockStepStack, nil) // Push nil parent
		newSteps := make([]Step, 0)
		l.currentSteps = &newSteps // Start new list
	} else {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps) // Push pointer to parent list
		newSteps := make([]Step, 0)
		l.currentSteps = &newSteps // Start new list for the block body
		l.logDebugAST("    %s: Pushed parent context %p. Stack size: %d. New steps: %p", blockType, l.blockStepStack[len(l.blockStepStack)-1], len(l.blockStepStack), l.currentSteps)
	}
}

// Exit block statement: Pops parent step list pointer to restore context
func (l *neuroScriptListenerImpl) exitBlockContext(blockType string) {
	l.logDebugAST("<<< Exit %s Statement Context - Restoring parent steps", blockType)
	stackSize := len(l.blockStepStack)
	if stackSize == 0 {
		l.logger.Error("%s: Cannot restore parent context, stack empty!", blockType)
		l.currentSteps = nil // Lost context
		return
	}
	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex] // Pop stack

	// Restore l.currentSteps
	l.currentSteps = parentStepsPtr

	if l.currentSteps == nil {
		l.logger.Warn("%s: Restored parent context, but it was nil (Stack size: %d)", blockType, len(l.blockStepStack))
	} else {
		l.logDebugAST("    %s: Restored parent context %p (Stack size: %d)", blockType, l.currentSteps, len(l.blockStepStack))
	}
}

// EnterStatement_list: Ensures currentSteps is initialized if needed
func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	// Ensure l.currentSteps is valid when entering a list, especially at procedure start
	if l.currentSteps == nil {
		l.logDebugAST(">>> Enter Statement_list: currentSteps was nil, initializing.")
		newSteps := make([]Step, 0)
		l.currentSteps = &newSteps
	} else {
		l.logDebugAST(">>> Enter Statement_list: currentSteps already exists (%p)", l.currentSteps)
	}
}

// ExitStatement_list: Pushes the completed list of steps onto the value stack
func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	if l.currentSteps == nil {
		l.logger.Warn("<<< Exit Statement_list: currentSteps is nil, pushing empty list.")
		l.pushValue([]Step{}) // Push empty slice
	} else {
		l.logDebugAST("<<< Exit Statement_list: Pushing %d steps onto value stack.", len(*l.currentSteps))
		l.pushValue(*l.currentSteps) // Push the slice value
	}
	// After pushing, we might want to reset currentSteps if the parent Exit method
	// doesn't immediately restore it, but the Enter/Exit block context handles this.
}

// --- IF Statement ---
func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.enterBlockContext("IF") // Push parent context, prepare for first statement list (THEN)
}

func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("--- Exit If_statement Finalization")

	// Stack order depends on whether ELSE exists:
	// With ELSE:    [ConditionExpr, ThenStepsSlice, ElseStepsSlice]
	// Without ELSE: [ConditionExpr, ThenStepsSlice]

	var elseSteps []Step = nil // Default to nil (no else block)
	var thenSteps []Step
	var conditionNode Expression

	// 1. Pop Else Steps (if they exist)
	if ctx.KW_ELSE() != nil {
		elseStepsRaw, ok := l.popValue()
		if !ok {
			l.addError(ctx, "Internal error: Stack error popping ELSE steps for IF")
			l.exitBlockContext("IF - Error") // Attempt to restore context
			return
		}
		// Assert it's a []Step slice
		elseStepsAsserted, ok := elseStepsRaw.([]Step)
		if !ok {
			l.addError(ctx, "Internal error: ELSE block value is not []Step (got %T)", elseStepsRaw)
			l.exitBlockContext("IF - Error")
			return
		}
		elseSteps = elseStepsAsserted
		l.logDebugAST("    Popped %d ELSE steps", len(elseSteps))
	}

	// 2. Pop Then Steps
	thenStepsRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping THEN steps for IF")
		l.exitBlockContext("IF - Error")
		return
	}
	thenStepsAsserted, ok := thenStepsRaw.([]Step)
	if !ok {
		l.addError(ctx, "Internal error: THEN block value is not []Step (got %T)", thenStepsRaw)
		l.exitBlockContext("IF - Error")
		return
	}
	thenSteps = thenStepsAsserted
	l.logDebugAST("    Popped %d THEN steps", len(thenSteps))

	// 3. Pop Condition Expression
	conditionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping condition for IF")
		l.exitBlockContext("IF - Error")
		return
	}
	conditionAsserted, ok := conditionRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: IF condition is not an Expression (got %T)", conditionRaw)
		l.exitBlockContext("IF - Error")
		return
	}
	conditionNode = conditionAsserted
	l.logDebugAST("    Popped IF condition: %T", conditionNode)

	// 4. Restore Parent Context *before* appending
	l.exitBlockContext("IF")
	if l.currentSteps == nil {
		l.logger.Error("ExitIf_statement: Parent step list nil after restore. Cannot append IF step.")
		// Cannot proceed if parent context is lost
		return
	}

	// 5. Create and Append IF Step to Parent
	ifStep := Step{
		Pos:       tokenToPosition(ctx.KW_IF().GetSymbol()), // Position of 'if' keyword
		Type:      "if",
		Cond:      conditionNode,
		Value:     thenSteps, // Assign the []Step slice
		ElseValue: elseSteps, // Assign the []Step slice (or nil)
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, ifStep)
	thenLen := len(thenSteps)
	elseLen := 0
	if elseSteps != nil { // Check if elseSteps slice itself is nil before getting len
		elseLen = len(elseSteps)
	}
	l.logDebugAST("    Appended complete IF Step to parent: Cond=%T, THEN Steps=%d, ELSE Steps=%d", ifStep.Cond, thenLen, elseLen)
}

// --- WHILE Statement ---
func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.enterBlockContext("WHILE") // Push parent, prepare for body steps
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST("--- Exit While_statement Finalization")
	// Stack: [ConditionExpr, BodyStepsSlice]

	// 1. Pop Body Steps
	bodyStepsRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping body steps for WHILE")
		l.exitBlockContext("WHILE - Error")
		return
	}
	bodySteps, ok := bodyStepsRaw.([]Step)
	if !ok {
		l.addError(ctx, "Internal error: WHILE body value is not []Step (got %T)", bodyStepsRaw)
		l.exitBlockContext("WHILE - Error")
		return
	}
	l.logDebugAST("    Popped %d WHILE body steps", len(bodySteps))

	// 2. Pop Condition Expression
	conditionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping condition for WHILE")
		l.exitBlockContext("WHILE - Error")
		return
	}
	conditionNode, ok := conditionRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: WHILE condition is not an Expression (got %T)", conditionRaw)
		l.exitBlockContext("WHILE - Error")
		return
	}
	l.logDebugAST("    Popped WHILE condition: %T", conditionNode)

	// 3. Restore Parent Context
	l.exitBlockContext("WHILE")
	if l.currentSteps == nil {
		l.logger.Error("ExitWhile_statement: Parent step list nil after restore. Cannot append WHILE step.")
		return
	}

	// 4. Create and Append WHILE Step to Parent
	whileStep := Step{
		Pos:       tokenToPosition(ctx.KW_WHILE().GetSymbol()), // Position of 'while' keyword
		Type:      "while",
		Cond:      conditionNode,
		Value:     bodySteps, // Assign the []Step slice
		ElseValue: nil,       // Not used
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, whileStep)
	l.logDebugAST("    Appended WHILE Step: Cond=%T, Steps=%d", conditionNode, len(bodySteps))
}

// --- FOR EACH Statement ---
func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	l.enterBlockContext("FOR") // Push parent, prepare for body steps
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST("--- Exit For_each_statement Finalization")
	// Stack: [CollectionExpr, BodyStepsSlice]

	// 1. Pop Body Steps
	bodyStepsRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping body steps for FOR")
		l.exitBlockContext("FOR - Error")
		return
	}
	bodySteps, ok := bodyStepsRaw.([]Step)
	if !ok {
		l.addError(ctx, "Internal error: FOR body value is not []Step (got %T)", bodyStepsRaw)
		l.exitBlockContext("FOR - Error")
		return
	}
	l.logDebugAST("    Popped %d FOR body steps", len(bodySteps))

	// 2. Pop Collection Expression
	collectionRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping collection expression for FOR")
		l.exitBlockContext("FOR - Error")
		return
	}
	collectionNode, ok := collectionRaw.(Expression)
	if !ok {
		l.addError(ctx, "Internal error: FOR collection is not an Expression (got %T)", collectionRaw)
		l.exitBlockContext("FOR - Error")
		return
	}
	l.logDebugAST("    Popped FOR collection: %T", collectionNode)

	// 3. Get Loop Variable Name
	loopVar := ""
	if id := ctx.IDENTIFIER(); id != nil {
		loopVar = id.GetText()
	} else {
		// Grammar requires IDENTIFIER, this is an internal error if missing
		l.addError(ctx, "Internal error: Missing IDENTIFIER for loop variable in FOR statement")
		l.exitBlockContext("FOR - Error")
		return
	}

	// 4. Restore Parent Context
	l.exitBlockContext("FOR")
	if l.currentSteps == nil {
		l.logger.Error("ExitFor_each_statement: Parent step list nil after restore. Cannot append FOR step.")
		return
	}

	// 5. Create and Append FOR Step to Parent
	forStep := Step{
		Pos:       tokenToPosition(ctx.KW_FOR().GetSymbol()), // Position of 'for' keyword
		Type:      "for",
		Target:    loopVar,        // Loop variable name
		Cond:      collectionNode, // Collection expression assigned to Cond field
		Value:     bodySteps,      // Body steps assigned to Value field
		ElseValue: nil,            // Not used
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, forStep)
	l.logDebugAST("    Appended FOR Step: Var=%q, Collection=%T, Steps=%d", loopVar, collectionNode, len(bodySteps))
}

// --- ON ERROR Statement ---
func (l *neuroScriptListenerImpl) EnterOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.enterBlockContext("ON_ERROR") // Push parent, prepare for handler steps
}

func (l *neuroScriptListenerImpl) ExitOnErrorStmt(ctx *gen.OnErrorStmtContext) {
	l.logDebugAST("--- Exit OnErrorStmt Finalization")
	// Stack: [HandlerStepsSlice]

	// 1. Pop Handler Steps
	handlerStepsRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Internal error: Stack error popping handler steps for ON_ERROR")
		l.exitBlockContext("ON_ERROR - Error")
		return
	}
	handlerSteps, ok := handlerStepsRaw.([]Step)
	if !ok {
		l.addError(ctx, "Internal error: ON_ERROR handler value is not []Step (got %T)", handlerStepsRaw)
		l.exitBlockContext("ON_ERROR - Error")
		return
	}
	l.logDebugAST("    Popped %d ON_ERROR handler steps", len(handlerSteps))

	// 2. Restore Parent Context
	l.exitBlockContext("ON_ERROR")
	if l.currentSteps == nil {
		l.logger.Error("ExitOnErrorStmt: Parent step list nil after restore. Cannot append ON_ERROR step.")
		return
	}

	// 3. Create and Append ON_ERROR Step to Parent
	onErrorStep := Step{
		Pos:       tokenToPosition(ctx.KW_ON_ERROR().GetSymbol()), // Position of 'on_error'
		Type:      "on_error",
		Target:    "",           // Not used
		Cond:      nil,          // Not used
		Value:     handlerSteps, // Handler steps assigned to Value
		ElseValue: nil,          // Not used
		Metadata:  make(map[string]string),
	}
	*l.currentSteps = append(*l.currentSteps, onErrorStep)
	l.logDebugAST("    Appended ON_ERROR Step: HandlerSteps=%d", len(handlerSteps))
}
