// filename: pkg/core/ast_builder_loops.go
// version: 1.3.0
// purpose: Implements listener methods for loop constructs, using correct Step slice type.
package core

import (
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Loop Statements ---

func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST(">>> EnterWhile_statement")
	// Block context is now handled by EnterStatement_list
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST("<<< ExitWhile_statement")
	bodyVal := l.pop()
	body, ok := bodyVal.([]Step) // MINIMAL CHANGE: Expect []Step, not *[]Step.
	if !ok {
		l.addError(ctx, "while statement expected a valid body on the stack, got %T", bodyVal)
		return
	}

	cond, ok := l.pop().(Expression)
	if !ok {
		l.addError(ctx, "while statement expected a valid condition on the stack")
		return
	}

	stmt := Step{
		Pos:  tokenToPosition(ctx.GetStart()),
		Type: "while",
		Cond: cond,
		Body: body,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST(">>> EnterFor_each_statement")
	// Block context is now handled by EnterStatement_list
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST("<<< ExitFor_each_statement")
	bodyVal := l.pop()
	body, ok := bodyVal.([]Step) // MINIMAL CHANGE: Expect []Step, not *[]Step.
	if !ok {
		l.addError(ctx, "for..each statement expected a valid body on the stack, got %T", bodyVal)
		return
	}

	collection, ok := l.pop().(Expression)
	if !ok {
		l.addError(ctx, "for..each statement expected a valid collection on the stack")
		return
	}

	stmt := Step{
		Pos:         tokenToPosition(ctx.GetStart()),
		Type:        "for",
		LoopVarName: ctx.IDENTIFIER().GetText(),
		Collection:  collection,
		Body:        body,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// --- Other Required Stubs ---

// The following methods are required to fully implement the NeuroScriptListener interface.
// For now, they are simple stubs. We can add logic to them as needed.

func (l *neuroScriptListenerImpl) EnterSignature_part(c *gen.Signature_partContext) {}
func (l *neuroScriptListenerImpl) ExitSignature_part(c *gen.Signature_partContext)  {}

func (l *neuroScriptListenerImpl) EnterNeeds_clause(c *gen.Needs_clauseContext) {}
func (l *neuroScriptListenerImpl) ExitNeeds_clause(c *gen.Needs_clauseContext)  {}

func (l *neuroScriptListenerImpl) EnterOptional_clause(c *gen.Optional_clauseContext) {}
func (l *neuroScriptListenerImpl) ExitOptional_clause(c *gen.Optional_clauseContext)  {}

func (l *neuroScriptListenerImpl) EnterReturns_clause(c *gen.Returns_clauseContext) {}
func (l *neuroScriptListenerImpl) ExitReturns_clause(c *gen.Returns_clauseContext)  {}

func (l *neuroScriptListenerImpl) EnterParam_list(c *gen.Param_listContext) {}
func (l *neuroScriptListenerImpl) ExitParam_list(c *gen.Param_listContext)  {}

func (l *neuroScriptListenerImpl) EnterMetadata_block(c *gen.Metadata_blockContext) {}
func (l *neuroScriptListenerImpl) ExitMetadata_block(c *gen.Metadata_blockContext)  {}

func (l *neuroScriptListenerImpl) EnterBody_line(c *gen.Body_lineContext) {}
func (l *neuroScriptListenerImpl) ExitBody_line(c *gen.Body_lineContext)  {}

func (l *neuroScriptListenerImpl) EnterStatement(c *gen.StatementContext) {}
func (l *neuroScriptListenerImpl) ExitStatement(c *gen.StatementContext)  {}

func (l *neuroScriptListenerImpl) EnterSimple_statement(c *gen.Simple_statementContext) {}
func (l *neuroScriptListenerImpl) ExitSimple_statement(c *gen.Simple_statementContext)  {}

func (l *neuroScriptListenerImpl) EnterBlock_statement(c *gen.Block_statementContext) {}
func (l *neuroScriptListenerImpl) ExitBlock_statement(c *gen.Block_statementContext)  {}

func (l *neuroScriptListenerImpl) EnterLvalue(c *gen.LvalueContext) {}

func (l *neuroScriptListenerImpl) EnterSet_statement(c *gen.Set_statementContext) {}

func (l *neuroScriptListenerImpl) EnterCall_statement(c *gen.Call_statementContext) {}

func (l *neuroScriptListenerImpl) EnterReturn_statement(c *gen.Return_statementContext) {}

func (l *neuroScriptListenerImpl) EnterEmit_statement(c *gen.Emit_statementContext) {}

func (l *neuroScriptListenerImpl) EnterMust_statement(c *gen.Must_statementContext) {}

func (l *neuroScriptListenerImpl) EnterFail_statement(c *gen.Fail_statementContext) {}

func (l *neuroScriptListenerImpl) EnterClearErrorStmt(c *gen.ClearErrorStmtContext) {}

func (l *neuroScriptListenerImpl) EnterAsk_stmt(c *gen.Ask_stmtContext) {}

func (l *neuroScriptListenerImpl) EnterBreak_statement(c *gen.Break_statementContext) {}

func (l *neuroScriptListenerImpl) EnterContinue_statement(c *gen.Continue_statementContext) {}

func (l *neuroScriptListenerImpl) EnterIf_statement(c *gen.If_statementContext) {}

func (l *neuroScriptListenerImpl) EnterOnErrorStmt(c *gen.OnErrorStmtContext) {
	l.enterBlockContext("on_error")
}
func (l *neuroScriptListenerImpl) ExitOnErrorStmt(c *gen.OnErrorStmtContext) {
	body := l.exitBlockContext("on_error")
	stmt := Step{
		Pos:  tokenToPosition(c.GetStart()),
		Type: "on_error",
		Body: body,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

func (l *neuroScriptListenerImpl) EnterQualified_identifier(c *gen.Qualified_identifierContext) {}
func (l *neuroScriptListenerImpl) ExitQualified_identifier(c *gen.Qualified_identifierContext)  {}

func (l *neuroScriptListenerImpl) EnterCall_target(c *gen.Call_targetContext) {}

func (l *neuroScriptListenerImpl) EnterLogical_or_expr(c *gen.Logical_or_exprContext) {}

func (l *neuroScriptListenerImpl) EnterLogical_and_expr(c *gen.Logical_and_exprContext) {}

func (l *neuroScriptListenerImpl) EnterBitwise_or_expr(c *gen.Bitwise_or_exprContext) {}

func (l *neuroScriptListenerImpl) EnterBitwise_xor_expr(c *gen.Bitwise_xor_exprContext) {}

func (l *neuroScriptListenerImpl) EnterBitwise_and_expr(c *gen.Bitwise_and_exprContext) {}

func (l *neuroScriptListenerImpl) EnterEquality_expr(c *gen.Equality_exprContext) {}

func (l *neuroScriptListenerImpl) EnterRelational_expr(c *gen.Relational_exprContext) {}

func (l *neuroScriptListenerImpl) EnterMultiplicative_expr(c *gen.Multiplicative_exprContext) {}

func (l *neuroScriptListenerImpl) EnterUnary_expr(c *gen.Unary_exprContext) {}

func (l *neuroScriptListenerImpl) EnterPower_expr(c *gen.Power_exprContext) {}

func (l *neuroScriptListenerImpl) EnterPrimary(c *gen.PrimaryContext) {}

func (l *neuroScriptListenerImpl) EnterCallable_expr(c *gen.Callable_exprContext) {}

func (l *neuroScriptListenerImpl) EnterPlaceholder(c *gen.PlaceholderContext) {}

func (l *neuroScriptListenerImpl) EnterLiteral(c *gen.LiteralContext) {}

func (l *neuroScriptListenerImpl) EnterNil_literal(c *gen.Nil_literalContext) {}

func (l *neuroScriptListenerImpl) EnterBoolean_literal(c *gen.Boolean_literalContext) {}

func (l *neuroScriptListenerImpl) EnterList_literal(c *gen.List_literalContext) {}

func (l *neuroScriptListenerImpl) EnterMap_literal(c *gen.Map_literalContext) {}

func (l *neuroScriptListenerImpl) EnterExpression_list_opt(c *gen.Expression_list_optContext) {}
func (l *neuroScriptListenerImpl) ExitExpression_list_opt(c *gen.Expression_list_optContext)  {}

func (l *neuroScriptListenerImpl) EnterExpression_list(c *gen.Expression_listContext) {}

func (l *neuroScriptListenerImpl) EnterMap_entry_list_opt(c *gen.Map_entry_list_optContext) {}
func (l *neuroScriptListenerImpl) ExitMap_entry_list_opt(c *gen.Map_entry_list_optContext)  {}

func (l *neuroScriptListenerImpl) EnterMap_entry_list(c *gen.Map_entry_listContext) {}
func (l *neuroScriptListenerImpl) ExitMap_entry_list(c *gen.Map_entry_listContext)  {}

func (l *neuroScriptListenerImpl) EnterMap_entry(c *gen.Map_entryContext) {}
