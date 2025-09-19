// NeuroScript Version: 0.6.0
// File version: 6
// Purpose: Implements all LSP request handler methods, separating them from the main server struct. Hover logic moved to hover.go.
// filename: pkg/nslsp/handlers.go
// nlines: 155
// risk_rating: MEDIUM

package nslsp

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

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

	filePath, err := uriToPath(params.TextDocument.URI)
	if err != nil {
		s.logger.Printf("Could not convert URI to path for symbol scanning: %v", err)
	} else {
		dirPath := filepath.Dir(filePath)
		s.symbolManager.ScanDirectory(dirPath)
	}

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
