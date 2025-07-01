// filename: pkg/parser/ast_builder_if_else.go
// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Implements the listener method for building 'if-then-else' statements, relying on the new block context handlers.
// nlines: 51
// risk_rating: HIGH

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// ExitIf_statement is called when exiting an if_statement rule in the grammar.
// It assembles the condition, then-block, and optional else-block into a single ast.Step.
// It relies on ExitStatement_list (via exitBlockContext) having already pushed the []ast.Step blocks
// for 'then' and 'else' onto the value stack.
func (l *neuroScriptListenerImpl) ExitIf_statement(c *gen.If_statementContext) {
	l.logDebugAST("<<< ExitIf_statement")

	var elseSteps []ast.Step

	// The 'else' block is optional. If it exists, its []ast.Step slice is on top of the stack.
	if c.KW_ELSE() != nil {
		rawElse, ok := l.poplang.Value()
		if !ok {
			l.addError(c, "stack underflow: could not pop else block for if statement")
			return
		}
		elseSteps, ok = rawElse.([]ast.Step)
		if !ok {
			l.addError(c, "internal ast error: expected else block to be []ast.Step, but got %T", rawElse)
			return
		}
	}

	// The 'then' block's []ast.Step slice was pushed before the else block.
	rawThen, ok := l.poplang.Value()
	if !ok {
		l.addError(c, "stack underflow: could not pop then block for if statement")
		return
	}
	thenSteps, ok := rawThen.([]ast.Step)
	if !ok {
		l.addError(c, "internal ast error: expected then block to be []ast.Step, but got %T", rawThen)
		return
	}

	// The condition expression was pushed onto the stack first.
	rawCond, ok := l.poplang.Value()
	if !ok {
		l.addError(c, "stack underflow: could not pop condition for if statement")
		return
	}
	cond, ok := rawCond.(ast.Expression)
	if !ok {
		l.addError(c, "internal ast error: expected condition to be ast.Expression, but got %T", rawCond)
		return
	}

	stmt := ast.Step{
		Position: tokenTolang.Position(c.GetStart()),
		Type:     "if",
		Cond:     cond,
		Body:     thenSteps,
		ElseBody: elseSteps,
	}

	*l.currentSteps = append(*l.currentSteps, stmt)
}
