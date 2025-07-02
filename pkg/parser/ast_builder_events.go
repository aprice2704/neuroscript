// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Corrected block handling in error handlers to prevent stack imbalance.
// filename: pkg/parser/ast_builder_events.go
// nlines: 51
// risk_rating: HIGH

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

func (l *neuroScriptListenerImpl) EnterEvent_handler(c *gen.Event_handlerContext) {
	l.logDebugAST(">>> EnterEvent_handler")
	// The non_empty_statement_list child will manage its own block context.
}

func (l *neuroScriptListenerImpl) ExitEvent_handler(c *gen.Event_handlerContext) {
	l.logDebugAST("<<< ExitEvent_handler")

	// Pop the event handler's body from the value stack.
	bodyVal, ok := l.pop()
	if !ok {
		l.addError(c, "stack underflow in event_handler: could not pop body")
		return
	}
	body, ok := bodyVal.([]ast.Step)
	if !ok {
		l.addError(c, "internal error in event_handler: body is not a []ast.Step, but %T", bodyVal)
		return
	}

	// Pop the event name expression.
	eventNameVal, ok := l.pop()
	if !ok {
		l.addError(c, "stack underflow in event_handler: could not pop event name")
		return
	}
	eventName, ok := eventNameVal.(ast.Expression)
	if !ok {
		l.addError(c, "internal error in event_handler: event name is not an ast.Expression, but %T", eventNameVal)
		return
	}

	pos := tokenToPosition(c.GetStart())
	// Create the OnEventDecl node.
	onEvent := &ast.OnEventDecl{
		Pos:		&pos,
		EventNameExpr:	eventName,
		Body:		body,
	}

	if c.STRING_LIT() != nil {
		onEvent.HandlerName, _ = unescapeString(c.STRING_LIT().GetText())
	}
	if c.IDENTIFIER() != nil {
		onEvent.EventVarName = c.IDENTIFIER().GetText()
	}

	l.program.Events = append(l.program.Events, onEvent)
}

func (l *neuroScriptListenerImpl) EnterError_handler(c *gen.Error_handlerContext) {
	l.logDebugAST(">>> EnterError_handler")
	// The non_empty_statement_list child will manage its own block context.
}

func (l *neuroScriptListenerImpl) ExitError_handler(c *gen.Error_handlerContext) {
	l.logDebugAST("<<< ExitError_handler")

	// Pop the error handler's body, which was pushed by ExitNon_empty_statement_list.
	bodyVal, ok := l.pop()
	if !ok {
		l.addError(c, "stack underflow in error_handler: could not pop body")
		return
	}
	body, ok := bodyVal.([]ast.Step)
	if !ok {
		l.addError(c, "internal error in error_handler: body is not a []ast.Step, but %T", bodyVal)
		return
	}

	// Add the 'on_error' step to the parent context (proc or command).
	pos := tokenToPosition(c.GetStart())
	l.addStep(ast.Step{
		Position:	pos,
		Type:		"on_error",
		Body:		body,
	})
}