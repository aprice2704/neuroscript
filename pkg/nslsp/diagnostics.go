// NeuroScript Version: 0.7.0
// File version: 0.1.6
// Purpose: FIX: Updated the NewSemanticAnalyzer call to pass the server's symbolManager.
// filename: pkg/nslsp/diagnostics.go
// nlines: 105
// risk_rating: HIGH

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
	logger.Println("DIAGNOSTICS: Starting diagnostics process...")

	if s == nil || s.coreParserAPI == nil || conn == nil || logger == nil {
		if logger != nil {
			logger.Println("ERROR: PublishDiagnostics called with nil server, parser, connection, or logger. Aborting.")
		}
		return
	}
	logger.Printf("DIAGNOSTICS: Publishing for URI: %s", uri)

	// 1. Get Syntax Errors from Parser
	logger.Println("DIAGNOSTICS: STEP 1: Parsing for syntax errors...")
	tree, structuredErrors := s.coreParserAPI.ParseForLSP(string(uri), content)
	logger.Printf("DIAGNOSTICS: STEP 1 DONE: Found %d syntax errors.", len(structuredErrors))

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
	logger.Println("DIAGNOSTICS: STEP 2: Starting semantic analysis...")
	if tree != nil && s.interpreter != nil {
		if s.interpreter.ToolRegistry() != nil || s.externalTools != nil {
			isDebug := os.Getenv("NSLSP_DEBUG_HOVER") != "" || os.Getenv("DEBUG_LSP_HOVER_TEST") != ""
			logger.Println("DIAGNOSTICS: Creating new semantic analyzer...")
			// THE FIX IS HERE
			semanticAnalyzer := NewSemanticAnalyzer(s.interpreter.ToolRegistry(), s.externalTools, s.symbolManager, isDebug)
			logger.Println("DIAGNOSTICS: Running semantic analyzer...")
			semanticDiagnostics := semanticAnalyzer.Analyze(tree)
			logger.Printf("DIAGNOSTICS: Semantic analysis found %d diagnostics.", len(semanticDiagnostics))
			if len(semanticDiagnostics) > 0 {
				allDiagnostics = append(allDiagnostics, semanticDiagnostics...)
			}
		} else {
			logger.Println("DIAGNOSTICS: SKIPPING semantic analysis: Tool Registries not available.")
		}
	} else {
		if tree == nil {
			logger.Println("DIAGNOSTICS: SKIPPING semantic analysis: AST is nil.")
		}
		if s.interpreter == nil {
			logger.Println("DIAGNOSTICS: SKIPPING semantic analysis: Interpreter is nil.")
		}
	}
	logger.Println("DIAGNOSTICS: STEP 2 DONE: Semantic analysis complete.")

	// 3. Publish Combined Diagnostics
	logger.Printf("DIAGNOSTICS: STEP 3: Publishing %d total diagnostics for %s.", len(allDiagnostics), uri)
	rpcErr := conn.Notify(ctx, "textDocument/publishDiagnostics", lsp.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: allDiagnostics,
	})
	if rpcErr != nil {
		logger.Printf("DIAGNOSTICS: ERROR during final notification send: %v", rpcErr)
	}
	logger.Println("DIAGNOSTICS: STEP 3 DONE: Diagnostics process finished.")
}
