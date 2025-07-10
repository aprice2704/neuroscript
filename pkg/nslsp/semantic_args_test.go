// NeuroScript Version: 0.3.1
// File version: 6
// Purpose: Corrected tests to use proper multiline strings and removed brittle checks for exact error messages.
// filename: pkg/nslsp/semantic_args_test.go
// nlines: 130
// risk_rating: MEDIUM

package nslsp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"sync"
	"testing"
	"time"

	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestSemanticDiagnostics_ArgumentCountMismatch(t *testing.T) {
	testCases := []struct {
		name           string
		content        string
		expectedNdiags int
	}{
		{
			name: "Too Few Arguments",
			content: `func M() means
  set x = tool.FS.Read()
endfunc`,
			expectedNdiags: 1,
		},
		{
			name: "Too Many Arguments",
			content: `func M() means
  set x = tool.FS.Read("/path", "extra")
endfunc`,
			expectedNdiags: 1,
		},
		{
			name: "Correct for Zero Arguments",
			content: `func M() means
  set x = tool.Meta.ListTools()
endfunc`,
			expectedNdiags: 0,
		},
		{
			name: "Incorrect for Zero Arguments",
			content: `func M() means
  set x = tool.Meta.ListTools("extra")
endfunc`,
			expectedNdiags: 1,
		},
		{
			name: "Nested Call with Errors",
			content: `func M() means
  set x = tool.FS.Read(tool.FS.Read())
endfunc`,
			expectedNdiags: 1, // The listener should find the single innermost error.
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- Setup ---
			t.Setenv("DEBUG_LSP_HOVER_TEST", "1")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			serverInstance := NewServer(log.New(io.Discard, "", 0))

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
			uri := lsp.DocumentURI("file:///args_test.ns")
			didOpenParams := lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{URI: uri, Text: tc.content},
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
			case <-ctx.Done():
				t.Fatal("Test timed out waiting for diagnostics")
			}
			wg.Wait()

			// Check only the number of diagnostics, not the exact message.
			// This is a more robust test, per your feedback.
			var semanticDiagnostics []lsp.Diagnostic
			for _, diag := range params.Diagnostics {
				if diag.Source == "nslsp-semantic" {
					semanticDiagnostics = append(semanticDiagnostics, diag)
				}
			}

			if len(semanticDiagnostics) != tc.expectedNdiags {
				t.Fatalf("Expected %d semantic diagnostic(s), but got %d. All Diagnostics: %+v", tc.expectedNdiags, len(semanticDiagnostics), params.Diagnostics)
			}
		})
	}
}
