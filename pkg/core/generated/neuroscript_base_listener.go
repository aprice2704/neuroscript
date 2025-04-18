// Code generated from NeuroScript.g4 by ANTLR 4.13.2. DO NOT EDIT.

package generated // NeuroScript
import "github.com/antlr4-go/antlr/v4"

// BaseNeuroScriptListener is a complete listener for a parse tree produced by NeuroScriptParser.
type BaseNeuroScriptListener struct{}

var _ NeuroScriptListener = &BaseNeuroScriptListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseNeuroScriptListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseNeuroScriptListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseNeuroScriptListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseNeuroScriptListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterProgram is called when production program is entered.
func (s *BaseNeuroScriptListener) EnterProgram(ctx *ProgramContext) {}

// ExitProgram is called when production program is exited.
func (s *BaseNeuroScriptListener) ExitProgram(ctx *ProgramContext) {}

// EnterOptional_newlines is called when production optional_newlines is entered.
func (s *BaseNeuroScriptListener) EnterOptional_newlines(ctx *Optional_newlinesContext) {}

// ExitOptional_newlines is called when production optional_newlines is exited.
func (s *BaseNeuroScriptListener) ExitOptional_newlines(ctx *Optional_newlinesContext) {}

// EnterFile_version_decl is called when production file_version_decl is entered.
func (s *BaseNeuroScriptListener) EnterFile_version_decl(ctx *File_version_declContext) {}

// ExitFile_version_decl is called when production file_version_decl is exited.
func (s *BaseNeuroScriptListener) ExitFile_version_decl(ctx *File_version_declContext) {}

// EnterProcedure_definition is called when production procedure_definition is entered.
func (s *BaseNeuroScriptListener) EnterProcedure_definition(ctx *Procedure_definitionContext) {}

// ExitProcedure_definition is called when production procedure_definition is exited.
func (s *BaseNeuroScriptListener) ExitProcedure_definition(ctx *Procedure_definitionContext) {}

// EnterParam_list_opt is called when production param_list_opt is entered.
func (s *BaseNeuroScriptListener) EnterParam_list_opt(ctx *Param_list_optContext) {}

// ExitParam_list_opt is called when production param_list_opt is exited.
func (s *BaseNeuroScriptListener) ExitParam_list_opt(ctx *Param_list_optContext) {}

// EnterParam_list is called when production param_list is entered.
func (s *BaseNeuroScriptListener) EnterParam_list(ctx *Param_listContext) {}

// ExitParam_list is called when production param_list is exited.
func (s *BaseNeuroScriptListener) ExitParam_list(ctx *Param_listContext) {}

// EnterStatement_list is called when production statement_list is entered.
func (s *BaseNeuroScriptListener) EnterStatement_list(ctx *Statement_listContext) {}

// ExitStatement_list is called when production statement_list is exited.
func (s *BaseNeuroScriptListener) ExitStatement_list(ctx *Statement_listContext) {}

// EnterBody_line is called when production body_line is entered.
func (s *BaseNeuroScriptListener) EnterBody_line(ctx *Body_lineContext) {}

// ExitBody_line is called when production body_line is exited.
func (s *BaseNeuroScriptListener) ExitBody_line(ctx *Body_lineContext) {}

// EnterStatement is called when production statement is entered.
func (s *BaseNeuroScriptListener) EnterStatement(ctx *StatementContext) {}

// ExitStatement is called when production statement is exited.
func (s *BaseNeuroScriptListener) ExitStatement(ctx *StatementContext) {}

// EnterSimple_statement is called when production simple_statement is entered.
func (s *BaseNeuroScriptListener) EnterSimple_statement(ctx *Simple_statementContext) {}

// ExitSimple_statement is called when production simple_statement is exited.
func (s *BaseNeuroScriptListener) ExitSimple_statement(ctx *Simple_statementContext) {}

// EnterBlock_statement is called when production block_statement is entered.
func (s *BaseNeuroScriptListener) EnterBlock_statement(ctx *Block_statementContext) {}

// ExitBlock_statement is called when production block_statement is exited.
func (s *BaseNeuroScriptListener) ExitBlock_statement(ctx *Block_statementContext) {}

// EnterSet_statement is called when production set_statement is entered.
func (s *BaseNeuroScriptListener) EnterSet_statement(ctx *Set_statementContext) {}

// ExitSet_statement is called when production set_statement is exited.
func (s *BaseNeuroScriptListener) ExitSet_statement(ctx *Set_statementContext) {}

// EnterCall_statement is called when production call_statement is entered.
func (s *BaseNeuroScriptListener) EnterCall_statement(ctx *Call_statementContext) {}

// ExitCall_statement is called when production call_statement is exited.
func (s *BaseNeuroScriptListener) ExitCall_statement(ctx *Call_statementContext) {}

// EnterReturn_statement is called when production return_statement is entered.
func (s *BaseNeuroScriptListener) EnterReturn_statement(ctx *Return_statementContext) {}

// ExitReturn_statement is called when production return_statement is exited.
func (s *BaseNeuroScriptListener) ExitReturn_statement(ctx *Return_statementContext) {}

// EnterEmit_statement is called when production emit_statement is entered.
func (s *BaseNeuroScriptListener) EnterEmit_statement(ctx *Emit_statementContext) {}

// ExitEmit_statement is called when production emit_statement is exited.
func (s *BaseNeuroScriptListener) ExitEmit_statement(ctx *Emit_statementContext) {}

// EnterIf_statement is called when production if_statement is entered.
func (s *BaseNeuroScriptListener) EnterIf_statement(ctx *If_statementContext) {}

// ExitIf_statement is called when production if_statement is exited.
func (s *BaseNeuroScriptListener) ExitIf_statement(ctx *If_statementContext) {}

// EnterWhile_statement is called when production while_statement is entered.
func (s *BaseNeuroScriptListener) EnterWhile_statement(ctx *While_statementContext) {}

// ExitWhile_statement is called when production while_statement is exited.
func (s *BaseNeuroScriptListener) ExitWhile_statement(ctx *While_statementContext) {}

// EnterFor_each_statement is called when production for_each_statement is entered.
func (s *BaseNeuroScriptListener) EnterFor_each_statement(ctx *For_each_statementContext) {}

// ExitFor_each_statement is called when production for_each_statement is exited.
func (s *BaseNeuroScriptListener) ExitFor_each_statement(ctx *For_each_statementContext) {}

// EnterCall_target is called when production call_target is entered.
func (s *BaseNeuroScriptListener) EnterCall_target(ctx *Call_targetContext) {}

// ExitCall_target is called when production call_target is exited.
func (s *BaseNeuroScriptListener) ExitCall_target(ctx *Call_targetContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseNeuroScriptListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseNeuroScriptListener) ExitExpression(ctx *ExpressionContext) {}

// EnterLogical_or_expr is called when production logical_or_expr is entered.
func (s *BaseNeuroScriptListener) EnterLogical_or_expr(ctx *Logical_or_exprContext) {}

// ExitLogical_or_expr is called when production logical_or_expr is exited.
func (s *BaseNeuroScriptListener) ExitLogical_or_expr(ctx *Logical_or_exprContext) {}

// EnterLogical_and_expr is called when production logical_and_expr is entered.
func (s *BaseNeuroScriptListener) EnterLogical_and_expr(ctx *Logical_and_exprContext) {}

// ExitLogical_and_expr is called when production logical_and_expr is exited.
func (s *BaseNeuroScriptListener) ExitLogical_and_expr(ctx *Logical_and_exprContext) {}

// EnterBitwise_or_expr is called when production bitwise_or_expr is entered.
func (s *BaseNeuroScriptListener) EnterBitwise_or_expr(ctx *Bitwise_or_exprContext) {}

// ExitBitwise_or_expr is called when production bitwise_or_expr is exited.
func (s *BaseNeuroScriptListener) ExitBitwise_or_expr(ctx *Bitwise_or_exprContext) {}

// EnterBitwise_xor_expr is called when production bitwise_xor_expr is entered.
func (s *BaseNeuroScriptListener) EnterBitwise_xor_expr(ctx *Bitwise_xor_exprContext) {}

// ExitBitwise_xor_expr is called when production bitwise_xor_expr is exited.
func (s *BaseNeuroScriptListener) ExitBitwise_xor_expr(ctx *Bitwise_xor_exprContext) {}

// EnterBitwise_and_expr is called when production bitwise_and_expr is entered.
func (s *BaseNeuroScriptListener) EnterBitwise_and_expr(ctx *Bitwise_and_exprContext) {}

// ExitBitwise_and_expr is called when production bitwise_and_expr is exited.
func (s *BaseNeuroScriptListener) ExitBitwise_and_expr(ctx *Bitwise_and_exprContext) {}

// EnterEquality_expr is called when production equality_expr is entered.
func (s *BaseNeuroScriptListener) EnterEquality_expr(ctx *Equality_exprContext) {}

// ExitEquality_expr is called when production equality_expr is exited.
func (s *BaseNeuroScriptListener) ExitEquality_expr(ctx *Equality_exprContext) {}

// EnterRelational_expr is called when production relational_expr is entered.
func (s *BaseNeuroScriptListener) EnterRelational_expr(ctx *Relational_exprContext) {}

// ExitRelational_expr is called when production relational_expr is exited.
func (s *BaseNeuroScriptListener) ExitRelational_expr(ctx *Relational_exprContext) {}

// EnterAdditive_expr is called when production additive_expr is entered.
func (s *BaseNeuroScriptListener) EnterAdditive_expr(ctx *Additive_exprContext) {}

// ExitAdditive_expr is called when production additive_expr is exited.
func (s *BaseNeuroScriptListener) ExitAdditive_expr(ctx *Additive_exprContext) {}

// EnterMultiplicative_expr is called when production multiplicative_expr is entered.
func (s *BaseNeuroScriptListener) EnterMultiplicative_expr(ctx *Multiplicative_exprContext) {}

// ExitMultiplicative_expr is called when production multiplicative_expr is exited.
func (s *BaseNeuroScriptListener) ExitMultiplicative_expr(ctx *Multiplicative_exprContext) {}

// EnterUnary_expr is called when production unary_expr is entered.
func (s *BaseNeuroScriptListener) EnterUnary_expr(ctx *Unary_exprContext) {}

// ExitUnary_expr is called when production unary_expr is exited.
func (s *BaseNeuroScriptListener) ExitUnary_expr(ctx *Unary_exprContext) {}

// EnterPower_expr is called when production power_expr is entered.
func (s *BaseNeuroScriptListener) EnterPower_expr(ctx *Power_exprContext) {}

// ExitPower_expr is called when production power_expr is exited.
func (s *BaseNeuroScriptListener) ExitPower_expr(ctx *Power_exprContext) {}

// EnterAccessor_expr is called when production accessor_expr is entered.
func (s *BaseNeuroScriptListener) EnterAccessor_expr(ctx *Accessor_exprContext) {}

// ExitAccessor_expr is called when production accessor_expr is exited.
func (s *BaseNeuroScriptListener) ExitAccessor_expr(ctx *Accessor_exprContext) {}

// EnterPrimary is called when production primary is entered.
func (s *BaseNeuroScriptListener) EnterPrimary(ctx *PrimaryContext) {}

// ExitPrimary is called when production primary is exited.
func (s *BaseNeuroScriptListener) ExitPrimary(ctx *PrimaryContext) {}

// EnterFunction_call is called when production function_call is entered.
func (s *BaseNeuroScriptListener) EnterFunction_call(ctx *Function_callContext) {}

// ExitFunction_call is called when production function_call is exited.
func (s *BaseNeuroScriptListener) ExitFunction_call(ctx *Function_callContext) {}

// EnterPlaceholder is called when production placeholder is entered.
func (s *BaseNeuroScriptListener) EnterPlaceholder(ctx *PlaceholderContext) {}

// ExitPlaceholder is called when production placeholder is exited.
func (s *BaseNeuroScriptListener) ExitPlaceholder(ctx *PlaceholderContext) {}

// EnterLiteral is called when production literal is entered.
func (s *BaseNeuroScriptListener) EnterLiteral(ctx *LiteralContext) {}

// ExitLiteral is called when production literal is exited.
func (s *BaseNeuroScriptListener) ExitLiteral(ctx *LiteralContext) {}

// EnterBoolean_literal is called when production boolean_literal is entered.
func (s *BaseNeuroScriptListener) EnterBoolean_literal(ctx *Boolean_literalContext) {}

// ExitBoolean_literal is called when production boolean_literal is exited.
func (s *BaseNeuroScriptListener) ExitBoolean_literal(ctx *Boolean_literalContext) {}

// EnterList_literal is called when production list_literal is entered.
func (s *BaseNeuroScriptListener) EnterList_literal(ctx *List_literalContext) {}

// ExitList_literal is called when production list_literal is exited.
func (s *BaseNeuroScriptListener) ExitList_literal(ctx *List_literalContext) {}

// EnterMap_literal is called when production map_literal is entered.
func (s *BaseNeuroScriptListener) EnterMap_literal(ctx *Map_literalContext) {}

// ExitMap_literal is called when production map_literal is exited.
func (s *BaseNeuroScriptListener) ExitMap_literal(ctx *Map_literalContext) {}

// EnterExpression_list_opt is called when production expression_list_opt is entered.
func (s *BaseNeuroScriptListener) EnterExpression_list_opt(ctx *Expression_list_optContext) {}

// ExitExpression_list_opt is called when production expression_list_opt is exited.
func (s *BaseNeuroScriptListener) ExitExpression_list_opt(ctx *Expression_list_optContext) {}

// EnterExpression_list is called when production expression_list is entered.
func (s *BaseNeuroScriptListener) EnterExpression_list(ctx *Expression_listContext) {}

// ExitExpression_list is called when production expression_list is exited.
func (s *BaseNeuroScriptListener) ExitExpression_list(ctx *Expression_listContext) {}

// EnterMap_entry_list_opt is called when production map_entry_list_opt is entered.
func (s *BaseNeuroScriptListener) EnterMap_entry_list_opt(ctx *Map_entry_list_optContext) {}

// ExitMap_entry_list_opt is called when production map_entry_list_opt is exited.
func (s *BaseNeuroScriptListener) ExitMap_entry_list_opt(ctx *Map_entry_list_optContext) {}

// EnterMap_entry_list is called when production map_entry_list is entered.
func (s *BaseNeuroScriptListener) EnterMap_entry_list(ctx *Map_entry_listContext) {}

// ExitMap_entry_list is called when production map_entry_list is exited.
func (s *BaseNeuroScriptListener) ExitMap_entry_list(ctx *Map_entry_listContext) {}

// EnterMap_entry is called when production map_entry is entered.
func (s *BaseNeuroScriptListener) EnterMap_entry(ctx *Map_entryContext) {}

// ExitMap_entry is called when production map_entry is exited.
func (s *BaseNeuroScriptListener) ExitMap_entry(ctx *Map_entryContext) {}
