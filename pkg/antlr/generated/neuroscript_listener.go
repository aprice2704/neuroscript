// Code generated from /home/aprice/dev/neuroscript/pkg/antlr/NeuroScript.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // NeuroScript
import "github.com/antlr4-go/antlr/v4"

// NeuroScriptListener is a complete listener for a parse tree produced by NeuroScriptParser.
type NeuroScriptListener interface {
	antlr.ParseTreeListener

	// EnterProgram is called when entering the program production.
	EnterProgram(c *ProgramContext)

	// EnterFile_header is called when entering the file_header production.
	EnterFile_header(c *File_headerContext)

	// EnterLibrary_script is called when entering the library_script production.
	EnterLibrary_script(c *Library_scriptContext)

	// EnterCommand_script is called when entering the command_script production.
	EnterCommand_script(c *Command_scriptContext)

	// EnterLibrary_block is called when entering the library_block production.
	EnterLibrary_block(c *Library_blockContext)

	// EnterCommand_block is called when entering the command_block production.
	EnterCommand_block(c *Command_blockContext)

	// EnterCommand_statement_list is called when entering the command_statement_list production.
	EnterCommand_statement_list(c *Command_statement_listContext)

	// EnterCommand_body_line is called when entering the command_body_line production.
	EnterCommand_body_line(c *Command_body_lineContext)

	// EnterCommand_statement is called when entering the command_statement production.
	EnterCommand_statement(c *Command_statementContext)

	// EnterOn_error_only_stmt is called when entering the on_error_only_stmt production.
	EnterOn_error_only_stmt(c *On_error_only_stmtContext)

	// EnterSimple_command_statement is called when entering the simple_command_statement production.
	EnterSimple_command_statement(c *Simple_command_statementContext)

	// EnterProcedure_definition is called when entering the procedure_definition production.
	EnterProcedure_definition(c *Procedure_definitionContext)

	// EnterSignature_part is called when entering the signature_part production.
	EnterSignature_part(c *Signature_partContext)

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

	// EnterNon_empty_statement_list is called when entering the non_empty_statement_list production.
	EnterNon_empty_statement_list(c *Non_empty_statement_listContext)

	// EnterStatement_list is called when entering the statement_list production.
	EnterStatement_list(c *Statement_listContext)

	// EnterBody_line is called when entering the body_line production.
	EnterBody_line(c *Body_lineContext)

	// EnterStatement is called when entering the statement production.
	EnterStatement(c *StatementContext)

	// EnterSimple_statement is called when entering the simple_statement production.
	EnterSimple_statement(c *Simple_statementContext)

	// EnterBlock_statement is called when entering the block_statement production.
	EnterBlock_statement(c *Block_statementContext)

	// EnterOn_stmt is called when entering the on_stmt production.
	EnterOn_stmt(c *On_stmtContext)

	// EnterError_handler is called when entering the error_handler production.
	EnterError_handler(c *Error_handlerContext)

	// EnterEvent_handler is called when entering the event_handler production.
	EnterEvent_handler(c *Event_handlerContext)

	// EnterClearEventStmt is called when entering the clearEventStmt production.
	EnterClearEventStmt(c *ClearEventStmtContext)

	// EnterLvalue is called when entering the lvalue production.
	EnterLvalue(c *LvalueContext)

	// EnterLvalue_list is called when entering the lvalue_list production.
	EnterLvalue_list(c *Lvalue_listContext)

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

	// EnterClearErrorStmt is called when entering the clearErrorStmt production.
	EnterClearErrorStmt(c *ClearErrorStmtContext)

	// EnterAsk_stmt is called when entering the ask_stmt production.
	EnterAsk_stmt(c *Ask_stmtContext)

	// EnterPromptuser_stmt is called when entering the promptuser_stmt production.
	EnterPromptuser_stmt(c *Promptuser_stmtContext)

	// EnterBreak_statement is called when entering the break_statement production.
	EnterBreak_statement(c *Break_statementContext)

	// EnterContinue_statement is called when entering the continue_statement production.
	EnterContinue_statement(c *Continue_statementContext)

	// EnterIf_statement is called when entering the if_statement production.
	EnterIf_statement(c *If_statementContext)

	// EnterWhile_statement is called when entering the while_statement production.
	EnterWhile_statement(c *While_statementContext)

	// EnterFor_each_statement is called when entering the for_each_statement production.
	EnterFor_each_statement(c *For_each_statementContext)

	// EnterQualified_identifier is called when entering the qualified_identifier production.
	EnterQualified_identifier(c *Qualified_identifierContext)

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

	// EnterCallable_expr is called when entering the callable_expr production.
	EnterCallable_expr(c *Callable_exprContext)

	// EnterPlaceholder is called when entering the placeholder production.
	EnterPlaceholder(c *PlaceholderContext)

	// EnterLiteral is called when entering the literal production.
	EnterLiteral(c *LiteralContext)

	// EnterNil_literal is called when entering the nil_literal production.
	EnterNil_literal(c *Nil_literalContext)

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

	// ExitFile_header is called when exiting the file_header production.
	ExitFile_header(c *File_headerContext)

	// ExitLibrary_script is called when exiting the library_script production.
	ExitLibrary_script(c *Library_scriptContext)

	// ExitCommand_script is called when exiting the command_script production.
	ExitCommand_script(c *Command_scriptContext)

	// ExitLibrary_block is called when exiting the library_block production.
	ExitLibrary_block(c *Library_blockContext)

	// ExitCommand_block is called when exiting the command_block production.
	ExitCommand_block(c *Command_blockContext)

	// ExitCommand_statement_list is called when exiting the command_statement_list production.
	ExitCommand_statement_list(c *Command_statement_listContext)

	// ExitCommand_body_line is called when exiting the command_body_line production.
	ExitCommand_body_line(c *Command_body_lineContext)

	// ExitCommand_statement is called when exiting the command_statement production.
	ExitCommand_statement(c *Command_statementContext)

	// ExitOn_error_only_stmt is called when exiting the on_error_only_stmt production.
	ExitOn_error_only_stmt(c *On_error_only_stmtContext)

	// ExitSimple_command_statement is called when exiting the simple_command_statement production.
	ExitSimple_command_statement(c *Simple_command_statementContext)

	// ExitProcedure_definition is called when exiting the procedure_definition production.
	ExitProcedure_definition(c *Procedure_definitionContext)

	// ExitSignature_part is called when exiting the signature_part production.
	ExitSignature_part(c *Signature_partContext)

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

	// ExitNon_empty_statement_list is called when exiting the non_empty_statement_list production.
	ExitNon_empty_statement_list(c *Non_empty_statement_listContext)

	// ExitStatement_list is called when exiting the statement_list production.
	ExitStatement_list(c *Statement_listContext)

	// ExitBody_line is called when exiting the body_line production.
	ExitBody_line(c *Body_lineContext)

	// ExitStatement is called when exiting the statement production.
	ExitStatement(c *StatementContext)

	// ExitSimple_statement is called when exiting the simple_statement production.
	ExitSimple_statement(c *Simple_statementContext)

	// ExitBlock_statement is called when exiting the block_statement production.
	ExitBlock_statement(c *Block_statementContext)

	// ExitOn_stmt is called when exiting the on_stmt production.
	ExitOn_stmt(c *On_stmtContext)

	// ExitError_handler is called when exiting the error_handler production.
	ExitError_handler(c *Error_handlerContext)

	// ExitEvent_handler is called when exiting the event_handler production.
	ExitEvent_handler(c *Event_handlerContext)

	// ExitClearEventStmt is called when exiting the clearEventStmt production.
	ExitClearEventStmt(c *ClearEventStmtContext)

	// ExitLvalue is called when exiting the lvalue production.
	ExitLvalue(c *LvalueContext)

	// ExitLvalue_list is called when exiting the lvalue_list production.
	ExitLvalue_list(c *Lvalue_listContext)

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

	// ExitClearErrorStmt is called when exiting the clearErrorStmt production.
	ExitClearErrorStmt(c *ClearErrorStmtContext)

	// ExitAsk_stmt is called when exiting the ask_stmt production.
	ExitAsk_stmt(c *Ask_stmtContext)

	// ExitPromptuser_stmt is called when exiting the promptuser_stmt production.
	ExitPromptuser_stmt(c *Promptuser_stmtContext)

	// ExitBreak_statement is called when exiting the break_statement production.
	ExitBreak_statement(c *Break_statementContext)

	// ExitContinue_statement is called when exiting the continue_statement production.
	ExitContinue_statement(c *Continue_statementContext)

	// ExitIf_statement is called when exiting the if_statement production.
	ExitIf_statement(c *If_statementContext)

	// ExitWhile_statement is called when exiting the while_statement production.
	ExitWhile_statement(c *While_statementContext)

	// ExitFor_each_statement is called when exiting the for_each_statement production.
	ExitFor_each_statement(c *For_each_statementContext)

	// ExitQualified_identifier is called when exiting the qualified_identifier production.
	ExitQualified_identifier(c *Qualified_identifierContext)

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

	// ExitCallable_expr is called when exiting the callable_expr production.
	ExitCallable_expr(c *Callable_exprContext)

	// ExitPlaceholder is called when exiting the placeholder production.
	ExitPlaceholder(c *PlaceholderContext)

	// ExitLiteral is called when exiting the literal production.
	ExitLiteral(c *LiteralContext)

	// ExitNil_literal is called when exiting the nil_literal production.
	ExitNil_literal(c *Nil_literalContext)

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
