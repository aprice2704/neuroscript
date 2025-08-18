// NeuroScript Version: 0.6.0
// File version: 11
// Purpose: Provides end-to-end tests for completion. FIX: Updated tests to use server.interpreter.ToolRegistry() method.
// filename: pkg/nslsp/completion_test.go
// nlines: 190
// risk_rating: MEDIUM

package nslsp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/tool"
	//	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// setupCompletionTest initializes a server and adds debug logging for registered tools.
func setupCompletionTest(t *testing.T) (*Server, lsp.DocumentURI, context.CancelFunc) {
	t.Helper()
	serverInstance := NewServer(log.New(io.Discard, "", 0))

	// --- DEBUG LOGGING ---
	t.Logf("--- Registered Tools For Test: %s ---", t.Name())
	// THE FIX IS HERE: Access the tool registry via the interpreter and its method.
	if serverInstance.interpreter != nil && serverInstance.interpreter.ToolRegistry() != nil && serverInstance.interpreter.ToolRegistry().NTools() > 0 {
		var toolNames []string
		for _, impl := range serverInstance.interpreter.ToolRegistry().ListTools() {
			toolNames = append(toolNames, string(impl.Spec.FullName))
		}
		sort.Strings(toolNames)
		for _, name := range toolNames {
			t.Logf("  - Found tool: %s", name)
		}
	} else {
		t.Log("  - WARNING: Tool registry is nil or empty.")
	}
	t.Log("--- End Registered Tools ---")
	// --- END DEBUG LOGGING ---

	uri := lsp.DocumentURI("file:///completion_test.ns")
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	return serverInstance, uri, cancel
}

func TestHandleTextDocumentCompletion_ToolGroups(t *testing.T) {
	server, uri, cancel := setupCompletionTest(t)
	defer cancel()

	content := "func M() means\n  set x = tool.\nendfunc"
	server.documentManager.Set(uri, content)

	params := lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: uri},
			Position:     lsp.Position{Line: 1, Character: 16},
		},
	}
	rawParams, _ := json.Marshal(params)
	req := &jsonrpc2.Request{Method: "textDocument/completion", Params: (*json.RawMessage)(&rawParams)}

	result, err := server.handleTextDocumentCompletion(context.Background(), nil, req)
	if err != nil {
		t.Fatalf("handleTextDocumentCompletion returned an error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected a completion list for tool groups, but got nil")
	}
	completionList, ok := result.(*lsp.CompletionList)
	if !ok {
		t.Fatalf("Expected result to be *lsp.CompletionList, got %T", result)
	}

	expectedGroups := map[string]bool{"fs": true, "meta": true, "shell": true, "gotools": true, "str": true, "math": true, "agentmodel": true, "io": true, "list": true, "os": true, "script": true, "syntax": true, "time": true, "tree": true}

	foundGroups := make(map[string]bool)
	for _, item := range completionList.Items {
		foundGroups[strings.ToLower(item.Label)] = true
	}

	if len(foundGroups) < len(expectedGroups) {
		t.Errorf("Expected at least %d tool groups, but found %d", len(expectedGroups), len(foundGroups))
	}

	for expected := range expectedGroups {
		if _, found := foundGroups[expected]; !found {
			t.Errorf("Expected to find tool group '%s' in completion results, but it was missing", expected)
		}
	}
}

func TestHandleTextDocumentCompletion_ToolNames(t *testing.T) {
	server, uri, cancel := setupCompletionTest(t)
	defer cancel()

	content := "func M() means\n  set x = tool.fs.\nendfunc"
	server.documentManager.Set(uri, content)

	params := lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: uri},
			Position:     lsp.Position{Line: 1, Character: 18},
		},
	}
	rawParams, _ := json.Marshal(params)
	req := &jsonrpc2.Request{Method: "textDocument/completion", Params: (*json.RawMessage)(&rawParams)}

	result, err := server.handleTextDocumentCompletion(context.Background(), nil, req)
	if err != nil {
		t.Fatalf("handleTextDocumentCompletion returned an error: %v", err)
	}

	completionList, ok := result.(*lsp.CompletionList)
	if !ok || completionList == nil || completionList.Items == nil {
		t.Fatalf("Expected a valid *lsp.CompletionList with non-nil Items, but got a different result type or nil items")
	}

	expectedTools := map[string]string{
		"Read":  "(filepath: string) -> string",
		"List":  "(path: string, recursive: bool?) -> slice_any",
		"Write": "(filepath: string, content: string) -> string",
	}

	foundTools := make(map[string]string)
	for _, item := range completionList.Items {
		if _, ok := expectedTools[item.Label]; ok {
			foundTools[item.Label] = item.Detail
		}
	}

	for name, detail := range expectedTools {
		if foundDetail, ok := foundTools[name]; !ok {
			t.Errorf("Expected to find tool '%s' in completion list, but it was missing", name)
		} else if detail != foundDetail {
			t.Errorf("For tool '%s', expected detail '%s' but got '%s'", name, detail, foundDetail)
		}
	}
}

func TestHandleTextDocumentCompletion_ExternalTools(t *testing.T) {
	server, uri, cancel := setupCompletionTest(t)
	defer cancel()

	tempDir := t.TempDir()
	toolsJSONPath := filepath.Join(tempDir, "fdm-tools.json")
	externalToolImpls := []tool.ToolImplementation{
		{Spec: tool.ToolSpec{Name: "SaveNode", Group: "fdm.core", ReturnType: "string", Args: []tool.ArgSpec{{Name: "handle", Type: "string", Required: true}}}},
	}
	toolsJSONContent, _ := json.Marshal(externalToolImpls)
	os.WriteFile(toolsJSONPath, toolsJSONContent, 0644)
	server.externalTools.LoadFromPaths(log.New(io.Discard, "", 0), tempDir, []string{"fdm-tools.json"})

	content := "func M() means\n  set x = tool.FDM.core.\nendfunc"
	server.documentManager.Set(uri, content)

	params := lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: uri},
			Position:     lsp.Position{Line: 1, Character: 25},
		},
	}
	rawParams, _ := json.Marshal(params)
	req := &jsonrpc2.Request{Method: "textDocument/completion", Params: (*json.RawMessage)(&rawParams)}

	result, err := server.handleTextDocumentCompletion(context.Background(), nil, req)
	if err != nil {
		t.Fatalf("handleTextDocumentCompletion returned an error: %v", err)
	}

	completionList, ok := result.(*lsp.CompletionList)
	if !ok || completionList == nil {
		t.Fatalf("Expected a valid *lsp.CompletionList, but got a different result type")
	}

	if len(completionList.Items) != 1 {
		t.Fatalf("Expected 1 external tool completion, got %d. Items: %+v", len(completionList.Items), completionList.Items)
	}
	item := completionList.Items[0]
	if item.Label != "SaveNode" {
		t.Errorf("Expected completion label to be 'SaveNode', got '%s'", item.Label)
	}
	expectedDetail := "(handle: string) -> string"
	if item.Detail != expectedDetail {
		t.Errorf("Expected completion detail to be '%s', got '%s'", expectedDetail, item.Detail)
	}
}
