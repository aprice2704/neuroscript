// NeuroScript Version: 0.6.0
// File version: 22
// Purpose: Integrated external tool implementations for diagnostics, correctly extracting the spec.
// filename: pkg/nslsp/semantic_analyzer.go
// nlines: 125
// risk_rating: HIGH

package nslsp

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
	lsp "github.com/sourcegraph/go-lsp"
)

// SemanticAnalyzer performs semantic checks on a parsed NeuroScript AST.
type SemanticAnalyzer struct {
	toolRegistry  tool.ToolRegistry
	externalTools *ExternalToolManager
	isDebug       bool
}

// NewSemanticAnalyzer creates a new analyzer instance.
func NewSemanticAnalyzer(registry tool.ToolRegistry, externalTools *ExternalToolManager, isDebug bool) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		toolRegistry:  registry,
		externalTools: externalTools,
		isDebug:       isDebug,
	}
}

// Analyze performs all configured semantic checks and returns a list of diagnostics.
func (sa *SemanticAnalyzer) Analyze(tree antlr.Tree) []lsp.Diagnostic {
	if tree == nil || (sa.toolRegistry == nil && sa.externalTools == nil) {
		return nil
	}
	parseTree, ok := tree.(antlr.ParseTree)
	if !ok {
		return nil
	}
	listener := &toolValidationListener{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		semanticAnalyzer:        sa,
		diagnostics:             []lsp.Diagnostic{},
	}
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, parseTree)
	return listener.diagnostics
}

// toolValidationListener walks the AST to find and validate tool calls.
type toolValidationListener struct {
	*gen.BaseNeuroScriptListener
	semanticAnalyzer *SemanticAnalyzer
	diagnostics      []lsp.Diagnostic
}

// EnterCallable_expr is called by the walker for every callable expression.
func (l *toolValidationListener) EnterCallable_expr(ctx *gen.Callable_exprContext) {
	callTarget := ctx.Call_target()
	if callTarget == nil || callTarget.KW_TOOL() == nil {
		return // Not a tool call.
	}

	qi := callTarget.Qualified_identifier()
	if qi == nil {
		return
	}

	astTextFullName := "tool." + qi.GetText()
	lookupName := types.FullName(strings.ToLower(astTextFullName))

	// --- 1. Find Tool Specification ---
	var spec tool.ToolSpec
	var found bool
	var impl tool.ToolImplementation

	if l.semanticAnalyzer.toolRegistry != nil {
		impl, found = l.semanticAnalyzer.toolRegistry.GetTool(lookupName)
	}
	if !found && l.semanticAnalyzer.externalTools != nil {
		impl, found = l.semanticAnalyzer.externalTools.GetTool(lookupName)
	}

	if found {
		spec = impl.Spec
	}

	// --- 2. Undefined Tool Check ---
	if !found {
		token := callTarget.GetStart()
		diagnostic := lsp.Diagnostic{
			Range:    lspRangeFromToken(token, callTarget.GetText()),
			Severity: lsp.Error,
			Source:   "nslsp-semantic",
			Message:  fmt.Sprintf("Tool '%s' is not defined.", astTextFullName),
		}
		l.diagnostics = append(l.diagnostics, diagnostic)
		return
	}

	// --- 3. Argument Count Check ---
	argList := ctx.Expression_list_opt()
	actualArgCount := 0
	if argList != nil && argList.Expression_list() != nil {
		actualArgCount = len(argList.Expression_list().AllExpression())
	}

	if !spec.Variadic && len(spec.Args) != actualArgCount {
		token := callTarget.GetStart()
		diagnostic := lsp.Diagnostic{
			Range:    lspRangeFromToken(token, callTarget.GetText()),
			Severity: lsp.Error,
			Source:   "nslsp-semantic",
			Message:  fmt.Sprintf("Expected %d argument(s) for tool '%s', but got %d.", len(spec.Args), astTextFullName, actualArgCount),
		}
		l.diagnostics = append(l.diagnostics, diagnostic)
	}
}

// lspRangeFromToken creates an LSP range from an ANTLR token.
func lspRangeFromToken(token antlr.Token, text string) lsp.Range {
	line := token.GetLine() - 1
	char := token.GetColumn()
	return lsp.Range{
		Start: lsp.Position{Line: line, Character: char},
		End:   lsp.Position{Line: line, Character: char + len(text)},
	}
}
