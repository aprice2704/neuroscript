// Code generated from NeuroScript.g4 by ANTLR 4.13.2. DO NOT EDIT.

package core // NeuroScript
import "github.com/antlr4-go/antlr/v4"

type BaseNeuroScriptVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseNeuroScriptVisitor) VisitProgram(ctx *ProgramContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitOptional_newlines(ctx *Optional_newlinesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitProcedure_definition(ctx *Procedure_definitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitParam_list_opt(ctx *Param_list_optContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitParam_list(ctx *Param_listContext) interface{} {
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

func (v *BaseNeuroScriptVisitor) VisitIf_statement(ctx *If_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitWhile_statement(ctx *While_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitFor_each_statement(ctx *For_each_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitCall_target(ctx *Call_targetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitCondition(ctx *ConditionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitExpression(ctx *ExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitTerm(ctx *TermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitPrimary(ctx *PrimaryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitPlaceholder(ctx *PlaceholderContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseNeuroScriptVisitor) VisitLiteral(ctx *LiteralContext) interface{} {
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
