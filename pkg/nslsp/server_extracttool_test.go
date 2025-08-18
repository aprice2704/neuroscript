// NeuroScript Version: 0.3.1
// File version: 10
// Purpose: Updated tests to use the server.interpreter.ToolRegistry() method.
// filename: pkg/nslsp/server_extracttool_test.go
// nlines: 245
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

	"github.com/aprice2704/neuroscript/pkg/parser"
	_ "github.com/aprice2704/neuroscript/pkg/toolbundles/all"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestExtractToolNameAtPosition(t *testing.T) {
	var testServerLogger *log.Logger
	if os.Getenv("DEBUG_LSP_HOVER_TEST") != "" {
		testServerLogger = log.New(os.Stderr, "[EXTRACT_TEST_VALIDATE] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
		testServerLogger.Println("Detailed debug logging enabled for hover validation test.")
	} else {
		testServerLogger = log.New(io.Discard, "", 0)
	}

	parserAPI := parser.NewParserAPI(nil)
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
		line       int
		char       int
		expected   string
		sourceName string
	}{
		{"On 'tool' in tool.FS.List", neuroscriptCodeSnippet, 1, 12, "tool.FS.List", "test.ns"},
		{"On first '.' in tool.FS.List", neuroscriptCodeSnippet, 1, 16, "tool.FS.List", "test.ns"},
		{"On 'FS' in tool.FS.List", neuroscriptCodeSnippet, 1, 17, "tool.FS.List", "test.ns"},
		{"On second '.' in tool.FS.List", neuroscriptCodeSnippet, 1, 19, "tool.FS.List", "test.ns"},
		{"On 'List' in tool.FS.List", neuroscriptCodeSnippet, 1, 20, "tool.FS.List", "test.ns"},
		{"On 'AnotherTool'", neuroscriptCodeSnippet, 2, 12, "", "test.ns"},
		{"On 'sin' (built-in)", neuroscriptCodeSnippet, 3, 12, "", "test.ns"},
		{"On 'tool' in tool.Meta.ListTools", neuroscriptCodeSnippet, 4, 15, "tool.Meta.ListTools", "test.ns"},
		{"On 'Meta' in tool.Meta.ListTools", neuroscriptCodeSnippet, 4, 20, "tool.Meta.ListTools", "test.ns"},
		{"On 'ListTools' in tool.Meta.ListTools", neuroscriptCodeSnippet, 4, 25, "tool.Meta.ListTools", "test.ns"},
		{"On 'MyFunc' (direct call)", neuroscriptCodeSnippet, 5, 12, "", "test.ns"},
		{"On 'tool' in tool.FS.NonExistentTool", neuroscriptCodeSnippet, 7, 12, "", "test.ns"},
		{"On 'FS' in tool.FS.NonExistentTool", neuroscriptCodeSnippet, 7, 17, "", "test.ns"},
		{"On 'NonExistentTool'", neuroscriptCodeSnippet, 7, 20, "", "test.ns"},
		{"On 'Read' in tool.FS.Read", neuroscriptCodeSnippet, 8, 20, "tool.FS.Read", "test.ns"},
		{"On 'List' in tool.FS.List (no parens)", neuroscriptCodeSnippet, 9, 20, "tool.FS.List", "test.ns"},
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

type dummyReadWriteCloser struct {
	reader io.Reader
	writer io.Writer
	closer io.Closer
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
  set x = tool.FS.List("/some/path")
  set y = tool.Meta.ListTools()
  set z = tool.FS.NonExistentTool()
  set w = MyFunc()
endfunc
`)
	serverInstance.documentManager.Set(uri, content)

	if debugHoverTest {
		testServerLogger.Println("Hover Test: Document set in manager.")
		testServerLogger.Println("Registered tools for this test run:")
		// THE FIX IS HERE: Access the tool registry via the interpreter and its method.
		for _, toolSpec := range serverInstance.interpreter.ToolRegistry().ListTools() {
			testServerLogger.Printf("- %s\n", string(toolSpec.Name()))
		}
	}

	testCases := []struct {
		name              string
		line              int
		char              int
		expectHover       bool
		minContentLength  int
		expectedToolName  string // This is the full name used for logging/debugging
		expectedSubstring string // This is what we check for in the hover content
	}{
		{"Hover on 'List' in tool.FS.List", 1, 20, true, 20, "tool.FS.List", "tool.FS.List"},
		{"Hover on 'tool' in tool.FS.List", 1, 12, true, 20, "tool.FS.List", "tool.FS.List"},
		{"Hover on 'ListTools' in tool.Meta.ListTools", 2, 25, true, 20, "tool.Meta.ListTools", "tool.Meta.ListTools"},
		{"Hover on 'FS' in tool.FS.NonExistentTool", 3, 17, false, 0, "", ""},
		{"Hover on 'NonExistentTool'", 3, 20, false, 0, "", ""},
		{"Hover on 'MyFunc'", 4, 12, false, 0, "", ""},
		{"Hover on keyword 'set'", 1, 2, false, 0, "", ""},
	}

	serverInputReader, clientWritesToServerPipe := io.Pipe()
	clientOutputReader, serverWritesToClientPipe := io.Pipe()

	rwc := &dummyReadWriteCloser{
		reader: serverInputReader,
		writer: serverWritesToClientPipe,
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
	dummyConn := jsonrpc2.NewConn(ctx, serverObjectStream, nil)
	defer dummyConn.Close()

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
			}

			if tc.expectedSubstring != "" && !strings.Contains(actualContent, tc.expectedSubstring) {
				t.Errorf("FAIL: Hover content for '%s' did not contain expected substring '%s'. Actual: '%s'",
					tc.expectedToolName, tc.expectedSubstring, actualContent)
			}
		})
	}
}

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
