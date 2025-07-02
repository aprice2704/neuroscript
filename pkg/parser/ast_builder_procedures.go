// NeuroScript Version: 0.5.2
// File version: 4
// Purpose: Removed redundant block context creation to fix stack imbalance.
// filename: pkg/parser/ast_builder_procedures.go
// nlines: 70
// risk_rating: MEDIUM

package parser

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
)

func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ctx.IDENTIFIER().GetText()
	l.logDebugAST(">>> Enter Procedure_definition for %s", procName)

	pos := tokenToPosition(ctx.KW_FUNC().GetSymbol())
	l.currentProc = &ast.Procedure{
		Position:	pos,
		Metadata:	make(map[string]string),
	}
	l.currentProc.SetName(procName)

	// DO NOT create a block here. The non_empty_statement_list rule handles the block context.
}

func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := l.currentProc.Name()
	l.logDebugAST("<<< Exit Procedure_definition for %s", procName)

	// The procedure body is the slice of steps now on top of the value stack.
	if bodyRaw, ok := l.pop(); ok {
		if bodySteps, isSteps := bodyRaw.([]ast.Step); isSteps {
			// Separate 'on error' handlers from the main body.
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
			l.push(bodyRaw)	// Push back the wrong type
		}
	} else {
		l.addError(ctx, "stack underflow: could not pop procedure body for '%s'", procName)
	}

	l.finalizeProcedure(ctx)
}

func (l *neuroScriptListenerImpl) finalizeProcedure(ctx antlr.ParserRuleContext) {
	if l.currentProc != nil {
		// Directly add the completed procedure to the program's map.
		if _, exists := l.program.Procedures[l.currentProc.Name()]; exists {
			l.addError(ctx, "duplicate procedure definition: '%s'", l.currentProc.Name())
		} else {
			l.program.Procedures[l.currentProc.Name()] = l.currentProc
		}
		l.currentProc = nil
	}
}

// getRuleText is a helper function to get the full text of a parser rule context.
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