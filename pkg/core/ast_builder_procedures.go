// NeuroScript Version: 0.3.1
// File version: 0.0.3 // Correctly initialize and assign OptionalParams as []ParamSpec.
// filename: pkg/core/ast_builder_procedures.go
package core

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	// Assuming logging.Logger is accessible via l.logger from ast_builder_main.go
)

// EnterProcedure_definition initializes the Procedure struct and processes the signature and metadata.
func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ""
	if id := ctx.IDENTIFIER(); id != nil {
		procName = id.GetText()
	} else {
		l.logger.Error("AST Builder: Procedure definition missing identifier!")
		l.addError(ctx, "procedure definition missing identifier")
		procName = "_INVALID_PROC_" // Default to an invalid name to avoid nil panics later
	}
	l.logDebugAST(">>> Enter Procedure_definition (func): %s", procName)

	procPos := tokenToPosition(ctx.GetStart())

	// 1. Initialize the current Procedure struct early
	l.currentProc = &Procedure{
		Pos:               procPos,
		Name:              procName,
		RequiredParams:    []string{},
		OptionalParams:    []ParamSpec{}, // <<< CORRECTED: Initialize as []ParamSpec{}
		ReturnVarNames:    []string{},
		Steps:             make([]Step, 0),
		Metadata:          make(map[string]string),
		Variadic:          false, // Initialize Variadic field
		VariadicParamName: "",    // Initialize VariadicParamName field
	}

	// Capture OriginalSignature
	if sigPartCtxAST := ctx.Signature_part(); sigPartCtxAST != nil {
		l.currentProc.OriginalSignature = getRuleText(sigPartCtxAST)
	} else if meansToken := ctx.KW_MEANS(); meansToken != nil { // Fallback for "func foo means"
		_ = ctx.GetStart()
		_ = meansToken.GetSymbol() // Get the "means" token
		// The signature is everything from "func" up to *before* "means"
		// This logic might need adjustment if KW_MEANS is inside Signature_part in some rule variants.
		// Assuming Signature_part is the preferred source if it exists.
		// If Signature_part is nil, but MEANS is there, it implies an empty signature before MEANS.
		// This OriginalSignature capture might need more refinement based on exact grammar alternatives.
		// For "func foo means", OriginalSignature should probably be empty or just "foo".
		// Let's assume getRuleText(ctx.Signature_part()) is the primary source if available.
		// If Signature_part is nil, OriginalSignature could be just the procName, or empty if no params.
		// The existing logic was:
		// if sig := getRuleText(ctx); sig != "" { ... } This gets the WHOLE rule text initially.
		// Let's simplify for now: if Signature_part is there, use it. Otherwise, it's more complex.
		// The original code used getRuleText(ctx.Signature_part()) or a substring before MEANS.
		// That substring logic was:
		// stopIndex := stopToken.GetStart() - 1
		// if stopIndex >= startToken.GetStart() {
		// l.currentProc.OriginalSignature = startToken.GetInputStream().GetText(startToken.GetStart(), stopIndex)
		// }
		// This seems complex and potentially error-prone if Signature_part isn't the direct source.
		// For now, relying on getRuleText(ctx.Signature_part()) is cleaner if that rule exists for all valid signatures.
		// If Signature_part is nil, it means the signature is effectively empty.
		if l.currentProc.OriginalSignature == "" && meansToken != nil {
			// This case indicates "func IDENTIFIER means" without explicit needs/optional/returns clauses.
			// The signature here is essentially just the function name and implies no params.
			// OriginalSignature might be just procName or empty depending on desired representation.
			// For simplicity, if Signature_part is nil, OriginalSignature remains empty or based on procName.
			// The previous complex logic for substringing before MEANS is removed for now as Signature_part should cover it.
		}
	}

	// 2. Process Signature Part
	l.logDebugAST("      Checking for signature part context for %s", procName)
	if sigPartCtx := ctx.Signature_part(); sigPartCtx != nil {
		l.logDebugAST("         Found signature_part context. Processing clauses...")
		if needsCtx := sigPartCtx.Needs_clause(); needsCtx != nil {
			if paramListCtx := needsCtx.Param_list(); paramListCtx != nil {
				l.currentProc.RequiredParams = l.extractParamNamesList(paramListCtx) // Use renamed helper
				l.logDebugAST("            Found 'needs' params: %v", l.currentProc.RequiredParams)
			}
		}
		if optionalCtx := sigPartCtx.Optional_clause(); optionalCtx != nil {
			if paramListCtx := optionalCtx.Param_list(); paramListCtx != nil {
				optionalParamNames := l.extractParamNamesList(paramListCtx) // Use renamed helper

				// Convert []string of names to []ParamSpec
				specs := make([]ParamSpec, len(optionalParamNames))
				for i, name := range optionalParamNames {
					specs[i] = ParamSpec{Name: name, DefaultValue: nil} // DefaultValue is nil for now
				}
				l.currentProc.OptionalParams = specs // <<< CORRECTED: Assign []ParamSpec
				l.logDebugAST("            Found 'optional' params (as ParamSpec): %v", l.currentProc.OptionalParams)
			}
		}
		// TODO: Implement parsing for variadic parameters (e.g., '...paramName')
		// If variadic parameter is found:
		//   l.currentProc.Variadic = true
		//   l.currentProc.VariadicParamName = "paramName" (extracted from signature)

		if returnsCtx := sigPartCtx.Returns_clause(); returnsCtx != nil {
			if paramListCtx := returnsCtx.Param_list(); paramListCtx != nil {
				l.currentProc.ReturnVarNames = l.extractParamNamesList(paramListCtx) // Use renamed helper
				l.logDebugAST("            Found 'returns' params: %v", l.currentProc.ReturnVarNames)
			}
		}
	} else {
		l.logDebugAST("         No signature_part context found for procedure '%s'. Assuming no parameters or return values.", procName)
	}

	// 3. Process Metadata Block (if it exists)
	if metaBlockCtx := ctx.Metadata_block(); metaBlockCtx != nil {
		l.logDebugAST("      Processing procedure metadata block for %s...", procName)
		for _, metaLineNode := range metaBlockCtx.AllMETADATA_LINE() {
			if termNode, ok := metaLineNode.(antlr.TerminalNode); ok {
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
							if l.currentProc.Metadata == nil { // Should have been initialized
								l.currentProc.Metadata = make(map[string]string)
							}
							l.currentProc.Metadata[key] = value
							l.logDebugAST("         Parsed proc metadata: [%s] = %q", key, value)
						} else {
							l.addErrorf(token, "Ignoring procedure metadata line with empty key")
						}
					} else {
						l.addErrorf(token, "Ignoring malformed procedure metadata line (missing ':'?) for key-value pair")
					}
				} else {
					l.addErrorf(token, "Unexpected procedure metadata line format (missing '::' prefix?)")
				}
			}
		}
	} else {
		l.logDebugAST("      No procedure metadata block found for %s.", procName)
	}

	// 4. Setup for processing steps
	if len(l.blockStepStack) != 0 {
		l.logger.Errorf("Block stack not empty at START of procedure definition! Clearing for procedure '%s'. Stack size: %d", procName, len(l.blockStepStack))
		l.errors = append(l.errors, fmt.Errorf("internal AST builder error: block stack not empty at start of procedure '%s'", procName))
		l.blockStepStack = l.blockStepStack[:0] // Clear it for recovery
	}
	l.currentSteps = &l.currentProc.Steps // Point currentSteps to the proc's main step slice
	l.logDebugAST("      Set currentSteps for procedure %s: (pointer: %p)", procName, l.currentSteps)

	// Clear value stack for the new procedure body
	l.valueStack = l.valueStack[:0]
}

// ExitProcedure_definition finalizes the procedure and adds it to the listener's temporary list.
func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := "(nil currentProc on exit)"
	if l.currentProc != nil {
		procName = l.currentProc.Name
	}
	l.logDebugAST("<<< Exit Procedure_definition (func): %s", procName)

	if l.currentProc == nil {
		l.logger.Error("AST Builder: Cannot append procedure, currentProc is nil on exit of procedure definition.")
		// Ensure currentSteps is also nil if currentProc is nil to prevent dangling pointers if error occurred mid-proc.
		if l.currentSteps != nil {
			l.logger.Debugf("Resetting currentSteps to nil as currentProc is nil on exit for %s.", procName)
			l.currentSteps = nil
		}
		return
	}

	// Final check of block stack state
	if len(l.blockStepStack) != 0 {
		l.logger.Errorf("Internal Error: Block stack not empty on exiting procedure definition for '%s'. Stack size: %d. This indicates unclosed blocks.", procName, len(l.blockStepStack))
		l.errors = append(l.errors, fmt.Errorf("internal AST builder error: block stack size is %d (should be 0) at end of procedure '%s'", len(l.blockStepStack), procName))
		l.blockStepStack = l.blockStepStack[:0] // Attempt recovery for subsequent parsing
		l.logger.Warnf("Cleared non-empty block stack for recovery after procedure '%s'.", procName)
	}

	// Add the fully constructed procedure to the program's list
	// Assuming l.program.Procedures is map[string]*Procedure as defined in ast.go
	if l.program.Procedures == nil {
		l.program.Procedures = make(map[string]*Procedure)
	}
	if _, exists := l.program.Procedures[l.currentProc.Name]; exists {
		//	l.addErrorf("Procedure '%s' already defined.", l.currentProc.Name)
		// Potentially overwrite or skip. For now, last one wins or error is just logged.
		// To be strict, this should probably prevent adding.
	}
	l.program.Procedures[l.currentProc.Name] = l.currentProc
	l.logDebugAST("      Stored procedure: %s", l.currentProc.Name)

	// Clean up for the next procedure or end of parsing
	l.currentProc = nil
	l.currentSteps = nil
	l.logDebugAST("      Reset currentProc and currentSteps to nil after exiting procedure %s.", procName)

	// Value stack check (should be empty if all expressions were correctly popped)
	if len(l.valueStack) > 0 {
		l.logger.Warnf("Value stack not empty at end of procedure '%s'. Stack size: %d. This might indicate an issue with expression evaluation or AST construction.", procName, len(l.valueStack))
		l.errors = append(l.errors, fmt.Errorf("internal AST builder error: value stack not empty (%d elements) at end of procedure '%s'", len(l.valueStack), procName))
		l.valueStack = l.valueStack[:0] // Clear for next procedure
	}
}

// extractParamNamesList extracts just the names of parameters as a string slice.
// Renamed from extractParamList.
func (l *neuroScriptListenerImpl) extractParamNamesList(ctx gen.IParam_listContext) []string {
	if ctx == nil {
		return []string{}
	}
	params := []string{}
	identifiers := ctx.AllIDENTIFIER()
	if identifiers == nil { // Should not happen if Param_listContext is valid and has identifiers
		return []string{}
	}
	for _, identifier := range identifiers {
		if identifier != nil { // TerminalNode IDENTIFIER
			params = append(params, identifier.GetText())
		}
	}
	return params
}

// getRuleText remains unchanged
func getRuleText(ctx antlr.ParserRuleContext) string {
	if ctx == nil {
		return ""
	}
	startToken := ctx.GetStart()
	stopToken := ctx.GetStop()
	if startToken == nil || stopToken == nil || startToken.GetInputStream() == nil {
		// This can happen if the context is somehow empty or represents an optional rule not taken.
		return ""
	}
	startIndex := startToken.GetStart()
	stopIndex := stopToken.GetStop()

	// Add check for valid indices to prevent panic with GetText
	if startIndex < 0 || stopIndex < 0 || stopIndex < startIndex {
		// This indicates an issue with the token stream or rule context bounds.
		return "" // Return empty string if indices are invalid
	}
	return startToken.GetInputStream().GetText(startIndex, stopIndex)
}
