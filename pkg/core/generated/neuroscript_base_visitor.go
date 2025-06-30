// Code generated from /home/aprice/dev/neuroscript/pkg/core/NeuroScript.g4 by ANTLR 4.13.2. DO NOT EDIT.

package core // NeuroScript
import "github.com/antlr4-go/antlr/v4"

type BaseNeuroScriptVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseNeuroScriptVisitor) VisitProgram(ctx *ProgramContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitFile_header(ctx *File_headerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitLibrary_script(ctx *Library_scriptContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitCommand_script(ctx *Command_scriptContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitLibrary_block(ctx *Library_blockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitCommand_block(ctx *Command_blockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitCommand_statement_list(ctx *Command_statement_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitCommand_body_line(ctx *Command_body_lineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitCommand_statement(ctx *Command_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitOn_error_only_stmt(ctx *On_error_only_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitSimple_command_statement(ctx *Simple_command_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitProcedure_definition(ctx *Procedure_definitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitSignature_part(ctx *Signature_partContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitNeeds_clause(ctx *Needs_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitOptional_clause(ctx *Optional_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitReturns_clause(ctx *Returns_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitParam_list(ctx *Param_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitMetadata_block(ctx *Metadata_blockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitNon_empty_statement_list(ctx *Non_empty_statement_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitStatement_list(ctx *Statement_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitBody_line(ctx *Body_lineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitStatement(ctx *StatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitSimple_statement(ctx *Simple_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitBlock_statement(ctx *Block_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitOn_stmt(ctx *On_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitError_handler(ctx *Error_handlerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitEvent_handler(ctx *Event_handlerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitClearEventStmt(ctx *ClearEventStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitLvalue(ctx *LvalueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitLvalue_list(ctx *Lvalue_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitSet_statement(ctx *Set_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitCall_statement(ctx *Call_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitReturn_statement(ctx *Return_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitEmit_statement(ctx *Emit_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitMust_statement(ctx *Must_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitFail_statement(ctx *Fail_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitClearErrorStmt(ctx *ClearErrorStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitAsk_stmt(ctx *Ask_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitBreak_statement(ctx *Break_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitContinue_statement(ctx *Continue_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitIf_statement(ctx *If_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitWhile_statement(ctx *While_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitFor_each_statement(ctx *For_each_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitQualified_identifier(ctx *Qualified_identifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitCall_target(ctx *Call_targetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitExpression(ctx *ExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitLogical_or_expr(ctx *Logical_or_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitLogical_and_expr(ctx *Logical_and_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitBitwise_or_expr(ctx *Bitwise_or_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitBitwise_xor_expr(ctx *Bitwise_xor_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitBitwise_and_expr(ctx *Bitwise_and_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitEquality_expr(ctx *Equality_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitRelational_expr(ctx *Relational_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitAdditive_expr(ctx *Additive_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitMultiplicative_expr(ctx *Multiplicative_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitUnary_expr(ctx *Unary_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitPower_expr(ctx *Power_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitAccessor_expr(ctx *Accessor_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitPrimary(ctx *PrimaryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitCallable_expr(ctx *Callable_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitPlaceholder(ctx *PlaceholderContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitLiteral(ctx *LiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitNil_literal(ctx *Nil_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitBoolean_literal(ctx *Boolean_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitList_literal(ctx *List_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitMap_literal(ctx *Map_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitExpression_list_opt(ctx *Expression_list_optContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitExpression_list(ctx *Expression_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitMap_entry_list_opt(ctx *Map_entry_list_optContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitMap_entry_list(ctx *Map_entry_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitMap_entry(ctx *Map_entryContext) interface{} {
	return v.VisitChildren(ctx)
}
