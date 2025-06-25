// ast_builder_events.go – Listener callbacks for `on event` and `on error`
// file version: 18
// Purpose: Replaced logging with fmt.Printf for forced debugging to identify AST node type.

package core

import (
	"fmt" // Ensure fmt is imported

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

	// Pop all potential values from the stack first
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

	var handlerName string
	if ctx.KW_NAMED() != nil {
		nameVal, _ := l.popValue()
		if nameNode, ok := nameVal.(*StringLiteralNode); ok {
			handlerName = nameNode.Value
		}
	}

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

	// --- FORCED DEBUGGING VIA PRINTF ---
	fmt.Printf("\n>>>> DEBUG | Event Expr Type: [%T] | Value: [%s] <<<<\n\n", eventExpr, eventExpr.String())
	// --- END DEBUGGING ---

	isErrorHandler := false
	if str, ok := eventExpr.(*StringLiteralNode); ok && str.Value == "error" {
		isErrorHandler = true
	}

	if isErrorHandler && l.currentProc != nil {
		l.logDebugAST("     SUCCESS: Identified as function-scoped 'on error' handler for procedure '%s'", l.currentProc.Name)
		marker, _ := l.popValue()
		if marker != onEventMarker {
			l.addError(ctx, "internal error: stack corruption, missing on_stmt marker for error handler")
		}
		handlerStep := &Step{
			Pos:  tokenToPosition(ctx.GetStart()),
			Type: "on_error_handler",
			Body: bodySteps,
		}
		l.currentProc.ErrorHandlers = append(l.currentProc.ErrorHandlers, handlerStep)
		return
	}

	// --- FORCED DEBUGGING VIA PRINTF ---
	if l.currentProc != nil && !isErrorHandler {
		fmt.Printf("\n>>>> DEBUG | Handler for event '%s' not identified as 'on error', treating as global <<<<\n\n", eventExpr.String())
	}
	// --- END DEBUGGING ---

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
	l.logDebugAST("     Added global 'on event' declaration (event: %q, name: %q)", eventExpr.String(), handlerName)
}
