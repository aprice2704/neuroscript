// NeuroScript Version: 0.3.0
// File version: 9
// Purpose: Adds block context handling for on_error statements to prevent panics.
// filename: pkg/core/ast_builder_blocks.go
// nlines: 110
// risk_rating: MEDIUM

package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// enterBlockContext sets up a new []Step slice for a new block and records the
// current value stack depth. This prepares the listener for collecting steps
// within a new lexical scope (like an if-body, loop-body, or proc-body).
func (l *neuroScriptListenerImpl) enterBlockContext(kind string) {
	l.logDebugAST(">>> Enter %s Block (valDepth: %d)", kind, len(l.valueStack))

	// 1. Record the current depth of the value stack.
	l.blockValueDepthStack = append(l.blockValueDepthStack, len(l.valueStack))

	// 2. Save the parent's step collector by pushing it onto the block stack.
	if l.currentSteps != nil {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps)
	}

	// 3. Create a fresh, empty slice for collecting the new block's steps.
	fresh := make([]Step, 0)
	l.currentSteps = &fresh
}

// exitBlockContext finalizes the current block's step collection. It pushes
// the completed slice of steps for the block onto the value stack, making it
// available for the parent grammar rule (e.g., ExitIf_statement) to consume.
// It then restores the previous (parent) step collector.
func (l *neuroScriptListenerImpl) exitBlockContext(kind string) {
	if l.currentSteps == nil {
		l.logger.Error("AST Builder FATAL: exitBlockContext called with nil currentSteps", "kind", kind)
		// To prevent further panics, we push an empty slice.
		l.pushValue([]Step{})
		return
	}

	completedChildSteps := *l.currentSteps
	l.logDebugAST("<<< Exit %s Block (items: %d, valDepth: %d)", kind, len(completedChildSteps), len(l.valueStack))

	// 1. Push the completed slice of steps onto the value stack for the parent rule.
	l.pushValue(completedChildSteps)

	// 2. Restore the parent's step collector from the block stack.
	if len(l.blockStepStack) > 0 {
		l.currentSteps = l.blockStepStack[len(l.blockStepStack)-1]
		l.blockStepStack = l.blockStepStack[:len(l.blockStepStack)-1]
	} else {
		l.currentSteps = nil // No parent context left.
	}

	// 3. Clean up any unexpected values left on the stack by child rules.
	// A correctly implemented block should leave exactly one item on the stack: its body.
	markerIdx := l.blockValueDepthStack[len(l.blockValueDepthStack)-1]
	allowedDepth := markerIdx + 1

	for len(l.valueStack) > allowedDepth {
		v := l.pop()
		l.logger.Warn("[AST] stray value purged during %s exit: %T", kind, v)
	}

	// 4. Pop the depth marker for this block.
	l.blockValueDepthStack = l.blockValueDepthStack[:len(l.blockValueDepthStack)-1]
}

// --- Statement_list listener wiring ---

// EnterStatement_list now treats every list of statements as a generic block.
// This simplifies logic and makes it more robust.
func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	var kind string
	switch ctx.GetParent().(type) {
	case *gen.Procedure_definitionContext:
		kind = "PROC_BODY"
	case *gen.If_statementContext:
		kind = "IF_ELSE_BODY"
	case *gen.For_each_statementContext:
		kind = "FOR_EACH_BODY"
	case *gen.While_statementContext:
		kind = "WHILE_BODY"
	case *gen.OnEventStmtContext:
		kind = "ON_EVENT_BODY"
	case *gen.OnErrorStmtContext: // BUG FIX: Added missing case for on_error blocks.
		kind = "ON_ERROR_BODY"
	default:
		kind = "UNKNOWN_BLOCK"
		l.addError(ctx, "statement list found inside unknown parent type: %T", ctx.GetParent())
	}
	l.enterBlockContext(kind)
}

// ExitStatement_list finalizes the generic block, pushing the collected
// steps onto the value stack for the parent rule to consume.
func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	var kind string
	switch ctx.GetParent().(type) {
	case *gen.Procedure_definitionContext:
		kind = "PROC_BODY"
	case *gen.If_statementContext:
		kind = "IF_ELSE_BODY"
	case *gen.For_each_statementContext:
		kind = "FOR_EACH_BODY"
	case *gen.While_statementContext:
		kind = "WHILE_BODY"
	case *gen.OnEventStmtContext:
		kind = "ON_EVENT_BODY"
	case *gen.OnErrorStmtContext: // BUG FIX: Added missing case for on_error blocks.
		kind = "ON_ERROR_BODY"
	default:
		kind = "UNKNOWN_BLOCK"
	}
	l.exitBlockContext(kind)
}
