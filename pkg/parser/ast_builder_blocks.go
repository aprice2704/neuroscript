// filename: pkg/parser/ast_builder_blocks.go
// NeuroScript Version: 0.5.2
// File version: 15
// Purpose: Corrected field and method names (e.g., ValueStack, blockStepStack, pushValue) to match the listener implementation.

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// blockContext is a helper struct to manage nested lists of steps during AST construction.
type blockContext struct {
	parentSteps *[]ast.Step
}

// enterBlockContext sets up a new []ast.Step slice for a new block.
func (l *neuroScriptListenerImpl) enterBlockContext(kind string) {
	l.logDebugAST(">>> Enter %s Block (valDepth: %d)", kind, len(l.ValueStack))
	l.blockStepStack = append(l.blockStepStack, &blockContext{
		parentSteps: l.currentSteps,
	})
	fresh := make([]ast.Step, 0)
	l.currentSteps = &fresh
}

// exitBlockContext finalizes the current block's step collection.
func (l *neuroScriptListenerImpl) exitBlockContext(kind string) {
	if l.currentSteps == nil {
		l.logger.Error("AST Builder FATAL: exitBlockContext called with nil currentSteps", "kind", kind)
		l.push([]ast.Step{}) // Corrected from pushValue
		return
	}

	completedChildSteps := *l.currentSteps
	l.logDebugAST("<<< Exit %s Block (items: %d, valDepth: %d)", kind, len(completedChildSteps), len(l.ValueStack))
	l.push(completedChildSteps) // Corrected from pushValue

	if len(l.blockStepStack) > 0 {
		// Restore the parent's step collector
		parentContext := l.blockStepStack[len(l.blockStepStack)-1]
		l.currentSteps = parentContext.parentSteps
		l.blockStepStack = l.blockStepStack[:len(l.blockStepStack)-1]
	} else {
		l.currentSteps = nil
	}
}

// --- Statement_list listener wiring ---

func (l *neuroScriptListenerImpl) EnterStatement_list(ctx *gen.Statement_listContext) {
	// This rule is now deprecated in favor of non_empty_statement_list, but we keep the wiring for safety.
}

func (l *neuroScriptListenerImpl) ExitStatement_list(ctx *gen.Statement_listContext) {
	// This rule is now deprecated in favor of non_empty_statement_list, but we keep the wiring for safety.
}

func (l *neuroScriptListenerImpl) EnterNon_empty_statement_list(ctx *gen.Non_empty_statement_listContext) {
	var kind string
	switch ctx.GetParent().(type) {
	case *gen.Procedure_definitionContext:
		kind = "PROC_BODY"
	case *gen.If_statementContext:
		kind = "IF_ELSE_BODY"
	case *gen.For_each_statementContext:
		kind = "FOR_EACH_BODY"
	case *gen.While_statementContext:
		kind = "WHILE_BODY"
	case *gen.Event_handlerContext:
		kind = "ON_EVENT_BODY"
	case *gen.Error_handlerContext:
		kind = "ON_ERROR_BODY"
	case *gen.Command_blockContext: // Added for command block support
		kind = "COMMAND_BODY"
	default:
		kind = "UNKNOWN_BLOCK"
		l.addError(ctx, "non_empty_statement_list found inside unknown parent type: %T", ctx.GetParent())
	}
	l.enterBlockContext(kind)
}

func (l *neuroScriptListenerImpl) ExitNon_empty_statement_list(ctx *gen.Non_empty_statement_listContext) {
	var kind string
	switch ctx.GetParent().(type) {
	case *gen.Procedure_definitionContext:
		kind = "PROC_BODY"
	case *gen.If_statementContext:
		kind = "IF_ELSE_BODY"
	case *gen.For_each_statementContext:
		kind = "FOR_EACH_BODY"
	case *gen.While_statementContext:
		kind = "WHILE_BODY"
	case *gen.Event_handlerContext:
		kind = "ON_EVENT_BODY"
	case *gen.Error_handlerContext:
		kind = "ON_ERROR_BODY"
	case *gen.Command_blockContext: // Added for command block support
		kind = "COMMAND_BODY"
	default:
		kind = "UNKNOWN_BLOCK"
	}
	l.exitBlockContext(kind)
}

// --- ADDED: Wiring for the command_statement_list used by command blocks ---

func (l *neuroScriptListenerImpl) EnterCommand_statement_list(ctx *gen.Command_statement_listContext) {
	l.enterBlockContext("COMMAND_BODY")
}

func (l *neuroScriptListenerImpl) ExitCommand_statement_list(ctx *gen.Command_statement_listContext) {
	l.exitBlockContext("COMMAND_BODY")
}
