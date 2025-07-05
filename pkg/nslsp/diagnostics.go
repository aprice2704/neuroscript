// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Purpose: Handles parsing of NeuroScript documents and publishing of LSP diagnostics.
// filename: pkg/nslsp/diagnostics.go
// nlines: 70 // Approximate, will vary with actual implementation
// risk_rating: MEDIUM // Involves parsing and error interpretation.

package nslsp

import (
	"context"
	"log"

	"github.com/aprice2704/neuroscript/pkg/parser"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func PublishDiagnostics(ctx context.Context, conn *jsonrpc2.Conn, logger *log.Logger, neuroParser *parser.ParserAPI, uri lsp.DocumentURI, content string) {
	logger.Printf("Publishing diagnostics for %s", uri)

	// Use the new ParseForLSP method
	_, structuredErrors := neuroParser.ParseForLSP(string(uri), content)

	diagnostics := make([]lsp.Diagnostic, len(structuredErrors))
	for i, err := range structuredErrors {
		// ANTLR lines are 1-based, columns are 0-based. LSP ranges are 0-based.
		lspLine := err.Line - 1
		if lspLine < 0 {
			lspLine = 0
		}
		lspChar := err.Column
		if lspChar < 0 {
			lspChar = 0
		}

		// Determine end character for the range
		// If OffendingSymbol is available, use its length. Otherwise, default to 1 char.
		endChar := lspChar + 1
		if len(err.OffendingSymbol) > 0 {
			endChar = lspChar + len(err.OffendingSymbol)
		}
		// Ensure endChar does not go below startChar if OffendingSymbol is empty
		if endChar <= lspChar {
			endChar = lspChar + 1
		}

		diagnostics[i] = lsp.Diagnostic{
			Range: lsp.Range{
				Start: lsp.Position{Line: lspLine, Character: lspChar},
				End:   lsp.Position{Line: lspLine, Character: endChar},
			},
			Severity: lsp.Error,
			Source:   "nslsp", // Or "NeuroScript Language Server"
			Message:  err.Msg,
		}
		logger.Printf("LSP Diagnostic: %s L%d:%d - %s", uri, err.Line, err.Column, err.Msg)
	}

	rpcErr := conn.Notify(ctx, "textDocument/publishDiagnostics", lsp.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
	if rpcErr != nil {
		logger.Printf("Error publishing diagnostics: %v", rpcErr)
	}
}
