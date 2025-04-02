// Code generated from NeuroScript.g4 by ANTLR 4.13.2. DO NOT EDIT.

package core // NeuroScript
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

// EnterCondition is called when production condition is entered.
func (s *BaseNeuroScriptListener) EnterCondition(ctx *ConditionContext) {}

// ExitCondition is called when production condition is exited.
func (s *BaseNeuroScriptListener) ExitCondition(ctx *ConditionContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseNeuroScriptListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseNeuroScriptListener) ExitExpression(ctx *ExpressionContext) {}

// EnterTerm is called when production term is entered.
func (s *BaseNeuroScriptListener) EnterTerm(ctx *TermContext) {}

// ExitTerm is called when production term is exited.
func (s *BaseNeuroScriptListener) ExitTerm(ctx *TermContext) {}

// EnterPrimary is called when production primary is entered.
func (s *BaseNeuroScriptListener) EnterPrimary(ctx *PrimaryContext) {}

// ExitPrimary is called when production primary is exited.
func (s *BaseNeuroScriptListener) ExitPrimary(ctx *PrimaryContext) {}

// EnterPlaceholder is called when production placeholder is entered.
func (s *BaseNeuroScriptListener) EnterPlaceholder(ctx *PlaceholderContext) {}

// ExitPlaceholder is called when production placeholder is exited.
func (s *BaseNeuroScriptListener) ExitPlaceholder(ctx *PlaceholderContext) {}

// EnterLiteral is called when production literal is entered.
func (s *BaseNeuroScriptListener) EnterLiteral(ctx *LiteralContext) {}

// ExitLiteral is called when production literal is exited.
func (s *BaseNeuroScriptListener) ExitLiteral(ctx *LiteralContext) {}

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
