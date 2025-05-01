// filename: pkg/core/ast_builder_procedures.go
package core

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	// Assuming interfaces logger is imported via ast_builder_main or here if needed
)

// --- Procedure Definition Handling (v0.2.0 - signature_part rule) ---

// EnterProcedure_definition initializes the Procedure struct and processes the signature and metadata.
func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ""
	if id := ctx.IDENTIFIER(); id != nil {
		procName = id.GetText()
	} else {
		l.logger.Error("AST Builder: Procedure definition missing identifier!")
		l.addError(ctx, "procedure definition missing identifier") // Add error
		procName = "_INVALID_PROC_"                                // Use a placeholder name
	}
	l.logDebugAST(">>> Enter Procedure_definition (func): %s", procName)

	// Set position for the procedure
	procPos := tokenToPosition(ctx.GetStart()) // Use start of 'func' keyword

	// 1. Initialize the current Procedure struct early
	l.currentProc = &Procedure{
		Pos:            procPos, // Set position
		Name:           procName,
		RequiredParams: []string{}, // Initialize as empty
		OptionalParams: []string{}, // Initialize as empty
		ReturnVarNames: []string{}, // Initialize as empty
		Steps:          make([]Step, 0),
		Metadata:       make(map[string]string), // Initialize metadata map
	}
	// Capture original signature text using helper (includes 'func' keyword and 'means')
	// We might want a more precise capture later if needed.
	if sig := getRuleText(ctx); sig != "" {
		// Attempt to capture just the signature part before 'means' for better representation
		if meansToken := ctx.KW_MEANS(); meansToken != nil {
			startToken := ctx.GetStart()
			stopToken := meansToken.GetSymbol()
			// Adjust stop index to be just before 'means'
			stopIndex := stopToken.GetStart() - 1
			if stopIndex >= startToken.GetStart() {
				l.currentProc.OriginalSignature = startToken.GetInputStream().GetText(startToken.GetStart(), stopIndex)
			} else {
				// Fallback if something is weird with indices
				l.currentProc.OriginalSignature = getRuleText(ctx.Signature_part())
			}
		} else {
			// Fallback if KW_MEANS isn't found somehow (grammar error?)
			l.currentProc.OriginalSignature = getRuleText(ctx.Signature_part())
		}
	}

	// 2. Process Signature Part (Check for signature_part context first)
	l.logDebugAST("    Checking for signature part context for %s", procName)

	// --- Access clauses via intermediate signature_part context ---
	// Use Signature_part() instead of Parameter_clauses()
	if sigPartCtx := ctx.Signature_part(); sigPartCtx != nil {
		l.logDebugAST("      Found signature_part context. Processing clauses...")
		// Access individual clauses via sigPartCtx
		if needsCtx := sigPartCtx.Needs_clause(); needsCtx != nil {
			// Ensure Param_list() is not nil before accessing (robustness)
			if paramListCtx := needsCtx.Param_list(); paramListCtx != nil {
				l.currentProc.RequiredParams = l.extractParamList(paramListCtx)
				l.logDebugAST("        Found 'needs' params: %v", l.currentProc.RequiredParams)
			}
		}
		if optionalCtx := sigPartCtx.Optional_clause(); optionalCtx != nil {
			// Ensure Param_list() is not nil
			if paramListCtx := optionalCtx.Param_list(); paramListCtx != nil {
				l.currentProc.OptionalParams = l.extractParamList(paramListCtx)
				l.logDebugAST("        Found 'optional' params: %v", l.currentProc.OptionalParams)
			}
		}
		if returnsCtx := sigPartCtx.Returns_clause(); returnsCtx != nil {
			// Ensure Param_list() is not nil
			if paramListCtx := returnsCtx.Param_list(); paramListCtx != nil {
				l.currentProc.ReturnVarNames = l.extractParamList(paramListCtx)
				l.logDebugAST("        Found 'returns' params: %v", l.currentProc.ReturnVarNames)
			}
		}
		// Note: The signature_part rule itself handles the structure (parens vs no parens).
		// We only need to check for the presence of the individual clause contexts here.
	} else {
		// This case should be less likely now as signature_part is not optional '?'
		// in the grammar rule `procedure_definition`, but handle defensively.
		l.logDebugAST("      No signature_part context found (grammar/parse issue?).")
		// Slices are already initialized empty, so nothing more needed here.
	}
	// --- END Signature Part Processing ---

	// 3. Process Metadata Block (if it exists) - Unchanged
	if metaBlockCtx := ctx.Metadata_block(); metaBlockCtx != nil {
		l.logDebugAST("    Processing procedure metadata block...")
		for _, metaLineNode := range metaBlockCtx.AllMETADATA_LINE() {
			if termNode, ok := metaLineNode.(antlr.TerminalNode); ok {
				fullLineText := termNode.GetText()
				token := termNode.GetSymbol() // Get token for error reporting
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
							l.logDebugAST("      Parsed proc metadata: [%s] = %q", key, value)
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
		l.logDebugAST("    No procedure metadata block found.")
	}

	// 4. Setup for processing steps - Unchanged
	l.currentSteps = &l.currentProc.Steps
	l.blockStepStack = append(l.blockStepStack, l.currentSteps)
	l.logDebugAST("    Pushed procedure step block onto stack (Stack size: %d)", len(l.blockStepStack))

	// Clear value stack for the new procedure body - Unchanged
	l.valueStack = l.valueStack[:0]
}

// ExitProcedure_definition finalizes the procedure and adds it to the listener's temporary list.
// --- Unchanged ---
func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := "(nil)"
	if l.currentProc != nil {
		procName = l.currentProc.Name
	}
	l.logDebugAST("<<< Exit Procedure_definition (func): %s", procName)

	if l.currentProc == nil {
		l.logger.Error("AST Builder: Cannot append procedure, currentProc is nil on exit.")
		if len(l.blockStepStack) > 0 {
			l.blockStepStack = l.blockStepStack[:len(l.blockStepStack)-1]
		}
		return // Cannot proceed
	}

	if len(l.blockStepStack) > 0 {
		if l.currentSteps != l.blockStepStack[len(l.blockStepStack)-1] {
			l.logger.Error("Internal Error: Block stack mismatch on exiting procedure", "procedure_name", procName)
			l.errors = append(l.errors, fmt.Errorf("internal AST builder error: block stack mismatch for procedure %s", procName))
		}
		l.blockStepStack = l.blockStepStack[:len(l.blockStepStack)-1]
		l.logDebugAST("    Popped procedure step block from stack (Stack size: %d)", len(l.blockStepStack))
	} else {
		l.logger.Error("Internal Error: Attempted to pop from empty block stack on exiting procedure", "procedure_name", procName)
		l.errors = append(l.errors, fmt.Errorf("internal AST builder error: empty block stack for procedure %s", procName))
	}

	l.logDebugAST("    Appending procedure pointer: %s (ReqParams: %d, OptParams: %d, Returns: %d, Steps: %d, Metadata: %d)",
		l.currentProc.Name, len(l.currentProc.RequiredParams), len(l.currentProc.OptionalParams),
		len(l.currentProc.ReturnVarNames), len(l.currentProc.Steps), len(l.currentProc.Metadata))

	l.procedures = append(l.procedures, l.currentProc)

	l.currentProc = nil
	if len(l.blockStepStack) > 0 {
		l.currentSteps = l.blockStepStack[len(l.blockStepStack)-1]
	} else {
		l.currentSteps = nil
	}

	if len(l.valueStack) > 0 {
		l.logger.Warn("Value stack not empty at end of procedure", "procedure", procName, "size", len(l.valueStack))
		l.errors = append(l.errors, fmt.Errorf("internal AST builder error: value stack not empty (%d elements) at end of procedure %s", len(l.valueStack), procName))
		l.valueStack = l.valueStack[:0]
	}
}

// --- Helper Methods ---

// extractParamList extracts a slice of strings from a Param_listContext.
// --- Unchanged ---
func (l *neuroScriptListenerImpl) extractParamList(ctx gen.IParam_listContext) []string {
	if ctx == nil {
		return []string{} // Return empty slice if context is nil
	}
	params := []string{}
	// Check if AllIDENTIFIER itself might be nil in some parse error cases
	identifiers := ctx.AllIDENTIFIER()
	if identifiers == nil {
		return []string{}
	}
	for _, identifier := range identifiers {
		if identifier != nil { // Also check individual identifiers
			params = append(params, identifier.GetText())
		}
	}
	return params
}

// getRuleText safely gets the text for a context using ctx.GetText().
// --- Unchanged ---
func getRuleText(ctx antlr.ParserRuleContext) string {
	if ctx == nil {
		return ""
	}
	startToken := ctx.GetStart()
	stopToken := ctx.GetStop()
	if startToken == nil || stopToken == nil || startToken.GetInputStream() == nil {
		return "" // Cannot get text
	}
	// Ensure indices are valid before accessing the stream
	startIndex := startToken.GetStart()
	stopIndex := stopToken.GetStop()
	if startIndex < 0 || stopIndex < 0 || stopIndex < startIndex {
		return "" // Invalid range
	}
	return startToken.GetInputStream().GetText(startIndex, stopIndex)
}

// --- REMOVED Empty Stubs for Obsolete/Unused Rules ---
// Removed Enter/Exit Parameter_clauses
// Removed Enter/Exit Needs_clause
// Removed Enter/Exit Optional_clause
// Removed Enter/Exit Returns_clause
// Removed Enter/Exit Param_list
