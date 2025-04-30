// filename: pkg/core/ast_builder_procedures.go
package core

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	// Assuming interfaces logger is imported via ast_builder_main or here if needed
)

// --- Procedure Definition Handling (v0.2.0 - Optional Parens via parameter_clauses) ---
// Grammar Version: 0.2.0-alpha-onerror-fix-2 // <<< Incremented Version

// EnterProcedure_definition initializes the Procedure struct and processes parameters and metadata.
func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ""
	if id := ctx.IDENTIFIER(); id != nil {
		procName = id.GetText()
	} else {
		l.logger.Error("AST Builder: Procedure definition missing identifier!")
		// TODO: Add error to listener's error list?
		procName = "_INVALID_PROC_" // Use a placeholder name
	}
	l.logDebugAST(">>> Enter Procedure_definition (func): %s", procName)

	// 1. Initialize the current Procedure struct early
	l.currentProc = &Procedure{
		Name:           procName,
		RequiredParams: []string{}, // Initialize as empty
		OptionalParams: []string{}, // Initialize as empty
		ReturnVarNames: []string{}, // Initialize as empty
		Steps:          make([]Step, 0),
		Metadata:       make(map[string]string), // Initialize metadata map
	}
	// Capture original signature text using helper
	if sig := getRuleText(ctx); sig != "" {
		l.currentProc.OriginalSignature = sig
	}

	// 2. Process Parameter Clauses (Check for parameter_clauses context first)
	l.logDebugAST("    Checking for parameter clauses context for %s", procName)

	// --- Access clauses via intermediate parameter_clauses context ---
	if paramClausesCtx := ctx.Parameter_clauses(); paramClausesCtx != nil {
		l.logDebugAST("      Found parameter_clauses context. Processing clauses...")
		// Now access individual clauses via paramClausesCtx
		if needsCtx := paramClausesCtx.Needs_clause(); needsCtx != nil {
			l.currentProc.RequiredParams = l.extractParamList(needsCtx.Param_list())
			l.logDebugAST("        Found 'needs' params: %v", l.currentProc.RequiredParams)
		}
		if optionalCtx := paramClausesCtx.Optional_clause(); optionalCtx != nil {
			l.currentProc.OptionalParams = l.extractParamList(optionalCtx.Param_list())
			l.logDebugAST("        Found 'optional' params: %v", l.currentProc.OptionalParams)
		}
		if returnsCtx := paramClausesCtx.Returns_clause(); returnsCtx != nil {
			l.currentProc.ReturnVarNames = l.extractParamList(returnsCtx.Param_list())
			l.logDebugAST("        Found 'returns' params: %v", l.currentProc.ReturnVarNames)
		}
	} else {
		// No parameter_clauses context means no needs/optional/returns clauses were present
		l.logDebugAST("      No parameter_clauses context found (no needs/optional/returns).")
		// Slices are already initialized empty, so nothing more needed here.
	}
	// --- END Parameter Clause Processing ---

	// 3. Process Metadata Block (if it exists) - THIS LOGIC REMAINS THE SAME
	if metaBlockCtx := ctx.Metadata_block(); metaBlockCtx != nil {
		l.logDebugAST("    Processing procedure metadata block...")
		for _, metaLineNode := range metaBlockCtx.AllMETADATA_LINE() {
			if termNode, ok := metaLineNode.(antlr.TerminalNode); ok {
				fullLineText := termNode.GetText()
				trimmedLine := strings.TrimSpace(fullLineText)
				if strings.HasPrefix(trimmedLine, "::") {
					content := strings.TrimSpace(trimmedLine[2:])
					// Assuming ParseMetadataLine exists in utils.go or similar
					key, value, ok := ParseMetadataLine(content)
					if ok {
						// Ensure metadata map is initialized (should be by step 1)
						if l.currentProc.Metadata == nil {
							l.currentProc.Metadata = make(map[string]string)
						}
						l.currentProc.Metadata[key] = value
						l.logDebugAST("      Parsed proc metadata: [%s] = %q", key, value)
					} else {
						l.logger.Warn("Malformed procedure metadata line ignored", "line", fullLineText)
						// TODO: Optionally add an error to l.errors here
					}
				} else {
					l.logger.Warn("Unexpected procedure metadata line format (missing '::'?)", "line", fullLineText)
				}
			}
		}
	} else {
		l.logDebugAST("    No procedure metadata block found.")
	}

	// 4. Setup for processing steps - THIS LOGIC REMAINS THE SAME
	l.currentSteps = &l.currentProc.Steps
	l.blockStepStack = append(l.blockStepStack, l.currentSteps)
	l.logDebugAST("    Pushed procedure step block onto stack (Stack size: %d)", len(l.blockStepStack))

	// Clear value stack for the new procedure body
	l.valueStack = l.valueStack[:0]
}

// ExitProcedure_definition finalizes the procedure and adds it to the list.
// (This function remains largely unchanged from the previous version)
func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := "(nil)"
	if l.currentProc != nil {
		procName = l.currentProc.Name
	}
	l.logDebugAST("<<< Exit Procedure_definition (func): %s", procName)

	if l.currentProc == nil {
		l.logger.Error("AST Builder: Cannot append procedure, currentProc is nil on exit.")
		// Attempt to pop stack anyway to prevent further issues?
		if len(l.blockStepStack) > 0 {
			l.blockStepStack = l.blockStepStack[:len(l.blockStepStack)-1]
		}
		return // Cannot proceed
	}

	// Pop the procedure's step block from the stack
	if len(l.blockStepStack) > 0 {
		// Verify top of stack matches currentSteps before popping
		if l.currentSteps != l.blockStepStack[len(l.blockStepStack)-1] {
			l.logger.Error("Internal Error: Block stack mismatch on exiting procedure", "procedure_name", procName)
			// State might be corrupted, attempt recovery by just popping
		}
		l.blockStepStack = l.blockStepStack[:len(l.blockStepStack)-1]
		l.logDebugAST("    Popped procedure step block from stack (Stack size: %d)", len(l.blockStepStack))
	} else {
		l.logger.Error("Internal Error: Attempted to pop from empty block stack on exiting procedure", "procedure_name", procName)
	}

	// Add the completed procedure to the list
	l.logDebugAST("    Appending procedure: %s (ReqParams: %d, OptParams: %d, Returns: %d, Steps: %d, Metadata: %d)",
		l.currentProc.Name, len(l.currentProc.RequiredParams), len(l.currentProc.OptionalParams),
		len(l.currentProc.ReturnVarNames), len(l.currentProc.Steps), len(l.currentProc.Metadata))
	l.procedures = append(l.procedures, *l.currentProc)

	// Reset current procedure and steps pointers
	l.currentProc = nil
	if len(l.blockStepStack) > 0 {
		// Restore the step slice pointer for the parent block (if any)
		l.currentSteps = l.blockStepStack[len(l.blockStepStack)-1]
	} else {
		l.currentSteps = nil // No longer inside any block
	}

	// Sanity check: Value stack should be empty after processing a procedure body
	if len(l.valueStack) > 0 {
		l.logger.Warn("Value stack not empty at end of procedure", "procedure", procName, "size", len(l.valueStack))
		l.valueStack = l.valueStack[:0] // Clear it to prevent issues
	}
}

// --- Helper Methods ---

// extractParamList extracts a slice of strings from a Param_listContext.
// (This helper remains the same)
func (l *neuroScriptListenerImpl) extractParamList(ctx gen.IParam_listContext) []string {
	if ctx == nil {
		return []string{}
	}
	params := []string{}
	for _, identifier := range ctx.AllIDENTIFIER() {
		params = append(params, identifier.GetText())
	}
	return params
}

// getRuleText safely gets the text for a context using ctx.GetText().
// (This helper remains the same)
func getRuleText(ctx antlr.ParserRuleContext) string {
	if ctx == nil {
		return ""
	}
	return ctx.GetText()
}

// --- Empty Stubs for Clause Rules (Not strictly needed if logic is in EnterProcedure_definition) ---
// We added Parameter_clauses here.
func (l *neuroScriptListenerImpl) EnterParameter_clauses(ctx *gen.Parameter_clausesContext) {}
func (l *neuroScriptListenerImpl) ExitParameter_clauses(ctx *gen.Parameter_clausesContext)  {}
func (l *neuroScriptListenerImpl) EnterNeeds_clause(ctx *gen.Needs_clauseContext)           {}
func (l *neuroScriptListenerImpl) ExitNeeds_clause(ctx *gen.Needs_clauseContext)            {}
func (l *neuroScriptListenerImpl) EnterOptional_clause(ctx *gen.Optional_clauseContext)     {}
func (l *neuroScriptListenerImpl) ExitOptional_clause(ctx *gen.Optional_clauseContext)      {}
func (l *neuroScriptListenerImpl) EnterReturns_clause(ctx *gen.Returns_clauseContext)       {}
func (l *neuroScriptListenerImpl) ExitReturns_clause(ctx *gen.Returns_clauseContext)        {}
func (l *neuroScriptListenerImpl) EnterParam_list(ctx *gen.Param_listContext)               {}
func (l *neuroScriptListenerImpl) ExitParam_list(ctx *gen.Param_listContext)                {}

// Assuming ParseMetadataLine exists in utils.go or similar:
// func ParseMetadataLine(lineContent string) (key, value string, ok bool) { ... }
