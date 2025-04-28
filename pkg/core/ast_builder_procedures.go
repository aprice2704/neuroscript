// pkg/core/ast_builder_procedures.go
package core

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

// --- Procedure Definition Handling (v0.2.0) ---

func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ""
	if id := ctx.IDENTIFIER(); id != nil {
		procName = id.GetText()
	} else {
		l.logger.Error("AST Builder: Procedure definition missing identifier!")
		procName = "_INVALID_PROC_"
	}
	l.logDebugAST(">>> Enter Procedure_definition (func): %s", procName)

	requiredParams := []string{}
	if needsCtx := ctx.Needs_clause(); needsCtx != nil {
		if paramList := needsCtx.Param_list(); paramList != nil {
			for _, id := range paramList.AllIDENTIFIER() {
				requiredParams = append(requiredParams, id.GetText())
			}
		}
	}
	optionalParams := []string{}
	if optionalCtx := ctx.Optional_clause(); optionalCtx != nil {
		if paramList := optionalCtx.Param_list(); paramList != nil {
			for _, id := range paramList.AllIDENTIFIER() {
				optionalParams = append(optionalParams, id.GetText())
			}
		}
	}
	returnVars := []string{}
	if returnsCtx := ctx.Returns_clause(); returnsCtx != nil {
		if paramList := returnsCtx.Param_list(); paramList != nil {
			for _, id := range paramList.AllIDENTIFIER() {
				returnVars = append(returnVars, id.GetText())
			}
		}
	}

	// FIX: Initialize Metadata map
	procMetadata := make(map[string]string)

	// FIX: Process the metadata block if it exists
	if metaBlockCtx := ctx.Metadata_block(); metaBlockCtx != nil {
		l.logDebugAST("    Processing metadata block...")
		for _, metaLineNode := range metaBlockCtx.AllMETADATA_LINE() {
			if termNode, ok := metaLineNode.(antlr.TerminalNode); ok {
				fullLineText := termNode.GetText()
				// The lexer rule now captures the line starting AFTER the ':: ' part
				// Let's refine the lexer rule slightly and adjust parsing here.
				// Assume lexer rule METADATA_LINE captures the full line `:: key: value`
				// We need to strip `:: ` and then parse.
				trimmedLine := strings.TrimSpace(fullLineText)
				if strings.HasPrefix(trimmedLine, "::") {
					// Strip '::' and leading space
					content := strings.TrimSpace(trimmedLine[2:])
					key, value, ok := parseMetadataLine(content) // Use helper from utils.go
					if ok {
						procMetadata[key] = value
						l.logDebugAST("      Parsed metadata: [%s] = %q", key, value)
					} else {
						l.logger.Warn("Malformed metadata line ignored: %s", fullLineText)
						// Optionally add an error to l.errors here
					}
				} else {
					l.logger.Warn("Unexpected metadata line format (missing '::'?): %s", fullLineText)
				}
			}
		}
	}

	l.currentProc = &Procedure{
		Name:              procName,
		RequiredParams:    requiredParams,
		OptionalParams:    optionalParams,
		ReturnVarNames:    returnVars,
		Metadata:          procMetadata, // FIX: Assign parsed metadata
		Steps:             make([]Step, 0),
		OriginalSignature: ctx.GetText(),
	}

	l.currentSteps = &l.currentProc.Steps
	l.valueStack = l.valueStack[:0] // Clear stack for new procedure body
}

// --- (Rest of the file remains the same) ---

func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := "(nil)"
	if l.currentProc != nil {
		procName = l.currentProc.Name
	}
	l.logDebugAST("<<< Exit Procedure_definition (func): %s", procName)
	if l.currentProc != nil {
		// Log metadata count for debugging
		l.logDebugAST("    Appending procedure: %s (ReqParams: %d, OptParams: %d, Returns: %d, Steps: %d, Metadata: %d)",
			l.currentProc.Name, len(l.currentProc.RequiredParams), len(l.currentProc.OptionalParams),
			len(l.currentProc.ReturnVarNames), len(l.currentProc.Steps), len(l.currentProc.Metadata))
		l.procedures = append(l.procedures, *l.currentProc)
	} else {
		l.logger.Error("AST Builder: Cannot append procedure, currentProc is nil on exit.")
	}
	l.currentProc = nil
	l.currentSteps = nil
	if len(l.valueStack) > 0 {
		l.logger.Warn("Value stack not empty at end of procedure %s (Size: %d)", procName, len(l.valueStack))
		l.valueStack = l.valueStack[:0]
	}
	if len(l.blockStepStack) > 0 {
		l.logger.Error("Block step stack not empty at end of procedure %s (Size: %d)", procName, len(l.blockStepStack))
		l.blockStepStack = l.blockStepStack[:0]
	}
}

// Clause/Param list enter/exit methods remain empty stubs
func (l *neuroScriptListenerImpl) EnterNeeds_clause(ctx *gen.Needs_clauseContext)       {}
func (l *neuroScriptListenerImpl) ExitNeeds_clause(ctx *gen.Needs_clauseContext)        {}
func (l *neuroScriptListenerImpl) EnterOptional_clause(ctx *gen.Optional_clauseContext) {}
func (l *neuroScriptListenerImpl) ExitOptional_clause(ctx *gen.Optional_clauseContext)  {}
func (l *neuroScriptListenerImpl) EnterReturns_clause(ctx *gen.Returns_clauseContext)   {}
func (l *neuroScriptListenerImpl) ExitReturns_clause(ctx *gen.Returns_clauseContext)    {}
func (l *neuroScriptListenerImpl) EnterParam_list(ctx *gen.Param_listContext)           {}
func (l *neuroScriptListenerImpl) ExitParam_list(ctx *gen.Param_listContext)            {}
