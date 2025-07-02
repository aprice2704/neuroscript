// filename: pkg/parser/ast_builder_loops.go
package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
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

	l.addStep(ast.Step{
		Position:	tokenToPosition(ctx.GetStart()),
		Type:		"while",
		Cond:		cond,
		Body:		body,
	})
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

	l.addStep(ast.Step{
		Position:	tokenToPosition(ctx.GetStart()),
		Type:		"for",
		LoopVarName:	ctx.IDENTIFIER().GetText(),
		Collection:	collection,
		Body:		body,
	})
}