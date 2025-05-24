// NeuroScript Version: 0.3.1
// File version: 0.2.6 // Fixed undefined lsp.LanguageSpecificMarkedString by using lsp.MarkedString.
// Purpose: Implements the main LSP server handler, routing requests, managing state,
//          providing diagnostics, and initial tool signature hover support.
// filename: pkg/nslsp/server.go
// nlines: 260 // Approximate
// risk_rating: MEDIUM

package nslsp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type Server struct {
	conn            *jsonrpc2.Conn
	logger          *log.Logger
	documentManager *DocumentManager
	coreParserAPI   *core.ParserAPI
	toolRegistry    *core.ToolRegistryImpl // Assuming ToolRegistryImpl is exported from core
}

func NewServer(logger *log.Logger) *Server {
	apiLoggerForCoreParser := logging.Logger(nil)

	registry := core.NewToolRegistry(nil)
	if registry == nil {
		logger.Println("CRITICAL: Failed to initialize tool registry in LSP server!")
	} else {
		logger.Printf("LSP Server: Tool registry initialized, %d global tool specs processed during its setup.", len(registry.ListTools()))
	}

	return &Server{
		logger:          logger,
		documentManager: NewDocumentManager(),
		coreParserAPI:   core.NewParserAPI(apiLoggerForCoreParser),
		toolRegistry:    registry,
	}
}

// isNotification checks if the JSON-RPC request is a notification.
// Notifications have an ID that is the zero value for jsonrpc2.ID (both Str and Num are zero/empty).
func isNotification(id jsonrpc2.ID) bool {
	return id.Str == "" && id.Num == 0
}

// Handle is the main RPC handler for LSP requests.
func (s *Server) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	s.conn = conn
	// Log ID by its components Str and Num
	s.logger.Printf("LSP Request: Method=%s, ID=(Str:'%s', Num:%d)", req.Method, req.ID.Str, req.ID.Num)

	switch req.Method {
	case "initialize":
		return s.handleInitialize(ctx, conn, req)
	case "initialized":
		s.logger.Println("LSP client 'initialized' notification received.")
		return nil, nil
	case "shutdown":
		return s.handleShutdown(ctx, conn, req)
	case "exit":
		s.handleExit(ctx, conn, req)
		return nil, nil
	case "textDocument/didOpen":
		return s.handleTextDocumentDidOpen(ctx, conn, req)
	case "textDocument/didChange":
		return s.handleTextDocumentDidChange(ctx, conn, req)
	case "textDocument/didSave":
		return s.handleTextDocumentDidSave(ctx, conn, req)
	case "textDocument/didClose":
		return s.handleTextDocumentDidClose(ctx, conn, req)
	case "textDocument/hover":
		return s.handleTextDocumentHover(ctx, conn, req)
	default:
		s.logger.Printf("Received unhandled method: %s", req.Method)
		if isNotification(req.ID) { // Use corrected isNotification
			return nil, nil
		}
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
	}
}

func (s *Server) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	s.logger.Println("Handling 'initialize' request")

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
		},
	}, nil
}

func (s *Server) handleShutdown(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	s.logger.Println("Handling 'shutdown' request")
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
	go PublishDiagnostics(ctx, s.conn, s.logger, s.coreParserAPI, params.TextDocument.URI, params.TextDocument.Text)
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
		go PublishDiagnostics(ctx, s.conn, s.logger, s.coreParserAPI, params.TextDocument.URI, content)
	}
	return nil, nil
}

func (s *Server) handleTextDocumentDidSave(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.DidSaveTextDocumentParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}
	s.logger.Printf("Document saved: %s", params.TextDocument.URI)
	if content, found := s.documentManager.Get(params.TextDocument.URI); found {
		go PublishDiagnostics(ctx, s.conn, s.logger, s.coreParserAPI, params.TextDocument.URI, content)
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
	go PublishDiagnostics(ctx, s.conn, s.logger, s.coreParserAPI, params.TextDocument.URI, "")
	return nil, nil
}

func (s *Server) handleTextDocumentHover(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.TextDocumentPositionParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}

	s.logger.Printf("Hover request for %s at L%d:%d", params.TextDocument.URI, params.Position.Line+1, params.Position.Character+1)

	content, found := s.documentManager.Get(params.TextDocument.URI)
	if !found {
		s.logger.Printf("Hover: Document not found %s", params.TextDocument.URI)
		return nil, nil
	}

	toolName := s.extractToolNameAtPosition(content, params.Position, string(params.TextDocument.URI))

	if toolName == "" {
		return nil, nil
	}

	s.logger.Printf("Hover: Identified potential tool name: %s", toolName)

	if s.toolRegistry == nil {
		s.logger.Printf("ERROR: Hover: ToolRegistry not available in LSP server.")
		return nil, nil
	}

	toolImpl, foundSpec := s.toolRegistry.GetTool(toolName)
	if !foundSpec {
		s.logger.Printf("Hover: ToolSpec not found for '%s'", toolName)
		return nil, nil
	}

	spec := toolImpl.Spec
	var hoverContent strings.Builder
	hoverContent.WriteString(fmt.Sprintf("#### `tool.%s`\n\n", spec.Name))
	if spec.Description != "" {
		hoverContent.WriteString(spec.Description + "\n\n")
	}

	hoverContent.WriteString("**Arguments:**\n")
	if len(spec.Args) == 0 {
		hoverContent.WriteString("*None*\n")
	} else {
		for _, arg := range spec.Args {
			reqStr := ""
			if arg.Required {
				reqStr = " (required)"
			}
			descStr := ""
			if arg.Description != "" {
				descStr = ": " + arg.Description
			}
			hoverContent.WriteString(fmt.Sprintf("* **`%s`** (`%s`)%s%s\n", arg.Name, arg.Type, descStr, reqStr))
		}
	}
	hoverContent.WriteString(fmt.Sprintf("\n**Returns:** `%s`\n", spec.ReturnType))

	// Corrected Hover Contents for sourcegraph/go-lsp
	// It expects []lsp.MarkedString where MarkedString can be a string or an object {Language: string, Value: string}
	return &lsp.Hover{ // Return a pointer to Hover
		Contents: []lsp.MarkedString{
			{Language: "markdown", Value: hoverContent.String()}, // Corrected line
		},
	}, nil
}

func UnmarshalParams(rawParams *json.RawMessage, v interface{}) error {
	if rawParams == nil {
		return nil
	}
	return json.Unmarshal(*rawParams, v)
}
