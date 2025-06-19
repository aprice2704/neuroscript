// ast_builder_events.go – Listener callbacks for `on event` declarations
// file version: 13
// Purpose: Corrected stack popping logic in ExitEvent_handler to be more robust.

package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

type onEventStackMarker struct{}

var onEventMarker = &onEventStackMarker{}

func (l *neuroScriptListenerImpl) EnterOn_stmt(ctx *gen.On_stmtContext) {
	l.logDebugAST(">>> EnterOn_stmt (push marker)")
	l.pushValue(onEventMarker)
}

func (l *neuroScriptListenerImpl) ExitOn_stmt(c *gen.On_stmtContext) {
	l.logDebugAST("<<< ExitOn_stmt")
}

func (l *neuroScriptListenerImpl) EnterEvent_handler(ctx *gen.Event_handlerContext) {}

func (l *neuroScriptListenerImpl) ExitEvent_handler(ctx *gen.Event_handlerContext) {
	l.logDebugAST("<<< ExitEvent_handler")

	// Pop the body, which was pushed by ExitStatement_list
	bodyVal, ok := l.popValue()
	if !ok || bodyVal == nil {
		l.addError(ctx, "internal error: stack underflow trying to pop event handler body")
		return
	}
	bodySteps, ok := bodyVal.([]Step)
	if !ok {
		l.addError(ctx, "internal error: expected []Step for event body, got %T", bodyVal)
		return
	}

	// Pop the handler name if it exists
	var handlerName string
	if ctx.KW_NAMED() != nil {
		nameVal, ok := l.popValue()
		if !ok {
			l.addError(ctx, "internal error: stack underflow trying to pop handler name")
			return
		}
		nameNode, ok := nameVal.(*StringLiteralNode)
		if !ok {
			l.addError(ctx, "internal error: expected StringLiteralNode for handler name, got %T", nameVal)
			return
		}
		handlerName = nameNode.Value
	}

	// Pop the event name expression
	eventExprVal, ok := l.popValue()
	if !ok {
		l.addError(ctx, "internal error: stack underflow trying to pop event name expression")
		return
	}
	eventExpr, ok := eventExprVal.(Expression)
	if !ok {
		l.addError(ctx, "internal error: expected Expression for event name, got %T", eventExprVal)
		return
	}

	// Pop the sentinel marker
	marker, ok := l.popValue()
	if !ok || marker != onEventMarker {
		l.addError(ctx, "internal error: stack corruption, missing on_stmt marker")
	}

	parentCtx, ok := ctx.GetParent().(*gen.On_stmtContext)
	if !ok {
		l.addError(ctx, "internal error: event_handler parent is not on_stmt")
		return
	}

	decl := &OnEventDecl{
		Pos:           tokenToPosition(parentCtx.KW_ON().GetSymbol()),
		EventNameExpr: eventExpr,
		Body:          bodySteps,
		HandlerName:   handlerName,
	}
	if id := ctx.IDENTIFIER(); id != nil {
		decl.EventVarName = id.GetText()
	}

	l.events = append(l.events, decl)
	l.logDebugAST("     Added on event declaration (event: %q, name: %q)", eventExpr.String(), handlerName)
}
