// NeuroScript Version: 0.7.0
// File version: 40
// Purpose: Refactored: Main analyzer struct and entry point. All listener logic moved to semantic_listeners.go, semantic_validate_vars.go, and semantic_validate_calls.go.
// filename: pkg/nslsp/semantic_analyzer.go
// nlines: 83

package nslsp

import (
	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/tool"
	lsp "github.com/sourcegraph/go-lsp"
)

// SemanticAnalyzer performs semantic checks on a parsed NeuroScript AST.
type SemanticAnalyzer struct {
	toolRegistry  tool.ToolRegistry
	externalTools *ExternalToolManager
	symbolManager *SymbolManager
	localSymbols  map[string]SymbolInfo
	isDebug       bool
}

// NewSemanticAnalyzer creates a new analyzer instance.
func NewSemanticAnalyzer(registry tool.ToolRegistry, externalTools *ExternalToolManager, symbolManager *SymbolManager, isDebug bool) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		toolRegistry:  registry,
		externalTools: externalTools,
		symbolManager: symbolManager,
		localSymbols:  make(map[string]SymbolInfo),
		isDebug:       isDebug,
	}
}

// Analyze performs all configured semantic checks and returns a list of diagnostics.
func (sa *SemanticAnalyzer) Analyze(tree antlr.Tree) []lsp.Diagnostic {
	if tree == nil {
		return nil
	}
	parseTree, ok := tree.(antlr.ParseTree)
	if !ok {
		return nil
	}

	// 1. Collect all procedure definitions in this file first.
	symbolCollector := &symbolCollectorListener{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		localSymbols:            sa.localSymbols,
	}
	walker := antlr.NewParseTreeWalker()
	walker.Walk(symbolCollector, parseTree)

	// 2. Run the main validation walker.
	validator := &validationListener{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		semanticAnalyzer:        sa,
		diagnostics:             []lsp.Diagnostic{},
		initializedVars:         nil, // Managed by scope handlers
	}
	walker.Walk(validator, parseTree)
	return validator.diagnostics
}
