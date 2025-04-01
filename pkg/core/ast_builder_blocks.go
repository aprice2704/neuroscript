// pkg/core/ast_builder_blocks.go
package core

import (
	// For logging

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Block Handling (Helpers and Statement Exits) ---

// enterBlock prepares the listener state for entering a nested block (IF, WHILE, FOR).
func (l *neuroScriptListenerImpl) enterBlock(blockType string, targetVar string) {
	l.logDebugAST(">>> enterBlock: %s (Target: %q)", blockType, targetVar)
	if l.currentSteps == nil {
		l.logger.Printf("[WARN] enterBlock (%s) called outside a valid step context", blockType)
		return
	}

	// 1. Create the placeholder step for the block itself, initially with nil Value.
	step := newStep(blockType, targetVar, nil, nil /* Body assigned later */, nil)

	// 2. Append this placeholder Step to the *current* (parent's) step list.
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("    Appended placeholder %s step to parent list (addr: %p)", blockType, l.currentSteps)

	// 3. Push the pointer to the *current* (parent's) step list onto the stack.
	l.blockStepStack = append(l.blockStepStack, l.currentSteps)

	// 4. Create a *new, separate* slice where the block's body steps will be built temporarily.
	newBlockSteps := make([]Step, 0)

	// 5. Set l.currentSteps to point to this *new, separate* slice.
	l.currentSteps = &newBlockSteps
	l.logDebugAST("    Started new block context. Parent stack size: %d, New currentSteps pointer: %p", len(l.blockStepStack), l.currentSteps)
}

// exitBlock finalizes a nested block.
func (l *neuroScriptListenerImpl) exitBlock(blockType string) {
	stackSize := len(l.blockStepStack)
	l.logDebugAST("<<< exitBlock: %s (Stack size before pop: %d)", blockType, stackSize)
	if stackSize == 0 {
		l.logger.Printf("[ERROR] exitBlock (%s) called with empty stack!", blockType)
		l.currentSteps = nil // Reset to avoid dangling pointer
		return
	}

	// 1. Get the slice of steps that were built for the block's body.
	finishedBlockSteps := *l.currentSteps // Dereference pointer to get the slice built in the block context

	// 2. Pop the pointer to the parent step list from the stack.
	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex] // Pop stack

	if parentStepsPtr == nil {
		l.logger.Printf("[ERROR] exitBlock (%s): Parent step list pointer from stack was nil", blockType)
		l.currentSteps = nil // Reset
		return
	}

	// 3. Find the placeholder Step struct that was added in enterBlock.
	parentStepIndex := len(*parentStepsPtr) - 1
	if parentStepIndex < 0 {
		l.logger.Printf("[ERROR] exitBlock (%s): Parent step list pointer was valid but list was empty?", blockType)
		l.currentSteps = parentStepsPtr // Restore parent pointer anyway
		return
	}
	blockStepToUpdate := &(*parentStepsPtr)[parentStepIndex]

	// 4. Assign the finished block body slice to the placeholder Step's Value field.
	if blockStepToUpdate.Type == blockType {
		blockStepToUpdate.Value = finishedBlockSteps // Assign the built slice
		l.logDebugAST("    Assigned %d steps to %s block step at parent index %d", len(finishedBlockSteps), blockType, parentStepIndex)
	} else {
		l.logger.Printf("[ERROR] exitBlock (%s): Mismatched block type on stack. Expected %s, found %s", blockType, blockType, blockStepToUpdate.Type)
	}

	// 5. Restore l.currentSteps to point to the parent list.
	l.currentSteps = parentStepsPtr
	l.logDebugAST("    Finished block. Restored currentSteps to parent: %p", l.currentSteps)
}

// Exit methods for blocks now pop the condition/collection node and store it
func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("<<< Exit If_statement")

	// --- FIX: Pop correct number of nodes based on condition structure ---
	var condNode interface{} // The node we will store (primary expression)
	var ok bool
	numConditionChildren := ctx.Condition().GetChildCount() // Includes expressions and potentially the operator

	if numConditionChildren == 1 { // Condition was just a single expression
		condNode, ok = l.popValue() // Pop the single expression node
		if !ok {
			l.logger.Println("[ERROR] AST Builder: Failed to pop condition for simple IF")
			l.exitBlock("IF") // Balance stack
			return
		}
		l.logDebugAST("    Popped 1 node for simple IF condition")
	} else if numConditionChildren == 3 { // Condition was expr OP expr
		// Pop the two expression nodes. Order matters: RHS then LHS.
		var node2, node1 interface{}
		node2, ok = l.popValue() // Pop RHS expression node
		if !ok {
			l.logger.Println("[ERROR] AST Builder: Failed to pop RHS condition for comparison IF")
			l.exitBlock("IF")
			return
		}
		node1, ok = l.popValue() // Pop LHS expression node
		if !ok {
			l.logger.Println("[ERROR] AST Builder: Failed to pop LHS condition for comparison IF")
			l.exitBlock("IF")
			return
		}
		// For now, we only store the LHS node in Step.Cond, as the interpreter
		// doesn't have a comparison node type yet. The main thing is cleaning the stack.
		condNode = node1
		l.logDebugAST("    Popped 2 nodes for comparison IF condition (node1=%T, node2=%T)", node1, node2)
		// TODO: Future - Create a ComparisonNode here using node1, node2, and the operator from ctx.Condition()
	} else {
		l.logger.Printf("[ERROR] AST Builder: Unexpected number of children (%d) in ConditionContext for IF", numConditionChildren)
		// Attempt to recover stack? Just exit block for now.
		condNode = nil // Cannot determine condition node
		l.exitBlock("IF")
		return
	}
	// --- END FIX ---

	l.exitBlock("IF") // Assign body steps to the Value field, restore context

	// Update the Cond field of the IF step (which is the last step in the *parent* list now)
	if l.currentSteps != nil && len(*l.currentSteps) > 0 {
		lastStepIndex := len(*l.currentSteps) - 1
		stepPtr := &(*l.currentSteps)[lastStepIndex]
		if stepPtr.Type == "IF" {
			stepPtr.Cond = condNode // Store the determined node (usually LHS) in Cond
			l.logDebugAST("    Updated IF Step %d Cond: %T %+v", lastStepIndex, stepPtr.Cond, stepPtr.Cond)
		} else {
			l.logger.Printf("[ERROR] Last step wasn't IF after exitBlock(IF). Type was %s", stepPtr.Type)
		}
	} else {
		l.logger.Println("[ERROR] Current step list was nil or empty after exitBlock(IF)")
	}
}

// NOTE: Similar logic might be needed for ExitWhile_statement if its condition
// also leaves multiple nodes on the stack. Let's fix IF first.
func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST("<<< Exit While_statement")

	// --- Apply similar fix as IF ---
	var condNode interface{}
	var ok bool
	numConditionChildren := ctx.Condition().GetChildCount()

	if numConditionChildren == 1 {
		condNode, ok = l.popValue()
		if !ok {
			l.logger.Println("[ERROR] AST Builder: Failed to pop condition for simple WHILE")
			l.exitBlock("WHILE")
			return
		}
		l.logDebugAST("    Popped 1 node for simple WHILE condition")
	} else if numConditionChildren == 3 {
		var node2, node1 interface{}
		node2, ok = l.popValue()
		if !ok {
			l.logger.Println("[ERROR] AST Builder: Failed to pop RHS condition for comparison WHILE")
			l.exitBlock("WHILE")
			return
		}
		node1, ok = l.popValue()
		if !ok {
			l.logger.Println("[ERROR] AST Builder: Failed to pop LHS condition for comparison WHILE")
			l.exitBlock("WHILE")
			return
		}
		condNode = node1
		l.logDebugAST("    Popped 2 nodes for comparison WHILE condition (node1=%T, node2=%T)", node1, node2)
		// TODO: Future - Create ComparisonNode
	} else {
		l.logger.Printf("[ERROR] AST Builder: Unexpected number of children (%d) in ConditionContext for WHILE", numConditionChildren)
		condNode = nil
		l.exitBlock("WHILE")
		return
	}
	// --- End Fix ---

	l.exitBlock("WHILE") // Assign body to Value, restore context

	// Update Cond field
	if l.currentSteps != nil && len(*l.currentSteps) > 0 {
		lastStepIndex := len(*l.currentSteps) - 1
		stepPtr := &(*l.currentSteps)[lastStepIndex]
		if stepPtr.Type == "WHILE" {
			stepPtr.Cond = condNode // Store node in Cond
			l.logDebugAST("    Updated WHILE Step %d Cond: %T %+v", lastStepIndex, stepPtr.Cond, stepPtr.Cond)
		} else {
			l.logger.Printf("[ERROR] Last step wasn't WHILE after exitBlock(WHILE). Type was %s", stepPtr.Type)
		}
	} else {
		l.logger.Println("[ERROR] Current step list was nil or empty after exitBlock(WHILE)")
	}
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST("<<< Exit For_each_statement")

	// Log stack *before* popping collection
	l.logDebugAST("[DEBUG-AST] AST Builder: Stack before popping collection for FOR EACH: %+v", l.valueStack)

	collectionNode, ok := l.popValue() // Pop collection node
	if !ok {
		l.logger.Println("[ERROR] AST Builder: Failed to pop collection node for FOR EACH")
		l.exitBlock("FOR") // Balance stack
		return
	}

	// Log popped collection node
	l.logDebugAST("[DEBUG-AST] AST Builder: Popped collection node for FOR EACH: %T %+v", collectionNode, collectionNode)

	loopVar := "" // Default loopVar to empty string if IDENTIFIER is nil
	if ctx.IDENTIFIER() != nil {
		loopVar = ctx.IDENTIFIER().GetText()
	}

	l.exitBlock("FOR") // Assign body to Value, restore context

	// Update Cond (collection node) and Target (loop var string) fields
	if l.currentSteps != nil && len(*l.currentSteps) > 0 {
		lastStepIndex := len(*l.currentSteps) - 1
		stepPtr := &(*l.currentSteps)[lastStepIndex] // Get pointer to the step in the slice

		if stepPtr.Type == "FOR" {
			// --- Original Assignment Order ---
			stepPtr.Target = loopVar      // Assign loop var name ("item") to Target
			stepPtr.Cond = collectionNode // Assign popped node (my_list) to Cond
			// --- End Original Assignment Order ---

			// Log AFTER assignment using logger
			l.logDebugAST("[DEBUG-AST] AST Builder: Assigned to FOR Step %d - Cond: %T %+v, Target: %q",
				lastStepIndex,
				stepPtr.Cond, // Log the value *just assigned*
				stepPtr.Cond,
				stepPtr.Target)

		} else {
			l.logger.Printf("[ERROR] Last step wasn't FOR after exitBlock(FOR). Type was %s", stepPtr.Type)
		}
	} else {
		l.logger.Println("[ERROR] Current step list was nil or empty after exitBlock(FOR)")
	}
}

// Enter methods for blocks add placeholder Step, manage block stack
func (l *neuroScriptListenerImpl) EnterIf_statement(ctx *gen.If_statementContext) {
	l.enterBlock("IF", "")
}
func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.enterBlock("WHILE", "")
}
func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	loopVar := ""
	if ctx.IDENTIFIER() != nil {
		loopVar = ctx.IDENTIFIER().GetText()
	}
	l.enterBlock("FOR", loopVar)
}
