// pkg/core/ast_builder_blocks.go
package core

import (
	// For logging

	"github.com/antlr4-go/antlr/v4" // Import antlr
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
	step := newStep(blockType, targetVar, nil, nil /* Body assigned later */, nil)
	*l.currentSteps = append(*l.currentSteps, step)
	l.logDebugAST("    Appended placeholder %s step to parent list (addr: %p)", blockType, l.currentSteps)
	l.blockStepStack = append(l.blockStepStack, l.currentSteps)
	newBlockSteps := make([]Step, 0)
	l.currentSteps = &newBlockSteps
	l.logDebugAST("    Started new block context. Parent stack size: %d, New currentSteps pointer: %p", len(l.blockStepStack), l.currentSteps)
}

// exitBlock finalizes a nested block.
func (l *neuroScriptListenerImpl) exitBlock(blockType string) {
	stackSize := len(l.blockStepStack)
	l.logDebugAST("<<< exitBlock: %s (Stack size before pop: %d)", blockType, stackSize)
	if stackSize == 0 {
		l.logger.Printf("[ERROR] exitBlock (%s) called with empty stack!", blockType)
		l.currentSteps = nil
		return
	}
	finishedBlockSteps := *l.currentSteps
	lastIndex := stackSize - 1
	parentStepsPtr := l.blockStepStack[lastIndex]
	l.blockStepStack = l.blockStepStack[:lastIndex]
	if parentStepsPtr == nil {
		l.logger.Printf("[ERROR] exitBlock (%s): Parent step list pointer from stack was nil", blockType)
		l.currentSteps = nil
		return
	}
	parentStepIndex := len(*parentStepsPtr) - 1
	if parentStepIndex < 0 {
		l.logger.Printf("[ERROR] exitBlock (%s): Parent step list pointer was valid but list was empty?", blockType)
		l.currentSteps = parentStepsPtr
		return
	}
	blockStepToUpdate := &(*parentStepsPtr)[parentStepIndex]
	if blockStepToUpdate.Type == blockType {
		blockStepToUpdate.Value = finishedBlockSteps
		l.logDebugAST("    Assigned %d steps to %s block step at parent index %d", len(finishedBlockSteps), blockType, parentStepIndex)
	} else {
		l.logger.Printf("[ERROR] exitBlock (%s): Mismatched block type on stack. Expected %s, found %s", blockType, blockType, blockStepToUpdate.Type)
	}
	l.currentSteps = parentStepsPtr
	l.logDebugAST("    Finished block. Restored currentSteps to parent: %p", l.currentSteps)
}

// --- FIX: Modified IF/WHILE to create ComparisonNode ---
func (l *neuroScriptListenerImpl) processCondition(ctx gen.IConditionContext) (interface{}, bool) {
	numConditionChildren := ctx.GetChildCount()
	opText := ""

	if numConditionChildren == 1 { // Condition was just a single expression
		condNode, ok := l.popValue()
		if !ok {
			return nil, false
		}
		l.logDebugAST("    Popped 1 node for simple condition")
		return condNode, true // Return the single expression node
	} else if numConditionChildren == 3 { // Condition was expr OP expr
		// Pop RHS first, then LHS
		nodeRHS, okRHS := l.popValue()
		if !okRHS {
			return nil, false
		}
		nodeLHS, okLHS := l.popValue()
		if !okLHS {
			return nil, false
		}

		// Get operator token text
		if opToken := ctx.GetChild(1).(antlr.TerminalNode); opToken != nil {
			opText = opToken.GetText()
		} else {
			l.logger.Printf("[ERROR] AST Builder: Could not get operator token text for comparison")
			return nil, false // Indicate error
		}

		l.logDebugAST("    Popped 2 nodes for comparison condition (LHS=%T, RHS=%T, Op=%q)", nodeLHS, nodeRHS, opText)
		// Create and return a ComparisonNode
		return ComparisonNode{Left: nodeLHS, Operator: opText, Right: nodeRHS}, true
	} else {
		l.logger.Printf("[ERROR] AST Builder: Unexpected number of children (%d) in ConditionContext", numConditionChildren)
		return nil, false // Indicate error
	}
}

func (l *neuroScriptListenerImpl) ExitIf_statement(ctx *gen.If_statementContext) {
	l.logDebugAST("<<< Exit If_statement")
	conditionNode, ok := l.processCondition(ctx.Condition()) // Use helper to process condition
	if !ok {
		l.logger.Println("[ERROR] AST Builder: Failed to process condition for IF")
		l.exitBlock("IF") // Balance stack
		return
	}

	l.exitBlock("IF") // Assign body steps to the Value field, restore context

	// Update the Cond field with the result from processCondition (either expression node or ComparisonNode)
	if l.currentSteps != nil && len(*l.currentSteps) > 0 {
		lastStepIndex := len(*l.currentSteps) - 1
		stepPtr := &(*l.currentSteps)[lastStepIndex]
		if stepPtr.Type == "IF" {
			stepPtr.Cond = conditionNode // Store the processed node in Cond
			l.logDebugAST("    Updated IF Step %d Cond: %T %+v", lastStepIndex, stepPtr.Cond, stepPtr.Cond)
		} else {
			l.logger.Printf("[ERROR] Last step wasn't IF after exitBlock(IF). Type was %s", stepPtr.Type)
		}
	} else {
		l.logger.Println("[ERROR] Current step list was nil or empty after exitBlock(IF)")
	}
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST("<<< Exit While_statement")
	conditionNode, ok := l.processCondition(ctx.Condition()) // Use helper
	if !ok {
		l.logger.Println("[ERROR] AST Builder: Failed to process condition for WHILE")
		l.exitBlock("WHILE") // Balance stack
		return
	}

	l.exitBlock("WHILE") // Assign body to Value, restore context

	// Update Cond field
	if l.currentSteps != nil && len(*l.currentSteps) > 0 {
		lastStepIndex := len(*l.currentSteps) - 1
		stepPtr := &(*l.currentSteps)[lastStepIndex]
		if stepPtr.Type == "WHILE" {
			stepPtr.Cond = conditionNode // Store node in Cond
			l.logDebugAST("    Updated WHILE Step %d Cond: %T %+v", lastStepIndex, stepPtr.Cond, stepPtr.Cond)
		} else {
			l.logger.Printf("[ERROR] Last step wasn't WHILE after exitBlock(WHILE). Type was %s", stepPtr.Type)
		}
	} else {
		l.logger.Println("[ERROR] Current step list was nil or empty after exitBlock(WHILE)")
	}
}

// --- END FIX ---

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST("<<< Exit For_each_statement")
	l.logDebugAST("[DEBUG-AST] AST Builder: Stack before popping collection for FOR EACH: %+v", l.valueStack)

	collectionNode, ok := l.popValue() // Pop collection node
	if !ok {
		l.logger.Println("[ERROR] AST Builder: Failed to pop collection node for FOR EACH")
		l.exitBlock("FOR") // Balance stack
		return
	}
	l.logDebugAST("[DEBUG-AST] AST Builder: Popped collection node for FOR EACH: %T %+v", collectionNode, collectionNode)

	loopVar := ""
	if ctx.IDENTIFIER() != nil {
		loopVar = ctx.IDENTIFIER().GetText()
	}

	l.exitBlock("FOR") // Assign body to Value, restore context

	if l.currentSteps != nil && len(*l.currentSteps) > 0 {
		lastStepIndex := len(*l.currentSteps) - 1
		stepPtr := &(*l.currentSteps)[lastStepIndex]
		if stepPtr.Type == "FOR" {
			stepPtr.Target = loopVar      // Assign loop var name ("item") to Target
			stepPtr.Cond = collectionNode // Assign popped node (my_list) to Cond
			l.logDebugAST("[DEBUG-AST] AST Builder: Assigned to FOR Step %d - Cond: %T %+v, Target: %q",
				lastStepIndex, stepPtr.Cond, stepPtr.Cond, stepPtr.Target)
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
