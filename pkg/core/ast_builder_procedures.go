// pkg/core/ast_builder_procedures.go
package core

import (
	// For logging warning
	"strings"

	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
)

func (l *neuroScriptListenerImpl) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ""
	if ctx.IDENTIFIER() != nil {
		procName = ctx.IDENTIFIER().GetText()
	}
	l.logDebugAST(">>> Enter Procedure_definition: %s", procName)
	params := []string{}
	if ctx.Param_list_opt() != nil && ctx.Param_list_opt().Param_list() != nil {
		for _, id := range ctx.Param_list_opt().Param_list().AllIDENTIFIER() {
			params = append(params, id.GetText())
		}
	}
	docstring := Docstring{}
	if ctx.COMMENT_BLOCK() != nil {
		commentContent := ctx.COMMENT_BLOCK().GetText()
		content := strings.TrimPrefix(commentContent, "COMMENT:")
		// Assuming lexer provides content without ENDCOMMENT marker
		// If not, add TrimSuffix here. Consider adding robustness.
		// Also trim ENDCOMMENT if present, robustly
		content = strings.TrimSuffix(strings.TrimSpace(content), "ENDCOMMENT") // Trim space AND suffix
		docstring = parseDocstring(strings.TrimSpace(content))
	} else {
		l.logger.Printf("[WARN] Procedure '%s' is missing COMMENT: block.", procName)
		docstring.Purpose = "(Docstring missing)"
	}
	l.currentProc = &Procedure{Name: procName, Params: params, Steps: make([]Step, 0), Docstring: docstring}
	l.currentSteps = &l.currentProc.Steps
	l.valueStack = l.valueStack[:0] // Ensure stack is clear for new procedure
}
func (l *neuroScriptListenerImpl) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	procName := ""
	if l.currentProc != nil {
		procName = l.currentProc.Name
	}
	l.logDebugAST("<<< Exit Procedure_definition: %s", procName)
	if l.currentProc != nil {
		l.logDebugAST("    Appending actual procedure: %s (Steps: %d)", l.currentProc.Name, len(l.currentProc.Steps))
		l.procedures = append(l.procedures, *l.currentProc)
		l.logDebugAST("    Procedures count after append: %d", len(l.procedures))
	} else {
		l.logDebugAST("    l.currentProc was nil, cannot append actual procedure.")
	}
	l.currentProc = nil
	l.currentSteps = nil
	if len(l.valueStack) > 0 { // Should be empty if parsing was correct
		l.logger.Printf("[WARN] Value stack not empty at end of procedure %s (Size: %d)", procName, len(l.valueStack))
		l.valueStack = l.valueStack[:0] // Clear anyway
	}
}
