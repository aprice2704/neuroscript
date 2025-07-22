// filename: pkg/parser/ast_builder_events.go
// NeuroScript Version: 0.6.0
// File version: 13
// Purpose: Removed obsolete blank line counting logic. Association is now handled by the LineInfo algorithm.

package parser

import (
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (l *neuroScriptListenerImpl) EnterEvent_handler(c *gen.Event_handlerContext) {
	l.logDebugAST(">>> EnterEvent_handler")
	onEvent := &ast.OnEventDecl{
		Metadata: make(map[string]string),
		Comments: make([]*ast.Comment, 0),
		// BlankLinesBefore is now set by the LineInfo algorithm in the builder.
	}
	newNode(onEvent, c.GetStart(), types.KindOnEventDecl)
	l.push(onEvent)
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
	if !ok {
		l.addError(c, "internal error in event_handler: stack node is not an *ast.OnEventDecl, but %T", nodeVal)
		return
	}

	onEvent.EventNameExpr = eventName
	onEvent.Body = body
	SetEndPos(onEvent, c.KW_ENDON().GetSymbol())

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
