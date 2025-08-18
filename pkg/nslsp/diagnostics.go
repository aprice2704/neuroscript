// NeuroScript Version: 0.6.0
// File version: 0.1.3
// Purpose: Integrate semantic analysis to publish both syntax and semantic diagnostics, now with external tool support. FIX: Use interpreter from server struct.
// filename: pkg/nslsp/diagnostics.go
// nlines: 80
// risk_rating: MEDIUM

package nslsp

import (
	"context"
	"log"
	"os"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// PublishDiagnostics parses a document, runs semantic analysis, and sends all findings to the client.
func PublishDiagnostics(ctx context.Context, conn *jsonrpc2.Conn, logger *log.Logger, s *Server, uri lsp.DocumentURI, content string) {
	logger.Printf("Publishing diagnostics for %s", uri)

	// 1. Get Syntax Errors from Parser
	tree, structuredErrors := s.coreParserAPI.ParseForLSP(string(uri), content)

	allDiagnostics := make([]lsp.Diagnostic, 0)

	// Convert parser errors to LSP diagnostics
	for _, err := range structuredErrors {
		lspLine := err.Line - 1
		if lspLine < 0 {
			lspLine = 0
		}
		endChar := err.Column + 1
		if len(err.OffendingSymbol) > 0 {
			endChar = err.Column + len(err.OffendingSymbol)
		}

		allDiagnostics = append(allDiagnostics, lsp.Diagnostic{
			Range: lsp.Range{
				Start: lsp.Position{Line: lspLine, Character: err.Column},
				End:   lsp.Position{Line: lspLine, Character: endChar},
			},
			Severity: lsp.Error,
			Source:   "nslsp-syntax",
			Message:  err.Msg,
		})
	}

	// 2. Get Semantic Errors from Analyzer
	// THE FIX IS HERE: Access the tool registry via the API interpreter facade.
	if tree != nil && s.interpreter != nil && (s.interpreter.ToolRegistry() != nil || s.externalTools != nil) {
		isDebug := os.Getenv("NSLSP_DEBUG_HOVER") != "" || os.Getenv("DEBUG_LSP_HOVER_TEST") != ""
		semanticAnalyzer := NewSemanticAnalyzer(s.interpreter.ToolRegistry(), s.externalTools, isDebug)
		semanticDiagnostics := semanticAnalyzer.Analyze(tree)
		if len(semanticDiagnostics) > 0 {
			allDiagnostics = append(allDiagnostics, semanticDiagnostics...)
		}
	} else {
		logger.Println("Skipping semantic analysis: AST, Interpreter, or Tool Registries not available.")
	}

	// 3. Publish Combined Diagnostics
	logger.Printf("Publishing %d total diagnostics for %s.", len(allDiagnostics), uri)
	rpcErr := conn.Notify(ctx, "textDocument/publishDiagnostics", lsp.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: allDiagnostics,
	})
	if rpcErr != nil {
		logger.Printf("Error publishing diagnostics: %v", rpcErr)
	}
}
