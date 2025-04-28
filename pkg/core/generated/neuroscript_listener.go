// Code generated from NeuroScript.g4 by ANTLR 4.13.2. DO NOT EDIT.

package core // NeuroScript
import "github.com/antlr4-go/antlr/v4"

// NeuroScriptListener is a complete listener for a parse tree produced by NeuroScriptParser.
type NeuroScriptListener interface {
	antlr.ParseTreeListener

	// EnterProgram is called when entering the program production.
	EnterProgram(c *ProgramContext)

	// EnterOptional_newlines is called when entering the optional_newlines production.
	EnterOptional_newlines(c *Optional_newlinesContext)

	// EnterFile_version_decl is called when entering the file_version_decl production.
	EnterFile_version_decl(c *File_version_declContext)

	// EnterProcedure_definition is called when entering the procedure_definition production.
	EnterProcedure_definition(c *Procedure_definitionContext)

	// EnterNeeds_clause is called when entering the needs_clause production.
	EnterNeeds_clause(c *Needs_clauseContext)

	// EnterOptional_clause is called when entering the optional_clause production.
	EnterOptional_clause(c *Optional_clauseContext)

	// EnterReturns_clause is called when entering the returns_clause production.
	EnterReturns_clause(c *Returns_clauseContext)

	// EnterParam_list is called when entering the param_list production.
	EnterParam_list(c *Param_listContext)

	// EnterMetadata_block is called when entering the metadata_block production.
	EnterMetadata_block(c *Metadata_blockContext)

	// EnterStatement_list is called when entering the statement_list production.
	EnterStatement_list(c *Statement_listContext)

	// EnterBody_line is called when entering the body_line production.
	EnterBody_line(c *Body_lineContext)

	// EnterStatement is called when entering the statement production.
	EnterStatement(c *StatementContext)

	// EnterMetadata_line_inline is called when entering the metadata_line_inline production.
	EnterMetadata_line_inline(c *Metadata_line_inlineContext)

	// EnterSimple_statement is called when entering the simple_statement production.
	EnterSimple_statement(c *Simple_statementContext)

	// EnterBlock_statement is called when entering the block_statement production.
	EnterBlock_statement(c *Block_statementContext)

	// EnterSet_statement is called when entering the set_statement production.
	EnterSet_statement(c *Set_statementContext)

	// EnterCall_statement is called when entering the call_statement production.
	EnterCall_statement(c *Call_statementContext)

	// EnterReturn_statement is called when entering the return_statement production.
	EnterReturn_statement(c *Return_statementContext)

	// EnterEmit_statement is called when entering the emit_statement production.
	EnterEmit_statement(c *Emit_statementContext)

	// EnterMust_statement is called when entering the must_statement production.
	EnterMust_statement(c *Must_statementContext)

	// EnterFail_statement is called when entering the fail_statement production.
	EnterFail_statement(c *Fail_statementContext)

	// EnterIf_statement is called when entering the if_statement production.
	EnterIf_statement(c *If_statementContext)

	// EnterWhile_statement is called when entering the while_statement production.
	EnterWhile_statement(c *While_statementContext)

	// EnterFor_each_statement is called when entering the for_each_statement production.
	EnterFor_each_statement(c *For_each_statementContext)

	// EnterTry_statement is called when entering the try_statement production.
	EnterTry_statement(c *Try_statementContext)

	// EnterCall_target is called when entering the call_target production.
	EnterCall_target(c *Call_targetContext)

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterLogical_or_expr is called when entering the logical_or_expr production.
	EnterLogical_or_expr(c *Logical_or_exprContext)

	// EnterLogical_and_expr is called when entering the logical_and_expr production.
	EnterLogical_and_expr(c *Logical_and_exprContext)

	// EnterBitwise_or_expr is called when entering the bitwise_or_expr production.
	EnterBitwise_or_expr(c *Bitwise_or_exprContext)

	// EnterBitwise_xor_expr is called when entering the bitwise_xor_expr production.
	EnterBitwise_xor_expr(c *Bitwise_xor_exprContext)

	// EnterBitwise_and_expr is called when entering the bitwise_and_expr production.
	EnterBitwise_and_expr(c *Bitwise_and_exprContext)

	// EnterEquality_expr is called when entering the equality_expr production.
	EnterEquality_expr(c *Equality_exprContext)

	// EnterRelational_expr is called when entering the relational_expr production.
	EnterRelational_expr(c *Relational_exprContext)

	// EnterAdditive_expr is called when entering the additive_expr production.
	EnterAdditive_expr(c *Additive_exprContext)

	// EnterMultiplicative_expr is called when entering the multiplicative_expr production.
	EnterMultiplicative_expr(c *Multiplicative_exprContext)

	// EnterUnary_expr is called when entering the unary_expr production.
	EnterUnary_expr(c *Unary_exprContext)

	// EnterPower_expr is called when entering the power_expr production.
	EnterPower_expr(c *Power_exprContext)

	// EnterAccessor_expr is called when entering the accessor_expr production.
	EnterAccessor_expr(c *Accessor_exprContext)

	// EnterPrimary is called when entering the primary production.
	EnterPrimary(c *PrimaryContext)

	// EnterFunction_call is called when entering the function_call production.
	EnterFunction_call(c *Function_callContext)

	// EnterPlaceholder is called when entering the placeholder production.
	EnterPlaceholder(c *PlaceholderContext)

	// EnterLiteral is called when entering the literal production.
	EnterLiteral(c *LiteralContext)

	// EnterBoolean_literal is called when entering the boolean_literal production.
	EnterBoolean_literal(c *Boolean_literalContext)

	// EnterList_literal is called when entering the list_literal production.
	EnterList_literal(c *List_literalContext)

	// EnterMap_literal is called when entering the map_literal production.
	EnterMap_literal(c *Map_literalContext)

	// EnterExpression_list_opt is called when entering the expression_list_opt production.
	EnterExpression_list_opt(c *Expression_list_optContext)

	// EnterExpression_list is called when entering the expression_list production.
	EnterExpression_list(c *Expression_listContext)

	// EnterMap_entry_list_opt is called when entering the map_entry_list_opt production.
	EnterMap_entry_list_opt(c *Map_entry_list_optContext)

	// EnterMap_entry_list is called when entering the map_entry_list production.
	EnterMap_entry_list(c *Map_entry_listContext)

	// EnterMap_entry is called when entering the map_entry production.
	EnterMap_entry(c *Map_entryContext)

	// ExitProgram is called when exiting the program production.
	ExitProgram(c *ProgramContext)

	// ExitOptional_newlines is called when exiting the optional_newlines production.
	ExitOptional_newlines(c *Optional_newlinesContext)

	// ExitFile_version_decl is called when exiting the file_version_decl production.
	ExitFile_version_decl(c *File_version_declContext)

	// ExitProcedure_definition is called when exiting the procedure_definition production.
	ExitProcedure_definition(c *Procedure_definitionContext)

	// ExitNeeds_clause is called when exiting the needs_clause production.
	ExitNeeds_clause(c *Needs_clauseContext)

	// ExitOptional_clause is called when exiting the optional_clause production.
	ExitOptional_clause(c *Optional_clauseContext)

	// ExitReturns_clause is called when exiting the returns_clause production.
	ExitReturns_clause(c *Returns_clauseContext)

	// ExitParam_list is called when exiting the param_list production.
	ExitParam_list(c *Param_listContext)

	// ExitMetadata_block is called when exiting the metadata_block production.
	ExitMetadata_block(c *Metadata_blockContext)

	// ExitStatement_list is called when exiting the statement_list production.
	ExitStatement_list(c *Statement_listContext)

	// ExitBody_line is called when exiting the body_line production.
	ExitBody_line(c *Body_lineContext)

	// ExitStatement is called when exiting the statement production.
	ExitStatement(c *StatementContext)

	// ExitMetadata_line_inline is called when exiting the metadata_line_inline production.
	ExitMetadata_line_inline(c *Metadata_line_inlineContext)

	// ExitSimple_statement is called when exiting the simple_statement production.
	ExitSimple_statement(c *Simple_statementContext)

	// ExitBlock_statement is called when exiting the block_statement production.
	ExitBlock_statement(c *Block_statementContext)

	// ExitSet_statement is called when exiting the set_statement production.
	ExitSet_statement(c *Set_statementContext)

	// ExitCall_statement is called when exiting the call_statement production.
	ExitCall_statement(c *Call_statementContext)

	// ExitReturn_statement is called when exiting the return_statement production.
	ExitReturn_statement(c *Return_statementContext)

	// ExitEmit_statement is called when exiting the emit_statement production.
	ExitEmit_statement(c *Emit_statementContext)

	// ExitMust_statement is called when exiting the must_statement production.
	ExitMust_statement(c *Must_statementContext)

	// ExitFail_statement is called when exiting the fail_statement production.
	ExitFail_statement(c *Fail_statementContext)

	// ExitIf_statement is called when exiting the if_statement production.
	ExitIf_statement(c *If_statementContext)

	// ExitWhile_statement is called when exiting the while_statement production.
	ExitWhile_statement(c *While_statementContext)

	// ExitFor_each_statement is called when exiting the for_each_statement production.
	ExitFor_each_statement(c *For_each_statementContext)

	// ExitTry_statement is called when exiting the try_statement production.
	ExitTry_statement(c *Try_statementContext)

	// ExitCall_target is called when exiting the call_target production.
	ExitCall_target(c *Call_targetContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitLogical_or_expr is called when exiting the logical_or_expr production.
	ExitLogical_or_expr(c *Logical_or_exprContext)

	// ExitLogical_and_expr is called when exiting the logical_and_expr production.
	ExitLogical_and_expr(c *Logical_and_exprContext)

	// ExitBitwise_or_expr is called when exiting the bitwise_or_expr production.
	ExitBitwise_or_expr(c *Bitwise_or_exprContext)

	// ExitBitwise_xor_expr is called when exiting the bitwise_xor_expr production.
	ExitBitwise_xor_expr(c *Bitwise_xor_exprContext)

	// ExitBitwise_and_expr is called when exiting the bitwise_and_expr production.
	ExitBitwise_and_expr(c *Bitwise_and_exprContext)

	// ExitEquality_expr is called when exiting the equality_expr production.
	ExitEquality_expr(c *Equality_exprContext)

	// ExitRelational_expr is called when exiting the relational_expr production.
	ExitRelational_expr(c *Relational_exprContext)

	// ExitAdditive_expr is called when exiting the additive_expr production.
	ExitAdditive_expr(c *Additive_exprContext)

	// ExitMultiplicative_expr is called when exiting the multiplicative_expr production.
	ExitMultiplicative_expr(c *Multiplicative_exprContext)

	// ExitUnary_expr is called when exiting the unary_expr production.
	ExitUnary_expr(c *Unary_exprContext)

	// ExitPower_expr is called when exiting the power_expr production.
	ExitPower_expr(c *Power_exprContext)

	// ExitAccessor_expr is called when exiting the accessor_expr production.
	ExitAccessor_expr(c *Accessor_exprContext)

	// ExitPrimary is called when exiting the primary production.
	ExitPrimary(c *PrimaryContext)

	// ExitFunction_call is called when exiting the function_call production.
	ExitFunction_call(c *Function_callContext)

	// ExitPlaceholder is called when exiting the placeholder production.
	ExitPlaceholder(c *PlaceholderContext)

	// ExitLiteral is called when exiting the literal production.
	ExitLiteral(c *LiteralContext)

	// ExitBoolean_literal is called when exiting the boolean_literal production.
	ExitBoolean_literal(c *Boolean_literalContext)

	// ExitList_literal is called when exiting the list_literal production.
	ExitList_literal(c *List_literalContext)

	// ExitMap_literal is called when exiting the map_literal production.
	ExitMap_literal(c *Map_literalContext)

	// ExitExpression_list_opt is called when exiting the expression_list_opt production.
	ExitExpression_list_opt(c *Expression_list_optContext)

	// ExitExpression_list is called when exiting the expression_list production.
	ExitExpression_list(c *Expression_listContext)

	// ExitMap_entry_list_opt is called when exiting the map_entry_list_opt production.
	ExitMap_entry_list_opt(c *Map_entry_list_optContext)

	// ExitMap_entry_list is called when exiting the map_entry_list production.
	ExitMap_entry_list(c *Map_entry_listContext)

	// ExitMap_entry is called when exiting the map_entry production.
	ExitMap_entry(c *Map_entryContext)
}
