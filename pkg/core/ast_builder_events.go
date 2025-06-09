// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Removed block context calls, delegating them to the Statement_list listener.
// filename: pkg/core/ast_builder_events.go
// nlines: 52
// risk_rating: MEDIUM

package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Listener Methods for 'on event' statements ---

func (l *neuroScriptListenerImpl) EnterOnEventStmt(ctx *gen.OnEventStmtContext) {
	l.logDebugAST(">>> EnterOnEventStmt: %s", getRuleText(ctx))
	// Block context is now handled by EnterStatement_list
}

func (l *neuroScriptListenerImpl) ExitOnEventStmt(ctx *gen.OnEventStmtContext) {
	l.logDebugAST("<<< ExitOnEventStmt")
	// Block context is handled by ExitStatement_list, which pushes the body steps to the stack.

	// 1. Pop the body steps ([]Step)
	bodyStepsRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Stack error: could not pop steps for 'on event' block.")
		return
	}
	bodySteps, isSteps := bodyStepsRaw.([]Step)
	if !isSteps {
		l.addError(ctx, "Type error: expected []Step for 'on event' body, got %T", bodyStepsRaw)
		l.pushValue(bodyStepsRaw) // Push back for debugging
		return
	}

	// 2. Pop the event name expression (Expression).
	eventNameExprRaw, ok := l.popValue()
	if !ok {
		l.addError(ctx, "Stack error: could not pop event name expression for 'on event' block.")
		return
	}
	eventNameExpr, isExpr := eventNameExprRaw.(Expression)
	if !isExpr {
		l.addError(ctx, "Type error: expected Expression for event name, got %T", eventNameExprRaw)
		l.pushValue(eventNameExprRaw) // Push back
		return
	}

	// 3. Construct the final OnEventNode.
	node := &OnEventNode{
		Pos:           tokenToPosition(ctx.KW_ON().GetSymbol()),
		EventNameExpr: eventNameExpr,
		Steps:         bodySteps,
	}

	if id := ctx.IDENTIFIER(); id != nil {
		node.PayloadVariable = id.GetText()
	}

	// 4. Add the new node to the program's list of event handlers.
	if l.program.EventHandlers == nil {
		l.program.EventHandlers = make([]*OnEventNode, 0)
	}
	l.program.EventHandlers = append(l.program.EventHandlers, node)
	l.logDebugAST("    Added 'on event' handler for: %s", getRuleText(ctx.Expression()))
}
