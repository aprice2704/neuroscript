// NeuroScript Version: 0.5.2
// File version: 5
// Purpose: Removed redundant block context creation to fix stack imbalance.
// filename: pkg/parser/ast_builder_commands.go
// nlines: 41
// risk_rating: MEDIUM

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

func (l *neuroScriptListenerImpl) EnterCommand_block(c *gen.Command_blockContext) {
	l.logDebugAST(">>> EnterCommand_block")
	pos := tokenToPosition(c.GetStart())
	l.currentCommand = &ast.CommandNode{
		Pos:           &pos,
		Metadata:      make(map[string]string),
		Body:          make([]ast.Step, 0),
		ErrorHandlers: make([]*ast.Step, 0),
	}
	// DO NOT create a block here. The command_statement_list rule handles the block context.
}

func (l *neuroScriptListenerImpl) ExitCommand_block(c *gen.Command_blockContext) {
	l.logDebugAST("<<< ExitCommand_block")
	if l.currentCommand == nil {
		l.addError(c, "Exiting command block but no current command context.")
		return
	}

	// The command body's []ast.Step was placed on the value stack by ExitCommand_statement_list.
	rawBody, ok := l.pop()
	if !ok {
		l.addError(c, "stack underflow: could not pop command block body")
		return
	}
	bodySteps, ok := rawBody.([]ast.Step)
	if !ok {
		l.addError(c, "internal ast error: command block body is not []ast.Step, got %T", rawBody)
		return
	}

	// Separate 'on error' handlers from the main body of steps.
	var regularSteps []ast.Step
	for i := range bodySteps {
		step := bodySteps[i] // Create a copy of the step
		if step.Type == "on_error" {
			l.currentCommand.ErrorHandlers = append(l.currentCommand.ErrorHandlers, &step)
		} else {
			regularSteps = append(regularSteps, step)
		}
	}
	l.currentCommand.Body = regularSteps

	// Add the completed command to the program's list.
	l.program.Commands = append(l.program.Commands, l.currentCommand)
	l.currentCommand = nil
}
