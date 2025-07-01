// filename: pkg/parser/ast_builder_commands.go
// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Corrected command block listener to integrate with central metadata and block handlers.

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// EnterCommand_block now only sets the currentCommand context.
// Metadata is handled by ExitMetadata_block and the body by the statement list listeners.
func (l *neuroScriptListenerImpl) EnterCommand_block(c *gen.Command_blockContext) {
	l.logDebugAST(">>> EnterCommand_block")
	l.currentCommand = &ast.CommandNode{
		Pos:           tokenTolang.Position(c.GetStart()),
		Metadata:      make(map[string]string),
		Body:          make([]ast.Step, 0),
		ErrorHandlers: make([]*ast.Step, 0),
	}
}

// ExitCommand_block finalizes the command node, retrieving the body from the value stack.
func (l *neuroScriptListenerImpl) ExitCommand_block(c *gen.Command_blockContext) {
	l.logDebugAST("<<< ExitCommand_block")
	if l.currentCommand == nil {
		l.addError(c, "Exiting command block but no current command context.")
		return
	}

	// The command body's []ast.Step was placed on the value stack by exitBlockContext.
	rawBody, ok := l.poplang.Value()
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
		step := &bodySteps[i]
		if step.Type == "on_error" {
			l.currentCommand.ErrorHandlers = append(l.currentCommand.ErrorHandlers, step)
		} else {
			regularSteps = append(regularSteps, *step)
		}
	}
	l.currentCommand.Body = regularSteps

	if len(l.currentCommand.Body) == 0 && len(l.currentCommand.ErrorHandlers) == 0 {
		l.addError(c, "command block cannot be empty")
	}

	l.commands = append(l.commands, l.currentCommand)
	l.currentCommand = nil
}
