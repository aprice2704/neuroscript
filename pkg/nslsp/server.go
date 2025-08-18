// NeuroScript Version: 0.6.0
// File version: 23
// Purpose: Simplifies server by removing the redundant tool registry, now that the api.Interpreter exposes it via a getter.
// filename: pkg/nslsp/server.go
// nlines: 300
// risk_rating: HIGH

package nslsp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/fsnotify/fsnotify"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

const serverVersion = "1.3.1"

type Server struct {
	conn            *jsonrpc2.Conn
	logger          *log.Logger
	documentManager *DocumentManager
	coreParserAPI   *parser.ParserAPI
	interpreter     *api.Interpreter // This now provides the tool registry.
	config          Config
	externalTools   *ExternalToolManager
	fileWatcher     *fsnotify.Watcher
}

func NewServer(logger *log.Logger) *Server {
	interp := api.New(api.WithLogger(nil))
	if interp == nil {
		logger.Fatal("CRITICAL: Failed to initialize NeuroScript interpreter via api.New()")
	}
	if interp.ToolRegistry() != nil {
		logger.Printf("LSP Server: Interpreter initialized, %d built-in tools loaded via API.", interp.ToolRegistry().NTools())
	}

	return &Server{
		logger:          logger,
		documentManager: NewDocumentManager(),
		coreParserAPI:   parser.NewParserAPI(nil),
		interpreter:     interp,
		externalTools:   NewExternalToolManager(),
	}
}

func isNotification(id jsonrpc2.ID) bool {
	return id.Str == "" && id.Num == 0
}

func (s *Server) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	s.conn = conn
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
	case "textDocument/completion":
		return s.handleTextDocumentCompletion(ctx, conn, req)
	default:
		s.logger.Printf("Received unhandled method: %s", req.Method)
		if isNotification(req.ID) {
			return nil, nil
		}
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
	}
}

func (s *Server) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.InitializeParams
	if err := UnmarshalParams(req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: err.Error()}
	}

	s.loadConfig(params.RootURI)
	s.startFileWatcher(ctx, params.RootURI)

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
	go PublishDiagnostics(ctx, s.conn, s.logger, s, params.TextDocument.URI, "")
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
	lookupName := types.FullName(strings.ToLower(toolName))

	var spec tool.ToolSpec
	var foundTool bool
	var impl tool.ToolImplementation

	if s.interpreter.ToolRegistry() != nil {
		impl, foundTool = s.interpreter.ToolRegistry().GetTool(lookupName)
	}
	if !foundTool && s.externalTools != nil {
		impl, foundTool = s.externalTools.GetTool(lookupName)
	}

	if !foundTool {
		s.logger.Printf("Hover: ToolSpec not found for '%s' in any registry.", toolName)
		return nil, nil
	}
	spec = impl.Spec

	var hoverContent strings.Builder
	hoverContent.WriteString(fmt.Sprintf("#### `%s`\n\n", toolName))
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

	return &lsp.Hover{
		Contents: []lsp.MarkedString{
			{Language: "markdown", Value: hoverContent.String()},
		},
	}, nil
}

func UnmarshalParams(rawParams *json.RawMessage, v interface{}) error {
	if rawParams == nil {
		return nil
	}
	return json.Unmarshal(*rawParams, v)
}
