// NeuroScript Version: 0.3.1
// File version: 5
// Purpose: Enable debug environment variable to get detailed trace from the semantic analyzer.
// filename: pkg/nslsp/diagnostics_test.go
// nlines: 135
// risk_rating: MEDIUM

package nslsp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// clientHandler is a mock LSP client handler for testing purposes.
type clientHandler struct {
	diagnosticsChan chan lsp.PublishDiagnosticsParams
	wg              *sync.WaitGroup
	t               *testing.T
}

func (h *clientHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Method == "textDocument/publishDiagnostics" {
		var params lsp.PublishDiagnosticsParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			h.t.Errorf("Client failed to unmarshal diagnostics params: %v", err)
		}
		h.diagnosticsChan <- params
		h.wg.Done()
	}
}

// TestSemanticDiagnostics_UndefinedTool verifies that a semantic error is reported for an undefined tool.
func TestSemanticDiagnostics_UndefinedTool(t *testing.T) {
	// --- Setup ---
	// FIX: Set the environment variable to enable detailed debug logging in the analyzer.
	t.Setenv("DEBUG_LSP_HOVER_TEST", "1")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testLogger := log.New(os.Stderr, "[DIAGNOSTICS_TEST] ", log.LstdFlags|log.Lshortfile)
	serverInstance := NewServer(testLogger)

	if serverInstance.toolRegistry != nil {
		t.Logf("DEBUG: Tool registry initialized with %d tools.", serverInstance.toolRegistry.NTools())
	} else {
		t.Fatal("FATAL: Tool registry is nil in the test server instance.")
	}

	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	serverConn := jsonrpc2.NewConn(ctx, newSimpleObjectStream(&dummyReadWriteCloser{reader: serverReader, writer: serverWriter}), nil)
	defer serverConn.Close()

	var wg sync.WaitGroup
	handler := &clientHandler{
		diagnosticsChan: make(chan lsp.PublishDiagnosticsParams, 1),
		wg:              &wg,
		t:               t,
	}

	clientConn := jsonrpc2.NewConn(ctx, newSimpleObjectStream(&dummyReadWriteCloser{reader: clientReader, writer: clientWriter}), handler)
	defer clientConn.Close()

	// --- Test Execution ---
	wg.Add(1)
	uri := lsp.DocumentURI("file:///test.ns")
	content := `func MyProc() means
  set x = tool.FS.Read("/valid/path")      # This is a valid tool
  set y = tool.FS.ThisToolDoesNotExist()  # This tool is invalid
endfunc
`
	didOpenParams := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{URI: uri, Text: content},
	}
	didOpenParamsBytes, _ := json.Marshal(didOpenParams)
	rawParams := json.RawMessage(didOpenParamsBytes)

	go serverInstance.Handle(ctx, serverConn, &jsonrpc2.Request{
		Method: "textDocument/didOpen",
		Params: &rawParams,
	})

	// --- Verification ---
	var params lsp.PublishDiagnosticsParams
	select {
	case params = <-handler.diagnosticsChan:
		// Diagnostics received successfully
	case <-ctx.Done():
		t.Fatal("Test timed out waiting for diagnostics to be published")
	}

	wg.Wait()

	if len(params.Diagnostics) != 1 {
		t.Fatalf("Expected 1 diagnostic, but got %d. Diagnostics: %+v", len(params.Diagnostics), params.Diagnostics)
	}

	diagnostic := params.Diagnostics[0]
	expectedMsg := "Tool 'tool.FS.ThisToolDoesNotExist' is not defined."
	if !strings.Contains(diagnostic.Message, expectedMsg) {
		t.Errorf("Diagnostic message is incorrect.\nGot:  '%s'\nWant: '%s'", diagnostic.Message, expectedMsg)
	}

	if diagnostic.Source != "nslsp-semantic" {
		t.Errorf("Expected diagnostic source to be 'nslsp-semantic', but got '%s'", diagnostic.Source)
	}

	if diagnostic.Range.Start.Line != 2 {
		t.Errorf("Expected diagnostic on line 3, but was on line %d", diagnostic.Range.Start.Line+1)
	}
}
