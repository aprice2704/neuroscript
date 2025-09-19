// NeuroScript Version: 0.7.0
// File version: 35
// Purpose: Implements full arity checking for user-defined procedures using min/max arg counts from the SymbolManager. FIX: Only report 'ProcNotFound' if the workspace-aware SymbolManager is active.
// filename: pkg/nslsp/semantic_analyzer.go
// nlines: 224
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

	symbolCollector := &symbolCollectorListener{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		localSymbols:            sa.localSymbols,
	}
	walker := antlr.NewParseTreeWalker()
	walker.Walk(symbolCollector, parseTree)

	validator := &validationListener{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		semanticAnalyzer:        sa,
		diagnostics:             []lsp.Diagnostic{},
	}
	walker.Walk(validator, parseTree)
	return validator.diagnostics
}

// symbolCollectorListener is a lightweight listener just for finding procedure definitions.
type symbolCollectorListener struct {
	*gen.BaseNeuroScriptListener
	localSymbols map[string]SymbolInfo
}

func (l *symbolCollectorListener) EnterProcedure_definition(ctx *gen.Procedure_definitionContext) {
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
}

// EnterCallable_expr is the single entry point for validating all tool and procedure calls.
func (l *validationListener) EnterCallable_expr(ctx *gen.Callable_exprContext) {
	callTarget := ctx.Call_target()
	if callTarget == nil {
		return
	}
	if callTarget.KW_TOOL() != nil {
		l.validateToolCall(ctx)
		return
	}
	if callTarget.IDENTIFIER() != nil {
		l.validateProcedureCall(ctx)
	}
}

func (l *validationListener) validateToolCall(ctx *gen.Callable_exprContext) {
	qi := ctx.Call_target().Qualified_identifier()
	if qi == nil {
		return
	}
	astTextFullName := "tool." + qi.GetText()
	lookupName := types.FullName(strings.ToLower(astTextFullName))
	var spec tool.ToolSpec
	var found bool
	var impl tool.ToolImplementation
	if l.semanticAnalyzer.toolRegistry != nil {
		impl, found = l.semanticAnalyzer.toolRegistry.GetTool(lookupName)
	}
	if !found && l.semanticAnalyzer.externalTools != nil {
		impl, found = l.semanticAnalyzer.externalTools.GetTool(lookupName)
	}
	if !found {
		token := ctx.Call_target().GetStart()
		l.diagnostics = append(l.diagnostics, lsp.Diagnostic{
			Range:    lspRangeFromToken(token, ctx.Call_target().GetText()),
			Severity: lsp.Error,
			Source:   "nslsp-semantic",
			Message:  fmt.Sprintf("Tool '%s' is not defined.", astTextFullName),
			Code:     string(DiagCodeToolNotFound),
		})
		return
	}
	spec = impl.Spec
	argList := ctx.Expression_list_opt()
	actualArgCount := 0
	if argList != nil && argList.Expression_list() != nil {
		actualArgCount = len(argList.Expression_list().AllExpression())
	}
	if !spec.Variadic && len(spec.Args) != actualArgCount {
		token := ctx.Call_target().GetStart()
		l.diagnostics = append(l.diagnostics, lsp.Diagnostic{
			Range:    lspRangeFromToken(token, ctx.Call_target().GetText()),
			Severity: lsp.Error,
			Source:   "nslsp-semantic",
			Message:  fmt.Sprintf("Expected %d argument(s) for tool '%s', but got %d.", len(spec.Args), astTextFullName, actualArgCount),
			Code:     string(DiagCodeArgCountMismatch),
		})
	}
}

func (l *validationListener) validateProcedureCall(ctx *gen.Callable_exprContext) {
	procName := ctx.Call_target().IDENTIFIER().GetText()
	actualArgCount := 0
	if ctx.Expression_list_opt() != nil && ctx.Expression_list_opt().Expression_list() != nil {
		actualArgCount = len(ctx.Expression_list_opt().Expression_list().AllExpression())
	}
	token := ctx.Call_target().GetStart()

	var symbolInfo SymbolInfo
	var isDefined bool

	if info, isLocal := l.semanticAnalyzer.localSymbols[procName]; isLocal {
		symbolInfo = info
		isDefined = true
	} else if l.semanticAnalyzer.symbolManager != nil {
		info, isGlobal := l.semanticAnalyzer.symbolManager.GetSymbolInfo(procName)
		if isGlobal {
			symbolInfo = info
			isDefined = true
			l.diagnostics = append(l.diagnostics, lsp.Diagnostic{
				Range:    lspRangeFromToken(token, procName),
				Severity: lsp.Information,
				Source:   "nslsp-semantic",
				Message:  fmt.Sprintf("Procedure '%s' is defined in another file: %s", procName, info.URI),
			})
		}
	}

	// THE FIX IS HERE: Only report "not defined" if we have a symbol manager
	// to actually check the workspace. If the manager is nil, it means we are
	// in a context (like the grammar test) that only cares about the current file.
	if !isDefined && l.semanticAnalyzer.symbolManager != nil {
		l.diagnostics = append(l.diagnostics, lsp.Diagnostic{
			Range:    lspRangeFromToken(token, procName),
			Severity: lsp.Warning,
			Source:   "nslsp-semantic",
			Message:  fmt.Sprintf("Procedure '%s' is not defined in the workspace.", procName),
			Code:     string(DiagCodeProcNotFound),
		})
		return
	}

	// If it's defined, we can proceed with arity checking.
	if isDefined {
		if actualArgCount < symbolInfo.MinArgs || actualArgCount > symbolInfo.MaxArgs {
			var expected string
			if symbolInfo.MinArgs == symbolInfo.MaxArgs {
				expected = fmt.Sprintf("%d", symbolInfo.MinArgs)
			} else {
				expected = fmt.Sprintf("%d to %d", symbolInfo.MinArgs, symbolInfo.MaxArgs)
			}
			l.diagnostics = append(l.diagnostics, lsp.Diagnostic{
				Range:    lspRangeFromToken(token, procName),
				Severity: lsp.Warning,
				Source:   "nslsp-semantic",
				Message:  fmt.Sprintf("Procedure '%s' expects %s arguments, but got %d.", procName, expected, actualArgCount),
				Code:     string(DiagCodeArgCountMismatch),
			})
		}
	}
}

func lspRangeFromToken(token antlr.Token, text string) lsp.Range {
	line := token.GetLine() - 1
	char := token.GetColumn()
	return lsp.Range{
		Start: lsp.Position{Line: line, Character: char},
		End:   lsp.Position{Line: line, Character: char + len(text)},
	}
}
