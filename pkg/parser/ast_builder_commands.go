// filename: pkg/parser/ast_builder_commands.go
// NeuroScript Version: 0.6.0
// File version: 12
// Purpose: Removed obsolete blank line counting logic. Association is now handled by the LineInfo algorithm.

package parser

import (
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (l *neuroScriptListenerImpl) EnterCommand_block(c *gen.Command_blockContext) {
	l.logDebugAST(">>> EnterCommand_block")
	token := c.GetStart()
	cmdNode := &ast.CommandNode{
		Metadata:      make(map[string]string),
		Body:          make([]ast.Step, 0),
		ErrorHandlers: make([]*ast.Step, 0),
		// BlankLinesBefore is now set by the LineInfo algorithm in the builder.
	}
	l.currentCommand = newNode(cmdNode, token, types.KindCommandBlock)
	l.currentCommand.Comments = l.associateCommentsToNode(l.currentCommand)
}

func (l *neuroScriptListenerImpl) ExitCommand_block(c *gen.Command_blockContext) {
	l.logDebugAST("<<< ExitCommand_block")
	if l.currentCommand == nil {
		l.addError(c, "Exiting command block but no current command context.")
		return
	}

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

	var regularSteps []ast.Step
	for i := range bodySteps {
		step := bodySteps[i]
		if step.Type == "on_error" {
			l.currentCommand.ErrorHandlers = append(l.currentCommand.ErrorHandlers, &step)
		} else {
			regularSteps = append(regularSteps, step)
		}
	}
	l.currentCommand.Body = regularSteps

	SetEndPos(l.currentCommand, c.KW_ENDCOMMAND().GetSymbol())
	l.program.Commands = append(l.program.Commands, l.currentCommand)
	l.currentCommand = nil
}
