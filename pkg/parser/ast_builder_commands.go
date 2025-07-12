// filename: pkg/parser/ast_builder_commands.go
// NeuroScript Version: 0.5.2
// File version: 6
// Purpose: Refactored command block creation to use newNode and BaseNode.

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

func (l *neuroScriptListenerImpl) EnterCommand_block(c *gen.Command_blockContext) {
	l.logDebugAST(">>> EnterCommand_block")
	token := c.GetStart()
	cmdNode := &ast.CommandNode{
		Metadata:      make(map[string]string),
		Body:          make([]ast.Step, 0),
		ErrorHandlers: make([]*ast.Step, 0),
	}
	l.currentCommand = newNode(cmdNode, token, ast.KindCommandBlock)
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

	l.program.Commands = append(l.program.Commands, l.currentCommand)
	l.currentCommand = nil
}
