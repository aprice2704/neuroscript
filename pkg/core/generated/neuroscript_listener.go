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

	// EnterParam_list_opt is called when entering the param_list_opt production.
	EnterParam_list_opt(c *Param_list_optContext)

	// EnterParam_list is called when entering the param_list production.
	EnterParam_list(c *Param_listContext)

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

	// EnterSet_statement is called when entering the set_statement production.
	EnterSet_statement(c *Set_statementContext)

	// EnterCall_statement is called when entering the call_statement production.
	EnterCall_statement(c *Call_statementContext)

	// EnterReturn_statement is called when entering the return_statement production.
	EnterReturn_statement(c *Return_statementContext)

	// EnterEmit_statement is called when entering the emit_statement production.
	EnterEmit_statement(c *Emit_statementContext)

	// EnterIf_statement is called when entering the if_statement production.
	EnterIf_statement(c *If_statementContext)

	// EnterWhile_statement is called when entering the while_statement production.
	EnterWhile_statement(c *While_statementContext)

	// EnterFor_each_statement is called when entering the for_each_statement production.
	EnterFor_each_statement(c *For_each_statementContext)

	// EnterCall_target is called when entering the call_target production.
	EnterCall_target(c *Call_targetContext)

	// EnterCondition is called when entering the condition production.
	EnterCondition(c *ConditionContext)

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterTerm is called when entering the term production.
	EnterTerm(c *TermContext)

	// EnterPrimary is called when entering the primary production.
	EnterPrimary(c *PrimaryContext)

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

	// ExitParam_list_opt is called when exiting the param_list_opt production.
	ExitParam_list_opt(c *Param_list_optContext)

	// ExitParam_list is called when exiting the param_list production.
	ExitParam_list(c *Param_listContext)

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

	// ExitSet_statement is called when exiting the set_statement production.
	ExitSet_statement(c *Set_statementContext)

	// ExitCall_statement is called when exiting the call_statement production.
	ExitCall_statement(c *Call_statementContext)

	// ExitReturn_statement is called when exiting the return_statement production.
	ExitReturn_statement(c *Return_statementContext)

	// ExitEmit_statement is called when exiting the emit_statement production.
	ExitEmit_statement(c *Emit_statementContext)

	// ExitIf_statement is called when exiting the if_statement production.
	ExitIf_statement(c *If_statementContext)

	// ExitWhile_statement is called when exiting the while_statement production.
	ExitWhile_statement(c *While_statementContext)

	// ExitFor_each_statement is called when exiting the for_each_statement production.
	ExitFor_each_statement(c *For_each_statementContext)

	// ExitCall_target is called when exiting the call_target production.
	ExitCall_target(c *Call_targetContext)

	// ExitCondition is called when exiting the condition production.
	ExitCondition(c *ConditionContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitTerm is called when exiting the term production.
	ExitTerm(c *TermContext)

	// ExitPrimary is called when exiting the primary production.
	ExitPrimary(c *PrimaryContext)

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
