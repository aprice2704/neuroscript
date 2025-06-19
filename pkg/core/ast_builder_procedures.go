// NeuroScript Version: 0.3.1
// File version: 4
// Purpose: Correctly iterate through all signature clauses regardless of order.
// filename: pkg/core/ast_builder_procedures.go
// nlines: 150
// risk_rating: LOW

package core

import (
	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ctx.IDENTIFIER().GetText()
	l.logDebugAST(">>> Enter Procedure_definition for %s. Parent currentSteps: %p, Stack depth: %d", procName, l.currentSteps, len(l.blockStepStack))

	l.currentProc = &Procedure{
		Pos:      tokenToPosition(ctx.KW_FUNC().GetSymbol()),
		Name:     procName,
		Metadata: make(map[string]string),
	}

	if sigCtx := ctx.Signature_part(); sigCtx != nil {
		l.currentProc.OriginalSignature = getRuleText(sigCtx)

		// Corrected: Iterate through all children of the signature to handle clauses in any order.
		for _, child := range sigCtx.GetChildren() {
			switch c := child.(type) {
			case *gen.Needs_clauseContext:
				if c.Param_list() != nil {
					l.currentProc.RequiredParams = append(l.currentProc.RequiredParams, l.extractParamNamesList(c.Param_list())...)
				}
			case *gen.Optional_clauseContext:
				if c.Param_list() != nil {
					paramNames := l.extractParamNamesList(c.Param_list())
					for _, name := range paramNames {
						l.currentProc.OptionalParams = append(l.currentProc.OptionalParams, ParamSpec{Name: name})
					}
				}
			case *gen.Returns_clauseContext:
				if c.Param_list() != nil {
					l.currentProc.ReturnVarNames = append(l.currentProc.ReturnVarNames, l.extractParamNamesList(c.Param_list())...)
				}
			}
		}
	}

	if metaBlockCtx := ctx.Metadata_block(); metaBlockCtx != nil {
		for _, metaLineToken := range metaBlockCtx.AllMETADATA_LINE() {
			line := metaLineToken.GetText()
			key, value, ok := ParseMetadataLine(line)
			if ok {
				l.currentProc.Metadata[key] = value
			} else {
				l.addErrorf(metaLineToken.GetSymbol(), "Malformed metadata line in procedure: %s", line)
			}
		}
	}
}

func (l *neuroScriptListenerImpl) extractParamNamesList(paramListCtx gen.IParam_listContext) []string {
	var names []string
	if paramListCtx == nil {
		return names
	}
	for _, idToken := range paramListCtx.AllIDENTIFIER() {
		names = append(names, idToken.GetText())
	}
	return names
}

func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ctx.IDENTIFIER().GetText()
	l.logDebugAST("--- ExitProcedure_definition for %s. Value stack size before body pop: %d", procName, len(l.valueStack))

	if l.currentProc == nil {
		l.addError(ctx, "Exiting procedure definition '%s' but no current procedure context (l.currentProc is nil).", procName)
		return
	}

	bodyRaw, ok := l.popValue()
	if !ok {
		l.addErrorf(ctx.KW_ENDFUNC().GetSymbol(), "Stack error: expected procedure body for '%s', but value stack was empty.", l.currentProc.Name)
		l.finalizeProcedure()
		return
	}

	// BUG FIX: Handle cases where the procedure body is a single expression, not a block.
	if expr, isExpr := bodyRaw.(Expression); isExpr {
		// This is a "bodiless" function that is just an expression.
		// We create an implicit return step for it.
		returnStep := Step{
			Pos:    expr.GetPos(),
			Type:   "return",
			Values: []Expression{expr},
		}
		l.currentProc.Steps = []Step{returnStep}
	} else if bodySteps, isSteps := bodyRaw.([]Step); isSteps {
		// This is the normal case with a block of statements.
		l.currentProc.Steps = bodySteps
	} else {
		l.addErrorf(ctx.KW_ENDFUNC().GetSymbol(), "Type error: procedure body for '%s' is not []Step or Expression (got %T).", l.currentProc.Name, bodyRaw)
		l.pushValue(bodyRaw) // Push back the wrong type
	}

	if sigCtx := ctx.Signature_part(); sigCtx != nil && l.currentProc.OriginalSignature == "" {
		l.currentProc.OriginalSignature = getRuleText(sigCtx)
	}

	l.finalizeProcedure()
	l.logDebugAST("<<< Exited Procedure_definition for %s", procName)
}

func getRuleText(ctx antlr.ParserRuleContext) string {
	if ctx == nil || ctx.GetStart() == nil || ctx.GetStop() == nil {
		return ""
	}
	startToken := ctx.GetStart()
	stopToken := ctx.GetStop()
	inputStream := startToken.GetInputStream()

	if inputStream == nil {
		return ""
	}

	startIndex := startToken.GetStart()
	stopIndex := stopToken.GetStop()

	if startIndex > stopIndex {
		return ""
	}

	return inputStream.GetText(startIndex, stopIndex)
}

func (l *neuroScriptListenerImpl) finalizeProcedure() {
	if l.currentProc == nil {
		l.logger.Error("finalizeProcedure called but l.currentProc is nil. This is unexpected.")
		return
	}

	l.procedures = append(l.procedures, l.currentProc)
	l.logDebugAST("   Added procedure %s to list. Total procedures: %d", l.currentProc.Name, len(l.procedures))
	l.currentProc = nil
}
