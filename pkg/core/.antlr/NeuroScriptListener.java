// Generated from /home/aprice/dev/neuroscript/pkg/core/NeuroScript.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.tree.ParseTreeListener;

/**
 * This interface defines a complete listener for a parse tree produced by
 * {@link NeuroScriptParser}.
 */
public interface NeuroScriptListener extends ParseTreeListener {
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#program}.
	 * @param ctx the parse tree
	 */
	void enterProgram(NeuroScriptParser.ProgramContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#program}.
	 * @param ctx the parse tree
	 */
	void exitProgram(NeuroScriptParser.ProgramContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#procedure_definition}.
	 * @param ctx the parse tree
	 */
	void enterProcedure_definition(NeuroScriptParser.Procedure_definitionContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#procedure_definition}.
	 * @param ctx the parse tree
	 */
	void exitProcedure_definition(NeuroScriptParser.Procedure_definitionContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#signature_part}.
	 * @param ctx the parse tree
	 */
	void enterSignature_part(NeuroScriptParser.Signature_partContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#signature_part}.
	 * @param ctx the parse tree
	 */
	void exitSignature_part(NeuroScriptParser.Signature_partContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#needs_clause}.
	 * @param ctx the parse tree
	 */
	void enterNeeds_clause(NeuroScriptParser.Needs_clauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#needs_clause}.
	 * @param ctx the parse tree
	 */
	void exitNeeds_clause(NeuroScriptParser.Needs_clauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#optional_clause}.
	 * @param ctx the parse tree
	 */
	void enterOptional_clause(NeuroScriptParser.Optional_clauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#optional_clause}.
	 * @param ctx the parse tree
	 */
	void exitOptional_clause(NeuroScriptParser.Optional_clauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#returns_clause}.
	 * @param ctx the parse tree
	 */
	void enterReturns_clause(NeuroScriptParser.Returns_clauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#returns_clause}.
	 * @param ctx the parse tree
	 */
	void exitReturns_clause(NeuroScriptParser.Returns_clauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#param_list}.
	 * @param ctx the parse tree
	 */
	void enterParam_list(NeuroScriptParser.Param_listContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#param_list}.
	 * @param ctx the parse tree
	 */
	void exitParam_list(NeuroScriptParser.Param_listContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#metadata_block}.
	 * @param ctx the parse tree
	 */
	void enterMetadata_block(NeuroScriptParser.Metadata_blockContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#metadata_block}.
	 * @param ctx the parse tree
	 */
	void exitMetadata_block(NeuroScriptParser.Metadata_blockContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#statement_list}.
	 * @param ctx the parse tree
	 */
	void enterStatement_list(NeuroScriptParser.Statement_listContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#statement_list}.
	 * @param ctx the parse tree
	 */
	void exitStatement_list(NeuroScriptParser.Statement_listContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#body_line}.
	 * @param ctx the parse tree
	 */
	void enterBody_line(NeuroScriptParser.Body_lineContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#body_line}.
	 * @param ctx the parse tree
	 */
	void exitBody_line(NeuroScriptParser.Body_lineContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#statement}.
	 * @param ctx the parse tree
	 */
	void enterStatement(NeuroScriptParser.StatementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#statement}.
	 * @param ctx the parse tree
	 */
	void exitStatement(NeuroScriptParser.StatementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#simple_statement}.
	 * @param ctx the parse tree
	 */
	void enterSimple_statement(NeuroScriptParser.Simple_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#simple_statement}.
	 * @param ctx the parse tree
	 */
	void exitSimple_statement(NeuroScriptParser.Simple_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#block_statement}.
	 * @param ctx the parse tree
	 */
	void enterBlock_statement(NeuroScriptParser.Block_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#block_statement}.
	 * @param ctx the parse tree
	 */
	void exitBlock_statement(NeuroScriptParser.Block_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#on_stmt}.
	 * @param ctx the parse tree
	 */
	void enterOn_stmt(NeuroScriptParser.On_stmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#on_stmt}.
	 * @param ctx the parse tree
	 */
	void exitOn_stmt(NeuroScriptParser.On_stmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#error_handler}.
	 * @param ctx the parse tree
	 */
	void enterError_handler(NeuroScriptParser.Error_handlerContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#error_handler}.
	 * @param ctx the parse tree
	 */
	void exitError_handler(NeuroScriptParser.Error_handlerContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#event_handler}.
	 * @param ctx the parse tree
	 */
	void enterEvent_handler(NeuroScriptParser.Event_handlerContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#event_handler}.
	 * @param ctx the parse tree
	 */
	void exitEvent_handler(NeuroScriptParser.Event_handlerContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#clearEventStmt}.
	 * @param ctx the parse tree
	 */
	void enterClearEventStmt(NeuroScriptParser.ClearEventStmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#clearEventStmt}.
	 * @param ctx the parse tree
	 */
	void exitClearEventStmt(NeuroScriptParser.ClearEventStmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#lvalue}.
	 * @param ctx the parse tree
	 */
	void enterLvalue(NeuroScriptParser.LvalueContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#lvalue}.
	 * @param ctx the parse tree
	 */
	void exitLvalue(NeuroScriptParser.LvalueContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#lvalue_list}.
	 * @param ctx the parse tree
	 */
	void enterLvalue_list(NeuroScriptParser.Lvalue_listContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#lvalue_list}.
	 * @param ctx the parse tree
	 */
	void exitLvalue_list(NeuroScriptParser.Lvalue_listContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#set_statement}.
	 * @param ctx the parse tree
	 */
	void enterSet_statement(NeuroScriptParser.Set_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#set_statement}.
	 * @param ctx the parse tree
	 */
	void exitSet_statement(NeuroScriptParser.Set_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#call_statement}.
	 * @param ctx the parse tree
	 */
	void enterCall_statement(NeuroScriptParser.Call_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#call_statement}.
	 * @param ctx the parse tree
	 */
	void exitCall_statement(NeuroScriptParser.Call_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#return_statement}.
	 * @param ctx the parse tree
	 */
	void enterReturn_statement(NeuroScriptParser.Return_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#return_statement}.
	 * @param ctx the parse tree
	 */
	void exitReturn_statement(NeuroScriptParser.Return_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#emit_statement}.
	 * @param ctx the parse tree
	 */
	void enterEmit_statement(NeuroScriptParser.Emit_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#emit_statement}.
	 * @param ctx the parse tree
	 */
	void exitEmit_statement(NeuroScriptParser.Emit_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#must_statement}.
	 * @param ctx the parse tree
	 */
	void enterMust_statement(NeuroScriptParser.Must_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#must_statement}.
	 * @param ctx the parse tree
	 */
	void exitMust_statement(NeuroScriptParser.Must_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#fail_statement}.
	 * @param ctx the parse tree
	 */
	void enterFail_statement(NeuroScriptParser.Fail_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#fail_statement}.
	 * @param ctx the parse tree
	 */
	void exitFail_statement(NeuroScriptParser.Fail_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#clearErrorStmt}.
	 * @param ctx the parse tree
	 */
	void enterClearErrorStmt(NeuroScriptParser.ClearErrorStmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#clearErrorStmt}.
	 * @param ctx the parse tree
	 */
	void exitClearErrorStmt(NeuroScriptParser.ClearErrorStmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#ask_stmt}.
	 * @param ctx the parse tree
	 */
	void enterAsk_stmt(NeuroScriptParser.Ask_stmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#ask_stmt}.
	 * @param ctx the parse tree
	 */
	void exitAsk_stmt(NeuroScriptParser.Ask_stmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#break_statement}.
	 * @param ctx the parse tree
	 */
	void enterBreak_statement(NeuroScriptParser.Break_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#break_statement}.
	 * @param ctx the parse tree
	 */
	void exitBreak_statement(NeuroScriptParser.Break_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#continue_statement}.
	 * @param ctx the parse tree
	 */
	void enterContinue_statement(NeuroScriptParser.Continue_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#continue_statement}.
	 * @param ctx the parse tree
	 */
	void exitContinue_statement(NeuroScriptParser.Continue_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#if_statement}.
	 * @param ctx the parse tree
	 */
	void enterIf_statement(NeuroScriptParser.If_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#if_statement}.
	 * @param ctx the parse tree
	 */
	void exitIf_statement(NeuroScriptParser.If_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#while_statement}.
	 * @param ctx the parse tree
	 */
	void enterWhile_statement(NeuroScriptParser.While_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#while_statement}.
	 * @param ctx the parse tree
	 */
	void exitWhile_statement(NeuroScriptParser.While_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#for_each_statement}.
	 * @param ctx the parse tree
	 */
	void enterFor_each_statement(NeuroScriptParser.For_each_statementContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#for_each_statement}.
	 * @param ctx the parse tree
	 */
	void exitFor_each_statement(NeuroScriptParser.For_each_statementContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#qualified_identifier}.
	 * @param ctx the parse tree
	 */
	void enterQualified_identifier(NeuroScriptParser.Qualified_identifierContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#qualified_identifier}.
	 * @param ctx the parse tree
	 */
	void exitQualified_identifier(NeuroScriptParser.Qualified_identifierContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#call_target}.
	 * @param ctx the parse tree
	 */
	void enterCall_target(NeuroScriptParser.Call_targetContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#call_target}.
	 * @param ctx the parse tree
	 */
	void exitCall_target(NeuroScriptParser.Call_targetContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#expression}.
	 * @param ctx the parse tree
	 */
	void enterExpression(NeuroScriptParser.ExpressionContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#expression}.
	 * @param ctx the parse tree
	 */
	void exitExpression(NeuroScriptParser.ExpressionContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#logical_or_expr}.
	 * @param ctx the parse tree
	 */
	void enterLogical_or_expr(NeuroScriptParser.Logical_or_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#logical_or_expr}.
	 * @param ctx the parse tree
	 */
	void exitLogical_or_expr(NeuroScriptParser.Logical_or_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#logical_and_expr}.
	 * @param ctx the parse tree
	 */
	void enterLogical_and_expr(NeuroScriptParser.Logical_and_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#logical_and_expr}.
	 * @param ctx the parse tree
	 */
	void exitLogical_and_expr(NeuroScriptParser.Logical_and_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#bitwise_or_expr}.
	 * @param ctx the parse tree
	 */
	void enterBitwise_or_expr(NeuroScriptParser.Bitwise_or_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#bitwise_or_expr}.
	 * @param ctx the parse tree
	 */
	void exitBitwise_or_expr(NeuroScriptParser.Bitwise_or_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#bitwise_xor_expr}.
	 * @param ctx the parse tree
	 */
	void enterBitwise_xor_expr(NeuroScriptParser.Bitwise_xor_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#bitwise_xor_expr}.
	 * @param ctx the parse tree
	 */
	void exitBitwise_xor_expr(NeuroScriptParser.Bitwise_xor_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#bitwise_and_expr}.
	 * @param ctx the parse tree
	 */
	void enterBitwise_and_expr(NeuroScriptParser.Bitwise_and_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#bitwise_and_expr}.
	 * @param ctx the parse tree
	 */
	void exitBitwise_and_expr(NeuroScriptParser.Bitwise_and_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#equality_expr}.
	 * @param ctx the parse tree
	 */
	void enterEquality_expr(NeuroScriptParser.Equality_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#equality_expr}.
	 * @param ctx the parse tree
	 */
	void exitEquality_expr(NeuroScriptParser.Equality_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#relational_expr}.
	 * @param ctx the parse tree
	 */
	void enterRelational_expr(NeuroScriptParser.Relational_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#relational_expr}.
	 * @param ctx the parse tree
	 */
	void exitRelational_expr(NeuroScriptParser.Relational_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#additive_expr}.
	 * @param ctx the parse tree
	 */
	void enterAdditive_expr(NeuroScriptParser.Additive_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#additive_expr}.
	 * @param ctx the parse tree
	 */
	void exitAdditive_expr(NeuroScriptParser.Additive_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#multiplicative_expr}.
	 * @param ctx the parse tree
	 */
	void enterMultiplicative_expr(NeuroScriptParser.Multiplicative_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#multiplicative_expr}.
	 * @param ctx the parse tree
	 */
	void exitMultiplicative_expr(NeuroScriptParser.Multiplicative_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#unary_expr}.
	 * @param ctx the parse tree
	 */
	void enterUnary_expr(NeuroScriptParser.Unary_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#unary_expr}.
	 * @param ctx the parse tree
	 */
	void exitUnary_expr(NeuroScriptParser.Unary_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#power_expr}.
	 * @param ctx the parse tree
	 */
	void enterPower_expr(NeuroScriptParser.Power_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#power_expr}.
	 * @param ctx the parse tree
	 */
	void exitPower_expr(NeuroScriptParser.Power_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#accessor_expr}.
	 * @param ctx the parse tree
	 */
	void enterAccessor_expr(NeuroScriptParser.Accessor_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#accessor_expr}.
	 * @param ctx the parse tree
	 */
	void exitAccessor_expr(NeuroScriptParser.Accessor_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#primary}.
	 * @param ctx the parse tree
	 */
	void enterPrimary(NeuroScriptParser.PrimaryContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#primary}.
	 * @param ctx the parse tree
	 */
	void exitPrimary(NeuroScriptParser.PrimaryContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#callable_expr}.
	 * @param ctx the parse tree
	 */
	void enterCallable_expr(NeuroScriptParser.Callable_exprContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#callable_expr}.
	 * @param ctx the parse tree
	 */
	void exitCallable_expr(NeuroScriptParser.Callable_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#placeholder}.
	 * @param ctx the parse tree
	 */
	void enterPlaceholder(NeuroScriptParser.PlaceholderContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#placeholder}.
	 * @param ctx the parse tree
	 */
	void exitPlaceholder(NeuroScriptParser.PlaceholderContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#literal}.
	 * @param ctx the parse tree
	 */
	void enterLiteral(NeuroScriptParser.LiteralContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#literal}.
	 * @param ctx the parse tree
	 */
	void exitLiteral(NeuroScriptParser.LiteralContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#nil_literal}.
	 * @param ctx the parse tree
	 */
	void enterNil_literal(NeuroScriptParser.Nil_literalContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#nil_literal}.
	 * @param ctx the parse tree
	 */
	void exitNil_literal(NeuroScriptParser.Nil_literalContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#boolean_literal}.
	 * @param ctx the parse tree
	 */
	void enterBoolean_literal(NeuroScriptParser.Boolean_literalContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#boolean_literal}.
	 * @param ctx the parse tree
	 */
	void exitBoolean_literal(NeuroScriptParser.Boolean_literalContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#list_literal}.
	 * @param ctx the parse tree
	 */
	void enterList_literal(NeuroScriptParser.List_literalContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#list_literal}.
	 * @param ctx the parse tree
	 */
	void exitList_literal(NeuroScriptParser.List_literalContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#map_literal}.
	 * @param ctx the parse tree
	 */
	void enterMap_literal(NeuroScriptParser.Map_literalContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#map_literal}.
	 * @param ctx the parse tree
	 */
	void exitMap_literal(NeuroScriptParser.Map_literalContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#expression_list_opt}.
	 * @param ctx the parse tree
	 */
	void enterExpression_list_opt(NeuroScriptParser.Expression_list_optContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#expression_list_opt}.
	 * @param ctx the parse tree
	 */
	void exitExpression_list_opt(NeuroScriptParser.Expression_list_optContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#expression_list}.
	 * @param ctx the parse tree
	 */
	void enterExpression_list(NeuroScriptParser.Expression_listContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#expression_list}.
	 * @param ctx the parse tree
	 */
	void exitExpression_list(NeuroScriptParser.Expression_listContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#map_entry_list_opt}.
	 * @param ctx the parse tree
	 */
	void enterMap_entry_list_opt(NeuroScriptParser.Map_entry_list_optContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#map_entry_list_opt}.
	 * @param ctx the parse tree
	 */
	void exitMap_entry_list_opt(NeuroScriptParser.Map_entry_list_optContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#map_entry_list}.
	 * @param ctx the parse tree
	 */
	void enterMap_entry_list(NeuroScriptParser.Map_entry_listContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#map_entry_list}.
	 * @param ctx the parse tree
	 */
	void exitMap_entry_list(NeuroScriptParser.Map_entry_listContext ctx);
	/**
	 * Enter a parse tree produced by {@link NeuroScriptParser#map_entry}.
	 * @param ctx the parse tree
	 */
	void enterMap_entry(NeuroScriptParser.Map_entryContext ctx);
	/**
	 * Exit a parse tree produced by {@link NeuroScriptParser#map_entry}.
	 * @param ctx the parse tree
	 */
	void exitMap_entry(NeuroScriptParser.Map_entryContext ctx);
}