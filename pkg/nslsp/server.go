// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 35
// :: description: Integrates the SymbolManager update on save.
// :: latestChange: Added textDocument/definition routing and extractProcedureNameAtPosition helper.
// :: filename: pkg/nslsp/server.go
// :: serialization: go
package nslsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/antlr/generated"
	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/fsnotify/fsnotify"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

const serverVersion = "1.3.5"

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

	// The interpreter's standard streams must be redirected to io.Discard.
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

// RefreshDiagnostics iterates over all open documents and re-publishes diagnostics.
func (s *Server) RefreshDiagnostics(ctx context.Context) {
	s.logger.Println("Refreshing diagnostics for all open documents...")
	openDocs := s.documentManager.GetAll()
	for uri, content := range openDocs {
		go PublishDiagnostics(ctx, s.conn, s.logger, s, uri, content)
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
	case "textDocument/formatting":
		return s.handleTextDocumentFormatting(ctx, conn, req)
	case "textDocument/definition":
		return s.handleTextDocumentDefinition(ctx, conn, req)
	default:
		s.logger.Printf("Received unhandled method: %s", req.Method)
		if isNotification(req.ID) {
			return nil, nil
		}
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
	}
}

// extractProcedureNameAtPosition finds the simple identifier at the cursor position
// that corresponds to a procedure call or definition.
func (s *Server) extractProcedureNameAtPosition(content string, position lsp.Position, sourceName string) string {
	debugHover := os.Getenv("NSLSP_DEBUG_HOVER") != ""
	var log loggerFunc = noOpLogger
	if debugHover {
		log = func(format string, args ...interface{}) {
			s.logger.Printf("[DEF_TRACE] "+format, args...)
		}
	}

	if s.coreParserAPI == nil {
		return ""
	}

	treeFromParser, _ := s.coreParserAPI.ParseForLSP(sourceName, content)
	if treeFromParser == nil {
		return ""
	}
	parseTreeRoot, ok := treeFromParser.(antlr.ParseTree)
	if !ok {
		return ""
	}

	foundTokenNode := findInitialNodeManually(parseTreeRoot, position.Line, position.Character, log)
	if foundTokenNode == nil {
		return ""
	}

	tokenType := foundTokenNode.GetSymbol().GetTokenType()
	// We only care about IDENTIFIER tokens
	if tokenType != gen.NeuroScriptLexerIDENTIFIER {
		return ""
	}

	// Verify context: Is this a tool call? If so, we ignore it (handled by extractToolName)
	// We want direct calls like `MyFunc()` or definitions like `func MyFunc`.
	// Simple check: The token text is the candidate name.
	return foundTokenNode.GetText()
}

// UnmarshalParams is a helper to decode JSON-RPC parameters.
func UnmarshalParams(rawParams *json.RawMessage, v interface{}) error {
	if rawParams == nil {
		return nil
	}
	return json.Unmarshal(*rawParams, v)
}
