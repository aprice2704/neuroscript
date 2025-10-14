// NeuroScript Version: 0.8.0
// File version: 28
// Purpose: Allows 'on error' handlers within command blocks and corrects the generated step type to 'on_error'.
// filename: pkg/parser/ast_builder_events.go
// nlines: 104
// risk_rating: HIGH

package parser

import (
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (l *neuroScriptListenerImpl) EnterEvent_handler(c *gen.Event_handlerContext) {
	l.logDebugAST(">>> EnterEvent_handler")
	token := c.GetStart()
	onEvent := &ast.OnEventDecl{
		Metadata: make(map[string]string),
		Comments: make([]*ast.Comment, 0),
	}
	l.assignPendingMetadata(token, onEvent.Metadata)

	newNode(onEvent, c.GetStart(), types.KindOnEventDecl)
	l.push(onEvent)
	l.currentEvent = onEvent
}

func (l *neuroScriptListenerImpl) ExitEvent_handler(c *gen.Event_handlerContext) {
	l.logDebugAST("<<< ExitEvent_handler")

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

	nodeVal, ok := l.pop()
	if !ok {
		l.addError(c, "stack underflow in event_handler: could not pop OnEventDecl node")
		return
	}
	onEvent, ok := nodeVal.(*ast.OnEventDecl)
	if !ok || onEvent == nil {
		l.addError(c, "internal error in event_handler: stack node is not a valid *ast.OnEventDecl, but %T", nodeVal)
		return
	}

	onEvent.EventNameExpr = eventName
	onEvent.Body = body
	SetEndPos(onEvent, c.GetStop())

	if c.STRING_LIT() != nil {
		onEvent.HandlerName, _ = unescapeString(c.STRING_LIT().GetText())
	}
	if c.IDENTIFIER() != nil {
		onEvent.EventVarName = c.IDENTIFIER().GetText()
	}

	l.program.Events = append(l.program.Events, onEvent)

	if l.eventHandlerCallback != nil {
		l.eventHandlerCallback(onEvent)
	}

	l.currentEvent = nil
}

func (l *neuroScriptListenerImpl) EnterError_handler(c *gen.Error_handlerContext) {
	l.logDebugAST(">>> EnterError_handler")
}

func (l *neuroScriptListenerImpl) ExitError_handler(c *gen.Error_handlerContext) {
	l.logDebugAST("<<< ExitError_handler")

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

	// Create a Step that represents the handler block.
	handlerStep := &ast.Step{
		// FIX: Correct the type to 'on_error' to match test expectations.
		Type: "on_error",
		Body: body,
	}
	newNode(handlerStep, c.GetStart(), types.KindStep)
	SetEndPos(handlerStep, c.GetStop())

	// FIX: Allow 'on error' handlers inside both procedures and command blocks.
	if l.currentProc != nil {
		l.currentProc.ErrorHandlers = append(l.currentProc.ErrorHandlers, handlerStep)
	} else if l.currentCommand != nil {
		l.currentCommand.ErrorHandlers = append(l.currentCommand.ErrorHandlers, handlerStep)
	} else {
		l.addError(c, "'on error' handler is only valid inside a func or command block")
	}
}
