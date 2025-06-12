// NeuroScript Version: 0.3.1
// File version: 1.2 // Corrected stack operations to expect []Step, not *[]Step.
// Purpose: Implements the AST builder logic for if/else statements.
// filename: pkg/core/ast_builder_if_else.go
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// ExitIf_statement is called when exiting an if_statement rule in the grammar.
// It assembles the condition, then-block, and optional else-block into a single Step,
// which is then added as a step to the current statement block.
func (l *neuroScriptListenerImpl) ExitIf_statement(c *gen.If_statementContext) {
	l.logDebugAST("<<< ExitIf_statement")

	var elseSteps []Step

	// The 'else' block is optional. If it exists, its statement list will have been
	// processed and the resulting []Step slice will be on top of the value stack.
	if c.KW_ELSE() != nil {
		elseVal := l.pop()
		// MINIMAL CHANGE: The type assertion now expects []Step, not *[]Step.
		if val, ok := elseVal.([]Step); ok {
			elseSteps = val
		} else {
			l.addError(c, "internal ast error: expected else block to be []Step, but got %T", elseVal)
			return // Cannot continue if the stack is corrupt
		}
	}

	// The 'then' block's statement list is always present and was pushed before the else block.
	thenVal := l.pop()
	// MINIMAL CHANGE: The type assertion now expects []Step, not *[]Step.
	thenSteps, ok := thenVal.([]Step)
	if !ok {
		l.addError(c, "internal ast error: expected then block to be []Step, but got %T", thenVal)
		return
	}

	// The condition expression was pushed onto the stack first.
	condVal := l.pop()
	cond, ok := condVal.(Expression)
	if !ok {
		l.addError(c, "internal ast error: expected condition to be Expression, but got %T", condVal)
		return
	}

	// Assemble the final Step node for the 'if' statement.
	ifStep := Step{
		Pos:  tokenToPosition(c.GetStart()),
		Type: "if",
		Cond: cond,
		Body: thenSteps,
		Else: elseSteps,
	}

	// Add the fully assembled if statement as a step to the current block.
	*l.currentSteps = append(*l.currentSteps, ifStep)
}
