// NeuroScript Version: 0.3.1
// File version: 0.0.8 // Corrected ANTLR Token method calls in getRuleText.
// Last Modified: 2025-05-27
package core

import (
	"fmt"

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
		if needs := sigCtx.Needs_clause(); needs != nil && needs.Param_list() != nil {
			l.currentProc.RequiredParams = l.extractParamNamesList(needs.Param_list())
		}
		if optional := sigCtx.Optional_clause(); optional != nil && optional.Param_list() != nil {
			paramNames := l.extractParamNamesList(optional.Param_list())
			for _, name := range paramNames {
				l.currentProc.OptionalParams = append(l.currentProc.OptionalParams, ParamSpec{Name: name})
			}
		}
		if returns := sigCtx.Returns_clause(); returns != nil && returns.Param_list() != nil {
			l.currentProc.ReturnVarNames = l.extractParamNamesList(returns.Param_list())
		}
	}

	if metaBlockCtx := ctx.Metadata_block(); metaBlockCtx != nil {
		for _, metaLineToken := range metaBlockCtx.AllMETADATA_LINE() {
			line := metaLineToken.GetText()
			key, value, ok := ParseMetadataLine(line)
			if ok {
				l.currentProc.Metadata[key] = value
			} else {
				l.addErrorf(metaLineToken.GetSymbol(), "Malformed metadata line: %s", line)
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
		l.addError(ctx, "Exiting procedure definition '%s' but no current procedure context (l.currentProc is nil). This indicates a prior critical error.", procName)
		return
	}
	if l.currentProc.Name != procName {
		l.addError(ctx, "Exiting procedure definition '%s', but l.currentProc.Name is '%s'. Mismatch!", procName, l.currentProc.Name)
	}

	bodyStepsRaw, ok := l.popValue()
	if !ok {
		l.addErrorf(ctx.KW_ENDFUNC().GetSymbol(), "Stack error: expected procedure body steps for '%s', but value stack was empty.", l.currentProc.Name)
		l.currentProc.Steps = []Step{}
		l.finalizeProcedure(ctx)
		return
	}

	bodySteps, isSteps := bodyStepsRaw.([]Step)
	if !isSteps {
		l.addErrorf(ctx.KW_ENDFUNC().GetSymbol(), "Type error: procedure body for '%s' is not []Step (got %T). Value: %v", l.currentProc.Name, bodyStepsRaw, bodyStepsRaw)
		l.pushValue(bodyStepsRaw)
		l.currentProc.Steps = []Step{}
		l.finalizeProcedure(ctx)
		return
	}
	l.currentProc.Steps = bodySteps
	l.logDebugAST("       Popped %d body steps for procedure %s. Value stack size after body pop: %d", len(l.currentProc.Steps), l.currentProc.Name, len(l.valueStack))

	if sigCtx := ctx.Signature_part(); sigCtx != nil && l.currentProc.OriginalSignature == "" {
		l.currentProc.OriginalSignature = getRuleText(sigCtx)
	}

	l.finalizeProcedure(ctx)
	l.logDebugAST("<<< Exited Procedure_definition for %s", procName)
}

// getRuleText is a helper to get the full text of a parser rule context.
// It uses standard ANTLR Go runtime API methods.
func getRuleText(ctx antlr.ParserRuleContext) string {
	if ctx == nil || ctx.GetStart() == nil || ctx.GetStop() == nil {
		return ""
	}
	startToken := ctx.GetStart() // antlr.Token
	stopToken := ctx.GetStop()   // antlr.Token
	inputStream := startToken.GetInputStream()
	if inputStream == nil {
		return ""
	}

	// Corrected to use GetStart() and GetStop() as per the antlr4-go/antlr Token API
	// These methods directly return the start and stop character indices.
	return inputStream.GetText(startToken.GetStart(), stopToken.GetStop())
}

func (l *neuroScriptListenerImpl) finalizeProcedure(procedureRuleCtx antlr.RuleContext) {
	if l.currentProc == nil {
		l.logger.Error("finalizeProcedure called but l.currentProc is nil. This is unexpected.")
		return
	}

	if len(l.valueStack) != 0 {
		errMsg := fmt.Sprintf("internal AST builder error: value stack not empty (%d elements) at end of procedure '%s'", len(l.valueStack), l.currentProc.Name)
		l.logger.Error(errMsg)

		for i, item := range l.valueStack {
			l.logger.Debug(fmt.Sprintf("    Stack item %d: %T - %v", i, item, item))
		}

		var errorToken antlr.Token
		if prc, ok := procedureRuleCtx.(antlr.ParserRuleContext); ok {
			errorToken = prc.GetStart()
		} else {
			l.logger.Warn(fmt.Sprintf("finalizeProcedure: procedureRuleCtx of type %T was not an antlr.ParserRuleContext", procedureRuleCtx))
		}

		if errorToken != nil {
			l.addErrorf(errorToken, "%s", errMsg)
		} else {
			if l.errors == nil {
				l.errors = make([]error, 0)
			}
			l.errors = append(l.errors, fmt.Errorf("%s", errMsg+" (procedure context start token unavailable for precise error location)"))
		}
		l.valueStack = []interface{}{}
	}

	l.procedures = append(l.procedures, l.currentProc)
	l.logDebugAST("   Added procedure %s to list. Total procedures: %d", l.currentProc.Name, len(l.procedures))
	l.currentProc = nil

	if len(l.blockStepStack) == 0 {
		l.currentSteps = nil
	} else {
		l.logger.Warn("finalizeProcedure: blockStepStack is not empty (size %d). This may indicate an issue with block context management.", len(l.blockStepStack))
	}
}

// Assume ParseMetadataLine and tokenToPosition are defined elsewhere
// e.g., in ast_builder_helpers.go or ast_builder_main.go

// func ParseMetadataLine(line string) (key, value string, ok bool) { ... }
// func tokenToPosition(token antlr.Token) *Position { ... }
