// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Implements all LSP request handler methods. ADDED textDocument/formatting handler. FIX: Moved workspace symbol scan from didOpen to initialize and removed the incorrect directory scanning logic.
// filename: pkg/nslsp/handlers.go
// nlines: 204

package nslsp

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/nsfmt" // Import the formatter
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func uriToPath(uri lsp.DocumentURI) (string, error) {
	u, err := url.ParseRequestURI(string(uri))
	if err != nil {
		return "", err
	}
	if u.Scheme != "file" {
		return "", fmt.Errorf("URI is not a file URI: %s", uri)
	}
	return u.Path, nil
}

func (s *Server) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	s.logger.Println("Handling 'initialize' request...")
	var params lsp.InitializeParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}
	s.logger.Printf("Initialization params: RootURI=%q", params.RootURI)

	s.loadConfig(params.RootURI)
	s.startFileWatcher(ctx, params.RootURI)

	// *** THE FIX IS HERE ***
	// This is the correct and only place to trigger the initial workspace scan.
	if workspacePath, err := uriToPath(params.RootURI); err == nil && workspacePath != "" {
		s.logger.Printf("Triggering initial workspace symbol scan for: %s", workspacePath)
		// We can run this in the background. The manager is (or should be)
		// thread-safe, and subsequent diagnostics will pick up symbols as
		// they become available.
		go s.symbolManager.ScanDirectory(workspacePath)
	} else if params.RootURI != "" {
		s.logger.Printf("WARN: Could not convert rootURI '%s' to path for symbol scan: %v", params.RootURI, err)
	}
	// *** END FIX ***

	s.logger.Println("'initialize' request handled successfully.")
	return lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
				Options: &lsp.TextDocumentSyncOptions{
					OpenClose: true,
					Change:    lsp.TDSKFull,
					Save:      &lsp.SaveOptions{IncludeText: false},
				},
			},
			HoverProvider: true,
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: []string{"."},
			},
			// ADDED: Advertise formatting capability
			DocumentFormattingProvider: true,
		},
	}, nil
}

func (s *Server) handleShutdown(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	s.logger.Println("Handling 'shutdown' request")
	if s.fileWatcher != nil {
		s.fileWatcher.Close()
	}
	return nil, nil
}

func (s *Server) handleExit(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	s.logger.Println("Handling 'exit' notification. Server will stop.")
	os.Exit(0)
}

func (s *Server) handleTextDocumentDidOpen(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.DidOpenTextDocumentParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}
	s.logger.Printf("Document opened: %s", params.TextDocument.URI)
	s.documentManager.Set(params.TextDocument.URI, params.TextDocument.Text)

	// *** THE FIX IS HERE ***
	// The incorrect, directory-based scanning logic has been removed.
	// The scan is now correctly triggered once in 'initialize'.
	// *** END FIX ***

	go PublishDiagnostics(ctx, s.conn, s.logger, s, params.TextDocument.URI, params.TextDocument.Text)
	return nil, nil
}

func (s *Server) handleTextDocumentDidChange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.DidChangeTextDocumentParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}
	s.logger.Printf("Document changed: %s", params.TextDocument.URI)
	if len(params.ContentChanges) > 0 {
		content := params.ContentChanges[0].Text
		s.documentManager.Set(params.TextDocument.URI, content)
		go PublishDiagnostics(ctx, s.conn, s.logger, s, params.TextDocument.URI, content)
	}
	return nil, nil
}

func (s *Server) handleTextDocumentDidSave(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.DidSaveTextDocumentParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}
	s.logger.Printf("didSave: SERVER VERSION '%s', BUILT-IN TOOL COUNT '%d'", serverVersion, s.interpreter.ToolRegistry().NTools())
	s.logger.Printf("Document saved: %s", params.TextDocument.URI)
	if content, found := s.documentManager.Get(params.TextDocument.URI); found {
		go PublishDiagnostics(ctx, s.conn, s.logger, s, params.TextDocument.URI, content)
	}
	return nil, nil
}

func (s *Server) handleTextDocumentDidClose(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.DidCloseTextDocumentParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}
	s.logger.Printf("Document closed: %s", params.TextDocument.URI)
	s.documentManager.Delete(params.TextDocument.URI)
	// Clear diagnostics for the closed file
	go PublishDiagnostics(ctx, s.conn, s.logger, s, params.TextDocument.URI, "")
	return nil, nil
}

// --- NEW HANDLER ---

// handleTextDocumentFormatting handles the 'textDocument/formatting' request.
func (s *Server) handleTextDocumentFormatting(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.DocumentFormattingParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}

	s.logger.Printf("Formatting request received for: %s", params.TextDocument.URI)

	// Get the document content from our manager
	content, found := s.documentManager.Get(params.TextDocument.URI)
	if !found {
		s.logger.Printf("Warning: textDocument/formatting request for unknown file: %s", params.TextDocument.URI)
		return nil, fmt.Errorf("document not found in manager: %s", params.TextDocument.URI)
	}

	// Call the nsfmt package
	formatted, err := nsfmt.Format([]byte(content))
	if err != nil {
		// If formatting fails (e.g., syntax error), return an error to the client.
		// This prevents "format on save" from wiping the file.
		s.logger.Printf("Formatting failed for %s: %v", params.TextDocument.URI, err)
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInternalError, Message: fmt.Sprintf("nsfmt failed: %v", err)}
	}

	// Calculate the full document range to replace
	lines := strings.Split(content, "\n")
	endLine := len(lines) - 1
	endChar := 0
	if endLine >= 0 {
		endChar = len(lines[endLine])
	}

	// Return a TextEdit to replace the entire document
	return []lsp.TextEdit{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: endLine, Character: endChar},
			},
			NewText: string(formatted),
		},
	}, nil
}
