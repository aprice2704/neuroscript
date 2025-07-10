// NeuroScript Version: 0.3.1
// File version: 18
// Purpose: Corrected the listener to use EnterCallable_expr, the correct ANTLR context for validating tool calls and their arguments.
// filename: pkg/nslsp/semantic_analyzer.go
// nlines: 110
// risk_rating: HIGH

package nslsp

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/parser/generated"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
	lsp "github.com/sourcegraph/go-lsp"
)

// SemanticAnalyzer performs semantic checks on a parsed NeuroScript AST.
type SemanticAnalyzer struct {
	toolRegistry tool.ToolRegistry
	isDebug      bool
}

// NewSemanticAnalyzer creates a new analyzer instance.
func NewSemanticAnalyzer(registry tool.ToolRegistry, isDebug bool) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		toolRegistry: registry,
		isDebug:      isDebug,
	}
}

// Analyze performs all configured semantic checks and returns a list of diagnostics.
func (sa *SemanticAnalyzer) Analyze(tree antlr.Tree) []lsp.Diagnostic {
	if tree == nil || sa.toolRegistry == nil {
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

	ids := getIdentifiersTextsFromQIGeneric(qi, l.semanticAnalyzer.isDebug)
	if len(ids) < 2 {
		return
	}

	group := strings.Join(ids[:len(ids)-1], ".")
	name := ids[len(ids)-1]
	fullName := types.MakeFullName(group, name)
	astTextFullName := "tool." + strings.Join(ids, ".")

	toolImpl, found := l.semanticAnalyzer.toolRegistry.GetTool(fullName)
	// --- 1. Undefined Tool Check ---
	if !found {
		token := qi.GetStart()
		diagnostic := lsp.Diagnostic{
			Range:    lsp.Range{Start: lsp.Position{Line: token.GetLine() - 1, Character: token.GetColumn()}, End: lsp.Position{Line: token.GetLine() - 1, Character: token.GetColumn() + len(qi.GetText())}},
			Severity: lsp.Error,
			Source:   "nslsp-semantic",
			Message:  fmt.Sprintf("Tool '%s' is not defined.", astTextFullName),
		}
		l.diagnostics = append(l.diagnostics, diagnostic)
		return
	}

	// --- 2. Argument Count Check ---
	spec := toolImpl.Spec
	argList := ctx.Expression_list_opt()
	actualArgCount := 0
	if argList != nil && argList.Expression_list() != nil {
		actualArgCount = len(argList.Expression_list().AllExpression())
	}

	if !spec.Variadic && len(spec.Args) != actualArgCount {
		token := callTarget.GetStart()
		diagnostic := lsp.Diagnostic{
			Range:    lsp.Range{Start: lsp.Position{Line: token.GetLine() - 1, Character: token.GetColumn()}, End: lsp.Position{Line: token.GetLine() - 1, Character: token.GetColumn() + len(callTarget.GetText())}},
			Severity: lsp.Error,
			Source:   "nslsp-semantic",
			Message:  fmt.Sprintf("Expected %d argument(s) for tool '%s', but got %d", len(spec.Args), astTextFullName, actualArgCount),
		}
		l.diagnostics = append(l.diagnostics, diagnostic)
	}
}
