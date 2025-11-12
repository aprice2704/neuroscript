// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Refactored: Contains all validation listener logic for variable scope, initialization tracking, and read validation. FIX: Corrected compiler errors by using IDENTIFIER(0) instead of IDENTIFIER(). FIX: Pre-initialize built-in variables 'stdout' and 'stderr'. FEAT: Added check for built-in functions to prevent false uninitialized var warnings.
// filename: pkg/nslsp/semantic_validate_vars.go
// nlines: 147

package nslsp

import (
	"fmt"

	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	lsp "github.com/sourcegraph/go-lsp"
)

// newBuiltInFunctionsSet creates a map of all built-in function names.
var newBuiltInFunctionsSet = map[string]bool{
	"ln":     true,
	"log":    true,
	"sin":    true,
	"cos":    true,
	"tan":    true,
	"asin":   true,
	"acos":   true,
	"atan":   true,
	"len":    true,
	"typeof": true,
	"eval":   true,
}

// --- Scope Management ---

// initializeScope creates a new map for tracking variables and pre-populates
// it with built-in global variables.
func (l *validationListener) initializeScope() {
	l.initializedVars = make(map[string]bool)
	l.initializedVars["stdout"] = true
	l.initializedVars["stderr"] = true
}

func (l *validationListener) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	l.initializeScope()
	// Add all parameters to the initialized set
	if sig := ctx.Signature_part(); sig != nil {
		if needs := sig.Needs_clause(0); needs != nil && needs.Param_list() != nil {
			for _, ident := range needs.Param_list().AllIDENTIFIER() {
				l.initializedVars[ident.GetText()] = true
			}
		}
		if optional := sig.Optional_clause(0); optional != nil && optional.Param_list() != nil {
			for _, ident := range optional.Param_list().AllIDENTIFIER() {
				l.initializedVars[ident.GetText()] = true
			}
		}
	}
}

func (l *validationListener) ExitProcedure_definition(ctx *gen.Procedure_definitionContext) {
	l.initializedVars = nil
}

func (l *validationListener) EnterCommand_block(ctx *gen.Command_blockContext) {
	l.initializeScope()
}

func (l *validationListener) ExitCommand_block(ctx *gen.Command_blockContext) {
	l.initializedVars = nil
}

func (l *validationListener) EnterEvent_handler(ctx *gen.Event_handlerContext) {
	l.initializeScope()
	// Add the 'as evt' variable, if it exists
	if ctx.IDENTIFIER() != nil {
		l.initializedVars[ctx.IDENTIFIER().GetText()] = true
	}
}

func (l *validationListener) ExitEvent_handler(ctx *gen.Event_handlerContext) {
	l.initializedVars = nil
}

// --- Variable Write Tracking ---

func (l *validationListener) EnterSet_statement(ctx *gen.Set_statementContext) {
	if l.initializedVars == nil {
		return // Not in a scope we are tracking
	}
	if ctx.Lvalue_list() != nil {
		for _, lvalCtx := range ctx.Lvalue_list().AllLvalue() {
			// We only care about simple variable assignments (e.g., `set x = ...`)
			// not `set x[0] = ...` or `set x.y = ...` for this check.
			// FIX: Use IDENTIFIER(0) to get the first (and only) identifier.
			if lvalCtx.GetChildCount() == 1 && lvalCtx.IDENTIFIER(0) != nil {
				l.initializedVars[lvalCtx.IDENTIFIER(0).GetText()] = true
			}
		}
	}
}

func (l *validationListener) EnterFor_each_statement(ctx *gen.For_each_statementContext) {
	if l.initializedVars == nil {
		return // Not in a scope
	}
	// Add the loop variable
	if ctx.IDENTIFIER() != nil {
		l.initializedVars[ctx.IDENTIFIER().GetText()] = true
	}
}

func (l *validationListener) EnterAsk_stmt(ctx *gen.Ask_stmtContext) {
	if l.initializedVars == nil {
		return
	}
	// Add the 'into res' variable
	if ctx.Lvalue() != nil {
		// Only track simple `into x`
		// FIX: Use IDENTIFIER(0)
		if ctx.Lvalue().GetChildCount() == 1 && ctx.Lvalue().IDENTIFIER(0) != nil {
			l.initializedVars[ctx.Lvalue().IDENTIFIER(0).GetText()] = true
		}
	}
}

func (l *validationListener) EnterPromptuser_stmt(ctx *gen.Promptuser_stmtContext) {
	if l.initializedVars == nil {
		return
	}
	// Add the 'into user_name' variable
	if ctx.Lvalue() != nil {
		// Only track simple `into x`
		// FIX: Use IDENTIFIER(0)
		if ctx.Lvalue().GetChildCount() == 1 && ctx.Lvalue().IDENTIFIER(0) != nil {
			l.initializedVars[ctx.Lvalue().IDENTIFIER(0).GetText()] = true
		}
	}
}

// --- Variable Read Tracking ---

// EnterPrimary is called for simple variable reads.
func (l *validationListener) EnterPrimary(ctx *gen.PrimaryContext) {
	// We are only interested in simple variable reads (IDENTIFIERs)
	// that are *not* part of a callable_expr (which is handled above).
	if l.initializedVars == nil || ctx.IDENTIFIER() == nil {
		return
	}

	// Check if this IDENTIFIER is part of a callable_expr. If so, it's a
	// procedure call, which is handled by validateProcedureCall, not here.
	parent := ctx.GetParent()
	if _, isCallable := parent.(*gen.Callable_exprContext); isCallable {
		return
	}

	varName := ctx.IDENTIFIER().GetText()

	// Check if it's initialized
	if _, isInitialized := l.initializedVars[varName]; isInitialized {
		return // It's fine
	}

	// Check if it's a built-in function
	if _, isBuiltIn := newBuiltInFunctionsSet[varName]; isBuiltIn {
		return
	}

	// Check if it's a procedure name (being used as a value, e.g. `set x = MyFunc`)
	if _, isLocalProc := l.semanticAnalyzer.localSymbols[varName]; isLocalProc {
		return
	}
	if l.semanticAnalyzer.symbolManager != nil {
		if _, isGlobalProc := l.semanticAnalyzer.symbolManager.GetSymbolInfo(varName); isGlobalProc {
			return
		}
	}

	// If we're here, it's an uninitialized variable.
	token := ctx.IDENTIFIER().GetSymbol()
	l.diagnostics = append(l.diagnostics, lsp.Diagnostic{
		Range:    lspRangeFromToken(token, varName),
		Severity: lsp.Warning, // As requested
		Source:   "nslsp-semantic",
		Message:  fmt.Sprintf("Variable '%s' is used before being initialized.", varName),
		Code:     string(DiagCodeUninitializedVar),
	})
}
