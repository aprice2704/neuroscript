// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Refactored: Contains struct definitions for semantic listeners and helpers.
// filename: pkg/nslsp/semantic_listeners.go
// nlines: 70

package nslsp

import (
	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	lsp "github.com/sourcegraph/go-lsp"
)

// symbolCollectorListener is a lightweight listener just for finding procedure definitions.
type symbolCollectorListener struct {
	*gen.BaseNeuroScriptListener
	localSymbols map[string]SymbolInfo
}

func (l *symbolCollectorListener) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
	// If the parser recovers from a syntax error (e.g., "func () means..."),
	// the IDENTIFIER token may be nil. We must check for this to prevent a panic.
	if ctx.IDENTIFIER() == nil {
		return // This is not a valid, named procedure to collect.
	}

	procName := ctx.IDENTIFIER().GetText()
	if procName == "" {
		return
	}
	needsCount := 0
	optionalCount := 0
	if sig := ctx.Signature_part(); sig != nil {
		if needs := sig.Needs_clause(0); needs != nil && needs.Param_list() != nil {
			needsCount = len(needs.Param_list().AllIDENTIFIER())
		}
		if optional := sig.Optional_clause(0); optional != nil && optional.Param_list() != nil {
			optionalCount = len(optional.Param_list().AllIDENTIFIER())
		}
	}
	token := ctx.IDENTIFIER().GetSymbol()
	l.localSymbols[procName] = SymbolInfo{
		Range:   lspRangeFromToken(token, procName),
		MinArgs: needsCount,
		MaxArgs: needsCount + optionalCount,
	}
}

// validationListener walks the AST to find and validate calls.
type validationListener struct {
	*gen.BaseNeuroScriptListener
	semanticAnalyzer *SemanticAnalyzer
	diagnostics      []lsp.Diagnostic
	initializedVars  map[string]bool // Tracks variables in the current scope
}

// lspRangeFromToken is a helper to create an lsp.Range from an ANTLR token.
func lspRangeFromToken(token antlr.Token, text string) lsp.Range {
	line := token.GetLine() - 1
	char := token.GetColumn()
	return lsp.Range{
		Start: lsp.Position{Line: line, Character: char},
		End:   lsp.Position{Line: line, Character: char + len(text)},
	}
}
