// filename: pkg/core/ast_builder_procedures.go
package core

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	// Assuming interfaces logger is imported via ast_builder_main or here if needed
)

// EnterProcedure_definition initializes the Procedure struct and processes the signature and metadata.
func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ""
	if id := ctx.IDENTIFIER(); id != nil {
		procName = id.GetText()
	} else {
		l.logger.Error("AST Builder: Procedure definition missing identifier!")
		l.addError(ctx, "procedure definition missing identifier")
		procName = "_INVALID_PROC_"
	}
	l.logDebugAST(">>> Enter Procedure_definition (func): %s", procName)

	procPos := tokenToPosition(ctx.GetStart())

	// 1. Initialize the current Procedure struct early
	l.currentProc = &Procedure{
		Pos:            procPos,
		Name:           procName,
		RequiredParams: []string{},
		OptionalParams: []string{},
		ReturnVarNames: []string{},
		Steps:          make([]Step, 0), // Initialize Steps slice
		Metadata:       make(map[string]string),
	}
	// ... (Capture OriginalSignature - unchanged) ...
	if sig := getRuleText(ctx); sig != "" {
		if meansToken := ctx.KW_MEANS(); meansToken != nil {
			startToken := ctx.GetStart()
			stopToken := meansToken.GetSymbol()
			stopIndex := stopToken.GetStart() - 1
			if stopIndex >= startToken.GetStart() {
				l.currentProc.OriginalSignature = startToken.GetInputStream().GetText(startToken.GetStart(), stopIndex)
			} else {
				l.currentProc.OriginalSignature = getRuleText(ctx.Signature_part())
			}
		} else {
			l.currentProc.OriginalSignature = getRuleText(ctx.Signature_part())
		}
	}

	// 2. Process Signature Part (unchanged)
	l.logDebugAST("      Checking for signature part context for %s", procName)
	if sigPartCtx := ctx.Signature_part(); sigPartCtx != nil {
		l.logDebugAST("         Found signature_part context. Processing clauses...")
		if needsCtx := sigPartCtx.Needs_clause(); needsCtx != nil {
			if paramListCtx := needsCtx.Param_list(); paramListCtx != nil {
				l.currentProc.RequiredParams = l.extractParamList(paramListCtx)
				l.logDebugAST("            Found 'needs' params: %v", l.currentProc.RequiredParams)
			}
		}
		if optionalCtx := sigPartCtx.Optional_clause(); optionalCtx != nil {
			if paramListCtx := optionalCtx.Param_list(); paramListCtx != nil {
				l.currentProc.OptionalParams = l.extractParamList(paramListCtx)
				l.logDebugAST("            Found 'optional' params: %v", l.currentProc.OptionalParams)
			}
		}
		if returnsCtx := sigPartCtx.Returns_clause(); returnsCtx != nil {
			if paramListCtx := returnsCtx.Param_list(); paramListCtx != nil {
				l.currentProc.ReturnVarNames = l.extractParamList(paramListCtx)
				l.logDebugAST("            Found 'returns' params: %v", l.currentProc.ReturnVarNames)
			}
		}
	} else {
		l.logDebugAST("         No signature_part context found (grammar/parse issue?).")
	}

	// 3. Process Metadata Block (if it exists) (unchanged)
	if metaBlockCtx := ctx.Metadata_block(); metaBlockCtx != nil {
		l.logDebugAST("      Processing procedure metadata block...")
		for _, metaLineNode := range metaBlockCtx.AllMETADATA_LINE() {
			if termNode, ok := metaLineNode.(antlr.TerminalNode); ok {
				// ... (metadata parsing logic - unchanged) ...
				fullLineText := termNode.GetText()
				token := termNode.GetSymbol()
				trimmedLine := strings.TrimSpace(fullLineText)
				if strings.HasPrefix(trimmedLine, "::") {
					content := strings.TrimSpace(trimmedLine[2:])
					parts := strings.SplitN(content, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						if key != "" {
							if l.currentProc.Metadata == nil {
								l.currentProc.Metadata = make(map[string]string)
							}
							l.currentProc.Metadata[key] = value
							l.logDebugAST("         Parsed proc metadata: [%s] = %q", key, value)
						} else {
							l.addErrorf(token, "Ignoring procedure metadata line with empty key")
						}
					} else {
						l.addErrorf(token, "Ignoring malformed procedure metadata line (missing ':'?)")
					}
				} else {
					l.addErrorf(token, "Unexpected procedure metadata line format (missing '::'?)")
				}
			}
		}
	} else {
		l.logDebugAST("      No procedure metadata block found.")
	}

	// 4. Setup for processing steps
	// *** DO NOT push base procedure steps onto blockStepStack ***
	if len(l.blockStepStack) != 0 { // Should always be empty here if previous proc exited cleanly
		l.logger.Error("Block stack not empty at START of procedure definition! Clearing.", "proc", procName, "size", len(l.blockStepStack))
		l.errors = append(l.errors, fmt.Errorf("internal AST builder error: block stack not empty at start of %s", procName))
		l.blockStepStack = l.blockStepStack[:0] // Clear it
	}
	l.currentSteps = &l.currentProc.Steps // Point currentSteps to the proc's main step slice
	l.logDebugAST("      Set currentSteps for procedure %s: %p", procName, l.currentSteps)

	// Clear value stack for the new procedure body
	l.valueStack = l.valueStack[:0]
}

// ExitProcedure_definition finalizes the procedure and adds it to the listener's temporary list.
// (Keep the version from the previous fix attempt - it checks stack is empty, sets currentSteps=nil)
func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := "(nil)"
	if l.currentProc != nil {
		procName = l.currentProc.Name
	}
	l.logDebugAST("<<< Exit Procedure_definition (func): %s", procName)

	if l.currentProc == nil {
		l.logger.Error("AST Builder: Cannot append procedure, currentProc is nil on exit.")
		l.currentSteps = nil
		return
	}

	// Check block stack state *before* finalizing procedure
	if len(l.blockStepStack) != 0 {
		l.logger.Error("Internal Error: Block stack not empty on exiting procedure definition", "procedure", procName, "size", len(l.blockStepStack))
		l.errors = append(l.errors, fmt.Errorf("internal AST builder error: block stack size is %d (should be 0) at end of procedure %s", len(l.blockStepStack), procName))
		l.blockStepStack = l.blockStepStack[:0] // Attempt recovery
		l.logger.Warn("Cleared non-empty block stack for recovery before finalizing procedure.", "procedure", procName)
	}

	l.logDebugAST("      Appending procedure pointer: %s (...)", procName)
	l.procedures = append(l.procedures, l.currentProc)

	l.currentProc = nil
	l.currentSteps = nil // <<< Set currentSteps to nil AFTER procedure is done

	// Value stack check
	if len(l.valueStack) > 0 {
		l.logger.Warn("Value stack not empty at end of procedure", "procedure", procName, "size", len(l.valueStack))
		l.errors = append(l.errors, fmt.Errorf("internal AST builder error: value stack not empty (%d elements) at end of procedure %s", len(l.valueStack), procName))
		l.valueStack = l.valueStack[:0]
	}
}

// --- Helper Methods --- (extractParamList, getRuleText unchanged)
func (l *neuroScriptListenerImpl) extractParamList(ctx gen.IParam_listContext) []string {
	if ctx == nil {
		return []string{}
	}
	params := []string{}
	identifiers := ctx.AllIDENTIFIER()
	if identifiers == nil {
		return []string{}
	}
	for _, identifier := range identifiers {
		if identifier != nil {
			params = append(params, identifier.GetText())
		}
	}
	return params
}
func getRuleText(ctx antlr.ParserRuleContext) string {
	if ctx == nil {
		return ""
	}
	startToken := ctx.GetStart()
	stopToken := ctx.GetStop()
	if startToken == nil || stopToken == nil || startToken.GetInputStream() == nil {
		return ""
	}
	startIndex := startToken.GetStart()
	stopIndex := stopToken.GetStop()
	if startIndex < 0 || stopIndex < 0 || stopIndex < startIndex {
		return ""
	}
	return startToken.GetInputStream().GetText(startIndex, stopIndex)
}
