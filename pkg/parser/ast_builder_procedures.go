// filename: pkg/parser/ast_builder_procedures.go
// NeuroScript Version: 0.6.0
// File version: 16
// Purpose: Corrected ParamSpec creation to use the newNode helper with the correct, existing NodeKind (types.KindVariable).

package parser

import (
	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ctx.IDENTIFIER().GetText()
	l.logDebugAST(">>> Enter Procedure_definition for %s", procName)

	token := ctx.KW_FUNC().GetSymbol()
	proc := &ast.Procedure{
		Metadata: make(map[string]string),
		Comments: make([]*ast.Comment, 0),
		// BlankLinesBefore is now set by the LineInfo algorithm in the builder.
	}
	proc.SetName(procName)
	l.currentProc = newNode(proc, token, types.KindProcedureDecl)
	l.currentProc.Comments = l.associateCommentsToNode(l.currentProc)
}

func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := l.currentProc.Name()
	l.logDebugAST("<<< Exit Procedure_definition for %s", procName)

	if bodyRaw, ok := l.pop(); ok {
		if bodySteps, isSteps := bodyRaw.([]ast.Step); isSteps {
			var regularSteps []ast.Step
			for i := range bodySteps {
				step := bodySteps[i]
				if step.Type == "on_error" {
					l.currentProc.ErrorHandlers = append(l.currentProc.ErrorHandlers, &step)
				} else {
					regularSteps = append(regularSteps, step)
				}
			}
			l.currentProc.Steps = regularSteps
		} else {
			l.addErrorf(ctx.KW_ENDFUNC().GetSymbol(), "Type error: procedure body for '%s' is not []ast.Step (got %T).", procName, bodyRaw)
			l.push(bodyRaw)
		}
	} else {
		l.addError(ctx, "stack underflow: could not pop procedure body for '%s'", procName)
	}

	SetEndPos(l.currentProc, ctx.KW_ENDFUNC().GetSymbol())
	l.finalizeProcedure(ctx)
}

func (l *neuroScriptListenerImpl) finalizeProcedure(ctx antlr.ParserRuleContext) {
	if l.currentProc != nil {
		if _, exists := l.program.Procedures[l.currentProc.Name()]; exists {
			l.addError(ctx, "duplicate procedure definition: '%s'", l.currentProc.Name())
		} else {
			l.program.Procedures[l.currentProc.Name()] = l.currentProc
		}
		l.currentProc = nil
	}
}

func (l *neuroScriptListenerImpl) ExitParam_list(ctx *gen.Param_listContext) {
	params := make([]string, 0, len(ctx.AllIDENTIFIER()))
	for _, ident := range ctx.AllIDENTIFIER() {
		params = append(params, ident.GetText())
	}
	l.push(params)
}

func (l *neuroScriptListenerImpl) ExitNeeds_clause(ctx *gen.Needs_clauseContext) {
	if l.currentProc == nil {
		l.addError(ctx, "found 'needs' clause outside of a procedure definition")
		return
	}
	val, ok := l.pop()
	if !ok {
		l.addError(ctx, "stack underflow reading params for 'needs' clause")
		return
	}
	params, ok := val.([]string)
	if !ok {
		l.addError(ctx, "internal error: 'needs' clause expected []string from stack, got %T", val)
		return
	}
	l.currentProc.RequiredParams = append(l.currentProc.RequiredParams, params...)
}

func (l *neuroScriptListenerImpl) ExitOptional_clause(ctx *gen.Optional_clauseContext) {
	if l.currentProc == nil {
		l.addError(ctx, "found 'optional' clause outside of a procedure definition")
		return
	}
	// We pop the value pushed by ExitParam_list to keep the stack clean, but we
	// don't use it. We get the identifiers directly from the context to have
	// access to their tokens for proper node initialization.
	val, ok := l.pop()
	if !ok {
		l.addError(ctx, "stack underflow reading params for 'optional' clause")
		return
	}
	if _, ok := val.([]string); !ok {
		l.addError(ctx, "internal error: 'optional' clause expected []string from stack, got %T", val)
		return
	}

	// Re-get the identifiers from the context to access their tokens.
	paramListCtx := ctx.Param_list()
	if paramListCtx == nil {
		return // No optional parameters present.
	}

	for _, identNode := range paramListCtx.AllIDENTIFIER() {
		// FIX: Create the ParamSpec using the newNode helper to ensure its
		// BaseNode (and thus NodeKind and StartPos) is fully initialized.
		param := &ast.ParamSpec{Name: identNode.GetText()}
		// A parameter is a form of variable, so KindVariable is the correct type.
		newNode(param, identNode.GetSymbol(), types.KindVariable)
		l.currentProc.OptionalParams = append(l.currentProc.OptionalParams, param)
	}
}

func (l *neuroScriptListenerImpl) ExitReturns_clause(ctx *gen.Returns_clauseContext) {
	if l.currentProc == nil {
		l.addError(ctx, "found 'returns' clause outside of a procedure definition")
		return
	}
	val, ok := l.pop()
	if !ok {
		l.addError(ctx, "stack underflow reading params for 'returns' clause")
		return
	}
	params, ok := val.([]string)
	if !ok {
		l.addError(ctx, "internal error: 'returns' clause expected []string from stack, got %T", val)
		return
	}
	l.currentProc.ReturnVarNames = append(l.currentProc.ReturnVarNames, params...)
}
