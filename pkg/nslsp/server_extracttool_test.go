// NeuroScript Version: 0.3.1
// File version: 0.1.6 // Implemented simpleObjectStream for jsonrpc2.Conn.
// Purpose: Test harness for LSP hover logic, including name extraction and full hover response.
// filename: pkg/nslsp/server_extracttool_test.go
// nlines: 240 // Approximate
// risk_rating: LOW

package nslsp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// TestExtractToolNameAtPosition (existing test - remains unchanged)
func TestExtractToolNameAtPosition(t *testing.T) {
	var testServerLogger *log.Logger
	if os.Getenv("DEBUG_LSP_HOVER_TEST") != "" {
		testServerLogger = log.New(os.Stderr, "[EXTRACT_TEST_VALIDATE] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
		testServerLogger.Println("Detailed debug logging enabled for hover validation test.")
	} else {
		testServerLogger = log.New(io.Discard, "", 0)
	}

	parserAPI := core.NewParserAPI(nil)
	serverInstance := NewServer(testServerLogger)
	serverInstance.coreParserAPI = parserAPI

	neuroscriptCodeSnippet := strings.TrimSpace(`
func MyProcedure() means
  set x = tool.FS.List("/some/path") 
  set y = AnotherTool()
  set z = sin(1.0)
  set path = tool.Meta.ListTools() 
  set val = MyFunc()
  set p = tool.AIWorker.LogPerformance()
  set q = tool.FS.NonExistentTool() 
  set r = tool.FS.Read("/file.txt") # Valid tool
  set s = tool.FS.List # Valid syntax, cursor on List
endfunc
`)
	snippetLines := strings.Split(neuroscriptCodeSnippet, "\n")
	if os.Getenv("DEBUG_LSP_HOVER_TEST") != "" {
		testServerLogger.Println("Using NeuroScript Code Snippet for tests:")
		for i, line := range snippetLines {
			testServerLogger.Printf("L%d: %s\n", i, line)
		}
	}

	testCases := []struct {
		name       string
		content    string
		line       int // 0-indexed
		char       int // 0-indexed
		expected   string
		sourceName string
	}{
		// Line 1: set x = tool.FS.List("/some/path")
		{"On 'tool' in tool.FS.List", neuroscriptCodeSnippet, 1, 12, "FS.List", "test.ns"},
		{"On first '.' in tool.FS.List", neuroscriptCodeSnippet, 1, 16, "FS.List", "test.ns"},
		{"On 'FS' in tool.FS.List", neuroscriptCodeSnippet, 1, 17, "FS.List", "test.ns"},
		{"On second '.' in tool.FS.List", neuroscriptCodeSnippet, 1, 19, "FS.List", "test.ns"},
		{"On 'List' in tool.FS.List", neuroscriptCodeSnippet, 1, 20, "FS.List", "test.ns"},
		{"On 'AnotherTool'", neuroscriptCodeSnippet, 2, 12, "", "test.ns"},
		{"On 'sin' (built-in)", neuroscriptCodeSnippet, 3, 12, "", "test.ns"},
		{"On 'tool' in tool.Meta.ListTools", neuroscriptCodeSnippet, 4, 15, "Meta.ListTools", "test.ns"},
		{"On 'Meta' in tool.Meta.ListTools", neuroscriptCodeSnippet, 4, 20, "Meta.ListTools", "test.ns"},
		{"On 'ListTools' in tool.Meta.ListTools", neuroscriptCodeSnippet, 4, 25, "Meta.ListTools", "test.ns"},
		{"On 'MyFunc' (direct call)", neuroscriptCodeSnippet, 5, 12, "", "test.ns"},
		{"On 'tool' in tool.AIWorker.LogPerformance", neuroscriptCodeSnippet, 6, 12, "AIWorker.LogPerformance", "test.ns"},
		{"On 'AIWorker' in tool.AIWorker.LogPerformance", neuroscriptCodeSnippet, 6, 17, "AIWorker.LogPerformance", "test.ns"},
		{"On 'LogPerformance' in tool.AIWorker.LogPerformance", neuroscriptCodeSnippet, 6, 26, "AIWorker.LogPerformance", "test.ns"},
		{"On 'tool' in tool.FS.NonExistentTool", neuroscriptCodeSnippet, 7, 12, "", "test.ns"},
		{"On 'FS' in tool.FS.NonExistentTool", neuroscriptCodeSnippet, 7, 17, "", "test.ns"},
		{"On 'NonExistentTool'", neuroscriptCodeSnippet, 7, 20, "", "test.ns"},
		{"On 'Read' in tool.FS.Read", neuroscriptCodeSnippet, 8, 20, "FS.Read", "test.ns"},
		{"On 'List' in tool.FS.List (no parens)", neuroscriptCodeSnippet, 9, 20, "FS.List", "test.ns"},
		{"On variable 'x'", neuroscriptCodeSnippet, 1, 6, "", "test.ns"},
		{"On keyword 'set'", neuroscriptCodeSnippet, 1, 2, "", "test.ns"},
		{"On string literal char '/'", neuroscriptCodeSnippet, 1, 28, "", "test.ns"},
		{"Inside string literal", neuroscriptCodeSnippet, 1, 30, "", "test.ns"},
		{"On '(' in tool.FS.List()", neuroscriptCodeSnippet, 1, 25, "", "test.ns"},
		{"On ')' in tool.FS.List()", neuroscriptCodeSnippet, 1, 39, "", "test.ns"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			extractedName := serverInstance.extractToolNameAtPosition(tc.content, lsp.Position{Line: tc.line, Character: tc.char}, tc.sourceName)
			if extractedName != tc.expected {
				t.Errorf("FAIL: For L%d:C%d ('%s' in '%s'), Expected '%s', but got '%s'",
					tc.line+1, tc.char+1, string(getRuneAt(tc.content, tc.line, tc.char)), strings.Split(tc.content, "\n")[tc.line], tc.expected, extractedName)
			} else {
				if os.Getenv("DEBUG_LSP_HOVER_TEST") != "" {
					t.Logf("PASS: For L%d:C%d ('%s' in '%s'), Correctly extracted '%s'",
						tc.line+1, tc.char+1, string(getRuneAt(tc.content, tc.line, tc.char)), strings.Split(tc.content, "\n")[tc.line], extractedName)
				}
			}
		})
	}
}

// dummyReadWriteCloser for jsonrpc2.Conn in tests
type dummyReadWriteCloser struct {
	reader io.Reader
	writer io.Writer
	closer io.Closer // Optional closer if the underlying pipes need it
}

func (d *dummyReadWriteCloser) Read(p []byte) (n int, err error) {
	return d.reader.Read(p)
}

func (d *dummyReadWriteCloser) Write(p []byte) (n int, err error) {
	return d.writer.Write(p)
}

func (d *dummyReadWriteCloser) Close() error {
	if d.closer != nil {
		return d.closer.Close()
	}
	// Attempt to close reader/writer if they are closers individually
	var errR, errW error
	if closer, ok := d.reader.(io.Closer); ok {
		errR = closer.Close()
	}
	if closer, ok := d.writer.(io.Closer); ok {
		errW = closer.Close()
	}
	if errR != nil {
		return errR
	}
	return errW
}

// simpleObjectStream implements jsonrpc2.ObjectStream using basic json.Encoder/Decoder
type simpleObjectStream struct {
	dec *json.Decoder
	enc *json.Encoder
	c   io.Closer
}

func newSimpleObjectStream(conn io.ReadWriteCloser) jsonrpc2.ObjectStream {
	return &simpleObjectStream{
		dec: json.NewDecoder(conn),
		enc: json.NewEncoder(conn),
		c:   conn,
	}
}

func (s *simpleObjectStream) WriteObject(obj interface{}) error {
	return s.enc.Encode(obj)
}

func (s *simpleObjectStream) ReadObject(obj interface{}) error {
	return s.dec.Decode(obj)
}

func (s *simpleObjectStream) Close() error {
	return s.c.Close()
}

func TestServerHandleHover(t *testing.T) {
	var testServerLogger *log.Logger
	debugHoverTest := os.Getenv("DEBUG_LSP_HOVER_TEST") != ""
	if debugHoverTest {
		testServerLogger = log.New(os.Stderr, "[HOVER_HANDLER_TEST] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
		testServerLogger.Println("Detailed debug logging enabled for hover handler test.")
	} else {
		testServerLogger = log.New(io.Discard, "", 0)
	}

	serverInstance := NewServer(testServerLogger)

	uri := lsp.DocumentURI("file:///test.ns")
	content := strings.TrimSpace(`
func MyProcedure() means
  set x = tool.FS.List("/some/path")       # Line 1 (0-indexed in snippet)
  set y = tool.Meta.ListTools()           # Line 2
  set z = tool.FS.NonExistentTool()       # Line 3
  set w = MyFunc()                        # Line 4
endfunc
`)
	serverInstance.documentManager.Set(uri, content)

	if debugHoverTest {
		testServerLogger.Println("Hover Test: Document set in manager.")
		testServerLogger.Println("Registered tools for this test run:")
		for _, toolSpec := range serverInstance.toolRegistry.ListTools() {
			testServerLogger.Printf("- %s\n", toolSpec.Name)
		}
	}

	testCases := []struct {
		name              string
		line              int
		char              int
		expectHover       bool
		minContentLength  int
		expectedToolName  string
		expectedSubstring string
	}{
		{"Hover on 'List' in tool.FS.List", 1, 20, true, 20, "FS.List", "tool.FS.List"},
		{"Hover on 'tool' in tool.FS.List", 1, 12, true, 20, "FS.List", "tool.FS.List"},
		{"Hover on 'ListTools' in tool.Meta.ListTools", 2, 25, true, 20, "Meta.ListTools", "tool.Meta.ListTools"},
		{"Hover on 'FS' in tool.FS.NonExistentTool", 3, 17, false, 0, "FS.NonExistentTool", ""},
		{"Hover on 'NonExistentTool'", 3, 20, false, 0, "FS.NonExistentTool", ""},
		{"Hover on 'MyFunc'", 4, 12, false, 0, "", ""},
		{"Hover on keyword 'set'", 1, 2, false, 0, "", ""},
	}

	// Use io.Pipe to create connected ReadWriteClosers
	// These pipes simulate the client-server communication channel.
	// serverInput is what the server reads from (client writes to this).
	// clientOutput is what the client reads from (server writes to this).
	serverInputReader, clientWritesToServerPipe := io.Pipe()
	clientOutputReader, serverWritesToClientPipe := io.Pipe()

	// Setup the dummyReadWriteCloser for the ObjectStream
	// The server's ObjectStream will read from serverInputReader and write to serverWritesToClientPipe.
	rwc := &dummyReadWriteCloser{
		reader: serverInputReader,        // Server reads from here
		writer: serverWritesToClientPipe, // Server writes here
		closer: struct{ io.Closer }{internalCloserFunc(func() error {
			var errFirst error
			errFirst = clientWritesToServerPipe.Close()
			errS := serverInputReader.Close()
			if errFirst == nil {
				errFirst = errS
			}
			errC := clientOutputReader.Close()
			if errFirst == nil {
				errFirst = errC
			}
			errSW := serverWritesToClientPipe.Close()
			if errFirst == nil {
				errFirst = errSW
			}
			return errFirst
		})},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverObjectStream := newSimpleObjectStream(rwc)
	// The handler (last arg for NewConn) can be nil for this test setup because
	// we are not testing the conn's ability to dispatch incoming requests to a handler.
	// We are directly calling serverInstance.Handle.
	dummyConn := jsonrpc2.NewConn(ctx, serverObjectStream, nil)
	defer dummyConn.Close() // This will close the rwc via simpleObjectStream.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: uri},
				Position:     lsp.Position{Line: tc.line, Character: tc.char},
			}
			paramsBytes, err := json.Marshal(params)
			if err != nil {
				t.Fatalf("Failed to marshal hover params: %v", err)
			}
			rawParams := json.RawMessage(paramsBytes)

			req := &jsonrpc2.Request{
				Method: "textDocument/hover",
				Params: &rawParams,
				ID:     jsonrpc2.ID{Num: 1},
			}

			result, err := serverInstance.Handle(ctx, dummyConn, req)

			if err != nil {
				if e, ok := err.(*jsonrpc2.Error); ok && e.Code == jsonrpc2.CodeMethodNotFound {
					t.Errorf("FAIL: Hover method not found by handler: %v", err)
					return
				}
				if tc.expectHover {
					t.Errorf("FAIL: Expected hover, but got error: %v", err)
				} else {
					if debugHoverTest {
						t.Logf("INFO: Correctly got error for non-hover case: %v (tool: %s)", err, tc.expectedToolName)
					}
				}
				return
			}

			if !tc.expectHover {
				if result != nil {
					t.Errorf("FAIL: Expected no hover (nil result), but got: %+v (tool: %s)", result, tc.expectedToolName)
				} else {
					if debugHoverTest {
						t.Logf("PASS: Correctly got no hover (nil result) for: %s", tc.name)
					}
				}
				return
			}

			if result == nil {
				t.Errorf("FAIL: Expected hover result, but got nil (tool: %s)", tc.expectedToolName)
				return
			}

			hoverResult, ok := result.(*lsp.Hover)
			if !ok {
				t.Errorf("FAIL: Expected result to be *lsp.Hover, but got %T (tool: %s)", result, tc.expectedToolName)
				return
			}

			if len(hoverResult.Contents) == 0 {
				t.Errorf("FAIL: Expected hover contents, but got empty array (tool: %s)", tc.expectedToolName)
				return
			}

			actualContent := hoverResult.Contents[0].Value
			if len(actualContent) < tc.minContentLength {
				t.Errorf("FAIL: Hover content for '%s' too short. Expected min %d, got %d. Content: '%s'",
					tc.expectedToolName, tc.minContentLength, len(actualContent), actualContent)
			} else {
				if debugHoverTest {
					t.Logf("PASS: Hover content for '%s' has sufficient length (%d). Content: '%s...'",
						tc.expectedToolName, len(actualContent), actualContent[:min(len(actualContent), 40)])
				}
			}

			if tc.expectedSubstring != "" && !strings.Contains(actualContent, tc.expectedSubstring) {
				t.Errorf("FAIL: Hover content for '%s' did not contain expected substring '%s'. Actual: '%s'",
					tc.expectedToolName, tc.expectedSubstring, actualContent)
			} else if tc.expectedSubstring != "" {
				if debugHoverTest {
					t.Logf("PASS: Hover content for '%s' contains expected substring '%s'", tc.expectedToolName, tc.expectedSubstring)
				}
			}
		})
	}
}

// Helper for creating a closer for the dummyReadWriteCloser from a function
type internalCloserFunc func() error

func (f internalCloserFunc) Close() error { return f() }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getRuneAt(content string, line0idx int, char0idx int) rune {
	lines := strings.Split(content, "\n")
	if line0idx >= 0 && line0idx < len(lines) {
		runesInLine := []rune(lines[line0idx])
		if char0idx >= 0 && char0idx < len(runesInLine) {
			return runesInLine[char0idx]
		}
	}
	return '?'
}
