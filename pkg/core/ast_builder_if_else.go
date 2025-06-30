// filename: pkg/core/ast_builder_if_else.go
// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Implements the listener method for building 'if-then-else' statements, relying on the new block context handlers.
// nlines: 51
// risk_rating: HIGH

package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// ExitIf_statement is called when exiting an if_statement rule in the grammar.
// It assembles the condition, then-block, and optional else-block into a single Step.
// It relies on ExitStatement_list (via exitBlockContext) having already pushed the []Step blocks
// for 'then' and 'else' onto the value stack.
func (l *neuroScriptListenerImpl) ExitIf_statement(c *gen.If_statementContext) {
	l.logDebugAST("<<< ExitIf_statement")

	var elseSteps []Step

	// The 'else' block is optional. If it exists, its []Step slice is on top of the stack.
	if c.KW_ELSE() != nil {
		rawElse, ok := l.popValue()
		if !ok {
			l.addError(c, "stack underflow: could not pop else block for if statement")
			return
		}
		elseSteps, ok = rawElse.([]Step)
		if !ok {
			l.addError(c, "internal ast error: expected else block to be []Step, but got %T", rawElse)
			return
		}
	}

	// The 'then' block's []Step slice was pushed before the else block.
	rawThen, ok := l.popValue()
	if !ok {
		l.addError(c, "stack underflow: could not pop then block for if statement")
		return
	}
	thenSteps, ok := rawThen.([]Step)
	if !ok {
		l.addError(c, "internal ast error: expected then block to be []Step, but got %T", rawThen)
		return
	}

	// The condition expression was pushed onto the stack first.
	rawCond, ok := l.popValue()
	if !ok {
		l.addError(c, "stack underflow: could not pop condition for if statement")
		return
	}
	cond, ok := rawCond.(Expression)
	if !ok {
		l.addError(c, "internal ast error: expected condition to be Expression, but got %T", rawCond)
		return
	}

	stmt := Step{
		Pos:  tokenToPosition(c.GetStart()),
		Type: "if",
		Cond: cond,
		Body: thenSteps,
		Else: elseSteps,
	}

	*l.currentSteps = append(*l.currentSteps, stmt)
}
