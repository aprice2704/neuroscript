// filename: pkg/core/ast_builder_events.go
// ast_builder_events.go – Listener callbacks for `on event` and `on error`
// file version: 22
// Purpose: Made ExitError_handler context-aware to handle command blocks correctly.

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// onEventStackMarker is a unique lang.Value pushed to the stack to identify 'on' blocks.
type onEventStackMarker struct{}

// EnterOn_stmt is called when entering an 'on' statement. It pushes a sentinel
// lang.Value to the stack for the exit handler to find, ensuring stack integrity.
func (l *neuroScriptListenerImpl) EnterOn_stmt(c *gen.On_stmtContext) {
	l.logDebugAST(">>> EnterOn_stmt, pushing marker")
	l.pushlang.Value(&onEventStackMarker{})
}

// ExitOn_stmt consumes the handler step (error or event) from the lang.Value stack,
// verifies and consumes the stack marker, and then appends the handler to the current step list.
func (l *neuroScriptListenerImpl) ExitOn_stmt(c *gen.On_stmtContext) {
	l.logDebugAST("<<< ExitOn_stmt")

	rawHandler, ok := l.poplang.Value()
	if !ok {
		l.addError(c, "stack underflow: could not pop handler for on_stmt")
		return
	}

	// Pop the marker that was pushed by EnterOn_stmt.
	marker, ok := l.poplang.Value()
	if !ok {
		l.addError(c, "stack underflow: could not pop on_stmt marker")
		return
	}
	if _, isMarker := marker.(*onEventStackMarker); !isMarker {
		l.addError(c, "internal error: stack corruption, missing on_stmt marker. Got %T instead.", marker)
		return
	}

	// Now process the handler we popped.
	if decl, isDecl := rawHandler.(*OnEventDecl); isDecl {
		// It's a global event handler declaration.
		l.events = append(l.events, decl)
		l.logDebugAST("     Added global 'on event' declaration (event: %q, name: %q)", decl.EventNameExpr.String(), decl.HandlerName)
	} else if step, isStep := rawHandler.(Step); isStep {
		// It's an in-procedure handler (on_error). Append it to the current block.
		*l.currentSteps = append(*l.currentSteps, step)
	} else {
		l.addError(c, "internal ast error: expected on_stmt handler to be ast.Step or OnEventDecl, got %T", rawHandler)
	}
}

// --- Error Handler ---

func (l *neuroScriptListenerImpl) EnterError_handler(c *gen.Error_handlerContext) {
	// Intentionally does nothing. Block context is handled by EnterNon_empty_statement_list.
}

// ExitError_handler builds the on_error step. It is now context-aware.
func (l *neuroScriptListenerImpl) ExitError_handler(c *gen.Error_handlerContext) {
	l.logDebugAST("--- ExitError_handler")

	rawBody, ok := l.poplang.Value()
	if !ok {
		l.addError(c, "stack underflow: could not pop body for on_error handler")
		return
	}
	body, ok := rawBody.([]Step)
	if !ok {
		l.addError(c, "internal ast error: on_error body is not []Step, got %T", rawBody)
		return
	}

	if len(body) == 0 {
		l.addError(c, "internal ast error: 'on error' block is empty despite grammar rule")
	}

	stmt := ast.Step{
		Position: tokenTolang.Position(c.GetStart()),
		Type:     "on_error",
		Body:     body,
	}

	// MODIFIED: Check if the parent is the command-specific 'on_error_only_stmt'.
	// This rule only appears inside command blocks and doesn't have the marker logic.
	if _, isCommandContext := c.GetParent().(*gen.On_error_only_stmtContext); isCommandContext {
		// In a command block, this is just a regular statement. Append it directly.
		*l.currentSteps = append(*l.currentSteps, stmt)
		l.logDebugAST("<<< Appended on_error step directly to command body")
	} else {
		// In a global or procedure context, push to the lang.Value stack for ExitOn_stmt to handle.
		l.pushlang.Value(stmt)
		l.logDebugAST("<<< Pushed generic on_error step to lang.Value stack")
	}
}

// --- Event Handler ---

func (l *neuroScriptListenerImpl) EnterEvent_handler(ctx *gen.Event_handlerContext) {
	// Intentionally does nothing. Block context is handled by EnterNon_empty_statement_list.
}

// ExitEvent_handler builds the 'on event' declaration and pushes it to the stack.
func (l *neuroScriptListenerImpl) ExitEvent_handler(ctx *gen.Event_handlerContext) {
	l.logDebugAST("<<< ExitEvent_handler")

	rawBody, ok := l.poplang.Value()
	if !ok {
		l.addError(ctx, "stack underflow: could not pop event handler body")
		return
	}
	body, ok := rawBody.([]Step)
	if !ok {
		l.addError(ctx, "internal error: expected []Step for event body, got %T", rawBody)
		return
	}

	var handlerName string
	if ctx.KW_NAMED() != nil {
		nameVal, _ := l.poplang.Value()
		if nameNode, ok := nameVal.(*ast.StringLiteralNode); ok {
			handlerName = nameNode.Value
		}
	}

	rawExpr, ok := l.poplang.Value()
	if !ok {
		l.addError(ctx, "internal error: stack underflow trying to pop event name expression")
		return
	}
	eventExpr, ok := rawExpr.(ast.Expression)
	if !ok {
		l.addError(ctx, "internal error: expected ast.Expression for event name, got %T", rawExpr)
		return
	}

	decl := &OnEventDecl{
		Position:      tokenTolang.Position(ctx.GetStart()),
		EventNameExpr: eventExpr,
		Body:          body,
		HandlerName:   handlerName,
	}
	if id := ctx.IDENTIFIER(); id != nil {
		decl.EventVarName = id.GetText()
	}

	l.pushlang.Value(decl)
}
