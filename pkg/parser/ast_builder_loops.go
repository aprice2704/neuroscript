// filename: pkg/core/ast_builder_loops.go
// NeuroScript Version: 0.5.2
// File version: 15
// Purpose: Removed event and error handler logic, which is now centralized in ast_builder_events.go.
// nlines: 90
// risk_rating: MEDIUM

package parser

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

// --- Loop Statements ---

func (l *neuroScriptListenerImpl) EnterWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST(">>> EnterWhile_statement")
	l.loopDepth++
}

func (l *neuroScriptListenerImpl) ExitWhile_statement(ctx *gen.While_statementContext) {
	l.logDebugAST("<<< ExitWhile_statement")
	l.loopDepth--

	bodyVal, ok := l.poplang.Value()
	if !ok {
		l.addError(ctx, "stack underflow in while statement body")
		return
	}
	body, ok := bodyVal.([]ast.Step)
	if !ok {
		l.addError(ctx, "while statement expected a valid body on the stack, got %T", bodyVal)
		return
	}

	condVal, ok := l.poplang.Value()
	if !ok {
		l.addError(ctx, "stack underflow in while statement condition")
		return
	}
	cond, ok := condVal.(ast.Expression)
	if !ok {
		l.addError(ctx, "while statement expected a valid condition on the stack")
		return
	}

	stmt := ast.Step{
		Position: tokenTolang.Position(ctx.GetStart()),
		Type:     "while",
		Cond:     cond,
		Body:     body,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

func (l *neuroScriptListenerImpl) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST(">>> EnterFor_each_statement")
	l.loopDepth++
}

func (l *neuroScriptListenerImpl) ExitFor_each_statement(ctx *gen.For_each_statementContext) {
	l.logDebugAST("<<< ExitFor_each_statement")
	l.loopDepth--

	bodyVal, ok := l.poplang.Value()
	if !ok {
		l.addError(ctx, "stack underflow in for..each body")
		return
	}
	body, ok := bodyVal.([]ast.Step)
	if !ok {
		l.addError(ctx, "for..each statement expected a valid body on the stack, got %T", bodyVal)
		return
	}

	collectionVal, ok := l.poplang.Value()
	if !ok {
		l.addError(ctx, "stack underflow in for..each collection")
		return
	}
	collection, ok := collectionVal.(ast.Expression)
	if !ok {
		l.addError(ctx, "for..each statement expected a valid collection on the stack")
		return
	}

	stmt := ast.Step{
		Position:    tokenTolang.Position(ctx.GetStart()),
		Type:        "for",
		LoopVarName: ctx.IDENTIFIER().GetText(),
		Collection:  collection,
		Body:        body,
	}
	*l.currentSteps = append(*l.currentSteps, stmt)
}

// --- Other Required Stubs ---

func (l *neuroScriptListenerImpl) EnterSignature_part(c *gen.Signature_partContext) {}
func (l *neuroScriptListenerImpl) ExitSignature_part(c *gen.Signature_partContext)  {}
func (l *neuroScriptListenerImpl) EnterNeeds_clause(c *gen.Needs_clauseContext)     {}
func (l *neuroScriptListenerImpl) ExitNeeds_clause(c *gen.Needs_clauseContext)      {}
func (l *neuroScriptListenerImpl) EnterOptional_clause(c *gen.Optional_clauseContext) {
}
func (l *neuroScriptListenerImpl) ExitOptional_clause(c *gen.Optional_clauseContext) {}
func (l *neuroScriptListenerImpl) EnterReturns_clause(c *gen.Returns_clauseContext)  {}
func (l *neuroScriptListenerImpl) ExitReturns_clause(c *gen.Returns_clauseContext)   {}
func (l *neuroScriptListenerImpl) EnterParam_list(c *gen.Param_listContext)          {}
func (l *neuroScriptListenerImpl) ExitParam_list(c *gen.Param_listContext)           {}
func (l *neuroScriptListenerImpl) EnterMetadata_block(c *gen.Metadata_blockContext)  {}
func (l *neuroScriptListenerImpl) EnterBody_line(c *gen.Body_lineContext)            {}
func (l *neuroScriptListenerImpl) ExitBody_line(c *gen.Body_lineContext)             {}
func (l *neuroScriptListenerImpl) EnterStatement(c *gen.StatementContext)            {}
func (l *neuroScriptListenerImpl) ExitStatement(c *gen.StatementContext)             {}
func (l *neuroScriptListenerImpl) EnterSimple_statement(c *gen.Simple_statementContext) {
}
func (l *neuroScriptListenerImpl) ExitSimple_statement(c *gen.Simple_statementContext) {
}
func (l *neuroScriptListenerImpl) EnterBlock_statement(c *gen.Block_statementContext) {
}
func (l *neuroScriptListenerImpl) ExitBlock_statement(c *gen.Block_statementContext) {}
func (l *neuroScriptListenerImpl) EnterLvalue(c *gen.LvalueContext)                  {}
func (l *neuroScriptListenerImpl) EnterIf_statement(c *gen.If_statementContext)      {}
