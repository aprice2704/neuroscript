// ast_builder_blocks.go – Block-context helpers and Statement_list handling
// file version: 5
//
// This file centralises the mechanics of entering and exiting nested statement
// blocks while walking the ANTLR parse tree.  It guarantees the two stack
// invariants documented in ast.go:
//
//   • blockStepStack        mirrors the chain of active parent []Step slices.
//   • valueStack            returns to its pre-block depth *plus exactly one*
//                            item (the child body []Step).
//
// Implementation
// --------------
//   enterBlockContext(kind)
//       1. record current valueStack depth in blockValueDepthStack
//       2. push currentSteps (if non-nil) onto blockStepStack
//       3. create a fresh []Step  -> currentSteps
//
//   exitBlockContext(kind)
//       1. push the completed child []Step onto valueStack
//       2. restore currentSteps from blockStepStack
//       3. flush (and log) any surplus operands above depth+1
//       4. drop the depth marker
//
// NOTE:  Ensure `neuroScriptListenerImpl` contains
//          blockValueDepthStack []int
//        in addition to blockStepStack.

package core

import (
	"fmt"

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// ------------------------------------------------------------------
// Low-level helpers
// ------------------------------------------------------------------

// enterBlockContext sets up a new []Step slice and records stack depth.
func (l *neuroScriptListenerImpl) enterBlockContext(kind string) {
	l.logDebugAST(">>> Enter %s Block (currSteps=%p, blockStepStack=%d, valDepth=%d)",
		kind, l.currentSteps, len(l.blockStepStack), len(l.valueStack))

	// 1. remember valueStack depth
	l.blockValueDepthStack = append(l.blockValueDepthStack, len(l.valueStack))

	// 2. save parent steps
	if l.currentSteps != nil {
		l.blockStepStack = append(l.blockStepStack, l.currentSteps)
	} else {
		l.logger.Warn(fmt.Sprintf("Entering %s with nil currentSteps (top-level body?)", kind))
	}

	// 3. fresh slice for this block
	fresh := make([]Step, 0)
	l.currentSteps = &fresh
}

// exitBlockContext finalises the block, flushes leaks, restores parent.
func (l *neuroScriptListenerImpl) exitBlockContext(kind string) []Step {
	l.logDebugAST("<<< Exit %s Block (items=%d, valDepth=%d)",
		kind, len(*l.currentSteps), len(l.valueStack))

	child := *l.currentSteps

	// 1. push child body
	l.pushValue(child)

	// 2. restore parent currentSteps
	if len(l.blockStepStack) > 0 {
		l.currentSteps = l.blockStepStack[len(l.blockStepStack)-1]
		l.blockStepStack = l.blockStepStack[:len(l.blockStepStack)-1]
	} else {
		l.currentSteps = nil
	}

	// 3. flush surplus operands ABOVE (marker + 2)
	markerIdx := l.blockValueDepthStack[len(l.blockValueDepthStack)-1]
	allowedDepth := markerIdx + 2 // ← was +1

	for len(l.valueStack) > allowedDepth {
		v, _ := l.popValue()
		l.logger.Warn("[AST] stray value purged during %s exit: %T", kind, v)
	}

	// 4. pop depth marker
	l.blockValueDepthStack = l.blockValueDepthStack[:len(l.blockValueDepthStack)-1]
	return child
}

// ------------------------------------------------------------------
// Statement_list listener wiring
// ------------------------------------------------------------------

func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	switch parent := ctx.GetParent().(type) {

	// Procedure body
	case *gen.Procedure_definitionContext:
		if l.currentProc == nil {
			l.addError(ctx, "internal: currentProc nil entering proc body")
			dummy := make([]Step, 0)
			l.currentSteps = &dummy
		} else {
			l.currentSteps = &l.currentProc.Steps
		}

	// IF-THEN / IF-ELSE
	case *gen.If_statementContext:
		if parent.Statement_list(0) == ctx {
			l.enterBlockContext("IF_THEN_BODY")
		} else if len(parent.AllStatement_list()) > 1 && parent.Statement_list(1) == ctx {
			l.enterBlockContext("IF_ELSE_BODY")
		}

	// ON_EVENT handler body
	case *gen.OnEventStmtContext:
		l.enterBlockContext("ON_EVENT_BODY")
	}
}

func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	switch parent := ctx.GetParent().(type) {

	// Procedure body
	case *gen.Procedure_definitionContext:
		l.pushValue(*l.currentSteps) // consumed by ExitProcedure_definition

	// IF-THEN / IF-ELSE
	case *gen.If_statementContext:
		if parent.Statement_list(0) == ctx {
			l.exitBlockContext("IF_THEN_BODY")
		} else if len(parent.AllStatement_list()) > 1 && parent.Statement_list(1) == ctx {
			l.exitBlockContext("IF_ELSE_BODY")
		}

	// ON_EVENT handler body
	case *gen.OnEventStmtContext:
		l.exitBlockContext("ON_EVENT_BODY")
	}
}
