// Code generated from NeuroScript.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // NeuroScript
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by NeuroScriptParser.
type NeuroScriptVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by NeuroScriptParser#program.
	VisitProgram(ctx *ProgramContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#optional_newlines.
	VisitOptional_newlines(ctx *Optional_newlinesContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#file_version_decl.
	VisitFile_version_decl(ctx *File_version_declContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#procedure_definition.
	VisitProcedure_definition(ctx *Procedure_definitionContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#param_list_opt.
	VisitParam_list_opt(ctx *Param_list_optContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#param_list.
	VisitParam_list(ctx *Param_listContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#statement_list.
	VisitStatement_list(ctx *Statement_listContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#body_line.
	VisitBody_line(ctx *Body_lineContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#statement.
	VisitStatement(ctx *StatementContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#simple_statement.
	VisitSimple_statement(ctx *Simple_statementContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#block_statement.
	VisitBlock_statement(ctx *Block_statementContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#set_statement.
	VisitSet_statement(ctx *Set_statementContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#call_statement.
	VisitCall_statement(ctx *Call_statementContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#return_statement.
	VisitReturn_statement(ctx *Return_statementContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#emit_statement.
	VisitEmit_statement(ctx *Emit_statementContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#if_statement.
	VisitIf_statement(ctx *If_statementContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#while_statement.
	VisitWhile_statement(ctx *While_statementContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#for_each_statement.
	VisitFor_each_statement(ctx *For_each_statementContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#call_target.
	VisitCall_target(ctx *Call_targetContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#expression.
	VisitExpression(ctx *ExpressionContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#logical_or_expr.
	VisitLogical_or_expr(ctx *Logical_or_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#logical_and_expr.
	VisitLogical_and_expr(ctx *Logical_and_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#bitwise_or_expr.
	VisitBitwise_or_expr(ctx *Bitwise_or_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#bitwise_xor_expr.
	VisitBitwise_xor_expr(ctx *Bitwise_xor_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#bitwise_and_expr.
	VisitBitwise_and_expr(ctx *Bitwise_and_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#equality_expr.
	VisitEquality_expr(ctx *Equality_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#relational_expr.
	VisitRelational_expr(ctx *Relational_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#additive_expr.
	VisitAdditive_expr(ctx *Additive_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#multiplicative_expr.
	VisitMultiplicative_expr(ctx *Multiplicative_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#unary_expr.
	VisitUnary_expr(ctx *Unary_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#power_expr.
	VisitPower_expr(ctx *Power_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#accessor_expr.
	VisitAccessor_expr(ctx *Accessor_exprContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#primary.
	VisitPrimary(ctx *PrimaryContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#function_call.
	VisitFunction_call(ctx *Function_callContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#placeholder.
	VisitPlaceholder(ctx *PlaceholderContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#literal.
	VisitLiteral(ctx *LiteralContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#boolean_literal.
	VisitBoolean_literal(ctx *Boolean_literalContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#list_literal.
	VisitList_literal(ctx *List_literalContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#map_literal.
	VisitMap_literal(ctx *Map_literalContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#expression_list_opt.
	VisitExpression_list_opt(ctx *Expression_list_optContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#expression_list.
	VisitExpression_list(ctx *Expression_listContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#map_entry_list_opt.
	VisitMap_entry_list_opt(ctx *Map_entry_list_optContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#map_entry_list.
	VisitMap_entry_list(ctx *Map_entry_listContext) interface{}

	// Visit a parse tree produced by NeuroScriptParser#map_entry.
	VisitMap_entry(ctx *Map_entryContext) interface{}
}
