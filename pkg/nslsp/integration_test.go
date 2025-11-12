// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: Provides a full end-to-end integration test. FIX: Updated 'MissingMeans' test to check for the correct, modern error message.
// filename: pkg/nslsp/integration_test.go
// nlines: 186
// risk_rating: LOW

package nslsp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// mockClientHandler simulates an LSP client for testing. It listens for
// `publishDiagnostics` notifications and uses a channel to signal when they arrive.
type mockClientHandler struct {
	t               *testing.T
	diagnosticsChan chan lsp.PublishDiagnosticsParams
	wg              *sync.WaitGroup
	// THE FIX IS HERE: An atomic flag to make Done() idempotent.
	// 0 = can call Done, 1 = Done already called for this wait cycle.
	doneFlag uint32
}

// PrepareForWait resets the idempotency flag so the handler can call Done() again.
func (h *mockClientHandler) PrepareForWait() {
	atomic.StoreUint32(&h.doneFlag, 0)
}

func (h *mockClientHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Method == "textDocument/publishDiagnostics" {
		var params lsp.PublishDiagnosticsParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			h.t.Errorf("Client failed to unmarshal diagnostics params: %v", err)
			return
		}
		h.diagnosticsChan <- params
		// THE FIX IS HERE: Atomically check and set the flag.
		// If it was 0, we set it to 1 and call Done(). If it was already 1, we do nothing.
		if atomic.CompareAndSwapUint32(&h.doneFlag, 0, 1) {
			h.wg.Done()
		}
	}
}

// setupIntegrationTest creates a full client-server connection in memory.
func setupIntegrationTest(t *testing.T, ctx context.Context) (serverConn, clientConn *jsonrpc2.Conn, handler *mockClientHandler) {
	// Create connected pipes
	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	// Create a real logger
	testLogger := log.New(os.Stderr, "[INTEGRATION_TEST] ", log.LstdFlags|log.Lshortfile)
	adapterLogger := NewServerLogger(testLogger)
	adapterLogger.SetLevel(interfaces.LogLevelDebug)

	// Create server
	serverInstance := NewServer(testLogger)

	// Create server connection
	serverStream := jsonrpc2.NewBufferedStream(&dummyReadWriteCloser{reader: serverReader, writer: serverWriter}, jsonrpc2.VSCodeObjectCodec{})
	serverConn = jsonrpc2.NewConn(ctx, serverStream, jsonrpc2.HandlerWithError(serverInstance.Handle))

	// Create client
	var wg sync.WaitGroup
	handler = &mockClientHandler{
		diagnosticsChan: make(chan lsp.PublishDiagnosticsParams, 2), // Buffer > 1 to handle rapid events
		wg:              &wg,
		t:               t,
	}
	clientStream := jsonrpc2.NewBufferedStream(&dummyReadWriteCloser{reader: clientReader, writer: clientWriter}, jsonrpc2.VSCodeObjectCodec{})
	clientConn = jsonrpc2.NewConn(ctx, clientStream, handler)

	return serverConn, clientConn, handler
}

// TestIntegration_DidOpen_PublishesDiagnostics verifies the full diagnostics flow.
func TestIntegration_DidOpen_PublishesDiagnostics(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverConn, clientConn, handler := setupIntegrationTest(t, ctx)
	defer serverConn.Close()
	defer clientConn.Close()

	// --- Test Execution ---
	handler.PrepareForWait()
	handler.wg.Add(1)
	uri := lsp.DocumentURI("file:///integration_test.ns")
	content := `func MyProc() means
  set x = tool.FS.ThisToolDoesNotExist()  # Semantic error
  set y = 1 + / 2                         # Syntax error
endfunc
`
	didOpenParams := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{URI: uri, Text: content},
	}
	paramsBytes, _ := json.Marshal(didOpenParams)
	rawParams := json.RawMessage(paramsBytes)

	if err := clientConn.Notify(ctx, "textDocument/didOpen", &rawParams); err != nil {
		t.Fatalf("Failed to send didOpen notification: %v", err)
	}

	// --- Verification ---
	var params lsp.PublishDiagnosticsParams
	select {
	case params = <-handler.diagnosticsChan:
		t.Log("Successfully received diagnostics from the server.")
	case <-ctx.Done():
		t.Fatal("Test timed out waiting for diagnostics to be published")
	}
	handler.wg.Wait()

	if len(params.Diagnostics) != 2 {
		var msgs []string
		for _, d := range params.Diagnostics {
			msgs = append(msgs, d.Message)
		}
		t.Fatalf("Expected 2 diagnostics, but got %d. Messages: %s", len(params.Diagnostics), strings.Join(msgs, ", "))
	}

	foundSemantic := false
	foundSyntax := false
	for _, d := range params.Diagnostics {
		if d.Code == string(DiagCodeToolNotFound) {
			foundSemantic = true
		}
		if d.Source == "nslsp-syntax" && strings.Contains(d.Message, "extraneous input '/' expecting") {
			foundSyntax = true
		}
	}

	if !foundSemantic {
		t.Error("Did not find the expected semantic error for the undefined tool (Code: ToolNotFound).")
	}
	if !foundSyntax {
		t.Error("Did not find the expected syntax error.")
	}
}

// TestIntegration_SyntaxError_MissingMeans verifies the "missing means" error.
func TestIntegration_SyntaxError_MissingMeans(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverConn, clientConn, handler := setupIntegrationTest(t, ctx)
	defer serverConn.Close()
	defer clientConn.Close()

	// --- Test Execution ---
	handler.PrepareForWait()
	handler.wg.Add(1)
	uri := lsp.DocumentURI("file:///missing_means.ns")
	content := `
func MyProc()
  # This line is where 'means' should be
  set x = 1
endfunc
`
	didOpenParams := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{URI: uri, Text: content},
	}
	paramsBytes, _ := json.Marshal(didOpenParams)
	rawParams := json.RawMessage(paramsBytes)

	if err := clientConn.Notify(ctx, "textDocument/didOpen", &rawParams); err != nil {
		t.Fatalf("Failed to send didOpen notification: %v", err)
	}

	// --- Verification ---
	var params lsp.PublishDiagnosticsParams
	select {
	case params = <-handler.diagnosticsChan:
		t.Log("Successfully received diagnostics from the server.")
	case <-ctx.Done():
		t.Fatal("Test timed out waiting for diagnostics to be published")
	}
	handler.wg.Wait()

	if len(params.Diagnostics) < 1 {
		t.Fatalf("Expected at least 1 syntax diagnostic, but got %d", len(params.Diagnostics))
	}

	foundSyntax := false
	for _, d := range params.Diagnostics {
		// THE FIX IS HERE:
		// We are now checking for the correct error message: "missing 'means'"
		if d.Source == "nslsp-syntax" && d.Range.Start.Line == 1 && strings.Contains(d.Message, "missing 'means'") {
			foundSyntax = true
			break
		}
	}

	if !foundSyntax {
		t.Errorf("Did not find the expected 'missing means' syntax error on line 2. Diagnostics: %+v", params.Diagnostics)
	}
}
