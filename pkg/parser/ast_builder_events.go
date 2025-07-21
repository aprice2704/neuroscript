// filename: pkg/parser/ast_builder_events.go
// NeuroScript Version: 0.6.0
// File version: 5
// Purpose: Sets the end position of event and error handler nodes using the StopPos field.

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (l *neuroScriptListenerImpl) EnterEvent_handler(c *gen.Event_handlerContext) {
	l.logDebugAST(">>> EnterEvent_handler")
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

	onEvent := &ast.OnEventDecl{
		EventNameExpr: eventName,
		Body:          body,
	}

	if c.STRING_LIT() != nil {
		onEvent.HandlerName, _ = unescapeString(c.STRING_LIT().GetText())
	}
	if c.IDENTIFIER() != nil {
		onEvent.EventVarName = c.IDENTIFIER().GetText()
	}

	newNode(onEvent, c.GetStart(), types.KindOnEventDecl)
	SetEndPos(onEvent, c.KW_ENDON().GetSymbol())

	l.program.Events = append(l.program.Events, onEvent)
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

	pos := tokenToPosition(c.GetStart())
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "on_error",
		Body:     body,
	}
	SetEndPos(&step, c.KW_ENDON().GetSymbol())
	l.addStep(step)
}
