// ast_builder_events.go – Listener callbacks for `on event` declarations
// file version: 10
//
// Design
// ------
// *Enter*: `EnterOnEventStmt` pushes a unique sentinel (`onEventMarker`) onto
// `valueStack`. The rest of the handler body is managed by the generic
// `Statement_list` logic, which pushes the collected body `[]Step` when it
// finishes.
//
// *Exit*: `ExitOnEventStmt` pops items until it reaches the sentinel,
// collecting at most *one* body slice and keeping the **last** `Expression`
// encountered as the event name. Any extra artefacts (e.g. alias identifiers)
// are discarded.  After assembling the `OnEventDecl`, nothing is pushed back on
// either stack, guaranteeing perfect balance regardless of success or early
// error return.
//
// Validation responsibility: The builder does *not* reject non‑literal event
// names here; that happens later during the semantic‑validation phase.  The
// builder’s job is strictly to produce a well‑formed AST and maintain stack
// invariants.
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// onEventStackMarker is a unique sentinel pushed at the start of every handler.
type onEventStackMarker struct{}

var onEventMarker = &onEventStackMarker{}

// --- Listener Methods for 'on event' statements ---

// EnterOnEventStmt pushes the sentinel so ExitOnEventStmt knows the boundary of
// this handler’s stack items.
func (l *neuroScriptListenerImpl) EnterOnEventStmt(ctx *gen.OnEventStmtContext) {
	l.logDebugAST(">>> EnterOnEventStmt (push marker)")
	l.pushValue(onEventMarker)
}

// ExitOnEventStmt assembles the handler node and ensures the stack is restored
// to its pre‑handler depth.
func (l *neuroScriptListenerImpl) ExitOnEventStmt(ctx *gen.OnEventStmtContext) {
	l.logDebugAST("<<< ExitOnEventStmt (pop to marker)")

	var (
		bodySteps []Step
		eventExpr Expression
	)

	// Pop until we hit the sentinel.
	for {
		v, ok := l.popValue()
		if !ok {
			l.addError(ctx, "internal error: value stack underflow in ExitOnEventStmt")
			return
		}
		switch tv := v.(type) {
		case *onEventStackMarker:
			goto donePop // sentinel reached
		case []Step:
			if bodySteps != nil {
				l.addError(ctx, "internal error: second body slice on stack for on event handler")
				// continue to sentinel to clean up
			} else {
				bodySteps = tv
			}
		case Expression:
			// Keep the MOST RECENT expression before sentinel as the event name.
			eventExpr = tv
		default:
			// Discard any other artefacts (identifiers, error nodes, etc.)
		}
	}

donePop:
	// Sanity checks
	if bodySteps == nil {
		l.addError(ctx, "internal error: missing body steps for on event handler")
		return
	}
	if eventExpr == nil {
		l.addError(ctx, "internal error: missing event name expression for on event handler")
		return
	}

	// --- Validation -----------------------------------------------
	// Require the event name to be a string literal; anything else is a
	// syntax/semantic error recorded on the builder.
	switch ev := eventExpr.(type) {
	case *StringLiteralNode:
		// OK – static name
	case *VariableNode:
		// Use the first token of the expression so the error is underlined
		l.addErrorf(
			ctx.Expression().GetStart(), // antlr.Token
			"Event name must be a static string literal, not a variable (%s)",
			ev.Name, // or ev.Name
		)
	default:
		l.addError(
			ctx,
			"Event name must be a static string literal (got %T)",
			ev,
		)
	}

	// Create the correct OnEventDecl node, which the interpreter expects.
	decl := &OnEventDecl{
		Pos:           tokenToPosition(ctx.KW_ON().GetSymbol()),
		EventNameExpr: eventExpr,
		Body:          bodySteps,
	}
	if id := ctx.IDENTIFIER(); id != nil {
		decl.EventVarName = id.GetText()
	}

	// Correctly append to the listener's temporary 'events' slice.
	l.events = append(l.events, decl)
	l.logDebugAST("     Added on event declaration (event expr %T)", eventExpr)
}
