// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Implemented statement list handlers to correctly manage block contexts and the value stack.
// filename: pkg/parser/ast_builder_blocks.go
// nlines: 65
// risk_rating: HIGH

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// blockContext is a helper struct to manage nested lists of steps during AST construction.
type blockContext struct {
	steps []ast.Step
}

// enterBlock sets up a new context for a new block of statements.
func (l *neuroScriptListenerImpl) enterBlock() {
	l.logDebugAST(">>> enterBlock (new context)")
	newCtx := &blockContext{
		steps: make([]ast.Step, 0),
	}
	l.blockStack = append(l.blockStack, newCtx)
}

// exitBlock finalizes the current block's step collection and pushes it to the value stack.
func (l *neuroScriptListenerImpl) exitBlock() {
	if len(l.blockStack) == 0 {
		l.logger.Error("AST Builder FATAL: exitBlock called with empty block stack")
		return
	}

	// Pop the current context.
	lastIndex := len(l.blockStack) - 1
	currentCtx := l.blockStack[lastIndex]
	l.blockStack = l.blockStack[:lastIndex]

	l.logDebugAST("<<< exitBlock (pushing %d steps to value stack)", len(currentCtx.steps))
	// Push the completed slice of steps onto the main value stack for the parent rule to consume.
	l.push(currentCtx.steps)
}

// --- Statement List Handlers ---

// EnterNon_empty_statement_list is called when entering a block that must contain statements.
func (l *neuroScriptListenerImpl) EnterNon_empty_statement_list(c *gen.Non_empty_statement_listContext) {
	l.enterBlock()
}

// ExitNon_empty_statement_list is called when exiting the block.
func (l *neuroScriptListenerImpl) ExitNon_empty_statement_list(c *gen.Non_empty_statement_listContext) {
	l.exitBlock()
}

// EnterCommand_statement_list is the equivalent for command blocks.
func (l *neuroScriptListenerImpl) EnterCommand_statement_list(c *gen.Command_statement_listContext) {
	l.enterBlock()
}

// ExitCommand_statement_list is called when exiting a command block's statements.
func (l *neuroScriptListenerImpl) ExitCommand_statement_list(c *gen.Command_statement_listContext) {
	l.exitBlock()
}