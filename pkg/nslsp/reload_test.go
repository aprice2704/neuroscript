// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: FIX: Made the test's client handler idempotent to prevent a panic from duplicate file-watcher notifications, resolving a race condition. FIX: Call the handler's PrepareForWait method to reset its state.
// filename: pkg/nslsp/reload_test.go
// nlines: 147
// risk_rating: HIGH

package nslsp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/tool"
	lsp "github.com/sourcegraph/go-lsp"
)

// TestIntegration_ExternalTool_LiveReload verifies the full file-watcher and reload flow.
func TestIntegration_ExternalTool_LiveReload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 1. --- Setup a temporary workspace and a server listening on it ---
	workspaceDir := t.TempDir()
	serverConn, clientConn, handler := setupIntegrationTest(t, ctx)
	defer serverConn.Close()
	defer clientConn.Close()

	initParams := lsp.InitializeParams{RootURI: lsp.DocumentURI("file://" + workspaceDir)}
	if err := clientConn.Call(ctx, "initialize", initParams, nil); err != nil {
		t.Fatalf("Failed to send initialize request: %v", err)
	}

	// 2. --- Open a file that uses a not-yet-defined tool ---
	handler.PrepareForWait() // THE FIX IS HERE
	handler.wg.Add(1)
	uri := lsp.DocumentURI("file://" + filepath.Join(workspaceDir, "main.ns"))
	content := `func Main() means
  call tool.custom.DoSomething()
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

	// 3. --- Verify we get the initial "undefined" diagnostic ---
	var initialDiagnostics lsp.PublishDiagnosticsParams
	select {
	case initialDiagnostics = <-handler.diagnosticsChan:
		t.Log("Successfully received initial diagnostics.")
	case <-ctx.Done():
		t.Fatal("Test timed out waiting for initial diagnostics")
	}
	handler.wg.Wait()

	if len(initialDiagnostics.Diagnostics) != 1 {
		t.Fatalf("Expected 1 initial diagnostic for undefined tool, but got %d.", len(initialDiagnostics.Diagnostics))
	}
	if !strings.Contains(initialDiagnostics.Diagnostics[0].Message, "Tool 'tool.custom.DoSomething' is not defined") {
		t.Fatalf("Initial diagnostic message was incorrect: %s", initialDiagnostics.Diagnostics[0].Message)
	}
	t.Log("Correctly received 'Tool not defined' diagnostic.")

	// 4. --- Dynamically create the tool definition file ---
	handler.PrepareForWait() // THE FIX IS HERE
	handler.wg.Add(1)
	toolsDir := filepath.Join(workspaceDir, "tools")
	if err := os.Mkdir(toolsDir, 0755); err != nil {
		t.Fatalf("Failed to create tools directory: %v", err)
	}
	// Give watcher a moment to pick up the directory creation
	time.Sleep(250 * time.Millisecond)

	toolDef := []tool.ToolImplementation{{
		Spec: tool.ToolSpec{
			Name:  "DoSomething",
			Group: "custom",
		},
	}}
	toolJSON, _ := json.Marshal(toolDef)
	toolsFilePath := filepath.Join(toolsDir, "custom.json")
	if err := os.WriteFile(toolsFilePath, toolJSON, 0644); err != nil {
		t.Fatalf("Failed to write tool definition file: %v", err)
	}
	t.Log("Created new tool definition file on disk.")

	// 5. --- Verify that the server sends an updated, empty diagnostic list ---
	var updatedDiagnostics lsp.PublishDiagnosticsParams
	// This loop drains any potential duplicate messages from the file watcher,
	// but the idempotent handler ensures wg.Done() is only called once.
	for {
		select {
		case updatedDiagnostics = <-handler.diagnosticsChan:
			t.Log("Received a diagnostic update...")
			if len(updatedDiagnostics.Diagnostics) == 0 {
				// We got the correct empty diagnostic list, break the loop and wait on the WG
				goto success
			}
			// If we received a non-empty list, it might be a stale event, so we continue waiting.
		case <-ctx.Done():
			t.Fatal("Test timed out waiting for updated diagnostics after tool file creation")
		}
	}

success:
	handler.wg.Wait() // Wait for the handler to have called Done().
	t.Log("Successfully received empty diagnostic list and WaitGroup returned, confirming live reload.")
}
