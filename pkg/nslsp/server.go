// NeuroScript Version: 0.7.0
// File version: 32
// Purpose: Integrates the SymbolManager to enable workspace-wide symbol scanning on initialization. FIX: Provide a non-nil Stdin to the interpreter to prevent initialization failures.
// filename: pkg/nslsp/server.go
// nlines: 129

package nslsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/fsnotify/fsnotify"
	"github.com/sourcegraph/jsonrpc2"
)

const serverVersion = "1.3.2"

// Server holds the state for the NeuroScript Language Server.
type Server struct {
	conn            *jsonrpc2.Conn
	logger          *log.Logger
	documentManager *DocumentManager
	coreParserAPI   *parser.ParserAPI
	interpreter     *api.Interpreter
	config          Config
	externalTools   *ExternalToolManager
	fileWatcher     *fsnotify.Watcher
	symbolManager   *SymbolManager
}

// NewServer creates and initializes a new LSP server instance.
func NewServer(logger *log.Logger) *Server {
	adapterLogger := NewServerLogger(logger)

	// THE FIX IS HERE: The interpreter's standard streams must be redirected
	// to io.Discard. This prevents any accidental writes from polluting the
	// LSP's stdout channel. Stdin must be non-nil, so we provide an empty reader.
	hostCtx, err := api.NewHostContextBuilder().
		WithLogger(adapterLogger).
		WithStdout(io.Discard).
		WithStderr(io.Discard).
		WithStdin(strings.NewReader("")). // Stdin must be non-nil.
		Build()
	if err != nil {
		logger.Fatalf("CRITICAL: Failed to create interpreter HostContext: %v", err)
	}

	interp := api.New(api.WithHostContext(hostCtx))
	if interp == nil {
		logger.Fatal("CRITICAL: Failed to initialize NeuroScript interpreter via api.New()")
	}
	if interp.ToolRegistry() != nil {
		logger.Printf("LSP Server: Interpreter initialized, %d built-in tools loaded via API.", interp.ToolRegistry().NTools())
	}

	return &Server{
		logger:          logger,
		documentManager: NewDocumentManager(),
		coreParserAPI:   parser.NewParserAPI(adapterLogger),
		interpreter:     interp,
		externalTools:   NewExternalToolManager(),
		symbolManager:   NewSymbolManager(logger),
	}
}

// isNotification checks if a JSON-RPC request is a notification (has no ID).
func isNotification(id jsonrpc2.ID) bool {
	return id.Str == "" && id.Num == 0
}

// Handle routes incoming JSON-RPC requests to the appropriate handler method.
func (s *Server) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	s.conn = conn
	s.logger.Printf("LSP Request: Method=%s, ID=(Str:'%s', Num:%d)", req.Method, req.ID.Str, req.ID.Num)

	// DEBUG: Log raw parameters to see what the client is sending.
	if req.Params != nil {
		rawParamsJSON, _ := json.Marshal(req.Params)
		s.logger.Printf("LSP Request PARAMS: %s", string(rawParamsJSON))
	}

	switch req.Method {
	case "initialize":
		// Per user request, workspace scan is now triggered by textDocument/didOpen
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

// UnmarshalParams is a helper to decode JSON-RPC parameters.
func UnmarshalParams(rawParams *json.RawMessage, v interface{}) error {
	if rawParams == nil {
		return nil
	}
	return json.Unmarshal(*rawParams, v)
}
