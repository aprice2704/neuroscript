// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 14
// :: description: Implements all LSP request handler methods.
// :: latestChange: FIX: defined InitializeParamsExtended to handle WorkspaceFolders which is missing from the older go-lsp struct.
// :: filename: pkg/nslsp/handlers.go
// :: serialization: go
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

// WorkspaceFolder represents a workspace folder as defined in LSP 3.6+
// We define it here because the sourcegraph/go-lsp version might not have it.
type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// InitializeParamsExtended extends the base params to support WorkspaceFolders
type InitializeParamsExtended struct {
	lsp.InitializeParams
	WorkspaceFolders []WorkspaceFolder `json:"workspaceFolders"`
}

func uriToPath(uri lsp.DocumentURI) (string, error) {
	u, err := url.ParseRequestURI(string(uri))
	if err != nil {
		return "", err
	}
	if u.Scheme != "file" {
		return "", fmt.Errorf("URI is not a file URI: %s", uri)
	}
	// Simple fix for Windows paths often coming in as /C:/...
	if len(u.Path) > 2 && u.Path[0] == '/' && u.Path[2] == ':' {
		return u.Path[1:], nil
	}
	return u.Path, nil
}

func (s *Server) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	s.logger.Println("Handling 'initialize' request...")

	// FIX: Use our extended struct to capture WorkspaceFolders
	var params InitializeParamsExtended
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}
	s.logger.Printf("Initialization params: RootURI=%q", params.RootURI)

	// Collect all workspace roots (Multi-root support)
	var workspacePaths []string
	if len(params.WorkspaceFolders) > 0 {
		for _, wf := range params.WorkspaceFolders {
			if path, err := uriToPath(lsp.DocumentURI(wf.URI)); err == nil && path != "" {
				workspacePaths = append(workspacePaths, path)
			}
		}
	}

	// Fallback to RootURI if no WorkspaceFolders provided
	if len(workspacePaths) == 0 && params.RootURI != "" {
		if path, err := uriToPath(params.RootURI); err == nil && path != "" {
			workspacePaths = append(workspacePaths, path)
		}
	}

	if len(workspacePaths) > 0 {
		s.logger.Printf("Initializing workspace with %d roots: %v", len(workspacePaths), workspacePaths)

		// Initialize watcher for the primary root.
		// Note: A more advanced watcher would watch all roots, but we start with RootURI for now.
		s.startFileWatcher(ctx, params.RootURI)

		go func() {
			for _, path := range workspacePaths {
				s.logger.Printf("Scanning workspace root: %s", path)
				// Load config for this root (to pick up tools.json)
				s.loadConfig(lsp.DocumentURI("file://" + path))

				// Scan symbols
				s.symbolManager.ScanDirectory(path)
			}
			// Refresh diagnostics once all scans are done
			s.RefreshDiagnostics(context.Background())
		}()
	} else {
		s.logger.Printf("WARN: No workspace roots found in initialize params.")
	}

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
			HoverProvider:      true,
			DefinitionProvider: true,
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: []string{"."},
			},
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

	var content string
	var found bool
	content, found = s.documentManager.Get(params.TextDocument.URI)
	if !found {
		if path, err := uriToPath(params.TextDocument.URI); err == nil {
			if bytes, err := os.ReadFile(path); err == nil {
				content = string(bytes)
			}
		}
	}

	if content != "" {
		go func() {
			s.symbolManager.UpdateSymbol(params.TextDocument.URI, content)
			PublishDiagnostics(ctx, s.conn, s.logger, s, params.TextDocument.URI, content)
		}()
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
	go PublishDiagnostics(ctx, s.conn, s.logger, s, params.TextDocument.URI, "")
	return nil, nil
}

func (s *Server) handleTextDocumentFormatting(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.DocumentFormattingParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}

	s.logger.Printf("Formatting request received for: %s", params.TextDocument.URI)

	content, found := s.documentManager.Get(params.TextDocument.URI)
	if !found {
		s.logger.Printf("Warning: textDocument/formatting request for unknown file: %s", params.TextDocument.URI)
		return nil, fmt.Errorf("document not found in manager: %s", params.TextDocument.URI)
	}

	formatted, err := nsfmt.Format([]byte(content))
	if err != nil {
		s.logger.Printf("Formatting failed for %s: %v", params.TextDocument.URI, err)
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInternalError, Message: fmt.Sprintf("nsfmt failed: %v", err)}
	}

	lines := strings.Split(content, "\n")
	endLine := len(lines) - 1
	endChar := 0
	if endLine >= 0 {
		endChar = len(lines[endLine])
	}

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

func (s *Server) handleTextDocumentDefinition(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.TextDocumentPositionParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}

	content, found := s.documentManager.Get(params.TextDocument.URI)
	if !found {
		return nil, nil
	}

	symbolName := s.extractProcedureNameAtPosition(content, params.Position, string(params.TextDocument.URI))
	if symbolName == "" {
		return nil, nil
	}

	info, found := s.symbolManager.GetSymbolInfo(symbolName)
	if !found {
		return nil, nil
	}

	return []lsp.Location{
		{
			URI:   info.URI,
			Range: info.Range,
		},
	}, nil
}
