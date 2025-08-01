// filename: pkg/parser/ast_builder_loops.go
// NeuroScript Version: 0.6.0
// File version: 8
// Purpose: Removed obsolete blank line counting logic. Association is now handled by the LineInfo algorithm.

package parser

import (
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// --- Loop Statements ---

func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST(">>> EnterWhile_statement")
	l.loopDepth++
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST("<<< ExitWhile_statement")
	l.loopDepth--

	bodyVal, ok := l.pop()
	if !ok {
		l.addError(ctx, "stack underflow in while statement body")
		return
	}
	body, ok := bodyVal.([]ast.Step)
	if !ok {
		l.addError(ctx, "while statement expected a valid body on the stack, got %T", bodyVal)
		return
	}

	condVal, ok := l.pop()
	if !ok {
		l.addError(ctx, "stack underflow in while statement condition")
		return
	}
	cond, ok := condVal.(ast.Expression)
	if !ok {
		l.addError(ctx, "while statement expected a valid condition on the stack, got %T", condVal)
		return
	}

	pos := tokenToPosition(ctx.GetStart())
	step := ast.Step{
		BaseNode: ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:     "while",
		Cond:     cond,
		Body:     body,
		// BlankLinesBefore is now set by the LineInfo algorithm in the builder.
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, ctx.KW_ENDWHILE().GetSymbol())
	l.addStep(step)
}

func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST(">>> EnterFor_each_statement")
	l.loopDepth++
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST("<<< ExitFor_each_statement")
	l.loopDepth--

	bodyVal, ok := l.pop()
	if !ok {
		l.addError(ctx, "stack underflow in for statement body")
		return
	}
	body, ok := bodyVal.([]ast.Step)
	if !ok {
		l.addError(ctx, "for statement expected a valid body on the stack, got %T", bodyVal)
		return
	}

	collectionVal, ok := l.pop()
	if !ok {
		l.addError(ctx, "stack underflow in for statement collection")
		return
	}
	collection, ok := collectionVal.(ast.Expression)
	if !ok {
		l.addError(ctx, "for statement expected a valid collection on the stack, got %T", collectionVal)
		return
	}

	pos := tokenToPosition(ctx.GetStart())
	step := ast.Step{
		BaseNode:    ast.BaseNode{StartPos: &pos, NodeKind: types.KindStep},
		Type:        "for",
		LoopVarName: ctx.IDENTIFIER().GetText(),
		Collection:  collection,
		Body:        body,
		// BlankLinesBefore is now set by the LineInfo algorithm in the builder.
	}
	step.Comments = l.associateCommentsToNode(&step)
	SetEndPos(&step, ctx.KW_ENDFOR().GetSymbol())
	l.addStep(step)
}
